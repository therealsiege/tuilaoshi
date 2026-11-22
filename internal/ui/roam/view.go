package roam

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/therealsiege/tuilaoshi/internal/journey"
	"github.com/therealsiege/tuilaoshi/internal/ui"
	"github.com/therealsiege/tuilaoshi/internal/wiki"
)

// ground is how a terrain draws its impassable and wild cells.
var ground = map[string][2]string{
	"mountain": {"^", ","},
	"forest":   {"Y", `"`},
	"river":    {"~", ","},
	"desert":   {"n", "."},
	"fire":     {"^", ","},
	"palace":   {"#", ","},
	"cave":     {"%", "."},
}

var (
	blockStyle = map[string]lipgloss.Style{
		"mountain": ui.MutedStyle, "forest": ui.SuccessStyle, "river": ui.PinyinStyle, "desert": ui.WarningStyle,
		"fire": ui.ErrorStyle, "palace": ui.MutedStyle, "cave": ui.MutedStyle,
	}
	pathStyle   = ui.MutedStyle
	wildStyle   = ui.SuccessStyle
	playerStyle = ui.WarningStyle.Bold(true)
	townStyle   = ui.ChineseStyle.Bold(true)
)

func (m Model) View() string {
	r := m.region.Region
	var b strings.Builder
	b.WriteString(ui.TitleStyle.MarginBottom(0).Render(r.Name+"  "+r.English) + "  " +
		ui.MutedStyle.Render(fmt.Sprintf("region %d of %d · tribulations %d–%d", m.region.Index+1, len(journey.Regions()), r.From, r.To)) + "\n")
	b.WriteString(ui.LoadError(m.loadErr))
	grid := m.gridView()
	side := m.sideView()
	if m.width == 0 || m.width >= m.region.W+40 {
		b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, grid, "   ", side))
	} else {
		b.WriteString(grid + "\n" + side)
	}
	b.WriteString("\n" + m.roadStrip() + "\n")
	if m.note != "" {
		b.WriteString(ui.WarningStyle.Render("  "+m.note) + "\n")
	}
	b.WriteString(ui.HelpStyle.Render("  arrows walk · enter act · c camp (study, party, shrine…) · ? help"))
	return b.String()
}

// roadStrip shows the whole road west in one line or two: every town in
// order, with its requests done for those reached and the tribulation that
// reaches the rest, and where the pilgrim stands among them.
func (m Model) roadStrip() string {
	current := 0 // 0: the pilgrimage is complete
	if m.stage != nil {
		current = m.stage.Number
	}
	var parts []string
	here := false
	for _, t := range journey.Towns() {
		if !here && !t.Reached(current) {
			parts = append(parts, ui.WarningStyle.Render(fmt.Sprintf("◆ you, %d", current)))
			here = true
		}
		if !t.Reached(current) {
			parts = append(parts, ui.MutedStyle.Render(fmt.Sprintf("○ ？？？ %d", t.Stage)))
			continue
		}
		done := 0
		for _, q := range t.Quests {
			if m.states[q.ID] == wiki.QuestDone {
				done++
			}
		}
		count := fmt.Sprintf("%d/%d", done, len(t.Quests))
		if done < len(t.Quests) {
			count = ui.WarningStyle.Render(count)
		}
		parts = append(parts, "● "+ui.ChineseStyle.Render(t.Name)+" "+count)
	}
	width := 100
	if m.width > 0 {
		width = m.width - 4
	}
	var lines []string
	line := ui.SubtitleStyle.Render("Road")
	for _, p := range parts {
		if lipgloss.Width(line)+3+lipgloss.Width(p) > width {
			lines = append(lines, line)
			line = "    "
		} else {
			line += ui.MutedStyle.Render(" · ")
		}
		line += p
	}
	return "  " + strings.Join(append(lines, line), "\n  ")
}

func (m Model) gridView() string {
	terrain := m.region.Region.Terrain
	g := ground[terrain]
	var b strings.Builder
	for y := 0; y < m.region.H; y++ {
		b.WriteString("  ")
		for x := 0; x < m.region.W; x++ {
			p := journey.Point{X: x, Y: y}
			if p == m.pos {
				b.WriteString(playerStyle.Render("@"))
				continue
			}
			switch m.region.At(p) {
			case journey.TileBlock:
				b.WriteString(blockStyle[terrain].Render(g[0]))
			case journey.TilePath:
				b.WriteString(pathStyle.Render("·"))
			case journey.TileWild:
				b.WriteString(wildStyle.Render(g[1]))
			case journey.TileTown:
				b.WriteString(townStyle.Render("H"))
			case journey.TileStage:
				b.WriteString(m.stageGlyph(p))
			case journey.TileEast:
				b.WriteString(ui.SubtitleStyle.Render(">"))
			case journey.TileWest:
				b.WriteString(ui.SubtitleStyle.Render("<"))
			}
		}
		b.WriteString("\n")
	}
	return b.String()
}

// stageGlyph draws a tribulation: cleared, ready to fight, or sealed; the
// region's last is the demon's lair.
func (m Model) stageGlyph(p journey.Point) string {
	n := m.region.Stages[p]
	glyph := "!"
	if n == m.region.Region.To {
		glyph = "&"
	}
	switch {
	case m.stageCleared(p):
		return ui.MutedStyle.Render("+")
	case m.stage != nil && m.stage.Number == n && m.stage.Status == journey.StageOpen:
		return ui.ErrorStyle.Bold(true).Render(glyph)
	}
	return ui.MutedStyle.Render(glyph)
}

// hpBar shows the pilgrim's HP, and how to heal when hurt.
func (m Model) hpBar() string {
	if m.max == 0 {
		return ""
	}
	const width = 16
	filled := width * m.hp / m.max
	style := ui.SuccessStyle
	if m.hp*3 < m.max {
		style = ui.ErrorStyle
	}
	bar := fmt.Sprintf("HP %s %d/%d", style.Render(strings.Repeat("█", filled))+ui.MutedStyle.Render(strings.Repeat("░", width-filled)), m.hp, m.max)
	if m.hp < m.max {
		bar += "\n" + ui.MutedStyle.Render("Heal: study at camp, or an inn")
	}
	return bar
}

// sideView says what is ahead, the purse, and the legend.
func (m Model) sideView() string {
	var b strings.Builder
	if s := m.stage; s != nil && s.Number <= m.region.Region.To && s.Number >= m.region.Region.From {
		b.WriteString(ui.SubtitleStyle.Render(fmt.Sprintf("Tribulation %d", s.Number)) + "\n")
		b.WriteString(ui.ChineseStyle.Render(s.Name) + " " + s.English + "\n")
		b.WriteString("vs " + s.Opponent.Name + " " + ui.MutedStyle.Render(s.Opponent.English) + "\n")
		if s.Status == journey.StageOpen {
			b.WriteString(ui.WarningStyle.Render("Ready: walk into the red marker") + "\n")
		} else {
			b.WriteString(ui.MutedStyle.Render("Sealed: "+journey.NeedText(*s, m.counts)) + "\n")
		}
	} else if s != nil && s.Number > m.region.Region.To {
		b.WriteString(ui.SuccessStyle.Render("This region is cleared.") + "\n" + ui.MutedStyle.Render("The road west (<) is open.") + "\n")
	} else if s == nil {
		b.WriteString(ui.SuccessStyle.Render("The scriptures are won.") + "\n")
	}
	b.WriteString("\n" + m.hpBar() + "\n")
	b.WriteString(ui.WarningStyle.Render(fmt.Sprintf("%d 铜钱 · %d 元宝", m.wallet.Tongqian, m.wallet.Yuanbao)) + "\n\n")
	if len(m.quests) > 0 {
		b.WriteString(ui.SubtitleStyle.Render("Hunts") + "\n" + strings.Join(m.quests, "\n") + "\n\n")
	}
	g := ground[m.region.Region.Terrain]
	legend := []string{
		playerStyle.Render("@") + " you",
		townStyle.Render("H") + " town",
		ui.ErrorStyle.Render("!") + " tribulation  " + ui.ErrorStyle.Render("&") + " demon's lair",
		ui.MutedStyle.Render("+") + " cleared",
		wildStyle.Render(g[1]) + " wild ground",
		ui.SubtitleStyle.Render("< >") + " roads to the next regions",
	}
	b.WriteString(ui.MutedStyle.Render(strings.Join(legend, "\n")))
	return b.String()
}
