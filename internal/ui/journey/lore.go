package journey

import (
	"fmt"
	"strings"

	"github.com/therealsiege/tuilaoshi/internal/ui"
)

// renderLore renders the story passages view (tab 6).
func (m Model) renderLore() string {
	var b strings.Builder

	b.WriteString(ui.SubtitleStyle.Render("  西游故事 Story Passages"))
	b.WriteString("\n\n")

	if len(m.state.Lore) == 0 {
		b.WriteString(ui.MutedStyle.Render("  No lore entries yet..."))
		b.WriteString(m.tabFooter())
		return b.String()
	}

	unlocked := 0
	for _, entry := range m.state.Lore {
		if entry.Unlocked {
			unlocked++
		}
	}
	summary := fmt.Sprintf("  %d/%d passages unlocked", unlocked, len(m.state.Lore))
	b.WriteString(ui.MutedStyle.Render(summary))
	b.WriteString("\n\n")

	for i, entry := range m.state.Lore {
		if i > 0 {
			b.WriteString("\n")
			b.WriteString(ui.Divider(50))
			b.WriteString("\n\n")
		}

		if entry.Unlocked {
			// Title line: Chinese title + English title
			titleLine := fmt.Sprintf("  %s  %s",
				ui.ChineseStyle.Render(entry.Chinese),
				entry.Title)
			b.WriteString(titleLine)
			b.WriteString("\n")
			b.WriteString(ui.MutedStyle.Render(fmt.Sprintf("  Tribulation %d", entry.Tribulation)))
			b.WriteString("\n\n")

			// Chinese passage
			b.WriteString("  " + ui.ChineseStyle.Render(entry.Passage))
			b.WriteString("\n\n")

			// Pinyin
			b.WriteString("  " + ui.Pinyin(entry.Pinyin))
			b.WriteString("\n\n")

			// English translation
			b.WriteString("  " + ui.MutedStyle.Render(entry.English))
			b.WriteString("\n")
		} else {
			// Locked entry
			lockText := fmt.Sprintf("  ??? -- Reach tribulation %d to unlock", entry.Tribulation)
			b.WriteString(ui.MutedStyle.Render(lockText))
			b.WriteString("\n")
		}
	}

	b.WriteString(m.tabFooter())

	return b.String()
}

// padToWidth pads a string with spaces to reach the target visible width.
// It counts runes rather than bytes so CJK characters are handled reasonably.
