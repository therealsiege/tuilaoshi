package world

import (
	"fmt"
	"math/rand"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/therealsiege/tuilaoshi/internal/journey"
	"github.com/therealsiege/tuilaoshi/internal/ui"
)

// Buying in Chinese (量词): at the store you ask for what you want, and the
// measure word is yours to choose: 一包斋饭, 五双草鞋. The right one buys at
// the town's price; a wrong one makes the shopkeeper laugh, say the right
// one, and charge a fifth more.

// wrongMeasureMarkup is the extra a wrong measure word costs, in percent.
const wrongMeasureMarkup = 20

// measures are the measure words offered, the right one among them.
var measures = []string{"双", "包", "个", "条", "件", "张", "本", "只"}

// numerals are counts as said before a measure word (两, not 二).
var numerals = []string{"零", "一", "两", "三", "四", "五", "六", "七", "八", "九", "十"}

// purchase is a purchase being asked for.
type purchase struct {
	sandals bool // straw sandals for the party; else a bundle of 斋饭
	count   int
	right   string
	choices []string
	picked  int // -1 until chosen
}

// item is what is being bought, in Chinese.
func (p purchase) item() string {
	if p.sandals {
		return "草鞋"
	}
	return "斋饭"
}

// openBuy asks for 斋饭 (or the party's sandals), measure word to choose.
func (m Model) openBuy(sandals bool) Model {
	p := purchase{sandals: sandals, count: 1, right: "包", picked: -1}
	if sandals {
		m.withTrail(func(_ *journey.Trail, ms []string) { p.count = len(ms) })
		p.right = "双"
	}
	rng := rand.New(rand.NewSource(m.now().UnixNano()))
	p.choices = []string{p.right}
	for _, i := range rng.Perm(len(measures)) {
		if len(p.choices) < 4 && measures[i] != p.right {
			p.choices = append(p.choices, measures[i])
		}
	}
	rng.Shuffle(len(p.choices), func(a, b int) { p.choices[a], p.choices[b] = p.choices[b], p.choices[a] })
	m.buying, m.screen, m.note = p, screenBuy, ""
	return m
}

// updateBuy takes 1-4 for the measure word; after buying, enter returns.
func (m Model) updateBuy(k tea.KeyMsg) Model {
	p := &m.buying
	if p.picked >= 0 {
		if key.Matches(k, m.gkeys.Enter) || k.Type == tea.KeySpace {
			m.screen = screenTown
		}
		return m
	}
	s := k.String()
	if len(s) != 1 || s[0] < '1' || int(s[0]-'1') >= len(p.choices) {
		return m
	}
	p.picked = int(s[0] - '1')
	markup := 0
	if p.choices[p.picked] != p.right {
		markup = wrongMeasureMarkup
	}
	if p.sandals {
		return m.buySandals(markup)
	}
	return m.buy(markup)
}

// asked is the request with a measure word in its blank.
func (p purchase) asked(measure string) string {
	return "老板，我要" + numerals[min(p.count, 10)] + measure + p.item() + "。"
}

func (m Model) buyView() string {
	p := m.buying
	var b strings.Builder
	b.WriteString(ui.TitleStyle.MarginBottom(0).Render("杂货铺 The store in "+m.towns[m.town].Name) + "\n\n")
	b.WriteString(ui.MutedStyle.Render("  The shopkeeper looks up: 要什么？ Say it with the right measure word.") + "\n\n")
	b.WriteString("  " + ui.ChineseStyle.Render(p.asked("___")) + "\n\n")
	for i, c := range p.choices {
		line := fmt.Sprintf("  %d  %s", i+1, c)
		switch {
		case p.picked >= 0 && c == p.right:
			line = ui.SuccessStyle.Render(line + "  ✓")
		case i == p.picked:
			line = ui.ErrorStyle.Render(line + "  ✗")
		}
		b.WriteString(line + "\n")
	}
	if p.picked < 0 {
		b.WriteString("\n" + ui.HelpStyle.Render("  1-4 choose · esc leave"))
		return b.String()
	}
	b.WriteString("\n")
	if p.choices[p.picked] == p.right {
		b.WriteString("  " + ui.ChineseStyle.Render("老板：好嘞！") + "\n")
	} else {
		b.WriteString("  " + ui.ChineseStyle.Render(fmt.Sprintf("老板笑了：“%s论%s！”", p.item(), p.right)) + "\n")
		b.WriteString(m.lines(ui.MutedStyle, fmt.Sprintf("The shopkeeper laughs: %s are counted with %s. He charges a fifth more.", p.item(), p.right), "  ", 76))
	}
	b.WriteString("\n" + ui.HelpStyle.Render("  enter back to town"))
	return b.String()
}
