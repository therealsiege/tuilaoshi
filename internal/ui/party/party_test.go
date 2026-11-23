package party

import (
	"context"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/therealsiege/tuilaoshi/internal/claude"
	"github.com/therealsiege/tuilaoshi/internal/journey"
	"github.com/therealsiege/tuilaoshi/internal/ui/uitest"
)

func partyWith(t *testing.T, journeyYAML string) Model {
	t.Helper()
	store := uitest.Wiki(t, map[string]string{"journey.yaml": journeyYAML})
	return New(store, nil, nil).SetSize(100, 30).Refresh()
}

func press(s string) tea.KeyMsg { return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)} }

func TestPartyScreen(t *testing.T) {
	m := partyWith(t, "cleared: {8: \"2026-11-01\"}\nbonds: {孙悟空: 16}\n")
	uitest.Golden(t, "party", m.View())
}

func TestBondScenesOpenByRank(t *testing.T) {
	m := partyWith(t, "cleared: {8: \"2026-11-01\"}\nbonds: {孙悟空: 6}\n") // 同行: the first scene
	if m.members[0].Name != "孙悟空" {
		t.Fatalf("first member %s", m.members[0].Name)
	}
	m, _ = m.Update(press("2"))
	if m.reading || !strings.Contains(m.note, "信任") {
		t.Fatalf("scene 2 at 6 points: reading %v, note %q", m.reading, m.note)
	}
	m, _ = m.Update(press("1"))
	if !m.reading || !strings.Contains(m.View(), "火眼金睛") {
		t.Fatalf("scene 1 didn't open:\n%s", m.View())
	}
	m, _ = m.Update(press("e"))
	if !strings.Contains(m.View(), "Fiery") && !strings.Contains(m.View(), "white cloud") {
		t.Errorf("e should show the translation:\n%s", m.View())
	}
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if m.reading {
		t.Error("esc should close the scene")
	}
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m, _ = m.Update(press("1"))
	if m.reading || !strings.Contains(m.note, "hasn't joined") {
		t.Errorf("a companion not yet joined: reading %v, note %q", m.reading, m.note)
	}
}

func TestTalkingRaisesTheBond(t *testing.T) {
	m := partyWith(t, "cleared: {8: \"2026-11-01\"}\n")
	var systems []string
	m.talker = func(_ context.Context, system string, history []claude.ChatMessage, msg string) (string, error) {
		systems = append(systems, system)
		if msg == journey.TalkOpener {
			return "小兄弟，今天走了多少路？\nBOND: 0", nil
		}
		return "不错！你还想走多远？\nBOND: 2", nil
	}
	m, cmd := m.Update(press("t"))
	if m.talk == nil || cmd == nil {
		t.Fatal("t didn't open a talk")
	}
	m = runTalk(t, m, cmd)
	if !strings.Contains(m.View(), "今天走了多少路") || strings.Contains(m.View(), "BOND") || strings.Contains(m.View(), journey.TalkOpener) {
		t.Fatalf("opener shown wrong:\n%s", m.View())
	}
	m, cmd = m.Update(tea.KeyMsg{Type: tea.KeyEnter}) // empty input sends nothing
	if cmd != nil {
		t.Fatal("sent an empty message")
	}
	m.talk.history = append(m.talk.history, claude.ChatMessage{Role: "user", Content: "我们走了很远"})
	m = m.onReply(talkReplyMsg{text: "不错！\nBOND: 2"})
	if bonds, err := m.store.BondPoints(); err != nil || bonds["孙悟空"] != 2 || !strings.Contains(m.note, "Bond +2") {
		t.Errorf("bond after a good reply: %v %v, note %q", bonds, err, m.note)
	}
	if !strings.Contains(systems[0], "孙悟空") {
		t.Errorf("the talk isn't in Wukong's voice")
	}
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if m.talk != nil {
		t.Error("esc should end the talk")
	}
}

// runTalk runs a talk command and delivers its reply.
func runTalk(t *testing.T, m Model, cmd tea.Cmd) Model {
	t.Helper()
	for _, msg := range flatten(cmd) {
		if r, ok := msg.(talkReplyMsg); ok {
			return m.onReply(r)
		}
	}
	t.Fatal("no reply")
	return m
}

// flatten runs a command and any batch inside it, skipping blink ticks.
func flatten(cmd tea.Cmd) []tea.Msg {
	if cmd == nil {
		return nil
	}
	msg := cmd()
	if batch, ok := msg.(tea.BatchMsg); ok {
		var out []tea.Msg
		for _, c := range batch {
			if c == nil {
				continue
			}
			done := make(chan tea.Msg, 1)
			go func() { done <- c() }()
			select {
			case r := <-done:
				out = append(out, r)
			case <-time.After(100 * time.Millisecond):
			}
		}
		return out
	}
	return []tea.Msg{msg}
}

func TestIdiomDuel(t *testing.T) {
	m := partyWith(t, "cleared: {}\n") // nobody has joined; 孙悟空 still duels
	turns := []claude.IdiomTurn{
		{Valid: true, Reply: claude.Idiom{Text: "一心一意", Pinyin: "yī xīn yī yì", Meaning: "wholeheartedly"}},
		{Valid: false, Note: "Not a real idiom."},
		{Valid: true, Played: claude.Idiom{Text: "意气风发", Pinyin: "yì qì fēng fā", Meaning: "high-spirited"}, Reply: claude.Idiom{}},
	}
	m.mover = func(_ context.Context, who string, chain []string, attempt string, hint bool) (claude.IdiomTurn, error) {
		if who != "孙悟空" {
			t.Errorf("duel with %s", who)
		}
		turn := turns[0]
		turns = turns[1:]
		return turn, nil
	}
	m, cmd := m.Update(press("i"))
	if m.duel == nil {
		t.Fatalf("no duel: %q", m.note)
	}
	m = m.onDuel(cmd().(tea.BatchMsg)[0]().(duelMsg))
	if m.duel.last.Text != "一心一意" || !strings.Contains(m.View(), "一心一意") {
		t.Fatalf("opening not shown:\n%s", m.View())
	}
	m.input.SetValue("意思意思")
	m, cmd = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = m.onDuel(cmd().(duelMsg))
	if m.duel.lives != duelLives-1 || !strings.Contains(m.View(), "Not a real idiom") {
		t.Fatalf("a miss: lives %d\n%s", m.duel.lives, m.View())
	}
	m.input.SetValue("意气风发")
	m, cmd = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = m.onDuel(cmd().(duelMsg))
	if !m.duel.won || !m.duel.over {
		t.Fatalf("stumping the companion should win: %+v", m.duel)
	}
	w, err := m.store.Wallet()
	bonds, berr := m.store.BondPoints()
	cards, cerr := m.store.DrillableCards()
	if err != nil || berr != nil || cerr != nil {
		t.Fatal(err, berr, cerr)
	}
	if w.Tongqian != duelWinPay+1 || bonds["孙悟空"] != 1 {
		t.Errorf("after the win: purse %+v (want %d), bond %d (want 1)", w, duelWinPay+1, bonds["孙悟空"])
	}
	idioms := 0
	for _, c := range cards {
		if c.Chinese == "一心一意" || c.Chinese == "意气风发" {
			idioms++
		}
	}
	if idioms != 2 {
		t.Errorf("%d of the idioms met joined the reviews, want 2", idioms)
	}
}
