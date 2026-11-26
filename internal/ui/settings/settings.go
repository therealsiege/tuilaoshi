// Package settings is the settings screen: the learner's preferences,
// changed with ←/→, saved at once to wiki/settings.yaml and put into effect
// for every screen (ui.Prefs).
package settings

import (
	"fmt"
	"slices"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/therealsiege/tuilaoshi/internal/ui"
	"github.com/therealsiege/tuilaoshi/internal/wiki"
)

// option is one setting: its choices and how to read and write it.
type option struct {
	label, help string
	choices     []string
	get         func(wiki.Settings) string
	set         func(*wiki.Settings, string)
}

var options = []option{
	{"Speech", "Speak answers and lines aloud with the voices (v at camp)",
		[]string{"on", "off"},
		func(s wiki.Settings) string { return map[bool]string{true: "on", false: "off"}[s.Speech] },
		func(s *wiki.Settings, v string) { s.Speech = v == "on" }},
	{"Sound effects", "Short cues for right and wrong answers, wins, encounters, level-ups and coins",
		[]string{"on", "off"},
		func(s wiki.Settings) string { return map[bool]string{true: "on", false: "off"}[s.Effects] },
		func(s *wiki.Settings, v string) { s.Effects = v == "on" }},
	{"Typing", "How answer boxes start: Chinese characters from candidates, or pinyin with tone numbers (ctrl+space switches)",
		[]string{"chinese", "pinyin"},
		func(s wiki.Settings) string { return s.Input },
		func(s *wiki.Settings, v string) { s.Input = v }},
	{"Battles", "Adaptive follows your recent accuracy; gentler and sterner fix the question mix",
		[]string{"adaptive", "gentler", "sterner"},
		func(s wiki.Settings) string { return s.Difficulty },
		func(s *wiki.Settings, v string) { s.Difficulty = v }},
	{"Wild encounters", "How often clans leap out on wild ground; off for a study-only walk",
		[]string{"normal", "fewer", "off"},
		func(s wiki.Settings) string { return s.Encounters },
		func(s *wiki.Settings, v string) { s.Encounters = v }},
	{"Portraits", "Enemy pictures in battle",
		[]string{"large", "small", "off"},
		func(s wiki.Settings) string { return s.Portraits },
		func(s *wiki.Settings, v string) { s.Portraits = v }},
	{"New words a day", "Words from your frequency list (make freqlist) that join your reviews each day, most frequent first",
		[]string{"0", "5", "10", "20"},
		func(s wiki.Settings) string { return fmt.Sprint(s.NewWords) },
		func(s *wiki.Settings, v string) { _, _ = fmt.Sscan(v, &s.NewWords) }}, //nolint:errcheck // choices are numbers
	{"Review chunk", "How many due words camp's Start here suggests at once (the daily bonus stays at 20)",
		[]string{"10", "20", "30", "50"},
		func(s wiki.Settings) string { return fmt.Sprint(s.ReviewChunk) },
		func(s *wiki.Settings, v string) { _, _ = fmt.Sscan(v, &s.ReviewChunk) }}, //nolint:errcheck // choices are numbers
	{"Tone colours", "Pinyin drawn in its tones' colours: first red, second orange, third green, fourth blue",
		[]string{"on", "off"},
		func(s wiki.Settings) string { return map[bool]string{true: "on", false: "off"}[s.ToneColours] },
		func(s *wiki.Settings, v string) { s.ToneColours = v == "on" }},
	{"Claude calls a day", "A cap on calls to Claude (chat, talks, grading, the storyteller, haggling, Fiery Eyes); My progress shows today's use",
		[]string{"no limit", "50", "100", "200"},
		func(s wiki.Settings) string {
			if s.ClaudeCap == 0 {
				return "no limit"
			}
			return fmt.Sprint(s.ClaudeCap)
		},
		func(s *wiki.Settings, v string) { s.ClaudeCap = 0; _, _ = fmt.Sscan(v, &s.ClaudeCap) }}, //nolint:errcheck // "no limit" reads as 0
}

// Model is the settings screen.
type Model struct {
	store   *wiki.Store
	gkeys   ui.GlobalKeys
	set     wiki.Settings
	cursor  int
	saved   bool
	loadErr error
	width   int
}

// New creates the settings screen.
func New(store *wiki.Store) Model { return Model{store: store, gkeys: ui.DefaultGlobalKeys()} }

func (m Model) Init() tea.Cmd { return nil }

// SetSize updates the terminal width.
func (m Model) SetSize(w, _ int) Model {
	m.width = w
	return m
}

// Refresh reads the settings.
func (m Model) Refresh() Model {
	m.set, m.loadErr = m.store.Settings()
	m.saved = false
	return m
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	k, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	switch {
	case key.Matches(k, m.gkeys.Back):
		return m, ui.SwitchTo(ui.ViewDashboard)
	case key.Matches(k, m.gkeys.Up):
		m.cursor = max(m.cursor-1, 0)
	case key.Matches(k, m.gkeys.Down):
		m.cursor = min(m.cursor+1, len(options)-1)
	case key.Matches(k, m.gkeys.Left):
		return m.step(-1), nil
	case key.Matches(k, m.gkeys.Right) || key.Matches(k, m.gkeys.Enter) || k.Type == tea.KeySpace:
		return m.step(1), nil
	}
	return m, nil
}

// step moves the selected setting to its next or previous choice, saves it
// and puts it into effect.
func (m Model) step(by int) Model {
	o := options[m.cursor]
	i := slices.Index(o.choices, o.get(m.set))
	o.set(&m.set, o.choices[(i+by+len(o.choices))%len(o.choices)])
	if err := m.store.SaveSettings(m.set); err != nil {
		m.loadErr = err
		return m
	}
	ui.Prefs, m.saved = m.set, true
	return m
}

func (m Model) View() string {
	var b strings.Builder
	b.WriteString(ui.TitleStyle.MarginBottom(0).Render("设置 Settings") + "\n")
	b.WriteString(ui.MutedStyle.Render("  Saved as you change them, for every journey.") + "\n\n")
	b.WriteString(ui.LoadError(m.loadErr))
	for i, o := range options {
		var choices []string
		for _, c := range o.choices {
			if c == o.get(m.set) {
				choices = append(choices, ui.SelectedStyle.UnsetPadding().UnsetMargins().Render(" "+c+" "))
			} else {
				choices = append(choices, ui.MutedStyle.Render(" "+c+" "))
			}
		}
		marker := "  "
		if i == m.cursor {
			marker = ui.WarningStyle.Render("▸ ")
		}
		fmt.Fprintf(&b, "  %s%-16s %s\n", marker, o.label, strings.Join(choices, ""))
	}
	b.WriteString("\n  " + ui.MutedStyle.Render(ui.WrapText(options[m.cursor].help, 72)) + "\n")
	if m.saved {
		b.WriteString(ui.SuccessStyle.Render("  Saved.") + "\n")
	}
	b.WriteString("\n" + ui.HelpStyle.Render("  ↑↓ choose · ←→ change · esc back"))
	return b.String()
}
