package dashboard

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/therealsiege/tuilaoshi/internal/journey"
	"github.com/therealsiege/tuilaoshi/internal/ui"
	"github.com/therealsiege/tuilaoshi/internal/wiki"
)

// menuColumnWidth is the width of one menu group; groups sit side by side
// when the terminal fits all three.
const menuColumnWidth = 28

// View shows, top to bottom: today's numbers, one suggested next step, the
// menu grouped by purpose, and where the journey stands.
func (m Model) View() string {
	var b strings.Builder
	b.WriteString(ui.TitleStyle.MarginBottom(0).Render("  营地 Camp"))
	b.WriteString(ui.MutedStyle.Render("  study by the fire, then back on the road (esc)"))
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

	footer := "  ? help · esc back on the road · T journeys · q quit"
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
	return "  " + ui.SubtitleStyle.Render("Today") + "  " + strings.Join(parts, " · ") + " · " + due + "\n" + m.renderPurse()
}

// renderPurse shows the coins saved, and how far today is from the daily bonus.
func (m Model) renderPurse() string {
	purse := fmt.Sprintf("%s 铜钱 · %s 元宝", thousands(m.wallet.Tongqian), thousands(m.wallet.Yuanbao))
	if left := wiki.DailyGoal - m.daily.CardsReviewed; left > 0 {
		purse += ui.MutedStyle.Render(fmt.Sprintf("   %d more %s today for the daily bonus", left, plural(left, "answer", "answers")))
	}
	return "  " + ui.SubtitleStyle.Render("Purse") + "  " + ui.WarningStyle.Render(purse) + "\n" + m.renderBounties()
}

// renderBounties lists today's bounties with progress, one per line.
func (m Model) renderBounties() string {
	var b strings.Builder
	for i, bt := range m.bounties {
		label := "        "
		if i == 0 {
			label = ui.SubtitleStyle.Render("Bounties")
		}
		status := ui.MutedStyle.Render(fmt.Sprintf("%d/%d", bt.Have, bt.Target))
		if bt.Done {
			status = ui.SuccessStyle.Render("✓")
		}
		fmt.Fprintf(&b, "  %s  %s %s %s\n", label, status, bt.Text, ui.MutedStyle.Render(fmt.Sprintf("+%d 铜钱", bt.Pay)))
	}
	note := "all three earn a streak shield"
	if m.shields > 0 {
		note += fmt.Sprintf(" · %d %s held", m.shields, plural(m.shields, "shield", "shields"))
	}
	return b.String() + "            " + ui.MutedStyle.Render(note) + m.renderFestival()
}

// renderFestival names the festival on today, or counts down to the next.
func (m Model) renderFestival() string {
	o, on, ok := journey.FestivalNow(m.now())
	switch {
	case !ok:
		return ""
	case on:
		return "\n  " + ui.SubtitleStyle.Render("节日") + "      " + ui.WarningStyle.Render(o.Name+"快乐！") + " " + o.English + ui.MutedStyle.Render(" · its story, questions and a 元宝 (y)")
	}
	days := o.DaysUntil(m.now())
	return "\n  " + ui.SubtitleStyle.Render("节日") + "      " + ui.MutedStyle.Render(fmt.Sprintf("next: %s %s in %d %s (y)", o.Name, o.English, days, plural(days, "day", "days")))
}

// renderStartHere suggests one next step: due reviews first, then the open
// journey stage, then reading.
func (m Model) renderStartHere() string {
	key, what := "t", "Read & listen to something new"
	switch s := m.journeyState.Stage; {
	case m.due > 0:
		if dailyChunk := ui.Prefs.ReviewChunk; m.due > dailyChunk {
			key, what = "r", fmt.Sprintf("Review %d of your %s due words", dailyChunk, thousands(m.due))
		} else {
			key, what = "r", fmt.Sprintf("Review your %s due %s", thousands(m.due), plural(m.due, "word", "words"))
		}
		if s != nil && s.Status == journey.StageLocked {
			what += fmt.Sprintf(": it opens tribulation %d (%s)", s.Number, journey.NeedText(*s, m.journeyState.Counts))
		}
	case s != nil && s.Status == journey.StageOpen:
		key, what = "esc", fmt.Sprintf("Tribulation %d is open: walk to its red marker on the road", s.Number)
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
	line := fmt.Sprintf("  %s  Level %d %s · %s %s · Tribulation %d %s", ui.SubtitleStyle.Render("Journey"),
		st.Level.N, ui.MutedStyle.Render(fmt.Sprintf("(%d/%d)", st.Level.Into, st.Level.Need)),
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
