package collection

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

	"github.com/therealsiege/tuilaoshi/internal/journey"
	"github.com/therealsiege/tuilaoshi/internal/ui"
)

const (
	listWidth  = 40
	chromeRows = 9 // title, tabs, blank, note, help and spacing
	maxShown   = 160
)

// listRows is how many sets fit in the list.
func (m Model) listRows() int {
	if m.height == 0 {
		return 20
	}
	return max(m.height-chromeRows, 5)
}

func (m Model) View() string {
	var b strings.Builder
	b.WriteString(ui.TitleStyle.MarginBottom(0).Render("图鉴 Collection") + "\n")
	b.WriteString(ui.MutedStyle.Render(ui.WrapIndent("Every enemy beaten and every character mastered is a card. Complete a set to claim its reward.", m.width)) + "\n")
	b.WriteString(m.tabsLine() + "\n\n")
	b.WriteString(ui.LoadError(m.loadErr))

	list, detail := m.listView(), m.detailView()
	if m.width == 0 || m.width >= listWidth+48 {
		b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, list, "    ", detail))
	} else {
		b.WriteString(list + "\n\n" + detail)
	}
	b.WriteString("\n\n")
	if m.note != "" {
		style := ui.SuccessStyle
		if m.noteErr {
			style = ui.WarningStyle
		}
		b.WriteString(style.Render("  "+m.note) + "\n")
	}
	b.WriteString(ui.HelpStyle.Render("  ↑↓ choose a set · tab enemies / characters · enter claim · esc back"))
	return b.String()
}

func (m Model) tabsLine() string {
	names := []string{"妖怪 Enemies", "字 Characters"}
	for t := range names {
		total, done := 0, 0
		for _, s := range m.sets[t] {
			total++
			if s.Have(m.got[t]) == len(s.Members) {
				done++
			}
		}
		label := fmt.Sprintf("%s  %d/%d sets", names[t], done, total)
		if tab(t) == m.tab {
			names[t] = ui.ActiveTabStyle.Render(label)
		} else {
			names[t] = ui.MutedStyle.Render(label)
		}
	}
	return "  " + strings.Join(names, "   ")
}

// listView shows the visible sets with their progress.
func (m Model) listView() string {
	sets := m.sets[m.tab]
	if len(sets) == 0 {
		return ui.MutedStyle.Render("  Nothing to collect yet.")
	}
	end := min(m.offset[m.tab]+m.listRows(), len(sets))
	rows := make([]string, 0, end-m.offset[m.tab])
	for i := m.offset[m.tab]; i < end; i++ {
		s := sets[i]
		have, complete, claimed := m.status(m.tab, s)
		mark := ui.MutedStyle.Render(fmt.Sprintf("%d/%d", have, len(s.Members)))
		switch {
		case claimed:
			mark = ui.SuccessStyle.Render("✓ claimed")
		case complete:
			mark = ui.WarningStyle.Render("★ claim")
		}
		label := ansi.Truncate(s.Name, listWidth-lipgloss.Width(mark)-3, "…")
		row := label + strings.Repeat(" ", max(listWidth-2-lipgloss.Width(label)-lipgloss.Width(mark), 1)) + mark
		if i == m.cursor[m.tab] {
			rows = append(rows, ui.SelectedStyle.UnsetPadding().UnsetMargins().Render("▸ ")+row)
		} else {
			rows = append(rows, "  "+row)
		}
	}
	return lipgloss.NewStyle().MarginLeft(2).Render(strings.Join(rows, "\n"))
}

// detailView shows the selected set's cards: collected ones named, the rest hidden.
func (m Model) detailView() string {
	sets := m.sets[m.tab]
	if len(sets) == 0 {
		return ""
	}
	s := sets[m.cursor[m.tab]]
	have, _, _ := m.status(m.tab, s)
	var b strings.Builder
	b.WriteString(ui.SubtitleStyle.Render(s.Name) + "\n")
	b.WriteString(ui.MutedStyle.Render(fmt.Sprintf("%d of %d collected · reward %s", have, len(s.Members), priceText(s.Reward))) + "\n\n")
	if m.tab == tabEnemies {
		for _, name := range s.Members {
			e := journey.EnemyLookup(name)
			switch {
			case m.got[m.tab][name] && e != nil:
				b.WriteString(ui.ChineseStyle.Render(name) + "  " + e.English + "\n")
			default:
				b.WriteString(ui.MutedStyle.Render("？？？  not yet met in battle") + "\n")
			}
		}
		return lipgloss.NewStyle().Width(44).Render(b.String())
	}
	var chars []string
	for i, ch := range s.Members {
		if i == maxShown {
			chars = append(chars, ui.MutedStyle.Render(fmt.Sprintf("…and %d more", len(s.Members)-maxShown)))
			break
		}
		if m.got[m.tab][ch] {
			chars = append(chars, ui.ChineseStyle.Render(ch))
		} else {
			chars = append(chars, ui.MutedStyle.Render(ch))
		}
	}
	b.WriteString(strings.Join(chars, " ") + "\n\n")
	b.WriteString(ui.MutedStyle.Render("Bright characters are mastered; review the dim ones to complete the family."))
	return lipgloss.NewStyle().Width(44).Render(b.String())
}

func priceText(p journey.Price) string {
	var parts []string
	if p.Tongqian > 0 {
		parts = append(parts, fmt.Sprintf("%d 铜钱", p.Tongqian))
	}
	if p.Yuanbao > 0 {
		parts = append(parts, fmt.Sprintf("%d 元宝", p.Yuanbao))
	}
	return strings.Join(parts, " + ")
}
