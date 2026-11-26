package diary

import (
	"fmt"
	"strings"

	"github.com/therealsiege/tuilaoshi/internal/ui"
)

// shown is how many past entries the diary shows.
const shown = 4

func (m Model) View() string {
	var b strings.Builder
	b.WriteString(ui.TitleStyle.MarginBottom(0).Render("日记 Travel diary") + "  " + ui.MutedStyle.Render(fmt.Sprintf("第%d天 day %d on the road", m.day, m.day)) + "\n\n")
	b.WriteString(ui.LoadError(m.loadErr))
	if !m.written || m.passed > 0 {
		b.WriteString("  " + ui.ChineseStyle.Render("今天走到了哪里？遇到了什么？心情怎么样？写两三句话。") + "\n")
		b.WriteString(m.muted("Where did you go today, what happened, how did you feel? Two or three sentences.") + "\n")
	}
	if !m.written {
		b.WriteString("  " + m.input.View() + "\n\n")
	}
	switch {
	case m.grading:
		b.WriteString(ui.MutedStyle.Render("  The master reads your entry…") + "\n")
	case m.passed > 0:
		b.WriteString(ui.SuccessStyle.Render(fmt.Sprintf("  Kept in the diary. Reading it by the fire lifts everyone's spirits (+%d).", gradeMorale[m.passed])) + "\n")
	case m.written:
		b.WriteString(ui.MutedStyle.Render("  Today's entry is written. Come back tomorrow for the next.") + "\n")
	case m.feedback != "":
		b.WriteString(ui.WarningStyle.Render("  Not yet: revise it and send again.") + "\n")
	}
	if m.feedback != "" {
		b.WriteString(m.muted(m.feedback))
	}
	if len(m.entries) > 0 {
		b.WriteString("\n" + ui.SubtitleStyle.Render("  往日 Earlier entries") + "\n")
		for i := len(m.entries) - 1; i >= max(len(m.entries)-shown, 0); i-- {
			e := m.entries[i]
			b.WriteString("  " + ui.MutedStyle.Render(fmt.Sprintf("%s · 第%d天", e.Date, e.Day)) + "\n")
			width := 72
			if m.width > 0 {
				width = min(width, m.width-6)
			}
			for _, l := range ui.WrapDisplay(e.Text, width) {
				b.WriteString("    " + ui.ChineseStyle.Render(l) + "\n")
			}
		}
	}
	if m.note != "" {
		b.WriteString("\n" + ui.WarningStyle.Render("  "+m.note) + "\n")
	}
	help := "  esc back to camp"
	if !m.written {
		help = "  enter send · ctrl+space pinyin/English · esc back to camp"
	}
	b.WriteString("\n" + ui.HelpStyle.Render(help))
	return b.String()
}

// muted renders English dimmed, two spaces in, wrapped to the window.
func (m Model) muted(text string) string {
	width := 76
	if m.width > 0 {
		width = min(width, m.width-4)
	}
	var b strings.Builder
	for _, l := range ui.WrapDisplay(text, width) {
		b.WriteString(ui.MutedStyle.Render("  "+l) + "\n")
	}
	return b.String()
}
