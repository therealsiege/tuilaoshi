package shrine

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

	"github.com/therealsiege/tuilaoshi/internal/journey"
	rpgengine "github.com/therealsiege/tuilaoshi/internal/rpg"
	"github.com/therealsiege/tuilaoshi/internal/ui"
)

// listWidth is the shelf's text width.
const listWidth = 44

func (m Model) View() string {
	var b strings.Builder
	b.WriteString(ui.TitleStyle.MarginBottom(0).Render("土地庙 Earth God's Shrine"))
	b.WriteString("  " + ui.WarningStyle.Render(fmt.Sprintf("%d 铜钱 · %d 元宝", m.wallet.Tongqian, m.wallet.Yuanbao)) + "\n")
	b.WriteString(ui.MutedStyle.Render("  The Earth God trades gear for the coins your study earns. Gear helps in battle; it never changes your reviews.") + "\n\n")
	b.WriteString(ui.LoadError(m.loadErr))

	shelf, sheet := m.shelfView(), m.sheetView()
	if m.width == 0 || m.width >= listWidth+46 {
		b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, shelf, "    ", sheet))
	} else {
		b.WriteString(shelf + "\n\n" + sheet)
	}
	b.WriteString("\n\n")
	if m.note != "" {
		style := ui.SuccessStyle
		if m.noteErr {
			style = ui.WarningStyle
		}
		b.WriteString(style.Render("  "+m.note) + "\n")
	}
	b.WriteString(ui.HelpStyle.Render("  ↑↓ choose · enter buy / put on / take off · esc back"))
	return b.String()
}

// shelfView lists every item by slot: worn, owned, for sale, or still to come.
func (m Model) shelfView() string {
	var rows []string
	slot := ""
	for i, g := range m.items {
		if g.Slot != slot {
			slot = g.Slot
			if len(rows) > 0 {
				rows = append(rows, "")
			}
			rows = append(rows, ui.SubtitleStyle.Render(journey.SlotNames[slot]))
		}
		var status string
		switch {
		case m.kit.Worn[g.Slot] == g.ID:
			status = ui.SuccessStyle.Render("worn")
		case m.owns(g.ID):
			status = ui.MutedStyle.Render("owned")
		case !g.OnShelf(m.cleared):
			status = ui.MutedStyle.Render(fmt.Sprintf("after %d", g.Opens))
		case g.Crafted():
			status = ui.SuccessStyle.Render("craft")
		default:
			status = ui.WarningStyle.Render(priceText(g.Price))
		}
		label := g.Name + " " + g.English
		if !g.OnShelf(m.cleared) && !m.owns(g.ID) {
			label = ui.MutedStyle.Render(label)
		}
		label = ansi.Truncate(label, listWidth-lipgloss.Width(status)-3, "…")
		gap := strings.Repeat(" ", max(listWidth-2-lipgloss.Width(label)-lipgloss.Width(status), 1))
		row := label + gap + status
		if i == m.cursor {
			rows = append(rows, ui.SelectedStyle.UnsetPadding().UnsetMargins().Render("▸ ")+row)
		} else {
			rows = append(rows, "  "+row)
		}
	}
	return lipgloss.NewStyle().MarginLeft(2).Render(strings.Join(rows, "\n"))
}

// sheetView describes the selected item, then the character as equipped.
func (m Model) sheetView() string {
	var b strings.Builder
	if len(m.items) > 0 {
		g := m.items[m.cursor]
		b.WriteString(ui.ChineseStyle.Render(g.Name) + "  " + g.English + "\n")
		b.WriteString(ui.WrapText(g.Desc, 40) + "\n")
		switch {
		case g.Crafted():
			b.WriteString(ui.MutedStyle.Render("Craft from: "+recipeText(g.Recipe, g.Price)) + "\n")
		case g.Price != (journey.Price{}):
			b.WriteString(ui.MutedStyle.Render("Price: "+priceText(g.Price)) + "\n")
		}
		b.WriteString("\n")
	}

	worn := journey.Worn(m.kit.Worn)
	title := "修行者 Pilgrim"
	var e rpgengine.Effects
	effects := rpgengine.AddGear(&e, worn)
	for _, g := range worn {
		if g.Slot == "title" {
			title = g.Name + " " + g.English
		}
	}
	b.WriteString(ui.SubtitleStyle.Render("Materials") + "  " + m.matsText() + "\n\n")
	b.WriteString(ui.SubtitleStyle.Render("Your pilgrim") + "\n")
	b.WriteString(ui.ChineseStyle.Render(title) + "\n")
	fmt.Fprintf(&b, "HP %d · +%d damage per hit\n", rpgengine.PlayerHP(m.mastered)+e.BonusHP, e.BonusDamage)
	for _, slot := range journey.GearSlots[:4] {
		name := ui.MutedStyle.Render("—")
		if g := journey.GearByID(m.kit.Worn[slot]); g != nil {
			name = g.Name + " " + g.English
		}
		fmt.Fprintf(&b, "%s  %s\n", ui.MutedStyle.Width(12).Render(journey.SlotNames[slot]), name)
	}
	if len(effects) > 0 {
		b.WriteString("\n" + ui.MutedStyle.Render(strings.Join(effects, "\n")))
	}
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

// matsText lists the materials held, or how to get them.
func (m Model) matsText() string {
	var parts []string
	for _, mat := range journey.Materials() {
		if n := m.mats[mat.ID]; n > 0 {
			parts = append(parts, fmt.Sprintf("%s ×%d", mat.Name, n))
		}
	}
	if len(parts) == 0 {
		return ui.MutedStyle.Render("none yet: road wins that recall shaky words drop them")
	}
	return strings.Join(parts, "  ")
}
