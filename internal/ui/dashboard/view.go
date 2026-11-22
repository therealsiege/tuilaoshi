package dashboard

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/therealsiege/tuilaoshi/internal/journey"
	"github.com/therealsiege/tuilaoshi/internal/ui"
)

// menuColumnWidth is the width of one menu group; groups sit side by side
// when the terminal fits all three.
const menuColumnWidth = 28

// View shows, top to bottom: today's numbers, one suggested next step, the
// menu grouped by purpose, and where the journey stands.
func (m Model) View() string {
	var b strings.Builder
	b.WriteString(ui.TitleStyle.MarginBottom(0).Render("  推老师 TuīLǎoShī"))
	b.WriteString(ui.MutedStyle.Render("  Mandarin tutor"))
	b.WriteString("\n\n")
	b.WriteString(ui.LoadError(m.loadErr))

	b.WriteString(m.renderToday())
	b.WriteString("\n\n")
	b.WriteString(m.renderStartHere())
	b.WriteString("\n\n")
	b.WriteString(m.renderMenu())
	b.WriteString("\n\n")
	b.WriteString(m.renderJourney())
	b.WriteString("\n")

	footer := "  ? help · esc back · q quit"
	if m.tts != nil && m.tts.Available() {
		footer += " · audio on (" + m.tts.ChineseVoice().Name + " / " + m.tts.EnglishVoice().Name + ")"
	}
	b.WriteString(ui.MutedStyle.Render(footer))
	return b.String()
}

// renderToday is one line: what was done today and what is waiting.
func (m Model) renderToday() string {
	parts := []string{fmt.Sprintf("%d reviewed", m.daily.CardsReviewed)}
	if m.daily.CardsReviewed > 0 {
		parts = append(parts, fmt.Sprintf("%.0f%% right", float64(m.daily.Correct)/float64(m.daily.CardsReviewed)*100))
	}
	if m.streak > 0 {
		parts = append(parts, fmt.Sprintf("%d-day streak", m.streak))
	}
	due := ui.SuccessStyle.Render("nothing due")
	if m.due > 0 {
		due = ui.WarningStyle.Render(thousands(m.due) + " due")
	}
	return "  " + ui.SubtitleStyle.Render("Today") + "  " + strings.Join(parts, " · ") + " · " + due
}

// renderStartHere suggests one next step: due reviews first, then the open
// journey stage, then reading.
func (m Model) renderStartHere() string {
	key, what := "t", "Read & listen to something new"
	switch s := m.journeyState.Stage; {
	case m.due > 0:
		key, what = "r", fmt.Sprintf("Review your %s due %s", thousands(m.due), plural(m.due, "word", "words"))
	case s != nil && s.Status == journey.StageOpen:
		key, what = "f", fmt.Sprintf("Clear tribulation %d: battle %s", s.Number, s.Opponent.Name)
	}
	return "  " + ui.WarningStyle.Render("▶ Start here") + "   " + ui.ChineseStyle.Render(key) + "  " + what
}

// renderMenu lays the menu groups side by side, or stacked when narrow.
func (m Model) renderMenu() string {
	cols := make([]string, 0, len(ui.DashboardMenu))
	for _, g := range ui.DashboardMenu {
		lines := []string{ui.SubtitleStyle.Render("  " + strings.ToUpper(g.Title))}
		for _, it := range g.Items {
			lines = append(lines, "  "+ui.ChineseStyle.Render(it.Key)+"  "+it.Label)
		}
		cols = append(cols, lipgloss.NewStyle().Width(menuColumnWidth).Render(strings.Join(lines, "\n")))
	}
	if m.width > 0 && m.width < menuColumnWidth*len(cols)+4 {
		return strings.Join(cols, "\n\n")
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, cols...)
}

// renderJourney is where the pilgrimage stands and who is next, with a
// companion's word.
func (m Model) renderJourney() string {
	st := m.journeyState
	line := fmt.Sprintf("  %s  %s %s · Tribulation %d %s", ui.SubtitleStyle.Render("Journey"),
		ui.ChineseStyle.Render(st.Rank.Name), st.Rank.English, st.Progress.Tribulation, st.Progress.Location)
	if m.xp > 0 {
		line += fmt.Sprintf(" · %s XP", thousands(m.xp))
	}
	if s := st.Stage; s != nil {
		next := fmt.Sprintf("\n  %s %s %s", ui.MutedStyle.Render("Next:"), s.Opponent.Name, s.Opponent.English)
		if s.Status == journey.StageOpen {
			next += ui.WarningStyle.Render(" (ready, f)")
		} else {
			next += ui.MutedStyle.Render(" (" + journey.NeedText(*s, st.Counts) + ")")
		}
		line += next
	}
	if d := journey.RandomDialogue("idle", st.Rank.Current); d != nil {
		line += "\n  " + ui.MutedStyle.Render(d.Speaker+"："+d.Line)
	}
	return line + "\n"
}

// thousands writes 10789 as "10,789".
func thousands(n int) string {
	s := fmt.Sprint(n)
	for i := len(s) - 3; i > 0; i -= 3 {
		s = s[:i] + "," + s[i:]
	}
	return s
}

func plural(n int, one, many string) string {
	if n == 1 {
		return one
	}
	return many
}
