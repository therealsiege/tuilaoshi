package roam

import (
	"fmt"
	"math/rand"
	"slices"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/therealsiege/tuilaoshi/internal/journey"
	"github.com/therealsiege/tuilaoshi/internal/ui"
	"github.com/therealsiege/tuilaoshi/internal/wiki"
)

// A trail event takes over the map until met: its Chinese text and
// choices; once chosen, the result, the English and what it changed.

// eventRoll draws 0-99 against journey.EventChance; tests fix it.
var eventRoll = func() int { return rand.Intn(100) }

// trailEvent is an event being met.
type trailEvent struct {
	journey.Event
	chose  int    // -1 until chosen
	effect string // what the choice changed, in short
	cross  bool   // the region's obstacle: crossing it once chosen
}

// maybeEvent may bring an event at the end of a day.
func (m Model) maybeEvent() Model {
	if eventRoll() >= journey.EventChance {
		return m
	}
	if e, ok := journey.PickEvent(m.region.Region.Terrain, m.members, rand.New(rand.NewSource(rand.Int63()))); ok {
		m.event = &trailEvent{Event: e, chose: -1}
	}
	return m
}

// updateEvent takes 1-3 for a choice, then enter to walk on.
func (m Model) updateEvent(k tea.KeyMsg) (Model, tea.Cmd) {
	e := m.event
	if e.chose >= 0 {
		if key.Matches(k, m.gkeys.Enter) || k.Type == tea.KeySpace || key.Matches(k, m.gkeys.Back) {
			m.event = nil
			if e.cross {
				m = m.crossed()
			}
		}
		return m, nil
	}
	s := k.String()
	if len(s) != 1 || s[0] < '1' || int(s[0]-'1') >= len(e.Choices) {
		return m, nil
	}
	e.chose = int(s[0] - '1')
	o := e.Choices[e.chose].Outcome
	switch {
	case o.Coins < 0:
		if err := m.store.Spend(wiki.Wallet{Tongqian: -o.Coins}); err != nil {
			e.effect = "Not enough 铜钱 to pay: nothing comes of it."
			return m, nil
		}
	case o.Coins > 0:
		if err := m.store.Earn(wiki.Wallet{Tongqian: o.Coins}); err != nil {
			m.loadErr = err
		}
	}
	spared := false
	if m.trail.Forewarned {
		if o, spared = journey.Forewarn(o); spared {
			m.trail.Forewarned = false // the warning is spent on this danger
		}
	}
	m.trail.Apply(o, m.members)
	m.saveTrail()
	e.effect = effectText(o)
	if spared {
		e.effect += " (forewarned: harm halved)"
	}
	return m.refreshPurse(), nil
}

// refreshPurse rereads the purse after an event paid or charged.
func (m Model) refreshPurse() Model {
	if w, err := m.store.Wallet(); err == nil {
		m.wallet = w
	}
	return m
}

// effectText says what an outcome changed: "斋饭 +12 · health -4".
func effectText(o journey.Outcome) string {
	var parts []string
	add := func(label string, v int) {
		if v != 0 {
			parts = append(parts, fmt.Sprintf("%s %+d", label, v))
		}
	}
	add("斋饭", o.Food)
	add("水", o.Water)
	add("铜钱", o.Coins)
	who := "party"
	if o.Who != "" {
		who = o.Who
	}
	add(who+" health", o.Health)
	add(who+" morale", o.Morale)
	if len(parts) == 0 {
		return "Nothing changes."
	}
	return strings.Join(parts, " · ")
}

func (m Model) eventView() string {
	e := m.event
	var b strings.Builder
	b.WriteString(ui.SubtitleStyle.Render(fmt.Sprintf("  Day %d on the road", m.trail.Day)) + "\n\n")
	b.WriteString(m.wrapped(ui.ChineseStyle, e.Text, "  ") + "\n")
	if e.chose >= 0 {
		b.WriteString(m.wrapped(ui.MutedStyle, e.English, "  ") + "\n")
	}
	for i, c := range e.Choices {
		line := fmt.Sprintf("%d  %s", i+1, c.Text)
		switch {
		case i == e.chose:
			b.WriteString(m.wrapped(ui.WarningStyle, line+"  ◀", "  "))
		case e.chose >= 0:
			b.WriteString(m.wrapped(ui.MutedStyle, line, "  "))
		default:
			b.WriteString(m.wrapped(ui.UnselectedStyle, line, "  "))
		}
		if e.chose >= 0 {
			b.WriteString(m.wrapped(ui.MutedStyle, c.English, "     "))
		}
	}
	if e.chose < 0 {
		if m.trail.Forewarned {
			b.WriteString("\n  " + ui.SuccessStyle.Render("Forewarned: whatever harm comes is halved."))
		}
		b.WriteString("\n" + ui.HelpStyle.Render("  1-3 choose"))
		return b.String()
	}
	c := e.Choices[e.chose]
	b.WriteString("\n" + m.wrapped(ui.ChineseStyle, c.Result, "  "))
	b.WriteString(m.wrapped(ui.MutedStyle, c.ResultEnglish, "  "))
	b.WriteString(m.wrapped(ui.WarningStyle, e.effect, "  "))
	b.WriteString("\n" + ui.HelpStyle.Render("  enter walk on"))
	return b.String()
}

// faceCrossing meets the region's obstacle: today's conditions and the choices.
func (m Model) faceCrossing() Model {
	e := m.region.Crossing.Event(rand.New(rand.NewSource(rand.Int63())))
	m.event = &trailEvent{Event: e, chose: -1, cross: true}
	return m
}

// crossed puts the party on the far side of the obstacle, for good.
func (m Model) crossed() Model {
	m.trail.Crossed = append(m.trail.Crossed, m.region.Region.ID)
	m.pos = m.region.CrossAt
	m.save()
	m.saveTrail()
	return m
}

// hasCrossed reports whether this region's obstacle is behind the party.
func (m Model) hasCrossed() bool { return slices.Contains(m.trail.Crossed, m.region.Region.ID) }

// Hear is the Chinese of the event or crossing open, to read aloud: its
// text, and once chosen what came of it; or the overheard demons' lines.
func (m Model) Hear() string {
	e := m.event
	if e == nil {
		if m.plot != nil {
			return m.plot.plotText()
		}
		return ""
	}
	if e.chose >= 0 {
		return e.Text + "\n" + e.Choices[e.chose].Result
	}
	return e.Text
}

// wrapped renders text in style wrapped to the window (80 columns at most
// is plenty), each line indented by pad. Beside the sidebar it wraps to
// the space left of it.
func (m Model) wrapped(style lipgloss.Style, text, pad string) string {
	room := 78
	switch {
	case m.width > 0 && !m.compact() && m.width >= m.region.W+40:
		room = max(m.region.W, m.width-53) // the sidebar and its gap take about 50
	case m.width > 0:
		room = min(room, m.width-1)
	}
	width := room - lipgloss.Width(pad)
	var b strings.Builder
	for _, l := range ui.WrapDisplay(text, width) {
		b.WriteString(pad + style.Render(l) + "\n")
	}
	return b.String()
}
