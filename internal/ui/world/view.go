package world

import (
	"fmt"
	"strings"

	"github.com/therealsiege/tuilaoshi/internal/journey"
	"github.com/therealsiege/tuilaoshi/internal/ui"
	"github.com/therealsiege/tuilaoshi/internal/wiki"
)

func (m Model) View() string {
	var b strings.Builder
	switch m.screen {
	case screenQuest:
		b.WriteString(m.questView())
	case screenTown:
		b.WriteString(m.townView())
	default:
		b.WriteString(m.mapView())
	}
	if m.note != "" {
		b.WriteString("\n" + ui.WarningStyle.Render("  "+m.note))
	}
	return b.String()
}

// mapView draws the road as a column of towns, marking where the
// pilgrimage stands between them.
func (m Model) mapView() string {
	var b strings.Builder
	b.WriteString(ui.TitleStyle.MarginBottom(0).Render("西行路 The Road West") + "\n")
	b.WriteString(ui.MutedStyle.Render("  Towns between the tribulations. Their notice boards hold requests to read and answer.") + "\n\n")
	b.WriteString(ui.LoadError(m.loadErr))
	// Connect the towns with a line when the terminal is tall enough.
	link := m.height == 0 || m.height >= 2*len(m.towns)+10
	here := false
	for i, t := range m.towns {
		reached := t.Reached(m.current)
		if !here && !reached {
			b.WriteString("     " + ui.WarningStyle.Render(fmt.Sprintf("◆ you are here, tribulation %d", m.current)) + "\n")
			if link {
				b.WriteString("     " + ui.MutedStyle.Render("│") + "\n")
			}
			here = true
		}
		dot, name := "●", ui.ChineseStyle.Render(t.Name)+" "+t.English
		if !reached {
			dot, name = "○", ui.MutedStyle.Render("？？？")
		}
		status := ui.MutedStyle.Render(fmt.Sprintf("tribulation %d", t.Stage))
		if reached {
			done := 0
			for _, q := range t.Quests {
				if m.states[q.ID] == wiki.QuestDone {
					done++
				}
			}
			status = ui.MutedStyle.Render(fmt.Sprintf("%d/%d requests", done, len(t.Quests)))
			if done < len(t.Quests) {
				status = ui.WarningStyle.Render(fmt.Sprintf("%d/%d requests", done, len(t.Quests)))
			}
		}
		marker := "  "
		if i == m.town {
			marker = ui.SelectedStyle.UnsetPadding().UnsetMargins().Render("▸ ")
		}
		fmt.Fprintf(&b, "   %s%s %s  %s\n", marker, dot, name, status)
		if link && i < len(m.towns)-1 {
			b.WriteString("     " + ui.MutedStyle.Render("│") + "\n")
		}
	}
	if !here && m.current != 0 {
		b.WriteString("     " + ui.WarningStyle.Render(fmt.Sprintf("◆ you are here, tribulation %d", m.current)) + "\n")
	}
	b.WriteString("\n" + ui.HelpStyle.Render("  ↑↓ choose a town · enter visit · esc back"))
	return b.String()
}

func (m Model) townView() string {
	t := m.towns[m.town]
	var b strings.Builder
	b.WriteString(ui.TitleStyle.MarginBottom(0).Render(t.Name+" "+t.English) + "\n")
	b.WriteString("  " + ui.ChineseStyle.Render(t.Desc) + "\n")
	b.WriteString(ui.MutedStyle.Render("  "+t.DescEnglish) + "\n\n")
	b.WriteString(ui.SubtitleStyle.Render("  告示栏 Notice board") + "\n")
	for i, q := range t.Quests {
		status := ui.WarningStyle.Render(fmt.Sprintf("+%d 铜钱", journey.QuestReward))
		switch m.states[q.ID] {
		case wiki.QuestDone:
			status = ui.SuccessStyle.Render("✓ done")
		case wiki.QuestLocked:
			status = ui.MutedStyle.Render("tomorrow")
		}
		marker := "  "
		if i == m.quest {
			marker = ui.SelectedStyle.UnsetPadding().UnsetMargins().Render("▸ ")
		}
		fmt.Fprintf(&b, "  %s%s · %s  %s\n", marker, ui.ChineseStyle.Render(q.Title), q.Giver, status)
	}
	b.WriteString("\n" + ui.HelpStyle.Render("  ↑↓ choose · enter read · g the Earth God's shrine · esc the road"))
	return b.String()
}

// questView shows the notice and its question; the English appears only
// once the question is answered.
func (m Model) questView() string {
	q := m.towns[m.town].Quests[m.quest]
	var b strings.Builder
	b.WriteString(ui.TitleStyle.MarginBottom(0).Render(q.Title) + "  " + ui.MutedStyle.Render(q.Giver+" · "+q.GiverEnglish) + "\n\n")
	for _, line := range strings.Split(strings.TrimSpace(q.Text), "\n") {
		b.WriteString("  " + ui.ChineseStyle.Render(line) + "\n")
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
	if m.answered < 0 {
		b.WriteString("\n" + ui.HelpStyle.Render("  1-4 answer · esc back"))
		return b.String()
	}
	b.WriteString("\n")
	if m.answered == m.right {
		b.WriteString(ui.SuccessStyle.Render(fmt.Sprintf("  Right! %s pays you %d 铜钱.", q.Giver, journey.QuestReward)) + "\n")
	} else {
		b.WriteString(ui.WarningStyle.Render("  Not quite. Come back tomorrow and read it again.") + "\n")
	}
	for _, line := range strings.Split(ui.WrapText(strings.TrimSpace(q.English), 72), "\n") {
		b.WriteString(ui.MutedStyle.Render("  "+line) + "\n")
	}
	b.WriteString("\n" + ui.HelpStyle.Render("  enter back to the notice board"))
	return b.String()
}
