package party

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/therealsiege/tuilaoshi/internal/ui/uitest"
)

func partyWith(t *testing.T, journeyYAML string) Model {
	t.Helper()
	store := uitest.Wiki(t, map[string]string{"journey.yaml": journeyYAML})
	return New(store).SetSize(100, 30).Refresh()
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
