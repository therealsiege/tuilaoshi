package reader

import (
	"fmt"
	"strings"

	"github.com/therealsiege/tuilaoshi/internal/ui"
)

func (m Model) View() string {
	var b strings.Builder
	b.WriteString(ui.TitleStyle.Render("Reading 读"))
	b.WriteString("\n\n")
	b.WriteString(ui.LoadError(m.loadErr))
	if m.saveErr != nil {
		b.WriteString(ui.ErrorStyle.Render(fmt.Sprintf("  Could not save progress: %v", m.saveErr)))
		b.WriteString("\n\n")
	}
	if m.open {
		b.WriteString(m.viewText())
	} else {
		b.WriteString(m.viewList())
	}
	return b.String()
}

func (m Model) viewList() string {
	var b strings.Builder
	if len(m.readings) == 0 {
		b.WriteString(ui.MutedStyle.Render("  No readings in wiki/readings yet."))
		b.WriteString("\n")
	}
	for i, p := range m.readings {
		line := fmt.Sprintf("%s  %s", p.Title, ui.MutedStyle.Render(p.Level))
		if i == m.listCursor {
			b.WriteString(ui.SelectedStyle.Render("▸ " + line))
		} else {
			b.WriteString(ui.UnselectedStyle.Render("  " + line))
		}
		b.WriteString("\n")
	}
	b.WriteString("\n")
	b.WriteString(ui.HelpStyle.Render("  ↑↓ choose · enter read · esc back"))
	return b.String()
}

func (m Model) viewText() string {
	var b strings.Builder
	b.WriteString(ui.SubtitleStyle.Render("  " + m.title))
	b.WriteString("\n\n")

	first, last := m.window()
	sel := place{-1, -1}
	if len(m.words) > 0 {
		sel = m.words[m.cursor]
	}
	for li := first; li < last; li++ {
		b.WriteString("  ")
		for ti, t := range m.lines[li] {
			switch {
			case li == sel.line && ti == sel.tok:
				b.WriteString(ui.SelectedStyle.Render(t.Text))
			case t.Word && m.marked[t.Text]:
				b.WriteString(ui.WarningStyle.Render(t.Text))
			default:
				b.WriteString(ui.ChineseStyle.Render(t.Text))
			}
		}
		b.WriteString("\n")
	}
	b.WriteString("\n")
	b.WriteString(ui.Divider(40))
	b.WriteString("\n")
	b.WriteString(m.viewGloss())
	b.WriteString("\n")
	b.WriteString(ui.HelpStyle.Render("  ←→ word · ↑↓ line · enter didn't know it (review tomorrow) · esc readings"))
	return b.String()
}

// window picks the text lines that fit on screen around the selected word.
func (m Model) window() (first, last int) {
	rows := m.height - 16
	if rows < 6 {
		rows = 6
	}
	if len(m.lines) <= rows {
		return 0, len(m.lines)
	}
	line := 0
	if len(m.words) > 0 {
		line = m.words[m.cursor].line
	}
	first = line - rows/2
	if first < 0 {
		first = 0
	}
	if first+rows > len(m.lines) {
		first = len(m.lines) - rows
	}
	return first, first + rows
}

// viewGloss shows the selected word's pinyin and meanings.
func (m Model) viewGloss() string {
	w := m.current()
	if w == "" {
		return ui.MutedStyle.Render("  This text has no words to look up.") + "\n"
	}
	var b strings.Builder
	cards := m.glosses(w)
	fmt.Fprintf(&b, "  %s", ui.ChineseStyle.Render(w))
	if len(cards) == 0 {
		b.WriteString(ui.MutedStyle.Render("  no gloss in the wiki"))
	}
	seen := map[string]bool{}
	for _, c := range cards {
		g := c.Pinyin + "  " + c.English
		if seen[g] {
			continue
		}
		seen[g] = true
		fmt.Fprintf(&b, "\n    %s  %s", ui.PinyinStyle.Render(c.Pinyin), c.English)
	}
	b.WriteString("\n")
	if m.status != "" {
		b.WriteString(ui.MutedStyle.Render("  " + m.status))
		b.WriteString("\n")
	}
	return b.String()
}
