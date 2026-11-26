package review

import (
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

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

func TestALongQueueFitsTheWindow(t *testing.T) {
	pages := map[string]string{}
	for i := range 40 {
		pages[fmt.Sprintf("characters/c%02d.md", i)] = strings.Replace(uitest.OverdueCharacter, "book", fmt.Sprintf("book number %d, with a meaning long enough to run past the edge of the window", i), 1)
	}
	m := New(uitest.Wiki(t, pages), claude.New(), nil, input.NewDict())
	for _, size := range [][2]int{{80, 24}, {120, 45}} {
		m = m.SetSize(size[0], size[1])
		m, _ = m.Update(loadDueMsg{})
		for range 20 {
			m, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
		}
		out := m.View()
		if n := strings.Count(out, "\n") + 1; n > size[1] || !strings.Contains(out, "复习") || !strings.Contains(out, "above") {
			t.Errorf("%dx%d: %d lines\n%s", size[0], size[1], n, out)
		}
		for _, l := range strings.Split(out, "\n") {
			if lipgloss.Width(l) > size[0] {
				t.Errorf("%dx%d: a line %d wide: %s", size[0], size[1], lipgloss.Width(l), l)
			}
		}
	}
}
