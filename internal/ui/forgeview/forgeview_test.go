package forgeview

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/therealsiege/tuilaoshi/internal/ui/uitest"
)

var parts = map[string][]string{"河": {"氵", "可"}, "林": {"木", "木"}, "好": {"女", "子"}}

func forgeWith(t *testing.T) Model {
	t.Helper()
	pages := map[string]string{}
	for ch, rad := range map[string]string{"河": "氵(water)", "林": "木(wood)", "好": "女(woman)"} {
		pages["characters/"+ch+".md"] = fmt.Sprintf("---\ntype: character\nreview_interval: 1\n---\n\n# %s\n\n- **Character**: %s\n- **Pinyin**: p\n- **Meaning**: m %s\n- **Radical**: %s\n", ch, ch, ch, rad)
	}
	m := New(uitest.Wiki(t, pages), nil)
	asked := 0
	m.decompose = func(_ context.Context, chars []string) (map[string][]string, error) {
		asked++
		out := map[string][]string{}
		for _, c := range chars {
			out[c] = parts[c]
		}
		return out, nil
	}
	m.now = func() time.Time { return time.Date(2026, 11, 23, 9, 0, 0, 0, time.Local) }
	m, cmd := m.SetSize(100, 30).Refresh()
	if m.state != stateLoading || cmd == nil {
		t.Fatalf("new characters should be looked up first: state %d", m.state)
	}
	m, _ = m.Update(cmd())
	if m.state != stateForging || len(m.queue) != 3 {
		t.Fatalf("after lookup: state %d, %d to forge", m.state, len(m.queue))
	}
	if again, cmd := New(m.store, nil).Refresh(); cmd != nil || again.state != stateForging {
		t.Error("cached parts should need no lookup")
	}
	return m
}

// pickParts presses the tray numbers for the current character's parts.
func pickParts(m Model, want []string) Model {
	used := map[int]bool{}
	for _, p := range want {
		for i, t := range m.tray {
			if t == p && (!used[i] || countIn(want, p) > 1) {
				used[i] = true
				m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(fmt.Sprint(i + 1))})
				break
			}
		}
	}
	return m
}

func countIn(s []string, x string) int {
	n := 0
	for _, v := range s {
		if v == x {
			n++
		}
	}
	return n
}

var enter = tea.KeyMsg{Type: tea.KeyEnter}

func TestForgeRightAndWrongThenSave(t *testing.T) {
	m := forgeWith(t)
	first := m.queue[0]
	m = pickParts(m, parts[first.Chinese])
	m, _ = m.Update(enter)
	if m.state != stateResult || !m.right || !strings.Contains(m.View(), "Forged!") || len(m.loot) != 1 {
		t.Fatalf("right parts: state %d right %v loot %v\n%s", m.state, m.right, m.loot, m.View())
	}
	m, _ = m.Update(enter)
	second := m.queue[1]
	wrong := parts[second.Chinese][:1]
	m = pickParts(m, wrong)
	m, _ = m.Update(enter)
	if m.right || !strings.Contains(m.View(), "cracks") {
		t.Fatalf("one part of two forged:\n%s", m.View())
	}
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	w, err := m.store.Wallet()
	mats, merr := m.store.Materials()
	if err != nil || merr != nil || w.Tongqian != 1 || len(mats) != 1 {
		t.Errorf("after the forge: purse %+v, materials %v (%v %v)", w, mats, err, merr)
	}
}
