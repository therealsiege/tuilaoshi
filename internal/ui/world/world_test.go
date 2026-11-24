package world

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/therealsiege/tuilaoshi/internal/journey"
	"github.com/therealsiege/tuilaoshi/internal/ui/uitest"
	"github.com/therealsiege/tuilaoshi/internal/wiki"
)

var day = time.Date(2026, 11, 22, 9, 0, 0, 0, time.Local)

// worldAt is the road with stages 1..cleared won and nothing mastered.
func worldAt(t *testing.T, cleared int) Model {
	t.Helper()
	var b strings.Builder
	b.WriteString("cleared:\n")
	for i := 1; i <= cleared; i++ {
		b.WriteString("    " + itoa(i) + ": \"2026-11-01\"\n")
	}
	store := uitest.Wiki(t, map[string]string{"journey.yaml": b.String()})
	m := New(store, nil, nil)
	m.now = func() time.Time { return day }
	return m.SetSize(100, 50).Refresh().OpenTown("changan")
}

func itoa(n int) string { return fmt.Sprint(n) }

func press(s string) tea.KeyMsg { return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)} }

var enter = tea.KeyMsg{Type: tea.KeyEnter}

func TestTownScreen(t *testing.T) {
	m := worldAt(t, 11).OpenTown("gao-village")
	uitest.Golden(t, "town", m.View())
	if _, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc}); cmd == nil {
		t.Error("esc in town should go back on the road")
	}
}

func TestAnsweringARequest(t *testing.T) {
	m := worldAt(t, 11)
	m, _ = m.Update(enter) // the bun seller's notice
	if m.screen != screenQuest || strings.Contains(m.View(), "Help wanted") {
		t.Fatalf("quest not open, or English shown before answering:\n%s", m.View())
	}
	m, _ = m.Update(press(string(rune('1' + m.right))))
	w, err := m.store.Wallet()
	if err != nil || w.Tongqian != 20 || !strings.Contains(m.View(), "Help wanted") {
		t.Fatalf("right answer: purse %+v %v\n%s", w, err, m.View())
	}
	m, _ = m.Update(enter)
	if m.screen != screenTown || m.states["changan-baker"] != wiki.QuestDone {
		t.Fatalf("back on the board: screen %d, state %v", m.screen, m.states["changan-baker"])
	}
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m, _ = m.Update(enter) // the gate notice: answer wrong
	m, _ = m.Update(press(string(rune('1' + (m.right+1)%len(m.choices)))))
	m, _ = m.Update(enter)
	if m.states["changan-gate"] != wiki.QuestLocked {
		t.Fatalf("wrong answer should lock the request today: %v", m.states["changan-gate"])
	}
	m, _ = m.Update(enter)
	if m.screen != screenTown || !strings.Contains(m.note, "tomorrow") {
		t.Errorf("a locked request opened: screen %d, note %q", m.screen, m.note)
	}
}

func TestWritingRequestPaysWhenItPasses(t *testing.T) {
	m := worldAt(t, 11)
	scores := []int{0, 2}
	m.grader = func(_ context.Context, task, criteria, answer string) (int, string, error) {
		if !strings.Contains(task, "招牌") || criteria == "" {
			t.Errorf("grader got task %q, criteria %q", task, criteria)
		}
		s := scores[0]
		scores = scores[1:]
		return s, "反馈 feedback", nil
	}
	for i, q := range m.towns[0].Quests {
		if q.Writing() {
			m.quest = i
		}
	}
	m, _ = m.Update(enter)
	if !m.Typing() {
		t.Fatal("the writing request didn't open for typing")
	}
	m, _ = m.Update(press("q")) // typing q must not quit or leave
	m.input.Reset()
	for _, r := range "dabianyi" {
		m, _ = m.Update(press(string(r)))
	}
	send := func() {
		t.Helper()
		var cmd tea.Cmd
		m, cmd = m.Update(enter)
		for m.input.Composing() && cmd == nil { // pick the first candidate, then send
			m, cmd = m.Update(enter)
		}
		if cmd == nil {
			t.Fatal("nothing sent")
		}
		m, _ = m.Update(cmd())
	}
	send()
	if m.answered > 0 || !strings.Contains(m.View(), "Not yet") {
		t.Fatalf("a failing answer passed:\n%s", m.View())
	}
	send()
	w, err := m.store.Wallet()
	if err != nil || w.Tongqian != 40 || !strings.Contains(m.View(), "pays you 40") {
		t.Fatalf("a passing answer: purse %+v %v\n%s", w, err, m.View())
	}
	m, _ = m.Update(enter)
	if m.screen != screenTown {
		t.Errorf("enter after passing should return to the board, at %d", m.screen)
	}
}

func TestStoneOpensToTheIdiomTypedFromMemory(t *testing.T) {
	m := worldAt(t, 11)
	if !strings.Contains(m.View(), "old stone") {
		t.Fatalf("the town doesn't mention its stone:\n%s", m.View())
	}
	m, _ = m.Update(press("s"))
	if m.screen != screenStone || !strings.Contains(m.View(), "一□□□") || strings.Contains(m.View(), "yí lù") {
		t.Fatalf("stone screen:\n%s", m.View())
	}
	m, _ = m.Update(press("一路"))
	m, _ = m.Update(enter)
	if m.answered > 0 || !strings.Contains(m.note, "stays shut") {
		t.Fatalf("half the idiom opened the stone: %q", m.note)
	}
	m, _ = m.Update(press("平安"))
	m, _ = m.Update(enter)
	w, err := m.store.Wallet()
	if err != nil || w.Yuanbao != 1 || !strings.Contains(m.View(), "slides open") {
		t.Fatalf("opened unaided: purse %+v %v\n%s", w, err, m.View())
	}
	m, _ = m.Update(enter)
	m, _ = m.Update(press("s"))
	if m.screen != screenTown || !strings.Contains(m.note, "already") {
		t.Errorf("an open stone reopened: screen %d, note %q", m.screen, m.note)
	}
}

func TestStoneHintPaysCopper(t *testing.T) {
	m := worldAt(t, 11)
	m, _ = m.Update(press("s"))
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyCtrlH})
	if !strings.Contains(m.View(), "yí lù píng ān") {
		t.Fatalf("hint not shown:\n%s", m.View())
	}
	m, _ = m.Update(press("一路平安"))
	m, _ = m.Update(enter)
	if w, err := m.store.Wallet(); err != nil || w.Yuanbao != 0 || w.Tongqian != stoneHintPay {
		t.Errorf("opened with the hint: purse %+v %v", w, err)
	}
}

func TestHuntAcceptBeatAndCollect(t *testing.T) {
	m := worldAt(t, 11)
	for i, q := range m.towns[0].Quests {
		if q.Hunting() {
			m.quest = i
		}
	}
	m, _ = m.Update(enter)
	if !strings.Contains(m.View(), "特处士") || !strings.Contains(m.View(), "enter accept") {
		t.Fatalf("hunt screen:\n%s", m.View())
	}
	m, _ = m.Update(enter)
	if m.screen != screenTown || !strings.Contains(m.View(), "hunting 0/2") {
		t.Fatalf("after accepting:\n%s", m.View())
	}
	for i := 0; i < 3; i++ {
		if err := m.store.AddKill("changan-hunt", 2); err != nil {
			t.Fatal(err)
		}
	}
	m = m.Refresh()
	if !strings.Contains(m.View(), "★ collect") {
		t.Fatalf("a finished hunt isn't ready to collect:\n%s", m.View())
	}
	m, _ = m.Update(enter)
	m, _ = m.Update(enter)
	w, err := m.store.Wallet()
	if err != nil || w.Tongqian != 40 || m.states["changan-hunt"] != wiki.QuestDone {
		t.Errorf("collecting: purse %+v %v, state %v", w, err, m.states["changan-hunt"])
	}
}

func TestBuyingFoodAndRestingThePartyInTown(t *testing.T) {
	m := worldAt(t, 11)
	m, _ = m.Update(press("b"))
	if !strings.Contains(m.note, "not enough") {
		t.Fatalf("bought with an empty purse: %q", m.note)
	}
	if err := m.store.Earn(wiki.Wallet{Tongqian: 40}); err != nil {
		t.Fatal(err)
	}
	m, _ = m.Update(press("b"))
	tr, _, err := journey.LoadTrail(m.store, nil)
	if err != nil || tr.Food != journey.StartFood+10 {
		t.Fatalf("after buying: food %d %v", tr.Food, err)
	}
	tr.Health[journey.Master] = 20
	if err := m.store.SaveTrail(tr); err != nil {
		t.Fatal(err)
	}
	if err := m.store.SetWounds(10); err != nil {
		t.Fatal(err)
	}
	m, _ = m.Update(press("r"))
	tr, _, err = journey.LoadTrail(m.store, nil)
	if err != nil || tr.Health[journey.Master] != 100 {
		t.Errorf("the inn didn't restore the master: %d %v (%q)", tr.Health[journey.Master], err, m.note)
	}
}
