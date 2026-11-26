package world

import (
	"fmt"
	"slices"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/therealsiege/tuilaoshi/internal/journey"
	"github.com/therealsiege/tuilaoshi/internal/ui"
	"github.com/therealsiege/tuilaoshi/internal/wiki"
)

func (m Model) View() string {
	var b strings.Builder
	switch m.screen {
	case screenHaggle:
		b.WriteString(m.haggleView())
	case screenGate:
		b.WriteString(m.gateView())
	case screenPassport:
		b.WriteString(m.passportView())
	case screenEyes:
		b.WriteString(m.eyesView())
	case screenRiddle:
		b.WriteString(m.riddleView())
	case screenAlms:
		b.WriteString(m.almsView())
	case screenStone:
		b.WriteString(m.stoneView())
	case screenQuest:
		if m.Typing() {
			b.WriteString(m.writingView())
		} else if m.towns[m.town].Quests[m.quest].Hunting() {
			b.WriteString(m.huntView())
		} else {
			b.WriteString(m.questView())
		}
	default:
		b.WriteString(m.townView())
	}
	if m.note != "" {
		b.WriteString("\n" + ui.WarningStyle.Render("  "+m.note))
	}
	return b.String()
}

func (m Model) townView() string {
	t := m.towns[m.town]
	var b strings.Builder
	b.WriteString(ui.TitleStyle.MarginBottom(0).Render(t.Name+" "+t.English) + "\n")
	b.WriteString("  " + ui.ChineseStyle.Render(t.Desc) + "\n")
	b.WriteString(m.wrap(ui.MutedStyle, t.DescEnglish) + "\n")
	b.WriteString(ui.SubtitleStyle.Render("  告示栏 Notice board") + "\n")
	for i, q := range t.Quests {
		status := ui.WarningStyle.Render(fmt.Sprintf("read · +%d 铜钱", q.Pay(1)))
		if q.Writing() {
			status = ui.WarningStyle.Render(fmt.Sprintf("write · up to +%d 铜钱", q.Pay(2)))
		}
		if q.Hunting() && m.states[q.ID] != wiki.QuestDone {
			status = ui.WarningStyle.Render(fmt.Sprintf("hunt · +%d 铜钱", q.Pay(1)))
			if taken, beaten := m.huntState(q); taken && beaten >= q.Count {
				status = ui.SuccessStyle.Render("★ collect")
			} else if taken {
				status = ui.MutedStyle.Render(fmt.Sprintf("hunting %d/%d", beaten, q.Count))
			}
		}
		switch m.states[q.ID] {
		case wiki.QuestDone:
			status = ui.SuccessStyle.Render("✓ done")
		case wiki.QuestLocked:
			status = ui.MutedStyle.Render("tomorrow")
		}
		marker := "  "
		if i == m.quest {
			marker = ui.SelectedStyle.UnsetPadding().UnsetMargins().Render("▸ ")
		}
		fmt.Fprintf(&b, "  %s%s · %s  %s\n", marker, ui.ChineseStyle.Render(q.Title), q.Giver, status)
	}
	b.WriteString("\n")
	if m.height > 0 && m.height < compactTown {
		b.WriteString(m.wrap(ui.MutedStyle, fmt.Sprintf("客栈 inn %d (r) · 斋饭 %d (b) · 草鞋 %d (x) · 寺庙 offering %d (o) · trade (t) · alms (a) · haggle (h) · riddle (l) · papers (p) · rumour (f)",
			t.Price(journey.InnPrice), t.Price(journey.FoodPrice), t.Price(journey.SandalPrice), t.Price(journey.OfferingPrice))))
	} else {
		b.WriteString(m.wrap(ui.MutedStyle, fmt.Sprintf("客栈 The inn: a night's rest heals you and the party for %d 铜钱 (r)", t.Price(journey.InnPrice))))
		b.WriteString(m.wrap(ui.MutedStyle, fmt.Sprintf("杂货铺 The store: %d 斋饭 for %d 铜钱 (b), or %d for a material (t); 草鞋 at %d a pair (x) · haggle for a sack (h) · the well fills your water",
			journey.FoodBundle, t.Price(journey.FoodPrice), journey.Barter, t.Price(journey.SandalPrice))))
		b.WriteString(m.wrap(ui.MutedStyle, fmt.Sprintf("寺庙 The temple: an offering of %d 铜钱 lifts the party's spirits, once a day (o)", t.Price(journey.OfferingPrice))))
		b.WriteString(m.wrap(ui.MutedStyle, fmt.Sprintf("化缘 Ask the townsfolk for alms, once a day: +%d 斋饭 if asked well (a)", journey.AlmsFood)))
		lantern := "灯谜 A lantern riddle hangs by the gate: guess the character for 铜钱 (l)"
		if m.feast() {
			lantern = "元宵灯会 The Lantern Festival fair: three riddles today, double pay (l)"
		}
		b.WriteString(m.wrap(ui.MutedStyle, lantern))
		b.WriteString(m.wrap(ui.MutedStyle, eyesLine()))
	}
	if _, gate := journey.GateFor(t.ID); gate && !slices.Contains(m.stamps, t.ID) {
		b.WriteString(m.wrap(ui.WarningStyle, "通关文牒 The gate officer asks to see your travel papers (p)"))
	} else if m.height == 0 || m.height >= compactTown {
		b.WriteString(m.wrap(ui.MutedStyle, fmt.Sprintf("通关文牒 Your travel rescript: %d of %d seals (p)", len(m.stamps), len(journey.Gates()))))
	}
	help := "↑↓ choose · enter read · r rest · b buy · x sandals · o offering · t trade · a alms · h haggle · l riddle · p papers · f rumour · g shrine · esc back"
	if t.Stone != nil {
		if slices.Contains(m.stones, t.ID) {
			b.WriteString("  " + ui.MutedStyle.Render("石碑 The roadside stone stands open: ") + ui.ChineseStyle.Render(t.Stone.Idiom) + "\n")
		} else {
			b.WriteString(m.wrap(ui.MutedStyle, "There's an old stone by the road, half covered in moss. (s)"))
			help += " · s the stone"
		}
	}
	if m.height > 0 && m.height < compactTown {
		help = "↑↓ choose · enter read · g shrine · ? all keys · esc back"
		if t.Stone != nil && !slices.Contains(m.stones, t.ID) {
			help += " · s the stone"
		}
	}
	b.WriteString("\n" + m.wrap(ui.HelpStyle.UnsetMarginTop(), help))
	return strings.TrimRight(b.String(), "\n")
}

// lines renders text in style, wrapped to the window (or to most columns,
// when the window is wider), each line indented by pad.
func (m Model) lines(style lipgloss.Style, text, pad string, most int) string {
	width := most
	if m.width > 0 {
		width = min(most, m.width-lipgloss.Width(pad)-1)
	}
	var b strings.Builder
	for _, l := range ui.WrapDisplay(text, width) {
		b.WriteString(pad + style.Render(l) + "\n")
	}
	return b.String()
}

// compactTown is the window height below which the town folds its services
// into a few lines of keys and prices.
const compactTown = 34

// wrap renders text in style, two spaces in, wrapped at the window's width
// (by display width, so Chinese counts double), continuing four spaces in.
func (m Model) wrap(style lipgloss.Style, text string) string {
	width := m.width - 2
	if m.width <= 0 {
		width = 1 << 30
	}
	var lines []string
	line := ""
	for _, w := range strings.Fields(text) {
		indent := 2
		if len(lines) > 0 {
			indent = 4
		}
		if line != "" && indent+lipgloss.Width(line)+1+lipgloss.Width(w) > width {
			lines = append(lines, line)
			line = w
			continue
		}
		if line != "" {
			line += " "
		}
		line += w
	}
	lines = append(lines, line)
	var b strings.Builder
	for i, l := range lines {
		pad := "  "
		if i > 0 {
			pad = "    "
		}
		b.WriteString(pad + style.Render(l) + "\n")
	}
	return b.String()
}

// questView shows the notice and its question; the English appears only
// once the question is answered.
func (m Model) questView() string {
	q := m.towns[m.town].Quests[m.quest]
	var b strings.Builder
	b.WriteString(ui.TitleStyle.MarginBottom(0).Render(q.Title) + "  " + ui.MutedStyle.Render(q.Giver+" · "+q.GiverEnglish) + "\n\n")
	for _, line := range strings.Split(strings.TrimSpace(q.Text), "\n") {
		b.WriteString("  " + ui.ChineseStyle.Render(line) + "\n")
	}
	b.WriteString("\n  " + ui.SubtitleStyle.Render(q.Question) + "\n")
	for i, c := range m.choices {
		line := fmt.Sprintf("  %d  %s", i+1, c)
		switch {
		case m.answered >= 0 && i == m.right:
			line = ui.SuccessStyle.Render(line + "  ✓")
		case m.answered == i:
			line = ui.ErrorStyle.Render(line + "  ✗")
		}
		b.WriteString(line + "\n")
	}
	if m.answered < 0 {
		b.WriteString("\n" + ui.HelpStyle.Render("  1-4 answer · esc back"))
		return b.String()
	}
	b.WriteString("\n")
	if m.answered == m.right {
		b.WriteString(ui.SuccessStyle.Render(fmt.Sprintf("  Right! %s pays you %d 铜钱.", q.Giver, q.Pay(1))) + "\n")
	} else {
		b.WriteString(ui.WarningStyle.Render("  Not quite. Come back tomorrow and read it again.") + "\n")
	}
	for _, line := range strings.Split(ui.WrapText(strings.TrimSpace(q.English), 72), "\n") {
		b.WriteString(ui.MutedStyle.Render("  "+line) + "\n")
	}
	b.WriteString("\n" + ui.HelpStyle.Render("  enter back to the notice board"))
	return b.String()
}
