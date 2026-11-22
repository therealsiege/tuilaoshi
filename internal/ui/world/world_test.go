package world

import (
	"fmt"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

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
	m := New(store)
	m.now = func() time.Time { return day }
	return m.SetSize(100, 50).Refresh()
}

func itoa(n int) string { return fmt.Sprint(n) }

func press(s string) tea.KeyMsg { return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)} }

var enter = tea.KeyMsg{Type: tea.KeyEnter}

func TestRoadScreens(t *testing.T) {
	m := worldAt(t, 11) // working on stage 12: Chang'an, the temple and Gao Village reached
	uitest.Golden(t, "road", m.View())
	m, _ = m.Update(enter)
	uitest.Golden(t, "town", m.View())
}

func TestAnsweringARequest(t *testing.T) {
	m := worldAt(t, 11)
	m.town = 0 // Chang'an
	m, _ = m.Update(enter)
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

func TestTownsAheadAreClosed(t *testing.T) {
	m := worldAt(t, 0)
	m.town = len(m.towns) - 1
	m, _ = m.Update(enter)
	if m.screen != screenMap || !strings.Contains(m.note, "tribulation 79") {
		t.Errorf("visited a town not yet reached: screen %d, note %q", m.screen, m.note)
	}
}
