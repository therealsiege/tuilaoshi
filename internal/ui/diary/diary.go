// Package diary is the travel diary (日记) kept at camp: once a day, a few
// sentences in Chinese about the day on the road, graded by Claude. A
// passing entry is kept in the diary and lifts the party's spirits on the
// trail; one that doesn't pass can be revised and sent again.
package diary

import (
	"context"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/therealsiege/tuilaoshi/internal/camp"
	"github.com/therealsiege/tuilaoshi/internal/claude"
	"github.com/therealsiege/tuilaoshi/internal/journey"
	"github.com/therealsiege/tuilaoshi/internal/ui"
	"github.com/therealsiege/tuilaoshi/internal/ui/input"
	"github.com/therealsiege/tuilaoshi/internal/wiki"
)

// Grader grades Chinese written for a task against criteria, 0-2, with feedback.
type Grader = camp.Grader

type gradeMsg struct {
	score    int
	feedback string
	err      error
}

// Model is the travel diary screen.
type Model struct {
	store  *wiki.Store
	gkeys  ui.GlobalKeys
	grader Grader
	input  input.Model
	now    func() time.Time

	entries  []wiki.DiaryEntry
	day      int  // the day on the trail
	written  bool // today's entry is in the diary
	grading  bool
	feedback string
	passed   int // the grade of the entry written on this visit, 0 until it passes

	note          string
	width, height int
	loadErr       error
}

// New creates the diary; with no Claude client, entries can't be graded.
func New(store *wiki.Store, client *claude.Client, dict *input.Dict) Model {
	in := input.NewWithDict(dict)
	in.SetPlaceholder("写今天的日记: type pinyin, then pick characters")
	in.SetCharLimit(300)
	m := Model{store: store, gkeys: ui.DefaultGlobalKeys(), input: in, now: time.Now}
	if client != nil {
		m.grader = client.GradeWriting
	}
	return m
}

func (m Model) Init() tea.Cmd { return nil }

// SetSize updates the terminal dimensions.
func (m Model) SetSize(w, h int) Model {
	m.width, m.height = w, h
	return m
}

// Typing reports whether the entry box has the keyboard.
func (m Model) Typing() bool { return !m.written }

// Composing reports whether a pinyin is being turned into characters.
func (m Model) Composing() bool { return m.input.Composing() }

// Refresh reloads the diary and the day on the trail, and opens today's
// entry when it isn't written yet.
func (m Model) Refresh() (Model, tea.Cmd) {
	m.note, m.feedback, m.passed, m.grading = "", "", 0, false
	entries, err := m.store.Diary()
	cleared, clearErr := m.store.ClearedStages()
	t, _, trailErr := journey.LoadTrail(m.store, cleared)
	m.entries, m.day, m.loadErr = entries, t.Day, firstErr(err, clearErr, trailErr)
	today := m.now().Format("2006-01-02")
	m.written = len(entries) > 0 && entries[len(entries)-1].Date == today
	if m.written {
		m.input.Blur()
		return m, nil
	}
	m.input.Reset()
	return m, m.input.Focus()
}

func firstErr(errs ...error) error {
	for _, err := range errs {
		if err != nil {
			return err
		}
	}
	return nil
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if g, ok := msg.(gradeMsg); ok {
		return m.onGrade(g), nil
	}
	k, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	switch {
	case key.Matches(k, m.gkeys.Back) && !m.input.Composing():
		return m, ui.SwitchTo(ui.ViewDashboard)
	case m.written:
		return m, nil
	case k.Type == tea.KeyEnter && !m.input.Composing():
		return m.submit()
	}
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(k)
	return m, cmd
}

// submit sends the entry for grading.
func (m Model) submit() (Model, tea.Cmd) {
	text := strings.TrimSpace(m.input.Value())
	if text == "" || m.grading {
		return m, nil
	}
	if m.grader == nil {
		m.note = "The diary needs ANTHROPIC_API_KEY (in ~/.config/tuilaoshi/env) to read your entry."
		return m, nil
	}
	m.grading, m.feedback, m.note = true, "", ""
	grader, prompt := m.grader, camp.DiaryTask(m.day)
	return m, func() tea.Msg {
		score, fb, err := grader(context.Background(), prompt, camp.DiaryCriteria, text)
		return gradeMsg{score, fb, err}
	}
}

// onGrade keeps a passing entry and lifts the party's spirits; one that
// doesn't pass stays open to revise.
func (m Model) onGrade(msg gradeMsg) Model {
	m.grading = false
	if msg.err != nil {
		m.note = "Couldn't reach Claude: " + msg.err.Error()
		return m
	}
	m.feedback = msg.feedback
	if msg.score == 0 {
		return m
	}
	e := wiki.DiaryEntry{Date: m.now().Format("2006-01-02"), Day: m.day, Text: strings.TrimSpace(m.input.Value()), Score: msg.score, Feedback: msg.feedback}
	m.loadErr = camp.KeepDiaryEntry(m.store, e)
	m.entries = append(m.entries, e)
	m.written, m.passed = true, msg.score
	m.input.Blur()
	return m
}

// Hear is the diary's Chinese, to read aloud: today's entry once written,
// else the day's prompt.
func (m Model) Hear() string {
	if m.written && len(m.entries) > 0 {
		return m.entries[len(m.entries)-1].Text
	}
	return "今天走到了哪里？遇到了什么？心情怎么样？"
}
