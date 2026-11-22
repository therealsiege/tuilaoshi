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
