package party

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/therealsiege/tuilaoshi/internal/journey"
	"github.com/therealsiege/tuilaoshi/internal/ui"
)

const listWidth = 34

func (m Model) View() string {
	if m.reading {
		return m.sceneView()
	}
	var b strings.Builder
	b.WriteString(ui.TitleStyle.MarginBottom(0).Render("师徒 Party & Bonds") + "\n")
	b.WriteString(ui.MutedStyle.Render("  Bonds grow when a companion's assist helps you win with words you're still recalling.") + "\n\n")
	b.WriteString(ui.LoadError(m.loadErr))
	list, detail := m.listView(), m.detailView()
	if m.width == 0 || m.width >= listWidth+50 {
		b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, list, "    ", detail))
	} else {
		b.WriteString(list + "\n\n" + detail)
	}
	b.WriteString("\n\n")
	if m.note != "" {
		b.WriteString(ui.WarningStyle.Render("  "+m.note) + "\n")
	}
	b.WriteString(ui.HelpStyle.Render("  ↑↓ companion · 1-4 read a bond scene · esc back"))
	return b.String()
}

func (m Model) listView() string {
	rows := make([]string, 0, len(m.members))
	for i, b := range m.members {
		status := ui.MutedStyle.Render(fmt.Sprintf("joins at %d", b.Stage))
		if b.Unlocked {
			status = ui.ChineseStyle.Render(journey.BondRanks[b.rank()].Name)
		}
		label := b.Name + " " + b.English
		row := label + strings.Repeat(" ", max(listWidth-2-lipgloss.Width(label)-lipgloss.Width(status), 1)) + status
		if i == m.cursor {
			rows = append(rows, ui.SelectedStyle.UnsetPadding().UnsetMargins().Render("▸ ")+row)
		} else {
			rows = append(rows, "  "+row)
		}
	}
	return lipgloss.NewStyle().MarginLeft(2).Render(strings.Join(rows, "\n"))
}

// detailView shows the selected bond: rank, progress, assist and scenes.
func (m Model) detailView() string {
	if len(m.members) == 0 {
		return ""
	}
	b := m.members[m.cursor]
	r := b.rank()
	var s strings.Builder
	s.WriteString(ui.SubtitleStyle.Render(b.Name+"  "+b.English) + "\n")
	s.WriteString(ui.MutedStyle.Render(b.Ability+": "+b.AbilityDesc) + "\n\n")
	rank := journey.BondRanks[r]
	fmt.Fprintf(&s, "Bond %s %s · %d points", ui.ChineseStyle.Render(rank.Name), rank.English, b.points)
	if r+1 < len(journey.BondRanks) {
		next := journey.BondRanks[r+1]
		s.WriteString(ui.MutedStyle.Render(fmt.Sprintf(" · %d more to %s", next.Points-b.points, next.Name)))
	}
	s.WriteString("\n")
	upgrade := journey.BondRanks[journey.BondUpgradeRank]
	if r >= journey.BondUpgradeRank {
		s.WriteString(ui.SuccessStyle.Render("Upgraded assist: "+b.bond.UpgradeDesc) + "\n\n")
	} else {
		s.WriteString(ui.MutedStyle.Render(fmt.Sprintf("At %s: %s", upgrade.Name, b.bond.UpgradeDesc)) + "\n\n")
	}
	for i, sc := range b.bond.Scenes {
		line := fmt.Sprintf("%d  %s %s", i+1, sc.Title, sc.EnglishTitle)
		if r >= i+1 && b.Unlocked {
			s.WriteString(line + "\n")
		} else {
			s.WriteString(ui.MutedStyle.Render(fmt.Sprintf("%d  ？？？  opens at %s", i+1, journey.BondRanks[i+1].Name)) + "\n")
		}
	}
	return lipgloss.NewStyle().Width(48).Render(s.String())
}

func (m Model) sceneView() string {
	var b strings.Builder
	b.WriteString(m.scene.TextView(""))
	if m.showEnglish {
		b.WriteString("\n")
		for _, line := range strings.Split(ui.WrapText(strings.TrimSpace(m.sceneOf.English), 70), "\n") {
			b.WriteString(ui.MutedStyle.Render("  "+line) + "\n")
		}
	}
	b.WriteString("\n" + ui.HelpStyle.Render("  ←→ word · ↑↓ line · e translation · esc close"))
	return b.String()
}
