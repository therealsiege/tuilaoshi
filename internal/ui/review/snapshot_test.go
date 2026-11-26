package review

import (
	"strings"
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

func TestLeavingEarlyKeepsTheAnswersGiven(t *testing.T) {
	t.Setenv("ANTHROPIC_API_KEY", "") // graded locally
	water := strings.NewReplacer("书", "水", "shū", "shuǐ", "book", "water").Replace(uitest.OverdueCharacter)
	store := uitest.Wiki(t, map[string]string{"characters/shu1.md": uitest.OverdueCharacter, "characters/shui3.md": water})
	m := New(store, claude.New(), nil, input.NewDict()).SetSize(80, 30)
	m, _ = m.Update(loadDueMsg{})
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m.input.SetValue(m.items[m.cursor].English)
	m, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m, _ = m.Update(cmd())
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("3")})
	if m.cursor != 1 {
		t.Fatalf("still on card %d", m.cursor)
	}
	if store.TodayStats().CardsReviewed != 0 {
		t.Fatal("saved before leaving")
	}
	m = m.Leave()
	if got := store.TodayStats(); got.CardsReviewed != 1 || got.Correct != 1 {
		t.Errorf("after leaving early: %+v", got)
	}
	if m = m.Leave(); store.TodayStats().CardsReviewed != 1 {
		t.Error("leaving twice saved the answer twice")
	}
}
