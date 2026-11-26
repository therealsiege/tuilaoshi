package assessment

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/therealsiege/tuilaoshi/internal/assess"
	"github.com/therealsiege/tuilaoshi/internal/ui"
	"github.com/therealsiege/tuilaoshi/internal/wiki"
)

// pastShown is how many past sittings are listed.
const pastShown = 5

// lastLines caps the feedback shown under a question.
const lastLines = 3

func (m Model) viewMock() string {
	switch m.phase {
	case phaseMockPick:
		return m.viewMockPick()
	case phaseMock:
		return m.viewMockSection()
	}
	return m.viewMockDone()
}

func (m Model) viewMockPick() string {
	var b strings.Builder
	b.WriteString("  " + ui.SubtitleStyle.Render("Mock test 模拟考") + "\n\n")
	b.WriteString(ui.WrapIndent("A timed paper in the HSK format, made from the wiki's own pages: "+
		"listening (10, spoken aloud), reading (15) and writing (3 sentences of word tiles, "+
		"and one sentence for Claude when it is set up). About 15 minutes; a pass is 60% of the maximum. "+
		"Sit it monthly to see whether the level check's estimate holds.", m.width) + "\n\n")
	fmt.Fprintf(&b, "  %s HSK4    %s HSK5\n", ui.ChineseStyle.Render("4"), ui.ChineseStyle.Render("5"))
	if !m.audible() {
		b.WriteString(ui.MutedStyle.Render("  No audio: the listening section will be skipped.") + "\n")
	}
	b.WriteString("\n" + m.viewSittings())
	b.WriteString(ui.HelpStyle.Render("  4/5 start · esc back"))
	return b.String()
}

// viewSittings lists the latest sittings, or says there are none yet.
func (m Model) viewSittings() string {
	var b strings.Builder
	if m.mock.saveErr != nil {
		b.WriteString(ui.ErrorStyle.Render(ui.WrapIndent(fmt.Sprintf("Could not save or read sittings: %v", m.mock.saveErr), m.width)) + "\n\n")
	}
	all := m.mock.sittings
	if len(all) == 0 {
		b.WriteString(ui.MutedStyle.Render("  No sittings yet.") + "\n\n")
		return b.String()
	}
	b.WriteString("  " + ui.SubtitleStyle.Render("Past sittings") + "\n")
	for i := len(all) - 1; i >= 0 && i >= len(all)-pastShown; i-- {
		b.WriteString(sittingLine(all[i]) + "\n")
	}
	return b.String() + "\n"
}

func sittingLine(s wiki.MockSitting) string {
	parts := []string{s.Date, fmt.Sprintf("HSK%d", s.Level)}
	for _, sec := range s.Sections {
		mark := "—"
		if !sec.Skipped {
			mark = fmt.Sprint(sec.Score)
		}
		parts = append(parts, fmt.Sprintf("%s %3s", strings.Fields(sec.Name)[0][:1], mark))
	}
	parts = append(parts, fmt.Sprintf("%d/%d", s.Total, s.Max), verdictWord(s.Passed))
	return "  " + strings.Join(parts, "  ")
}

func verdictWord(passed bool) string {
	if passed {
		return ui.SuccessStyle.Render("合格 pass")
	}
	return ui.ErrorStyle.Render("不合格 not yet")
}

// timeLeft is the section's time left as m:ss.
func (m Model) timeLeft() string {
	left := m.mock.deadline.Sub(m.clock())
	if left < 0 {
		left = 0
	}
	secs := int((left + time.Second - 1) / time.Second)
	return fmt.Sprintf("%d:%02d left", secs/60, secs%60)
}

func (m Model) viewMockSection() string {
	var b strings.Builder
	t := m.mock
	it := m.current()
	fmt.Fprintf(&b, "  %s  %s\n\n", ui.SubtitleStyle.Render(fmt.Sprintf("%s · HSK%d · %d of %d",
		assess.SectionNames[t.section], t.test.Level, t.item+1, len(t.items[t.section]))), ui.MutedStyle.Render(m.timeLeft()))
	if it.Text != "" {
		b.WriteString(ui.ChineseStyle.Render(ui.WrapIndent(it.Text, m.width)) + "\n\n")
	}
	if it.Audio != "" {
		b.WriteString(ui.MutedStyle.Render("  (listen: tab hears it again)") + "\n")
	}
	b.WriteString(ui.SubtitleStyle.Render(ui.WrapIndent(it.Prompt, m.width)) + "\n\n")
	help := "  ↑↓ choose · enter/1-4 answer · esc stop (nothing saved)"
	switch it.Kind {
	case assess.Tiles:
		b.WriteString(m.viewTiles(it))
		help = "  1-6 place a tile · backspace undo · esc stop (nothing saved)"
	case assess.Compose:
		b.WriteString("  " + t.input.View() + "\n")
		if t.grading {
			b.WriteString(ui.MutedStyle.Render("  Claude is grading…") + "\n")
		}
		help = "  enter send · esc stop (nothing saved)"
	default:
		b.WriteString(m.viewChoices(it))
		if it.Audio != "" {
			help = "  ↑↓ choose · 1-4 answer · tab hear again · esc stop (nothing saved)"
		}
	}
	if t.last != "" {
		lines := strings.Split(ui.WrapIndent("last: "+t.last, m.width), "\n")
		b.WriteString("\n" + ui.MutedStyle.Render(strings.Join(lines[:min(lastLines, len(lines))], "\n")) + "\n")
	}
	if t.audioErr != nil {
		b.WriteString("\n" + ui.ErrorStyle.Render(ui.WrapIndent("Audio failed: "+t.audioErr.Error(), m.width)) + "\n")
	}
	b.WriteString("\n" + ui.HelpStyle.Render(help))
	return b.String()
}

func (m Model) viewChoices(it assess.Item) string {
	var b strings.Builder
	width := 76
	if m.width > 0 {
		width = min(width, m.width-4)
	}
	for i, c := range it.Choices {
		lines := ui.WrapDisplay(fmt.Sprintf("%d. %s", i+1, c), width-5)
		line := strings.Join(lines, "\n       ")
		if i == m.mock.selected {
			b.WriteString(ui.SelectedStyle.Render("  ▸ " + line))
		} else {
			b.WriteString(ui.UnselectedStyle.Render("    " + line))
		}
		b.WriteString("\n")
	}
	return b.String()
}

func (m Model) viewTiles(it assess.Item) string {
	placed := map[int]bool{}
	sentence := ""
	for _, p := range m.mock.placed {
		placed[p] = true
		sentence += it.Tiles[p]
	}
	tiles := make([]string, 0, len(it.Tiles))
	for i, tile := range it.Tiles {
		if placed[i] {
			tiles = append(tiles, ui.MutedStyle.Render(fmt.Sprintf("%d %s", i+1, strings.Repeat("·", len([]rune(tile))))))
		} else {
			tiles = append(tiles, fmt.Sprintf("%s %s", ui.ChineseStyle.Render(fmt.Sprint(i+1)), tile))
		}
	}
	width := 78
	if m.width > 0 {
		width = min(width, m.width-2)
	}
	var b strings.Builder
	line := ""
	for _, tile := range tiles {
		if lipgloss.Width(line+"  "+tile) > width {
			b.WriteString(line + "\n")
			line = ""
		}
		line += "  " + tile
	}
	b.WriteString(line + "\n\n")
	b.WriteString("  ▸ " + ui.ChineseStyle.Render(sentence) + ui.MutedStyle.Render("_") + "\n")
	return b.String()
}

func (m Model) viewMockDone() string {
	var b strings.Builder
	t := m.mock
	fmt.Fprintf(&b, "  %s\n\n", ui.SubtitleStyle.Render(fmt.Sprintf("Mock test 模拟考 · HSK%d", t.test.Level)))
	for s, sc := range t.scores {
		mark := fmt.Sprintf("%3d / 100", sc.Score())
		if sc.Skipped {
			mark = ui.MutedStyle.Render("skipped: " + t.why[s])
		}
		fmt.Fprintf(&b, "  %-16s %s\n", assess.SectionNames[s], mark)
	}
	b.WriteString("\n")
	if t.most == 0 {
		b.WriteString(ui.WrapIndent("Nothing could be sat: the wiki needs more pages at this level.", m.width) + "\n\n")
	} else {
		fmt.Fprintf(&b, "  Total %d / %d · pass mark %d · %s\n", t.total, t.most, (t.most*assess.MockPassPercent+99)/100, verdictWord(t.passed))
		b.WriteString(ui.WrapIndent(assess.Verdict(t.test.Level, t.passed, t.estimate), m.width) + "\n\n")
	}
	b.WriteString(m.viewSittings())
	b.WriteString(ui.HelpStyle.Render("  m another sitting · enter dashboard · esc back"))
	return b.String()
}
