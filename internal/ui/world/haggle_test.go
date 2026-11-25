package world

import (
	"context"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/therealsiege/tuilaoshi/internal/claude"
	"github.com/therealsiege/tuilaoshi/internal/journey"
	"github.com/therealsiege/tuilaoshi/internal/ui/uitest"
	"github.com/therealsiege/tuilaoshi/internal/wiki"
)

func TestParseHaggle(t *testing.T) {
	zh, en, price := parseHaggle("好吧，二十五个铜钱。\nEN: Fine, 25 copper.\nPRICE: 25", 39, 18)
	if zh != "好吧，二十五个铜钱。" || en != "Fine, 25 copper." || price != 25 {
		t.Errorf("%q %q %d", zh, en, price)
	}
	if _, _, price := parseHaggle("五个！\nPRICE: 5", 39, 18); price != 18 {
		t.Errorf("below the floor: %d", price)
	}
	if _, _, price := parseHaggle("嗯……", 30, 18); price != 30 {
		t.Errorf("no PRICE line keeps the price: %d", price)
	}
}

func TestHagglingForASack(t *testing.T) {
	m := worldAt(t, 11)
	m.talker = nil
	if m, _ = m.Update(press("h")); !strings.Contains(m.note, "ANTHROPIC_API_KEY") {
		t.Fatalf("no Claude: %q", m.note)
	}
	m.talker = func(_ context.Context, system string, _ []claude.ChatMessage, said string) (string, error) {
		if !strings.Contains(system, "never go below") || said != "我们是出家人，便宜一点儿吧。" {
			t.Errorf("merchant asked %q with %q", said, system)
		}
		return "出家人啊，那就二十五个吧。\nEN: Monks, eh? Then 25.\nPRICE: 25", nil
	}
	m, _ = m.Update(press("h"))
	base := m.towns[m.town].Price(journey.FoodPrice) * haggleBundle / journey.FoodBundle
	if m.screen != screenHaggle || m.haggle.price != base*haggleAsk/100 {
		t.Fatalf("asking price %d", m.haggle.price)
	}
	m.input.SetValue("我们是出家人，便宜一点儿吧。")
	m, cmd := m.Update(enter)
	if cmd == nil {
		t.Fatal("nothing said")
	}
	m, _ = m.Update(cmd())
	uitest.Golden(t, "haggle", m.View())
	if m.haggle.price != 25 || !strings.Contains(m.View(), "Monks, eh?") {
		t.Fatalf("price %d\n%s", m.haggle.price, m.View())
	}
	if err := m.store.Earn(wiki.Wallet{Tongqian: 30}); err != nil {
		t.Fatal(err)
	}
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyCtrlY})
	tr, _, err := journey.LoadTrail(m.store, nil)
	w, werr := m.store.Wallet()
	if err != nil || werr != nil || tr.Food != journey.StartFood+haggleBundle || w.Tongqian != 5 {
		t.Errorf("bought: food %d purse %+v (%v %v)", tr.Food, w, err, werr)
	}
}
