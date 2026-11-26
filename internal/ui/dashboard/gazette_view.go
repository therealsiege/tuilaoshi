package dashboard

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/therealsiege/tuilaoshi/internal/ui"
)

// updateGazette reads the open issue: e swaps Chinese and English when the
// window holds only one, ↑↓ scroll, enter or esc go back to camp.
func (m Model) updateGazette(k tea.KeyMsg) Model {
	switch {
	case key.Matches(k, m.gkeys.Back), key.Matches(k, m.gkeys.Enter):
		m.gazetteOpen, m.gazetteEN, m.gazetteTop, m.note = false, false, 0, ""
	case k.String() == "e":
		m.gazetteEN, m.gazetteTop = !m.gazetteEN, 0
	case key.Matches(k, m.gkeys.Up):
		m.gazetteTop = max(m.gazetteTop-1, 0)
	case key.Matches(k, m.gkeys.Down):
		m.gazetteTop++
	}
	return m
}

// gazetteBody is the issue's lines, wrapped to the window: the Chinese,
// then its English.
func (m Model) gazetteBody() (zh, en []string) {
	width := m.bodyWidth()
	for _, l := range strings.Split(m.gazette.Text, "\n") {
		for _, w := range ui.WrapDisplay(l, width) {
			zh = append(zh, "  "+ui.ChineseStyle.Render(w))
		}
	}
	for _, l := range strings.Split(m.gazette.English, "\n") {
		for _, w := range ui.WrapDisplay(l, width) {
			en = append(en, ui.MutedStyle.Render("  "+w))
		}
	}
	return zh, en
}

// bodyWidth is how wide the issue's lines run.
func (m Model) bodyWidth() int {
	if m.width > 0 {
		return min(m.width-4, 100)
	}
	return 76
}

// gazetteView is the open issue, with its English below when the window
// holds both, else one at a time (e swaps) and scrolled with ↑↓.
func (m Model) gazetteView() string {
	var head strings.Builder
	head.WriteString(ui.TitleStyle.MarginBottom(0).Render("邸报 The court gazette") + "  " + ui.MutedStyle.Render("week of "+m.gazette.Week))
	if m.gazette.ByHand {
		head.WriteString(ui.MutedStyle.Render(" · set by the clerk"))
	}
	head.WriteString("\n\n")
	var foot strings.Builder
	zh, en := m.gazetteBody()
	if m.note != "" {
		for _, l := range ui.WrapDisplay(m.note, m.bodyWidth()) {
			foot.WriteString(ui.WarningStyle.Render("  "+l) + "\n")
		}
	}
	body := append(append(append([]string{}, zh...), ""), en...)
	help := "  ctrl+r hear it · enter or esc back to camp"
	room := 1 << 30
	if m.height > 0 {
		room = m.height - 4 - strings.Count(foot.String(), "\n")
	}
	if len(body) > room {
		body = zh
		if m.gazetteEN {
			body = en
		}
		help = "  ctrl+r hear it · e English/中文 · enter or esc back to camp"
		if len(body) > room {
			top := min(m.gazetteTop, len(body)-room+1)
			body = append(body[top:top+room-1], ui.MutedStyle.Render("  … ↑↓ for more"))
			help = "  ctrl+r hear it · e English/中文 · ↑↓ scroll · esc back"
		}
	}
	return head.String() + strings.Join(body, "\n") + "\n\n" + foot.String() + ui.HelpStyle.UnsetMarginTop().Render(help)
}
