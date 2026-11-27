package dashboard

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/therealsiege/tuilaoshi/internal/camp"
	"github.com/therealsiege/tuilaoshi/internal/claude"
	"github.com/therealsiege/tuilaoshi/internal/wiki"
)

// The court gazette (邸报): once a week, from Monday, a short newspaper
// report of the learner's own journey in Chinese at HSK4-5 with its
// English: the kingdoms that have stamped the travel papers, the week's
// tribulations cleared and rumours settled, errands done and sacks
// haggled for, and the diary's latest words. Claude writes it from those
// facts; without a key the court's clerk sets it from a template. Each
// issue is kept with the journey.

// GazetteWriter writes an issue from the week's facts, returning its
// Chinese and English (claude.Client.WriteGazette).
type GazetteWriter func(ctx context.Context, facts string) (string, string, error)

// WithGazette lets Claude write the gazette; a nil client leaves it to the
// template.
func (m Model) WithGazette(c *claude.Client) Model {
	if c != nil {
		m.writer = c.WriteGazette
	}
	return m
}

// gazetteMsg is an issue written by Claude.
type gazetteMsg struct {
	week    string
	zh, en  string
	err     error
	byClerk wiki.Gazette // the template's issue, in case Claude fails
}

// weekGazette is this week's issue if it has come out.
func (m Model) weekGazette() (*wiki.Gazette, error) {
	return camp.WeekGazette(m.store, m.now())
}

// openGazette shows this week's issue, printing it first if it hasn't come
// out: by Claude when there is a writer, else from the template.
func (m Model) openGazette() (Model, tea.Cmd) {
	if m.printing {
		return m, nil
	}
	issue, err := m.weekGazette()
	if err != nil {
		m.loadErr = err
		return m, nil
	}
	if issue != nil {
		m.gazette, m.gazetteOpen = issue, true
		return m, nil
	}
	clerk, facts, err := camp.ClerkGazette(m.store, m.now())
	if err != nil {
		m.loadErr = err
		return m, nil
	}
	week := clerk.Week
	if m.writer == nil {
		return m.printGazette(clerk), nil
	}
	m.printing, m.note = true, "The court gazette is being printed…"
	write, prompt := m.writer, facts.Prompt(week)
	return m, func() tea.Msg {
		zh, en, err := write(context.Background(), prompt)
		return gazetteMsg{week: week, zh: zh, en: en, err: err, byClerk: clerk}
	}
}

// onGazette keeps Claude's issue, or the clerk's when Claude failed.
func (m Model) onGazette(msg gazetteMsg) Model {
	m.printing, m.note = false, ""
	if msg.err != nil {
		m = m.printGazette(msg.byClerk)
		m.note = "Couldn't reach Claude (" + msg.err.Error() + "): the court's clerk set this issue by hand."
		return m
	}
	return m.printGazette(wiki.Gazette{Week: msg.week, Text: msg.zh, English: msg.en})
}

// printGazette keeps an issue and opens it.
func (m Model) printGazette(g wiki.Gazette) Model {
	if err := m.store.SaveGazette(g); err != nil {
		m.loadErr = err
		return m
	}
	m.gazette, m.gazetteOpen = &g, true
	return m
}
