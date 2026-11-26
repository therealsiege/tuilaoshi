package world

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/therealsiege/tuilaoshi/internal/journey"
	"github.com/therealsiege/tuilaoshi/internal/ui"
)

// errandView shows the goal and its points, checked off as they are got
// across, and as much of the talk as the window holds, newest last.
func (m Model) errandView() string {
	e := m.errand
	er := e.errand
	var head strings.Builder
	head.WriteString(ui.TitleStyle.MarginBottom(0).Render("任务 An errand in "+m.towns[m.town].Name) + "  " +
		ui.MutedStyle.Render(er.Role+" "+er.Name+" · "+er.RoleEnglish) + "\n")
	head.WriteString(m.lines(ui.ChineseStyle, "目标："+er.Goal, "  ", 76))
	head.WriteString(m.lines(ui.MutedStyle, er.GoalEnglish, "  ", 76))
	for _, p := range er.Points {
		mark, style := "☐", ui.MutedStyle
		if e.done[p.ID] {
			mark, style = "✓", ui.SuccessStyle
		}
		head.WriteString(m.lines(style, mark+" "+p.Chinese+"  "+p.English, "  ", 76))
	}

	var foot strings.Builder
	left := journey.ErrandTurns - e.turns
	switch {
	case e.finished:
		foot.WriteString(m.lines(ui.SuccessStyle, fmt.Sprintf("任务完成！%s thanks you: +%d 铜钱, and the party's spirits lift.", er.Name, er.Reward), "  ", 76))
		foot.WriteString("\n" + ui.HelpStyle.Render("  enter back to town"))
	case e.over():
		foot.WriteString(m.lines(ui.WarningStyle, er.Name+" has someone else to see to. The errand is left undone; try another tomorrow.", "  ", 76))
		foot.WriteString("\n" + ui.HelpStyle.Render("  enter back to town"))
	case e.waiting:
		foot.WriteString(ui.MutedStyle.Render("  "+er.Name+" thinks…") + "\n\n")
		foot.WriteString(ui.HelpStyle.Render("  esc leave (the day's errand is spent)"))
	default:
		foot.WriteString("  " + m.input.View() + "\n\n")
		foot.WriteString(m.lines(ui.HelpStyle.UnsetMarginTop(), fmt.Sprintf("enter say it in Chinese (%d left) · ctrl+space pinyin/English · esc leave", left), "  ", 200))
	}

	// Each line is labelled by who said it, the label padded so the words
	// line up under one another.
	label := max(lipgloss.Width(er.Name), 4)
	pad := strings.Repeat(" ", label+3)
	var talk []string
	for _, l := range e.lines {
		who, style := ui.SubtitleStyle.Render("你"+strings.Repeat(" ", label-2)), ui.UnselectedStyle.UnsetPadding()
		if l.merchant {
			who, style = ui.ChineseStyle.Render(er.Name+strings.Repeat(" ", label-lipgloss.Width(er.Name))), ui.ChineseStyle
		}
		said := strings.Split(strings.TrimRight(m.lines(style, l.zh, pad, 100), "\n"), "\n")
		said[0] = "  " + who + " " + strings.TrimPrefix(said[0], pad)
		talk = append(talk, said...)
		if l.en != "" {
			talk = append(talk, strings.Split(strings.TrimRight(m.lines(ui.MutedStyle, l.en, pad, 100), "\n"), "\n")...)
		}
	}
	// The talk gets the rows left once the goal, the box and a note fit.
	if m.height > 0 {
		used := strings.Count(head.String(), "\n") + strings.Count(foot.String(), "\n") + 4
		if m.note != "" {
			used += strings.Count(m.lines(ui.WarningStyle, m.note, "  ", 76), "\n")
		}
		talk = talk[max(len(talk)-max(m.height-used, 2), 0):]
	}
	return head.String() + "\n" + strings.Join(talk, "\n") + "\n\n" + strings.TrimRight(foot.String(), "\n")
}
