package drill

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/therealsiege/tuilaoshi/internal/claude"
	"github.com/therealsiege/tuilaoshi/internal/ui/input"
	"github.com/therealsiege/tuilaoshi/internal/ui/uitest"
)

func TestDrillScreens(t *testing.T) {
	store := uitest.Wiki(t, map[string]string{"characters/shu1.md": uitest.OverdueCharacter})
	m := New(store, claude.New(), nil, input.NewDict()).SetSize(80, 30)
	uitest.Golden(t, "mode_picker", m.View())

	m, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m, _ = m.Update(cmd())
	if m.state != statePrompt {
		t.Fatalf("state = %v after loading cards, want prompt", m.state)
	}
	uitest.Golden(t, "prompt", m.View())
}
