package diary

import (
	"context"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/therealsiege/tuilaoshi/internal/journey"
	"github.com/therealsiege/tuilaoshi/internal/ui/uitest"
)

var day = time.Date(2026, 11, 24, 20, 0, 0, 0, time.Local)

func diaryWith(t *testing.T, score int) Model {
	t.Helper()
	m := New(uitest.Wiki(t, nil), nil, nil)
	m.now = func() time.Time { return day }
	m.grader = func(_ context.Context, task, _, answer string) (int, string, error) {
		if !strings.Contains(task, "第1天") || answer == "" {
			t.Errorf("graded %q for %q", answer, task)
		}
		return score, "很好。", nil
	}
	m, _ = m.SetSize(100, 40).Refresh()
	return m
}

// write types an entry and sends it to the grader.
func write(t *testing.T, m Model, text string) Model {
	t.Helper()
	m.input.SetValue(text)
	m, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("enter didn't send the entry")
	}
	m, _ = m.Update(cmd())
	return m
}

func TestADiaryEntryLiftsSpirits(t *testing.T) {
	m := diaryWith(t, 2)
	uitest.Golden(t, "empty", m.View())
	tr, _, err := journey.LoadTrail(m.store, nil)
	if err != nil {
		t.Fatal(err)
	}
	tr.Morale[journey.Master] = 50
	if err := m.store.SaveTrail(tr); err != nil {
		t.Fatal(err)
	}
	m = write(t, m, "今天我们走了很远的路。")
	tr, _, err = journey.LoadTrail(m.store, nil)
	if err != nil || tr.Morale[journey.Master] != 60 {
		t.Errorf("a top grade should lift morale 50 -> 60: %d %v", tr.Morale[journey.Master], err)
	}
	entries, err := m.store.Diary()
	if err != nil || len(entries) != 1 || entries[0].Day != 1 || entries[0].Score != 2 {
		t.Fatalf("entries %+v %v", entries, err)
	}
	if m, _ = m.Refresh(); !m.written || m.Typing() || !strings.Contains(m.View(), "今天我们走了很远的路") {
		t.Errorf("today's entry should be written and shown:\n%s", m.View())
	}
	m.now = func() time.Time { return day.AddDate(0, 0, 1) }
	if m, _ = m.Refresh(); m.written {
		t.Error("tomorrow's entry should be open")
	}
}

func TestAFailingEntryCanBeRevised(t *testing.T) {
	m := write(t, diaryWith(t, 0), "hello")
	if entries, err := m.store.Diary(); err != nil || len(entries) != 0 || m.written || !strings.Contains(m.View(), "revise") {
		t.Errorf("a failing entry was kept: %+v\n%s", entries, m.View())
	}
}

func TestWithoutClaudeTheDiarySaysWhy(t *testing.T) {
	m := diaryWith(t, 2)
	m.grader = nil
	m.input.SetValue("今天下雨了。")
	if m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter}); !strings.Contains(m.note, "ANTHROPIC_API_KEY") {
		t.Errorf("note %q", m.note)
	}
}
