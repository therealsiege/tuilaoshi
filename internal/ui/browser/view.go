package browser

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

	"github.com/therealsiege/tuilaoshi/internal/ui"
)

// Layout: a title line, a search line, the two panes, and a help line.
const (
	chromeRows = 6 // title, search, blank, help, and the panes' borders
	minPaneH   = 6
)

// listWidth is the list pane's text width: a third of the screen, 24-44.
func (m Model) listWidth() int {
	if m.width == 0 {
		return 34
	}
	return min(max(m.width/3, 24), 44)
}

// paneHeight is the inner height of both panes.
func (m Model) paneHeight() int {
	if m.height == 0 {
		return 20
	}
	return max(m.height-chromeRows, minPaneH)
}

// listRows is how many list entries fit in the list pane.
func (m Model) listRows() int { return m.paneHeight() }

// layoutViewport sizes the page pane to what is left beside the list.
func (m *Model) layoutViewport() {
	w := 60
	if m.width > 0 {
		w = max(m.width-m.listWidth()-8, 24) // two panes' borders and padding
	}
	m.viewport.Width, m.viewport.Height = w, m.paneHeight()
}

func (m Model) View() string {
	var b strings.Builder
	count := fmt.Sprintf("%s pages", thousands(len(m.entries)))
	if m.query != "" {
		count = fmt.Sprintf("%s of %s pages match", thousands(len(m.shown)), thousands(len(m.entries)))
	}
	b.WriteString(ui.TitleStyle.MarginBottom(0).Render("Wiki Browser 浏览") + "  " + ui.MutedStyle.Render(count) + "\n")
	b.WriteString(m.searchLine() + "\n")
	b.WriteString(ui.LoadError(m.loadErr))

	// Slim panes: the shared border style's padding would eat rows and columns.
	listStyle, pageStyle := ui.BorderStyle.Padding(0, 1), ui.BorderStyle.Padding(0, 1)
	if m.focus == focusList {
		listStyle = ui.ActiveBorderStyle.Padding(0, 1)
	} else {
		pageStyle = ui.ActiveBorderStyle.Padding(0, 1)
	}
	h := m.paneHeight()
	list := listStyle.Width(m.listWidth() + 2).Height(h).MaxHeight(h + 2).Render(m.listView())
	page := pageStyle.Width(m.viewport.Width + 2).Height(h).MaxHeight(h + 2).Render(m.viewport.View())
	b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, list, page) + "\n")
	b.WriteString(ui.HelpStyle.Render("  " + m.help()))
	return b.String()
}

func (m Model) searchLine() string {
	switch {
	case m.searching:
		return "  " + ui.WarningStyle.Render("/ ") + m.query + ui.SelectedStyle.Render(" ")
	case m.query != "":
		return "  " + ui.MutedStyle.Render("/ "+m.query+"   (/ to change, esc to clear)")
	}
	return "  " + ui.MutedStyle.Render("/ to search titles, pinyin (tones optional), meanings and tags")
}

// listView renders only the rows that fit, from the scroll offset.
func (m Model) listView() string {
	if len(m.shown) == 0 {
		return ui.MutedStyle.Render(" No pages match.")
	}
	w := m.listWidth()
	end := min(m.offset+m.listRows(), len(m.shown))
	rows := make([]string, 0, end-m.offset)
	for i := m.offset; i < end; i++ {
		e := m.entries[m.shown[i]]
		label := e.icon + " " + e.title
		if e.level != "" {
			label += " " + ui.MutedStyle.Render(e.level)
		}
		label = ansi.Truncate(label, w-1, "…") // one column for the ▸ marker
		if i == m.cursor {
			rows = append(rows, ui.SelectedStyle.UnsetPadding().UnsetMargins().Render("▸"+label))
		} else {
			rows = append(rows, " "+label)
		}
	}
	return strings.Join(rows, "\n")
}

func (m Model) help() string {
	switch {
	case m.searching:
		return "type to search · ↑↓ results · enter read · esc clear"
	case m.focus == focusContent:
		return "↑↓ scroll · tab/← list · / search · esc back"
	}
	return "↑↓ move · pgup/pgdn page · enter read · / search · esc back"
}

// thousands writes 6017 as "6,017".
func thousands(n int) string {
	s := fmt.Sprint(n)
	for i := len(s) - 3; i > 0; i -= 3 {
		s = s[:i] + "," + s[i:]
	}
	return s
}
