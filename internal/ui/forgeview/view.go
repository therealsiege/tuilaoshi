package forgeview

import (
	"fmt"
	"strings"

	"github.com/therealsiege/tuilaoshi/internal/journey"
	"github.com/therealsiege/tuilaoshi/internal/ui"
)

func (m Model) View() string {
	var b strings.Builder
	b.WriteString(ui.TitleStyle.MarginBottom(0).Render("字炉 Character Forge") + "\n")
	b.WriteString(ui.MutedStyle.Render("  Forge a character you're learning from its written parts.") + "\n\n")
	b.WriteString(ui.LoadError(m.loadErr))
	switch m.state {
	case stateLoading:
		b.WriteString(ui.MutedStyle.Render("  The forge heats up… (looking up new characters' parts)") + "\n")
	case stateEmpty:
		b.WriteString(ui.MutedStyle.Render("  Nothing to forge: no characters under review with known parts.") + "\n")
		b.WriteString(ui.MutedStyle.Render("  New characters' parts come from Claude (needs ANTHROPIC_API_KEY).") + "\n")
	case stateForging, stateResult:
		b.WriteString(m.puzzleView())
	}
	if m.note != "" {
		b.WriteString("\n" + ui.WarningStyle.Render("  "+m.note) + "\n")
	}
	help := "  esc back to camp"
	switch m.state {
	case stateForging:
		help = "  1-8 pick a part · backspace undo · enter forge · esc back"
	case stateResult:
		help = "  enter next character · esc back (saves what you forged)"
	}
	b.WriteString("\n" + ui.HelpStyle.Render(help))
	return b.String()
}

func (m Model) puzzleView() string {
	c := m.queue[m.cur]
	var b strings.Builder
	fmt.Fprintf(&b, "  %s  %s   %s\n\n", ui.MutedStyle.Render(fmt.Sprintf("%d/%d", m.cur+1, len(m.queue))),
		ui.PinyinStyle.Render(c.Pinyin), c.English)
	parts := m.comps[c.Chinese]
	var slots []string
	for i := range parts {
		if i < len(m.picked) {
			slots = append(slots, ui.ChineseStyle.Render(m.tray[m.picked[i]]))
		} else {
			slots = append(slots, ui.MutedStyle.Render("□"))
		}
	}
	extra := ""
	if len(m.picked) > len(parts) {
		extra = ui.ErrorStyle.Render(fmt.Sprintf("  (+%d too many)", len(m.picked)-len(parts)))
	}
	b.WriteString("  Parts: " + strings.Join(slots, " + ") + extra + "\n\n")
	var tray []string
	for i, p := range m.tray {
		tray = append(tray, fmt.Sprintf("%d %s", i+1, ui.ChineseStyle.Render(p)))
	}
	b.WriteString("  " + strings.Join(tray, "   ") + "\n")
	if m.state != stateResult {
		return b.String()
	}
	b.WriteString("\n")
	if m.right {
		b.WriteString(ui.SuccessStyle.Render(fmt.Sprintf("  Forged! %s = %s", c.Chinese, strings.Join(parts, " + "))) + "\n")
		if n := len(m.loot); n > 0 && m.paid[c.ID] {
			if mat := journey.MaterialByID(m.loot[n-1]); mat != nil {
				b.WriteString(ui.SuccessStyle.Render("  The forge leaves a "+mat.Name+" "+mat.English) + "\n")
			}
		}
	} else {
		b.WriteString(ui.WarningStyle.Render(fmt.Sprintf("  It cracks. %s is %s", c.Chinese, strings.Join(parts, " + "))) + "\n")
	}
	if r := m.radicals[c.Chinese]; r != "" {
		b.WriteString(ui.MutedStyle.Render("  radical "+r) + "\n")
	}
	return b.String()
}
