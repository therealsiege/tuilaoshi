package world

import (
	"context"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/therealsiege/tuilaoshi/internal/ui"
	"github.com/therealsiege/tuilaoshi/internal/wiki"
)

// Grader grades Chinese written for a task against criteria, 0-2, with feedback.
type Grader func(ctx context.Context, task, criteria, answer string) (int, string, error)

type gradeMsg struct {
	score    int
	feedback string
	err      error
}

// submitWriting sends the written answer for grading.
func (m Model) submitWriting() (Model, tea.Cmd) {
	text := strings.TrimSpace(m.input.Value())
	if text == "" || m.grading {
		return m, nil
	}
	if m.grader == nil {
		m.note = "Writing requests need ANTHROPIC_API_KEY (in ~/.config/tuilaoshi/env)."
		return m, nil
	}
	q := m.towns[m.town].Quests[m.quest]
	m.grading, m.feedback, m.note = true, "", ""
	grader := m.grader
	return m, func() tea.Msg {
		score, fb, err := grader(context.Background(), q.Text+" ("+q.English+")", q.Criteria, text)
		return gradeMsg{score, fb, err}
	}
}

// onGrade pays a passing answer once; a failing one can be improved and
// sent again.
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
	q := m.towns[m.town].Quests[m.quest]
	if err := m.store.FinishQuest(q.ID, wiki.Wallet{Tongqian: q.Pay(msg.score)}); err != nil {
		m.loadErr = err
	}
	m.answered = msg.score
	m.input.Blur()
	return m
}

// updateWriting types the answer; enter sends it, or returns once passed.
func (m Model) updateWriting(k tea.KeyMsg) (Model, tea.Cmd) {
	switch {
	case m.answered > 0 && (k.Type == tea.KeyEnter || k.Type == tea.KeySpace):
		m.screen = screenTown
		return m.Refresh(), nil
	case m.answered > 0:
		return m, nil
	case k.Type == tea.KeyEnter && !m.input.Composing():
		return m.submitWriting()
	}
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(k)
	return m, cmd
}

// writingView shows the request, the answer box, and the grade.
func (m Model) writingView() string {
	q := m.towns[m.town].Quests[m.quest]
	var b strings.Builder
	b.WriteString(ui.TitleStyle.MarginBottom(0).Render(q.Title) + "  " + ui.MutedStyle.Render(q.Giver+" · "+q.GiverEnglish) + "\n\n")
	b.WriteString("  " + ui.ChineseStyle.Render(q.Text) + "\n")
	b.WriteString(ui.MutedStyle.Render("  "+q.English) + "\n\n")
	b.WriteString("  " + m.input.View() + "\n\n")
	switch {
	case m.grading:
		b.WriteString(ui.MutedStyle.Render("  "+q.Giver+" reads it…") + "\n")
	case m.answered > 0:
		b.WriteString(ui.SuccessStyle.Render(fmt.Sprintf("  %s is pleased and pays you %d 铜钱.", q.Giver, q.Pay(m.answered))) + "\n")
	case m.feedback != "":
		b.WriteString(ui.WarningStyle.Render("  Not yet: edit it and send again.") + "\n")
	}
	if m.feedback != "" {
		b.WriteString(ui.MutedStyle.Render("  "+ui.WrapText(m.feedback, 72)) + "\n")
	}
	help := "  enter send · ctrl+space pinyin/English · esc back"
	if m.answered > 0 {
		help = "  enter back to the notice board"
	}
	b.WriteString("\n" + ui.HelpStyle.Render(help))
	return b.String()
}
