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
	if m.gazetteOpen && m.gazette != nil {
		return m.gazetteView()
	}
	if m.reading && m.letter != nil {
		return m.letterView()
	}
	var b strings.Builder
	b.WriteString(ui.TitleStyle.MarginBottom(0).Render("  营地 Camp"))
	b.WriteString(ui.MutedStyle.Render("  study by the fire, then back on the road (esc)"))
	b.WriteString("\n\n")
	b.WriteString(ui.LoadError(m.loadErr))

	if m.compact() {
		b.WriteString(m.renderCompact())
	} else {
		b.WriteString(m.renderToday())
		b.WriteString("\n\n")
		b.WriteString(m.renderStartHere())
		b.WriteString("\n\n")
		b.WriteString(m.renderMenu())
		b.WriteString("\n\n")
		b.WriteString(m.renderJourney())
		b.WriteString("\n")
	}

	if m.note != "" {
		b.WriteString(ui.WarningStyle.Render("  "+m.note) + "\n")
	}
	footer := "  ? help · esc back on the road · T journeys · q quit"
	if m.tts != nil && m.tts.Available() && !m.compact() {
		footer += " · audio on (" + m.tts.ChineseVoice().Name + " / " + m.tts.EnglishVoice().Name + ")"
	}
	b.WriteString(ui.MutedStyle.Render(footer))
	return b.String()
}

// compactRows is the terminal height below which camp folds into its
// short layout; the full one needs about 32 rows.
const compactRows = 32

// compact reports whether camp should use its short layout.
func (m Model) compact() bool { return m.height > 0 && m.height < compactRows }

// renderCompact is camp in about fifteen lines: today and the purse, one
// line each for bounties and the week's events, the next step, the menu as
// wrapped keys, and the journey.
func (m Model) renderCompact() string {
	var b strings.Builder
	b.WriteString(strings.SplitN(m.renderToday(), "\n", 2)[0] + "\n")
	purse := fmt.Sprintf("%s 铜钱 · %s 元宝", thousands(m.wallet.Tongqian), thousands(m.wallet.Yuanbao))
	done := 0
	for _, bt := range m.bounties {
		if bt.Done {
			done++
		}
	}
	b.WriteString("  " + ui.SubtitleStyle.Render("Purse") + "  " + ui.WarningStyle.Render(purse) +
		ui.MutedStyle.Render(fmt.Sprintf(" · bounties %d/%d", done, len(m.bounties))) + "\n")
	week := []string{}
	if o, on, ok := journey.FestivalNow(m.now()); ok {
		if on {
			week = append(week, ui.WarningStyle.Render(o.Name+"快乐！")+" (y)")
		} else {
			week = append(week, fmt.Sprintf("%s in %d days", o.Name, o.DaysUntil(m.now())))
		}
	}
	if t, _ := journey.TermOn(m.now()); t.Name != "" {
		week = append(week, "节气 "+t.Name)
	}
	if m.letter != nil && !m.letter.Read {
		week = append(week, ui.WarningStyle.Render("来信 a letter (n)"))
	}
	if m.missed >= journey.InnerDemonWakes {
		week = append(week, ui.ErrorStyle.Render("心魔 wakes (m)"))
	} else {
		week = append(week, fmt.Sprintf("心魔 %d/%d", m.missed, journey.InnerDemonWakes))
	}
	b.WriteString("  " + ui.MutedStyle.Render(strings.Join(week, " · ")) + "\n\n")
	key, what := m.startHere()
	startWidth := 76
	if m.width > 0 {
		startWidth = m.width - 22
	}
	for i, l := range ui.WrapDisplay(what, startWidth) {
		if i == 0 {
			b.WriteString("  " + ui.WarningStyle.Render("▶ Start here") + "   " + ui.ChineseStyle.Render(key) + "  " + l + "\n")
		} else {
			b.WriteString(strings.Repeat(" ", 20) + l + "\n")
		}
	}
	b.WriteString("\n")
	var keys []string
	for _, g := range ui.DashboardMenu {
		for _, it := range g.Items {
			keys = append(keys, ui.ChineseStyle.Render(it.Key)+" "+it.Label)
		}
	}
	width := m.width - 4
	if width <= 0 {
		width = 76
	}
	line := ""
	for _, k := range keys {
		if line != "" && lipgloss.Width(line)+3+lipgloss.Width(k) > width {
			b.WriteString("  " + line + "\n")
			line = ""
		}
		if line != "" {
			line += ui.MutedStyle.Render(" · ")
		}
		line += k
	}
	b.WriteString("  " + line + "\n\n")
	st := m.journeyState
	fmt.Fprintf(&b, "  %s  Level %d · %s · Tribulation %d %s\n", ui.SubtitleStyle.Render("Journey"),
		st.Level.N, ui.ChineseStyle.Render(st.Rank.Name), st.Progress.Tribulation, st.Progress.Location)
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
	today := "  " + ui.SubtitleStyle.Render("Today") + "  " + strings.Join(parts, " · ") + " · " + due + "\n"
	if m.newWords > 0 {
		today += ui.SuccessStyle.Render(fmt.Sprintf("         +%d new words from your frequency list joined your reviews", m.newWords)) + "\n"
	} else if m.paused != "" {
		today += ui.MutedStyle.Render("         New words wait: "+m.paused+". Review to bring them back.") + "\n"
	}
	return today + m.renderPurse()
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
	return b.String() + "            " + ui.MutedStyle.Render(note) + m.renderFestival() + m.renderTerm() + m.renderInnerDemon() + m.renderLetterLine()
}

// weatherNote says what a term's weather does on the road.
var weatherNote = map[string]string{
	"cold": "cold: the party eats more",
	"heat": "hot: the party drinks more",
	"rain": "rain fills the gourds",
}

// renderTerm is the solar term (节气) in effect, and its weather on the road.
func (m Model) renderTerm() string {
	t, starts := journey.TermOn(m.now())
	if t.Name == "" {
		return ""
	}
	name := t.Name
	if starts {
		name = "今日" + t.Name
	}
	line := "\n  " + ui.SubtitleStyle.Render("节气") + "      " + ui.ChineseStyle.Render(name) + " " + t.English + " " + ui.MutedStyle.Render(t.Text)
	if w := weatherNote[t.Weather]; w != "" {
		line += ui.MutedStyle.Render(" · " + w)
	}
	return line
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

// renderInnerDemon says whether the week's missed words have woken the
// Inner Demon.
func (m Model) renderInnerDemon() string {
	label := "\n  " + ui.SubtitleStyle.Render("心魔") + "      "
	if m.missed >= journey.InnerDemonWakes {
		return label + ui.ErrorStyle.Render(fmt.Sprintf("the Inner Demon wakes: %d words you missed this week (m)", m.missed))
	}
	return label + ui.MutedStyle.Render(fmt.Sprintf("sleeps: %d of %d missed words this week wake it", m.missed, journey.InnerDemonWakes))
}

// firstCheckDue is the backlog at which a learner who hasn't taken the
// level check is pointed to it first: it moves the words they know out of
// the way.
const firstCheckDue = 500

// renderStartHere suggests one next step: the level check for a new
// learner facing a big backlog, due reviews, then the open journey stage,
// then reading.
func (m Model) renderStartHere() string {
	key, what := m.startHere()
	return "  " + ui.WarningStyle.Render("▶ Start here") + "   " + ui.ChineseStyle.Render(key) + "  " + what
}

// startHere is the one next step: its key and what it does.
func (m Model) startHere() (key, what string) {
	key, what = "t", "Read & listen to something new"
	switch s := m.journeyState.Stage; {
	case !m.assessed && m.due > firstCheckDue:
		key, what = "a", fmt.Sprintf("Check your level first: words you already know leave today's %s due", thousands(m.due))
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
	return key, what
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
			next += ui.WarningStyle.Render(" (ready: walk to its red ! on the road)")
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
