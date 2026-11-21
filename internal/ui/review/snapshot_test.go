package review

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/therealsiege/tuilaoshi/internal/claude"
	"github.com/therealsiege/tuilaoshi/internal/ui/input"
	"github.com/therealsiege/tuilaoshi/internal/ui/uitest"
)

func TestReviewScreens(t *testing.T) {
	store := uitest.Wiki(t, map[string]string{"characters/shu1.md": uitest.OverdueCharacter})
	m := New(store, claude.New(), nil, input.NewDict()).SetSize(80, 30)
	m, _ = m.Update(loadDueMsg{})
	if len(m.items) != 1 {
		t.Fatalf("got %d due items, want 1", len(m.items))
	}
	uitest.Golden(t, "list", m.View())

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if m.state != stateDrill {
		t.Fatalf("state = %v after enter, want drill", m.state)
	}
	uitest.Golden(t, "prompt", m.View())
}

func TestReviewCtrlRStaysOnCard(t *testing.T) {
	store := uitest.Wiki(t, map[string]string{"characters/shu1.md": uitest.OverdueCharacter})
	m := New(store, claude.New(), nil, input.NewDict())
	m, _ = m.Update(loadDueMsg{})
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyCtrlR})
	if m.state != stateDrill || m.input.Value() != "" {
		t.Errorf("ctrl+r changed the drill: state %v, input %q", m.state, m.input.Value())
	}
}
