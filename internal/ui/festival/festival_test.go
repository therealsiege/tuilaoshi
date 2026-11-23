package festival

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/therealsiege/tuilaoshi/internal/ui/uitest"
)

func festivalOn(t *testing.T, day string) Model {
	t.Helper()
	m := New(uitest.Wiki(t, nil))
	at, err := time.ParseInLocation("2006-01-02", day, time.Local)
	if err != nil {
		t.Fatal(err)
	}
	m.now = func() time.Time { return at.Add(10 * time.Hour) }
	return m.SetSize(100, 40).Refresh()
}

func press(s string) tea.KeyMsg { return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)} }

func TestBeforeAFestivalCountsDown(t *testing.T) {
	m := festivalOn(t, "2026-11-23")
	uitest.Golden(t, "upcoming", m.View())
	if m.on || !strings.Contains(m.View(), "in 29 days") {
		t.Errorf("late November:\n%s", m.View())
	}
	if m, _ = m.Update(press("1")); m.asking >= 0 {
		t.Error("a question opened before the festival")
	}
}

func TestDuringAFestival(t *testing.T) {
	m := festivalOn(t, "2026-12-22")
	uitest.Golden(t, "solstice", m.View())
	if !m.on || m.added != len(m.occ.Words) {
		t.Fatalf("on the solstice: on %v, %d words added", m.on, m.added)
	}
	cards, err := m.store.DrillableCards()
	if err != nil || len(cards) < len(m.occ.Words) {
		t.Fatalf("festival words aren't review cards: %d %v", len(cards), err)
	}
	m, _ = m.Update(press("1"))
	m, _ = m.Update(press(string(rune('1' + (m.right+1)%len(m.choices)))))
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if m, _ = m.Update(press("1")); m.asking >= 0 {
		t.Fatal("a missed question reopened the same day")
	}
	m.now = func() time.Time { return time.Date(2026, 12, 23, 10, 0, 0, 0, time.Local) }
	m = m.Refresh()
	for i := 0; i < 2; i++ {
		m, _ = m.Update(press(string(rune('1' + i))))
		m, _ = m.Update(press(string(rune('1' + m.right))))
		m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	}
	w, err := m.store.Wallet()
	if err != nil || w.Tongqian != 2*QuestionPay+FestivalPay || w.Yuanbao != FestivalIngots {
		t.Errorf("both answered: purse %+v %v", w, err)
	}
}
