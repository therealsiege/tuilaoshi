package shrine

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/therealsiege/tuilaoshi/internal/ui/uitest"
	"github.com/therealsiege/tuilaoshi/internal/wiki"
)

func shrineWith(t *testing.T, copper int) Model {
	t.Helper()
	store := uitest.Wiki(t, map[string]string{"journey.yaml": "cleared: {}\n"})
	if err := store.Earn(wiki.Wallet{Tongqian: copper}); err != nil {
		t.Fatal(err)
	}
	return New(store).SetSize(100, 40).Refresh()
}

// pick moves the cursor to an item by id.
func pick(t *testing.T, m Model, id string) Model {
	t.Helper()
	for i, g := range m.items {
		if g.ID == id {
			m.cursor = i
			return m
		}
	}
	t.Fatalf("no item %s", id)
	return m
}

var enter = tea.KeyMsg{Type: tea.KeyEnter}

func TestShrineScreen(t *testing.T) {
	uitest.Golden(t, "shrine", shrineWith(t, 100).View())
}

func TestBuyWearAndTakeOff(t *testing.T) {
	m := pick(t, shrineWith(t, 100), "peach-staff")
	m, _ = m.Update(enter)
	if m.noteErr || m.kit.Worn["weapon"] != "peach-staff" || m.wallet.Tongqian != 20 {
		t.Fatalf("after buying the staff: %q, worn %v, purse %+v", m.note, m.kit.Worn, m.wallet)
	}
	if !strings.Contains(m.View(), "+3 damage per hit") {
		t.Errorf("character sheet doesn't show the staff:\n%s", m.View())
	}
	m, _ = m.Update(enter)
	if m.kit.Worn["weapon"] != "" || m.wallet.Tongqian != 20 {
		t.Fatalf("enter on worn gear should take it off for free: worn %v, purse %+v", m.kit.Worn, m.wallet)
	}
	m = pick(t, m, "talisman")
	m, _ = m.Update(enter)
	if !m.noteErr || !strings.Contains(m.note, "100 more 铜钱") || len(m.kit.Owned) != 1 {
		t.Errorf("buying beyond the purse: %q, owned %v", m.note, m.kit.Owned)
	}
	m = pick(t, m, "demon-mirror")
	m, _ = m.Update(enter)
	if !m.noteErr || !strings.Contains(m.note, "after tribulation 25") {
		t.Errorf("an item not yet on the shelf: %q", m.note)
	}
}
