package journey

import (
	"fmt"
	"strings"

	"github.com/therealsiege/tuilaoshi/internal/journey"
	"github.com/therealsiege/tuilaoshi/internal/ui"
	"github.com/therealsiege/tuilaoshi/internal/ui/imgview"
)

// renderBestiary renders the demon catalog view (tab 7).
func (m Model) renderBestiary() string {
	var b strings.Builder

	b.WriteString(ui.TitleStyle.Render("  妖怪图鉴 Demon Bestiary"))
	b.WriteString("\n\n")

	entries := journey.Bestiary()
	if len(entries) == 0 {
		b.WriteString(ui.MutedStyle.Render("  No demons cataloged..."))
		b.WriteString(m.tabFooter())
		return b.String()
	}

	// Build a set of defeated boss names for quick lookup.
	defeated := make(map[string]bool)
	for _, boss := range m.state.Bosses {
		if boss.Defeated {
			defeated[boss.Name] = true
		}
	}

	// Clamp cursor.
	cursor := m.bestiaryCursor
	if cursor >= len(entries) {
		cursor = len(entries) - 1
	}
	if cursor < 0 {
		cursor = 0
	}

	// LEFT pane: scrollable demon name list.
	var left strings.Builder
	for i, entry := range entries {
		mark := "·"
		if defeated[entry.Name] {
			mark = ui.SuccessStyle.Render("✓")
		}

		name := fmt.Sprintf("%s %s  %s", mark, entry.Name, entry.English)
		if i == cursor {
			left.WriteString(ui.SelectedStyle.Render(" > " + name))
		} else {
			left.WriteString("   " + name)
		}
		left.WriteString("\n")
	}

	// RIGHT pane: detail panel for the selected demon.
	selected := entries[cursor]
	var right strings.Builder

	// Demon image: try terminal image first, fall back to ASCII art.
	rendered := imgview.RenderWithFallback(selected.ImagePath, 30, 15, selected.Art)
	if rendered != "" {
		for _, line := range strings.Split(rendered, "\n") {
			right.WriteString("  " + line + "\n")
		}
		right.WriteString("\n")
	}

	// Title / epithet
	right.WriteString("  ")
	right.WriteString(ui.ChineseStyle.Render(selected.Title))
	right.WriteString("  ")
	right.WriteString(ui.MutedStyle.Render(selected.TitleEng))
	right.WriteString("\n\n")

	// Origin
	right.WriteString(ui.SubtitleStyle.Render("  Origin"))
	right.WriteString("\n")
	right.WriteString("  " + ui.MutedStyle.Render(ui.WrapText(selected.Origin, 44)))
	right.WriteString("\n\n")

	// Powers
	right.WriteString(ui.SubtitleStyle.Render("  Powers"))
	right.WriteString("\n")
	right.WriteString("  " + ui.MutedStyle.Render(ui.WrapText(selected.Powers, 44)))
	right.WriteString("\n\n")

	// Weakness
	right.WriteString(ui.SubtitleStyle.Render("  Weakness"))
	right.WriteString("\n")
	right.WriteString("  " + ui.MutedStyle.Render(ui.WrapText(selected.Weakness, 44)))
	right.WriteString("\n\n")

	// Chapter
	right.WriteString(ui.SubtitleStyle.Render("  Chapter"))
	right.WriteString("\n")
	right.WriteString("  " + ui.MutedStyle.Render(selected.Chapter))
	right.WriteString("\n")

	// Compose the two-pane layout side by side.
	leftLines := strings.Split(strings.TrimRight(left.String(), "\n"), "\n")
	rightLines := strings.Split(strings.TrimRight(right.String(), "\n"), "\n")

	// Fixed left column width for alignment.
	leftWidth := 36

	maxLines := len(leftLines)
	if len(rightLines) > maxLines {
		maxLines = len(rightLines)
	}

	sep := ui.MutedStyle.Render(" │ ")
	for i := 0; i < maxLines; i++ {
		lc := ""
		if i < len(leftLines) {
			lc = leftLines[i]
		}
		rc := ""
		if i < len(rightLines) {
			rc = rightLines[i]
		}
		// Pad the left column to a fixed width (using rune count for CJK).
		padded := padToWidth(lc, leftWidth)
		b.WriteString(padded)
		b.WriteString(sep)
		b.WriteString(rc)
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(ui.HelpStyle.Render("  ↑↓/jk: scroll demons | tab/←→: switch tabs | esc: dashboard"))

	return b.String()
}
