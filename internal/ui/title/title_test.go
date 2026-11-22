package title

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/therealsiege/tuilaoshi/internal/ui/uitest"
	"github.com/therealsiege/tuilaoshi/internal/wiki"
)

var day = time.Date(2026, 11, 22, 9, 0, 0, 0, time.Local)

func titleFor(t *testing.T, store *wiki.Store) Model {
	t.Helper()
	m := New(store, nil)
	m.now = func() time.Time { return day }
	return m.SetSize(100, 40).Refresh()
}

var (
	enter = tea.KeyMsg{Type: tea.KeyEnter}
	down  = tea.KeyMsg{Type: tea.KeyDown}
)

func press(s string) tea.KeyMsg { return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)} }

func isStarted(cmd tea.Cmd) bool {
	if cmd == nil {
		return false
	}
	_, ok := cmd().(StartedMsg)
	return ok
}

func TestFirstLaunchOffersANewJourney(t *testing.T) {
	m := titleFor(t, uitest.Wiki(t, nil))
	uitest.Golden(t, "first", m.View())
	if m.menu[0] != choiceNew {
		t.Fatalf("menu %v, want New first with no saves", m.menu)
	}
	m, _ = m.Update(enter)
	m, cmd := m.Update(enter) // the default name
	if !isStarted(cmd) || m.store.Save == "" {
		t.Fatal("a new journey didn't start")
	}
	saves, err := m.store.Saves()
	if err != nil || len(saves) != 1 || saves[0].Name != "第1次西行" {
		t.Fatalf("saves %+v %v", saves, err)
	}
}

func TestContinueAndLoad(t *testing.T) {
	store := uitest.Wiki(t, nil)
	older, err := store.NewSave("older", day.Add(-48*time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.NewSave("newer", day.Add(-time.Hour)); err != nil {
		t.Fatal(err)
	}
	m := titleFor(t, store)
	uitest.Golden(t, "saves", m.View())
	if !strings.Contains(m.View(), "newer") {
		t.Fatalf("continue doesn't name the latest journey:\n%s", m.View())
	}
	m, _ = m.Update(down)
	m, _ = m.Update(down) // Load
	m, _ = m.Update(enter)
	m, _ = m.Update(down) // older
	m, _ = m.Update(press("d"))
	if !strings.Contains(m.View(), "Press d again") {
		t.Fatalf("delete didn't ask to confirm:\n%s", m.View())
	}
	m, _ = m.Update(press("x")) // anything else cancels
	_, cmd := m.Update(enter)
	if !isStarted(cmd) || store.Save != older {
		t.Fatalf("loading the older journey: started %v, save %q", isStarted(cmd), store.Save)
	}
}
