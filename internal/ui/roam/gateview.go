package roam

import (
	"fmt"
	"slices"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/therealsiege/tuilaoshi/internal/journey"
	"github.com/therealsiege/tuilaoshi/internal/ui"
)

// gateView shows the gate: its pattern, what it means and model
// sentences, then the round under way (tiles to lay, or the sentence to
// type), what came of the last try, and the keys.
func (m Model) gateView() string {
	g := m.gate
	var b strings.Builder
	head := fmt.Sprintf("  语法关 Grammar gate: %s %s", g.Pattern, g.Pinyin)
	if !g.passed() {
		head += fmt.Sprintf(" · round %d of %d", g.round+1, gateRounds)
	}
	b.WriteString(ui.SubtitleStyle.Render(head) + "\n")
	b.WriteString(m.wrapped(ui.MutedStyle, g.Explain, "  "))
	examples := g.Examples
	if m.compact() && len(examples) > 2 {
		examples = examples[:2] // a short window keeps room for the round
	}
	for _, e := range examples {
		b.WriteString(m.wrapped(ui.ChineseStyle, "例 "+e.Chinese, "  "))
		b.WriteString(m.wrapped(ui.MutedStyle, e.English, "     "))
	}
	b.WriteString("\n")
	switch {
	case g.passed():
		b.WriteString(m.gateNotes())
		b.WriteString("  " + ui.SuccessStyle.Render("语法关已过！ The gate opens: the demon's lair lies ahead.") + "\n")
		b.WriteString("\n" + ui.HelpStyle.Render("  enter walk on"))
	case g.typed():
		b.WriteString(m.wrapped(ui.SubtitleStyle, fmt.Sprintf("Now a sentence of your own with %s, on anything:", g.Pattern), "  "))
		b.WriteString("  " + m.gateIn.View() + "\n\n")
		b.WriteString(m.gateNotes())
		b.WriteString(ui.HelpStyle.Render("  type pinyin, pick characters · enter send · esc leave"))
	default:
		s := g.sentence()
		b.WriteString(m.wrapped(ui.SubtitleStyle, "Lay the tiles in order: "+s.English, "  "))
		built := strings.Join(g.chosen(g.picked), "")
		if built == "" {
			built = ui.MutedStyle.Render("…")
		} else {
			built = ui.WarningStyle.Render(built)
		}
		b.WriteString("  → " + built + "\n")
		b.WriteString(m.tileRow() + "\n")
		b.WriteString(m.gateNotes())
		b.WriteString(ui.HelpStyle.Render("  1-9 lay a tile · backspace undo · esc leave"))
	}
	return b.String()
}

// tileRow lays out the round's tiles with their numbers, those laid
// already greyed, wrapped to the room there is.
func (m Model) tileRow() string {
	g := m.gate
	words := g.sentence().Words
	room := m.room()
	var lines []string
	line := " "
	for p, w := range g.tiles {
		tile := fmt.Sprintf("%d %s", p+1, words[w])
		style := ui.ChineseStyle
		if slices.Contains(g.picked, p) {
			style = ui.MutedStyle
		}
		if lipgloss.Width(line)+2+lipgloss.Width(tile) > room {
			lines = append(lines, line)
			line = " "
		}
		line += " " + style.Render(tile) + " "
	}
	return strings.Join(append(lines, line), "\n")
}

// room is the width an event's, plot's or gate's lines may take: the
// window (80 columns at most is plenty), or beside the sidebar the space
// left of it.
func (m Model) room() int {
	room := 78
	switch {
	case m.width > 0 && !m.compact() && m.width >= m.region.W+40:
		room = max(m.region.W, m.width-lipgloss.Width(m.sideView())-4) // left of the sidebar and its gap
	case m.width > 0:
		room = min(room, m.width-1)
	}
	return room
}

// gateNotes is what came of the last try and Claude's word on it, if any,
// with a blank line after.
func (m Model) gateNotes() string {
	g := m.gate
	var b strings.Builder
	if g.note != "" {
		style := ui.WarningStyle
		if strings.HasPrefix(g.note, "对") || strings.HasPrefix(g.note, "好") {
			style = ui.SuccessStyle
		}
		b.WriteString(m.wrapped(style, g.note, "  "))
	}
	if g.feedback != "" {
		lines := strings.Split(strings.TrimRight(m.wrapped(ui.MutedStyle, g.feedback, "  "), "\n"), "\n")
		b.WriteString(strings.Join(lines[:min(len(lines), 3)], "\n") + "\n") // Claude's word, at most three lines
	}
	if b.Len() > 0 {
		b.WriteString("\n")
	}
	return b.String()
}

// gateSide is the sidebar's word on the region's gate while its lair stands.
func (m Model) gateSide() string {
	g, ok := journey.GrammarGateFor(m.region.Region.ID)
	if _, won := m.cleared[m.region.Region.To]; !ok || won {
		return ""
	}
	state := ui.WarningStyle.Render("Guards the & lair: pass it first")
	if m.gatePassed() {
		state = ui.SuccessStyle.Render("Passed: the lair is open")
	}
	return "\n" + ui.SubtitleStyle.Render("语法关 Grammar gate") + "\n" + ui.ChineseStyle.Render(g.Pattern) + " " + ui.PinyinStyle.Render(g.Pinyin) + "\n" + state + "\n"
}

// gateLine is the gate in one line for the short map.
func (m Model) gateLine() string {
	g, ok := journey.GrammarGateFor(m.region.Region.ID)
	if _, won := m.cleared[m.region.Region.To]; !ok || won {
		return ""
	}
	state := ui.WarningStyle.Render("guards the & lair: walk into it to take the test")
	if m.gatePassed() {
		state = ui.SuccessStyle.Render("passed: the lair is open")
	}
	return ui.SubtitleStyle.Render("语法关") + " " + ui.ChineseStyle.Render(g.Pattern) + " " + state
}
