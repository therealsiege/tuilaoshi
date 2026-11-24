package world

import (
	"fmt"
	"slices"
	"strings"

	"github.com/therealsiege/tuilaoshi/internal/journey"
	"github.com/therealsiege/tuilaoshi/internal/ui"
	"github.com/therealsiege/tuilaoshi/internal/wiki"
)

func (m Model) View() string {
	var b strings.Builder
	switch m.screen {
	case screenAlms:
		b.WriteString(m.almsView())
	case screenStone:
		b.WriteString(m.stoneView())
	case screenQuest:
		if m.Typing() {
			b.WriteString(m.writingView())
		} else if m.towns[m.town].Quests[m.quest].Hunting() {
			b.WriteString(m.huntView())
		} else {
			b.WriteString(m.questView())
		}
	default:
		b.WriteString(m.townView())
	}
	if m.note != "" {
		b.WriteString("\n" + ui.WarningStyle.Render("  "+m.note))
	}
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
		status := ui.WarningStyle.Render(fmt.Sprintf("read · +%d 铜钱", q.Pay(1)))
		if q.Writing() {
			status = ui.WarningStyle.Render(fmt.Sprintf("write · up to +%d 铜钱", q.Pay(2)))
		}
		if q.Hunting() && m.states[q.ID] != wiki.QuestDone {
			status = ui.WarningStyle.Render(fmt.Sprintf("hunt · +%d 铜钱", q.Pay(1)))
			if taken, beaten := m.huntState(q); taken && beaten >= q.Count {
				status = ui.SuccessStyle.Render("★ collect")
			} else if taken {
				status = ui.MutedStyle.Render(fmt.Sprintf("hunting %d/%d", beaten, q.Count))
			}
		}
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
	b.WriteString("\n  " + ui.MutedStyle.Render(fmt.Sprintf("客栈 The inn: a night's rest heals you and the party for %d 铜钱 (r)", wiki.InnPrice)) + "\n")
	b.WriteString("  " + ui.MutedStyle.Render(fmt.Sprintf("杂货铺 The store: %d 斋饭 for %d 铜钱 (b), or %d for a material (t) · the well fills your water", foodBundle, foodPrice, journey.Barter)) + "\n")
	b.WriteString("  " + ui.MutedStyle.Render(fmt.Sprintf("化缘 Ask the townsfolk for alms, once a day: +%d 斋饭 if asked well (a)", journey.AlmsFood)) + "\n")
	help := "  ↑↓ choose · enter read · r rest · b buy · t trade · a alms · g shrine · esc back on the road"
	if t.Stone != nil {
		if slices.Contains(m.stones, t.ID) {
			b.WriteString("\n  " + ui.MutedStyle.Render("石碑 The roadside stone stands open: ") + ui.ChineseStyle.Render(t.Stone.Idiom) + "\n")
		} else {
			b.WriteString("\n  " + ui.MutedStyle.Render("There's an old stone by the road, half covered in moss. (s)") + "\n")
			help += " · s the stone"
		}
	}
	b.WriteString("\n" + ui.HelpStyle.Render(help))
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
		b.WriteString(ui.SuccessStyle.Render(fmt.Sprintf("  Right! %s pays you %d 铜钱.", q.Giver, q.Pay(1))) + "\n")
	} else {
		b.WriteString(ui.WarningStyle.Render("  Not quite. Come back tomorrow and read it again.") + "\n")
	}
	for _, line := range strings.Split(ui.WrapText(strings.TrimSpace(q.English), 72), "\n") {
		b.WriteString(ui.MutedStyle.Render("  "+line) + "\n")
	}
	b.WriteString("\n" + ui.HelpStyle.Render("  enter back to the notice board"))
	return b.String()
}
