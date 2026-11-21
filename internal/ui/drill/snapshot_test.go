package drill

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/therealsiege/tuilaoshi/internal/claude"
	"github.com/therealsiege/tuilaoshi/internal/model"
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

func TestExactChineseIsGradedLocally(t *testing.T) {
	c := model.Card{Chinese: "书", English: "book"}
	tests := []struct {
		answer string
		mode   model.DrillMode
		want   bool
	}{
		{"书", model.ModeProduction, true},
		{" 书 ", model.ModeListening, true},
		{"书", model.ModeSentence, true},
		{"书", model.ModeRecognition, false}, // recognition expects English
		{"書", model.ModeProduction, false},  // a variant goes to Claude
	}
	for _, tt := range tests {
		if r, ok := exactChinese(c, tt.answer, tt.mode); ok != tt.want || (ok && !r.Correct) {
			t.Errorf("exactChinese(%q, %s) = %v, %v; want local=%v", tt.answer, tt.mode, r, ok, tt.want)
		}
	}
}
