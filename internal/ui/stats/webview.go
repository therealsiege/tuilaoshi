package stats

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

	"github.com/therealsiege/tuilaoshi/internal/camp"
	"github.com/therealsiege/tuilaoshi/internal/ui"
)

// Drawing the spider's web: the centre word, its strands to the words
// sharing a character, and their strands one step further out.

// spider is the web's place on the progress screen.
type spider struct {
	on     bool // the web is shown instead of the numbers
	g      camp.Web
	centre int      // the word in the middle, -1 for an empty web
	trail  []string // centres walked from, for going back
	sel    int      // the chosen strand from the centre
	walk   walk     // the quick review of a cluster, while one runs
	note   string   // a line under the web, e.g. nothing to revive
	err    error    // why the words or a review could not be read or saved
}

// Walking reports whether a walk of the web is under way; it uses esc
// itself, to go back to the web.
func (m Model) Walking() bool { return m.spider.walk.active() }

// loadWeb reads the studied words and rebuilds the web, keeping the same
// centre word when it is still there.
func (m Model) loadWeb() Model {
	words, err := m.store.StudiedWords()
	keep := m.centreWord()
	m.spider.g = camp.BuildWeb(words, m.now())
	m.spider.err = err
	if i, ok := m.spider.g.Index[keep]; ok && keep != "" {
		m.spider.centre = i
	} else {
		m.spider.centre, m.spider.trail, m.spider.sel = m.spider.g.Weakest(), nil, 0
	}
	m.spider.sel = min(m.spider.sel, max(0, len(m.strands())-1))
	return m
}

// centreWord is the word in the middle of the web, or "".
func (m Model) centreWord() string {
	g := m.spider.g
	if m.spider.centre < 0 || m.spider.centre >= len(g.Words) {
		return ""
	}
	return g.Words[m.spider.centre].Card.Chinese
}

// strands lists the words tied to the centre.
func (m Model) strands() []int {
	if m.centreWord() == "" {
		return nil
	}
	return m.spider.g.Links[m.spider.centre]
}

// updateWeb handles keys while the web is shown.
func (m Model) updateWeb(msg tea.KeyMsg) (Model, tea.Cmd) {
	sp := &m.spider
	n := len(m.strands())
	sp.note = ""
	switch msg.String() {
	case "w":
		sp.on = false
	case "up", "k":
		sp.sel = max(0, sp.sel-1)
	case "down", "j":
		sp.sel = min(max(0, n-1), sp.sel+1)
	case "right", "l", "tab":
		if n > 0 {
			sp.trail = append(sp.trail, m.centreWord())
			sp.centre, sp.sel = m.strands()[sp.sel], 0
		}
	case "left", "h", "backspace":
		if t := len(sp.trail); t > 0 {
			sp.centre, sp.sel = sp.g.Index[sp.trail[t-1]], 0
			sp.trail = sp.trail[:t-1]
		}
	case "c":
		sp.centre, sp.trail, sp.sel = sp.g.Weakest(), nil, 0
	case "enter", " ":
		return m.startWalk(), nil
	}
	return m, nil
}

// size is the screen size, defaulting to 80x24 before the first resize.
func (m Model) size() (w, h int) {
	w, h = m.width, m.height
	if w <= 0 {
		w = 80
	}
	if h <= 0 {
		h = 24
	}
	return w, h
}

// webView draws the web around the centre word.
func (m Model) webView() string {
	g := m.spider.g
	width, height := m.size()
	var b strings.Builder
	title := ui.TitleStyle.MarginBottom(0).Render("盘丝洞 The Spider's Web")
	counts := fmt.Sprintf("  %d words · %d strands · %d weak ╳", len(g.Words), g.Strands, g.WeakStrands)
	b.WriteString(fit(title+ui.MutedStyle.Render(counts), width) + "\n\n")
	if m.spider.err != nil {
		b.WriteString(fit(ui.ErrorStyle.Render("  "+m.spider.err.Error()), width) + "\n\n")
	}
	help := "  ↑↓ pick · → go · ← back · c weakest · enter revive · w numbers · esc camp"
	if m.centreWord() == "" {
		b.WriteString(ui.MutedStyle.Render("  No studied words yet: words join the web once reviewed.") + "\n")
		b.WriteString(ui.HelpStyle.Render(fit("  w numbers · esc camp", width)))
		return b.String()
	}
	centre := g.Words[m.spider.centre]
	b.WriteString(fit(ui.MutedStyle.Render("  centre ")+wordLine(centre), width) + "\n\n")

	links := m.strands()
	if len(links) == 0 {
		b.WriteString(ui.MutedStyle.Render(fmt.Sprintf("  No other studied word shares a character with %s yet.", centre.Card.Chinese)) + "\n")
	} else {
		rows := max(3, height-9)
		if m.spider.err != nil {
			rows = max(3, rows-2)
		}
		for _, r := range m.graphRows(links, rows, width) {
			b.WriteString(r + "\n")
		}
		nb := g.Words[links[m.spider.sel]]
		b.WriteString("\n" + fit(ui.MutedStyle.Render("  ▸ ")+wordLine(nb)+
			ui.MutedStyle.Render(" · shares "+camp.Shared(centre.Card.Chinese, nb.Card.Chinese)), width) + "\n")
	}
	if m.spider.note != "" {
		b.WriteString(fit(ui.WarningStyle.Render("  "+m.spider.note), width) + "\n")
	}
	if n := len(g.Cluster(m.spider.centre)); n > 0 {
		help = strings.Replace(help, "enter revive", fmt.Sprintf("enter revive %d", n), 1)
	}
	b.WriteString(ui.HelpStyle.Render(fit(help, width)))
	return b.String()
}

// graphRows draws the centre's strands as a tree, one strand a row, each
// followed by that word's own strands to words not already shown.
func (m Model) graphRows(links []int, rows, width int) []string {
	g := m.spider.g
	centre := g.Words[m.spider.centre]
	shown := min(len(links), rows)
	if len(links) > rows {
		shown = rows - 1 // the last row says how many more
	}
	off := max(0, m.spider.sel-(shown-1))
	total := shown
	if len(links) > shown {
		total++
	}
	onScreen := map[int]bool{m.spider.centre: true}
	for _, j := range links {
		onScreen[j] = true
	}
	cw := lipgloss.Width(centre.Card.Chinese)
	wordW := 4
	for _, j := range links[off : off+shown] {
		wordW = max(wordW, min(8, lipgloss.Width(g.Words[j].Card.Chinese)))
	}
	pinyin := width >= 100
	var out []string
	for k := 0; k < total; k++ {
		lead := "  " + strings.Repeat(" ", cw) + "  "
		branch := "├"
		if k == 0 {
			lead = "  " + wordStyle(centre).Render(centre.Card.Chinese) + ui.MutedStyle.Render(" ─")
			branch = "┬"
		}
		if k == total-1 {
			branch = "└"
		}
		if total == 1 {
			branch = "─"
		}
		if k == shown {
			out = append(out, lead+ui.MutedStyle.Render(fmt.Sprintf("%s─ … %d more", branch, len(links)-shown)))
			continue
		}
		j := links[off+k]
		nb := g.Words[j]
		strand := ui.SuccessStyle.Render("──") // ╳ marks a strand into a slipping word
		if nb.Weak {
			strand = ui.ErrorStyle.Render("╳─")
		}
		cursor := " "
		if off+k == m.spider.sel {
			cursor = ui.SelectedStyle.Render("▸")
		}
		word := ansi.Truncate(nb.Card.Chinese, 8, "…")
		row := lead + ui.MutedStyle.Render(branch) + strand + cursor + wordStyle(nb).Render(word) +
			strings.Repeat(" ", wordW-lipgloss.Width(word))
		if pinyin {
			py := ansi.Truncate(nb.Card.Pinyin, 14, "…")
			row += " " + ui.Pinyin(py) + strings.Repeat(" ", 14-lipgloss.Width(py))
		}
		out = append(out, row+m.outerRing(j, onScreen, width-lipgloss.Width(row)-1))
	}
	return out
}

// outerRing lists word j's strands to words not already on screen, as
// many as fit in room cells, with a count of the rest.
func (m Model) outerRing(j int, onScreen map[int]bool, room int) string {
	g := m.spider.g
	var far []int
	for _, k := range g.Links[j] {
		if !onScreen[k] {
			far = append(far, k)
		}
	}
	if len(far) == 0 || room < 8 {
		return ""
	}
	var b strings.Builder
	used := 1
	b.WriteString(" ")
	for i, k := range far {
		ww := g.Words[k]
		item := ansi.Truncate(ww.Card.Chinese, 8, "…")
		need := 2 + lipgloss.Width(item)
		reserve := 0
		if i < len(far)-1 {
			reserve = 5 // room for " +NNN"
		}
		if used+need+reserve > room {
			b.WriteString(ui.MutedStyle.Render(fmt.Sprintf(" +%d", len(far)-i)))
			break
		}
		mark := ui.SuccessStyle.Render("─")
		if ww.Weak {
			mark = ui.ErrorStyle.Render("╳")
		}
		b.WriteString(" " + mark + wordStyle(ww).Render(item))
		used += need
	}
	return b.String()
}

// wordStyle shows slipping words in red and firm ones in the usual ink.
func wordStyle(w camp.WebWord) lipgloss.Style {
	if w.Weak {
		return ui.ErrorStyle.Bold(true)
	}
	return ui.ChineseStyle
}

// wordLine is a word with its pinyin, meaning and how firmly it is held.
func wordLine(w camp.WebWord) string {
	return wordStyle(w).Render(w.Card.Chinese) + " " + ui.Pinyin(w.Card.Pinyin) + "  " +
		w.Card.English + ui.MutedStyle.Render(" · "+holdText(w))
}

// holdText says in words how firmly a word is held, so the web reads
// without colour.
func holdText(w camp.WebWord) string {
	parts := []string{fmt.Sprintf("every %dd", w.Card.ReviewInterval)}
	if w.Overdue {
		parts = append(parts, "due")
	}
	if w.Lapses > 0 {
		parts = append(parts, fmt.Sprintf("missed %d×", w.Lapses))
	}
	if w.Weak {
		parts = append(parts, "╳ slipping")
	} else {
		parts = append(parts, "firm")
	}
	return strings.Join(parts, ", ")
}

// fit cuts a line to the screen width.
func fit(s string, width int) string {
	return ansi.Truncate(s, width-1, "…")
}
