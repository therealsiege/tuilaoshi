package dashboard

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/therealsiege/tuilaoshi/internal/model"
	"github.com/therealsiege/tuilaoshi/internal/ui/uitest"
)

func TestComposeLetter(t *testing.T) {
	zh, en := composeLetter(40, 90, []string{"已", "未"}, 12, 3)
	for _, want := range []string{"上个星期你复习了40次", "百分之90", "已、未", "答错了3个词", "有12个词", "南海 观音"} {
		if !strings.Contains(zh, want) {
			t.Errorf("letter lacks %q:\n%s", want, zh)
		}
	}
	if strings.Count(zh, "\n") != strings.Count(en, "\n") {
		t.Errorf("Chinese and English have different lines:\n%s\n%s", zh, en)
	}
	if zh, _ := composeLetter(10, 50, nil, 0, 0); strings.Contains(zh, "常常忘记") || strings.Contains(zh, "心魔") || !strings.Contains(zh, "慢慢来") {
		t.Errorf("a weak week with nothing hard or missed:\n%s", zh)
	}
}

func TestAWeeklyLetterComesAfterAWeekOfReviews(t *testing.T) {
	store := uitest.Wiki(t, map[string]string{"characters/shu1.md": uitest.OverdueCharacter})
	monday := time.Date(2026, 11, 30, 9, 0, 0, 0, time.Local)
	cards, err := store.DrillableCards()
	if err != nil {
		t.Fatal(err)
	}
	if err := store.RecordReview(&cards[0], model.DiffGood, monday.AddDate(0, 0, -3)); err != nil {
		t.Fatal(err)
	}
	m := New(store, nil)
	m.now = func() time.Time { return monday }
	m = m.SetSize(110, 40).Refresh()
	if m.letter == nil || m.letter.Read || !strings.Contains(m.View(), "A letter from Guanyin has come (n)") {
		t.Fatalf("no unread letter at camp:\n%s", m.View())
	}
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("n")})
	if out := m.View(); !strings.Contains(out, "圣僧") || !strings.Contains(out, "上个星期你复习了1次") || m.Hear() == "" {
		t.Errorf("letter view:\n%s", out)
	}
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if m.reading {
		t.Error("esc should close the letter")
	}
	letters, err := store.Letters()
	if err != nil || len(letters) != 1 || !letters[0].Read || letters[0].Week != "2026-11-30" {
		t.Errorf("letters %+v %v", letters, err)
	}
	m.Refresh()
	if letters, err = store.Letters(); err != nil || len(letters) != 1 {
		t.Errorf("a second letter the same week: %d %v", len(letters), err)
	}
}
