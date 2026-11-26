package world

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/therealsiege/tuilaoshi/internal/claude"
	"github.com/therealsiege/tuilaoshi/internal/journey"
	"github.com/therealsiege/tuilaoshi/internal/ui"
	"github.com/therealsiege/tuilaoshi/internal/wiki"
)

// Haggling (讨价还价) at the store: a merchant, played by Claude, asks too
// much for a big bundle of 斋饭 and bargains in Chinese. Good reasons said
// politely (we are monks, we buy a lot, your food is good) bring the price
// down, never below a floor; rudeness can push it up. ctrl+y takes the
// price on offer; after haggleRounds the merchant's word is final.

// The bundle haggled over, and how the price is set from the town's price
// for a bundle of 斋饭: asking 30% over, never below 60%.
const (
	haggleBundle = 20
	haggleRounds = 5
	haggleAsk    = 130 // percent of the town's price
	haggleFloor  = 60  // percent
)

// haggleLine is a line said while haggling, and its English.
type haggleLine struct {
	merchant bool
	zh, en   string
}

// haggling is a haggle in progress.
type haggling struct {
	price, floor int
	rounds       int
	history      []claude.ChatMessage
	lines        []haggleLine
	waiting      bool
	bought       bool
}

type haggleMsg struct {
	reply string
	err   error
}

// merchantPrompt sets the merchant's character and the reply format.
const merchantPrompt = `You are a shrewd but good-natured merchant in %s, a market town on the road west in Journey to the West, selling a sack of %d portions of 斋饭 (vegetarian food) to a travelling monk who is learning Chinese.
Your current price is %d 铜钱. You will never go below %d 铜钱.
Lower the price a little for good reasons said politely in Chinese (they are monks, they have little money, they buy a lot, they praise your food, they will tell others), more for very good or well-said ones. If they are rude or make no sense, keep the price or raise it a little. Bargain a bit before giving in.
Answer in one or two short sentences of simple spoken Chinese (HSK3-4), in character, naming your price.
Then on separate lines, exactly:
EN: <the English of what you said>
PRICE: <your price now, a whole number>`

// parseHaggle reads the merchant's reply: the Chinese, its English and the
// price, which can't go below floor or above the asking price.
func parseHaggle(reply string, current, floor int) (zh, en string, price int) {
	price = current
	var said []string
	for _, line := range strings.Split(reply, "\n") {
		line = strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(line, "EN:"):
			en = strings.TrimSpace(strings.TrimPrefix(line, "EN:"))
		case strings.HasPrefix(line, "PRICE:"):
			if n, err := strconv.Atoi(strings.TrimSpace(strings.TrimPrefix(line, "PRICE:"))); err == nil {
				price = n
			}
		case line != "":
			said = append(said, line)
		}
	}
	return strings.Join(said, " "), en, max(price, floor)
}

// openHaggle meets the merchant, asking too much.
func (m Model) openHaggle() (Model, tea.Cmd) {
	if m.talker == nil {
		m.note = "Haggling " + claude.NeedsKey + "."
		return m, nil
	}
	base := m.towns[m.town].Price(journey.FoodPrice) * haggleBundle / journey.FoodBundle
	h := &haggling{price: base * haggleAsk / 100, floor: base * haggleFloor / 100}
	h.lines = []haggleLine{{true, fmt.Sprintf("这袋斋饭有%d份，%d个铜钱，便宜吧？", haggleBundle, h.price),
		fmt.Sprintf("This sack holds %d portions, %d copper. Cheap, isn't it?", haggleBundle, h.price)}}
	h.history = []claude.ChatMessage{{Role: "user", Content: "（顾客走过来）"}, {Role: "assistant", Content: h.lines[0].zh}}
	m.haggle, m.screen, m.note = h, screenHaggle, ""
	m.input.Reset()
	return m, m.input.Focus()
}

// updateHaggle types an argument; enter says it, ctrl+y buys at the price.
func (m Model) updateHaggle(k tea.KeyMsg) (Model, tea.Cmd) {
	h := m.haggle
	switch {
	case h.bought:
		if k.Type == tea.KeyEnter || k.Type == tea.KeySpace {
			m.screen = screenTown
		}
		return m, nil
	case k.Type == tea.KeyCtrlY:
		return m.buyHaggled(), nil
	case h.waiting:
		return m, nil
	case k.Type == tea.KeyEnter && !m.input.Composing():
		said := strings.TrimSpace(m.input.Value())
		if said == "" {
			return m, nil
		}
		if h.rounds >= haggleRounds {
			m.note = "The merchant folds his arms: that is his last price. ctrl+y to buy, esc to leave."
			return m, nil
		}
		h.rounds++
		h.lines = append(h.lines, haggleLine{false, said, ""})
		h.waiting = true
		m.input.Reset()
		talker, history := m.talker, append([]claude.ChatMessage{}, h.history...)
		system := fmt.Sprintf(merchantPrompt, m.towns[m.town].Name, haggleBundle, h.price, h.floor)
		h.history = append(h.history, claude.ChatMessage{Role: "user", Content: said})
		return m, func() tea.Msg {
			reply, err := talker(context.Background(), system, history, said)
			return haggleMsg{reply, err}
		}
	}
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(k)
	return m, cmd
}

// onHaggle takes the merchant's answer.
func (m Model) onHaggle(msg haggleMsg) Model {
	h := m.haggle
	if h == nil {
		return m
	}
	h.waiting = false
	if msg.err != nil {
		m.note = "Couldn't reach Claude: " + msg.err.Error()
		return m
	}
	zh, en, price := parseHaggle(msg.reply, h.price, h.floor)
	h.price = min(price, h.price*12/10) // a rude buyer pays at most a fifth more
	h.lines = append(h.lines, haggleLine{true, zh, en})
	h.history = append(h.history, claude.ChatMessage{Role: "assistant", Content: zh})
	return m
}

// buyHaggled buys the sack at the price on offer.
func (m Model) buyHaggled() Model {
	h := m.haggle
	if err := m.store.Spend(wiki.Wallet{Tongqian: h.price}); err != nil {
		m.note = fmt.Sprintf("The sack is %d 铜钱: not enough in the purse.", h.price)
		return m
	}
	food := 0
	m.withTrail(func(t *journey.Trail, _ []string) {
		t.Food += haggleBundle
		food = t.Food
	})
	h.bought = true
	m.input.Blur()
	m.note = fmt.Sprintf("Bought %d 斋饭 for %d 铜钱: %d now.", haggleBundle, h.price, food)
	return m
}

func (m Model) haggleView() string {
	h := m.haggle
	var b strings.Builder
	b.WriteString(ui.TitleStyle.MarginBottom(0).Render("讨价还价 Haggling at the store in "+m.towns[m.town].Name) + "  " +
		ui.WarningStyle.Render(fmt.Sprintf("price now: %d 铜钱 for %d 斋饭", h.price, haggleBundle)) + "\n\n")
	for _, l := range h.lines[max(len(h.lines)-8, 0):] {
		who, style := "你  ", ui.UnselectedStyle.UnsetPadding()
		if l.merchant {
			who, style = "商人", ui.ChineseStyle
		}
		said := m.lines(style, l.zh, "       ", 100)
		b.WriteString("  " + ui.SubtitleStyle.Render(who) + " " + strings.TrimPrefix(said, "       "))
		if l.en != "" {
			b.WriteString(m.lines(ui.MutedStyle, l.en, "       ", 100))
		}
	}
	b.WriteString("\n")
	switch {
	case h.bought:
		b.WriteString(ui.SuccessStyle.Render("  The merchant ties up the sack: 慢走！") + "\n\n" + ui.HelpStyle.Render("  enter back to town"))
		return b.String()
	case h.waiting:
		b.WriteString(ui.MutedStyle.Render("  The merchant strokes his beard…") + "\n\n")
	default:
		b.WriteString("  " + m.input.View() + "\n\n")
	}
	b.WriteString(ui.HelpStyle.Render(fmt.Sprintf("  bargain in Chinese · enter say it (%d left) · ctrl+y buy at %d · esc leave", haggleRounds-h.rounds, h.price)))
	return b.String()
}
