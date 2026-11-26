package festival

import (
	"fmt"
	"strings"

	"github.com/therealsiege/tuilaoshi/internal/ui"
	"github.com/therealsiege/tuilaoshi/internal/wiki"
)

func (m Model) View() string {
	var b strings.Builder
	b.WriteString(ui.TitleStyle.MarginBottom(0).Render("节日 Festivals") + "\n")
	b.WriteString(ui.LoadError(m.loadErr))
	if !m.ok {
		b.WriteString(ui.MutedStyle.Render("  No festival dates ahead: the calendar in festivals.yaml needs extending.") + "\n")
		return b.String()
	}
	o := m.occ
	if !m.on {
		days := o.DaysUntil(m.now())
		fmt.Fprintf(&b, "  Next: %s %s, %s · in %d %s\n\n", ui.ChineseStyle.Render(o.Name), o.English, o.Date.Format("2 Jan 2006"), days, plural(days, "day", "days"))
		for _, line := range strings.Split(ui.WrapText(o.Desc, 74), "\n") {
			b.WriteString("  " + line + "\n")
		}
		b.WriteString("\n")
		var words []string
		for _, w := range o.Words {
			words = append(words, w.Chinese)
		}
		b.WriteString(ui.MutedStyle.Render("  Its words: "+strings.Join(words, " ")) + "\n")
		b.WriteString(ui.MutedStyle.Render(ui.WrapIndent("When it comes, its story, two questions and a 元宝 will be here, and its words join your reviews.", m.width)) + "\n")
		b.WriteString("\n" + ui.HelpStyle.Render("  esc back"))
		return b.String()
	}
	fmt.Fprintf(&b, "  %s %s %s\n", ui.WarningStyle.Render("Today:"), ui.ChineseStyle.Bold(true).Render(o.Name+"快乐！"), o.English)
	if m.added > 0 {
		b.WriteString(ui.SuccessStyle.Render(fmt.Sprintf("  %d festival words joined your reviews.", m.added)) + "\n")
	}
	b.WriteString("\n")
	if m.asking >= 0 {
		b.WriteString(m.questionView())
		return b.String()
	}
	for _, line := range strings.Split(strings.TrimSpace(o.Text), "\n") {
		b.WriteString("  " + ui.ChineseStyle.Render(line) + "\n")
	}
	if m.english {
		b.WriteString("\n")
		for _, line := range strings.Split(ui.WrapText(strings.TrimSpace(o.TextEnglish), 74), "\n") {
			b.WriteString(ui.MutedStyle.Render("  "+line) + "\n")
		}
	}
	b.WriteString("\n" + ui.SubtitleStyle.Render("  Questions") + ui.MutedStyle.Render(fmt.Sprintf("  %d 铜钱 each; both: +%d 铜钱 +%d 元宝", QuestionPay, FestivalPay, FestivalIngots)) + "\n")
	for i, q := range o.Questions {
		status := ""
		switch m.states[m.questID(i)] {
		case wiki.QuestDone:
			status = ui.SuccessStyle.Render("  ✓")
		case wiki.QuestLocked:
			status = ui.MutedStyle.Render("  tomorrow")
		}
		fmt.Fprintf(&b, "  %d  %s%s\n", i+1, q.Question, status)
	}
	var words []string
	for _, w := range o.Words {
		words = append(words, w.Chinese+" "+ui.MutedStyle.Render(w.Pinyin))
	}
	b.WriteString("\n" + ui.MutedStyle.Render("  Words: ") + strings.Join(words, " · ") + "\n")
	if m.note != "" {
		b.WriteString("\n" + ui.WarningStyle.Render("  "+m.note) + "\n")
	}
	b.WriteString("\n" + ui.HelpStyle.Render("  1-2 answer a question · e English · esc back"))
	return b.String()
}

func (m Model) questionView() string {
	q := m.occ.Questions[m.asking]
	var b strings.Builder
	for _, line := range strings.Split(strings.TrimSpace(m.occ.Text), "\n") {
		b.WriteString("  " + ui.MutedStyle.Render(line) + "\n")
	}
	b.WriteString("\n  " + ui.SubtitleStyle.Render(q.Question) + "\n")
	for i, c := range m.choices {
		line := fmt.Sprintf("  %d  %s", i+1, c)
		switch {
		case m.answered >= 0 && i == m.right:
			line = ui.SuccessStyle.Render(line + "  ✓")
		case m.answered == i:
			line = ui.ErrorStyle.Render(line + "  ✗")
		}
		b.WriteString(line + "\n")
	}
	switch {
	case m.answered < 0:
		b.WriteString("\n" + ui.HelpStyle.Render("  1-4 answer · esc back"))
	case m.answered == m.right:
		b.WriteString("\n" + ui.SuccessStyle.Render(fmt.Sprintf("  Right! +%d 铜钱", QuestionPay)) + "\n" + ui.HelpStyle.Render("  enter back"))
	default:
		b.WriteString("\n" + ui.WarningStyle.Render("  Not quite. Read it again and try tomorrow.") + "\n" + ui.HelpStyle.Render("  enter back"))
	}
	if m.note != "" {
		b.WriteString("\n" + ui.SuccessStyle.Render("  "+m.note))
	}
	return b.String()
}

func plural(n int, one, many string) string {
	if n == 1 {
		return one
	}
	return many
}
