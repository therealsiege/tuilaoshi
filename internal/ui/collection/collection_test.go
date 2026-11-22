package collection

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/therealsiege/tuilaoshi/internal/ui/uitest"
)

func collectionWiki(t *testing.T) Model {
	t.Helper()
	page := func(ch, rad, interval string) string {
		return "---\ntype: character\nreview_interval: " + interval +
			"\n---\n\n# " + ch + "\n\n- **Character**: " + ch + "\n- **Radical**: " + rad + "\n"
	}
	store := uitest.Wiki(t, map[string]string{
		"characters/he2.md":  page("河", "氵 water", "7"),
		"characters/hai3.md": page("海", "氵 water", "1"),
		"characters/huo3.md": page("火", "火 fire", "7"),
		"journey.yaml":       "cleared: {}\ndefeated: [熊山君]\n",
	})
	return New(store).SetSize(100, 30).Refresh()
}

var enter = tea.KeyMsg{Type: tea.KeyEnter}

func TestCollectionScreens(t *testing.T) {
	m := collectionWiki(t)
	uitest.Golden(t, "enemies", m.View())
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	uitest.Golden(t, "characters", m.View())
}

func TestClaimACompleteSetOnce(t *testing.T) {
	m := collectionWiki(t)
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	if s := m.sets[tabCharacters][m.cursor[tabCharacters]]; s.ID != "radical-氵" {
		t.Fatalf("first set is %s, want the larger water family", s.ID)
	}
	m, _ = m.Update(enter)
	if !m.noteErr || !strings.Contains(m.note, "1 more") {
		t.Errorf("claiming an incomplete set: %q", m.note)
	}
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown}) // fire: 火 is mastered
	m, _ = m.Update(enter)
	w, err := m.store.Wallet()
	if m.noteErr || err != nil || w.Yuanbao != 1 || w.Tongqian != 10 {
		t.Fatalf("claiming fire: %q, purse %+v %v", m.note, w, err)
	}
	m, _ = m.Update(enter)
	if w, err := m.store.Wallet(); err != nil || w.Yuanbao != 1 || !strings.Contains(m.note, "Already") {
		t.Errorf("claimed twice: %q, purse %+v", m.note, w)
	}
}
