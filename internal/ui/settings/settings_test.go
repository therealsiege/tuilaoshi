package settings

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/therealsiege/tuilaoshi/internal/ui"
	"github.com/therealsiege/tuilaoshi/internal/ui/uitest"
	"github.com/therealsiege/tuilaoshi/internal/wiki"
)

func TestChangingASettingSavesAndAppliesIt(t *testing.T) {
	defer func(p wiki.Settings) { ui.Prefs = p }(ui.Prefs)
	m := New(uitest.Wiki(t, nil)).Refresh()
	uitest.Golden(t, "settings", m.View())
	m.cursor = row(t, "Wild encounters")
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRight})
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRight})
	if ui.Prefs.Encounters != "off" || !strings.Contains(m.View(), "Saved") {
		t.Fatalf("encounters %q after two steps right", ui.Prefs.Encounters)
	}
	saved, err := m.store.Settings()
	if err != nil || saved.Encounters != "off" {
		t.Errorf("saved %+v %v", saved, err)
	}
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRight}) // wraps to normal
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyLeft})  // and back to off
	if ui.Prefs.Encounters != "off" {
		t.Errorf("left and right don't cycle: %q", ui.Prefs.Encounters)
	}
	m.cursor = row(t, "Review chunk")
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRight})
	if ui.Prefs.ReviewChunk != 30 {
		t.Errorf("review chunk %d, want 30", ui.Prefs.ReviewChunk)
	}
}

// row finds a setting's row by its label.
func row(t *testing.T, label string) int {
	t.Helper()
	for i, o := range options {
		if o.label == label {
			return i
		}
	}
	t.Fatalf("no setting %q", label)
	return 0
}
