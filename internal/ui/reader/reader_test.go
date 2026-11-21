package reader

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/therealsiege/tuilaoshi/internal/ui/uitest"
)

const restaurant = `---
type: reading
level: HSK3
---

# 在餐厅 At the Restaurant

## Text

服务员：这是菜单。

小张：谢谢。

## Vocabulary

| Word | Pinyin | Meaning |
|------|--------|---------|
| 服务员 | fúwùyuán | waiter, server |
`

const food = `---
type: vocabulary
level: HSK2
review_interval: 10
next_review: 2026-12-30
---

# Food

| Word | Pinyin | English |
|------|--------|---------|
| 菜单 | càidān | menu |
`

func openRestaurant(t *testing.T) Model {
	t.Helper()
	store := uitest.Wiki(t, map[string]string{"readings/restaurant.md": restaurant, "vocabulary/food.md": food})
	m := New(store).SetSize(80, 30).Refresh()
	if m.loadErr != nil || len(m.readings) != 1 {
		t.Fatalf("readings %d, err %v", len(m.readings), m.loadErr)
	}
	uitest.Golden(t, "list", m.View())
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if !m.Open() {
		t.Fatal("enter did not open the reading")
	}
	return m
}

func press(m Model, k tea.KeyType) Model {
	m, _ = m.Update(tea.KeyMsg{Type: k})
	return m
}

func TestReadingGlossesWords(t *testing.T) {
	m := openRestaurant(t)
	if m.current() != "服务员" {
		t.Fatalf("first word %q, want 服务员 (from the reading's vocabulary)", m.current())
	}
	uitest.Golden(t, "text", m.View())

	for m.current() != "菜单" && m.cursor < len(m.words)-1 {
		m = press(m, tea.KeyRight)
	}
	if !strings.Contains(m.View(), "càidān") {
		t.Errorf("菜单 gloss missing:\n%s", m.View())
	}
	if m = press(m, tea.KeyDown); m.current() != "小" {
		t.Errorf("down moved to %q, want the first word of the next line", m.current())
	}
	if m = press(m, tea.KeyEsc); m.Open() {
		t.Error("esc did not return to the list")
	}
}

func TestMarkingWordSchedulesItForTomorrow(t *testing.T) {
	m := openRestaurant(t)
	for m.current() != "菜单" {
		m = press(m, tea.KeyRight)
	}
	m = press(m, tea.KeyEnter)
	if m.saveErr != nil || !m.marked["菜单"] {
		t.Fatalf("mark failed: err %v, marked %v", m.saveErr, m.marked)
	}
	cards, err := m.store.DrillableCards()
	if err != nil {
		t.Fatal(err)
	}
	tomorrow := time.Now().AddDate(0, 0, 1).Format("2006-01-02")
	for _, c := range cards {
		if c.Chinese == "菜单" && c.NextReview.Format("2006-01-02") != tomorrow {
			t.Errorf("菜单 next review %s, want %s", c.NextReview.Format("2006-01-02"), tomorrow)
		}
	}

	m = press(m, tea.KeyLeft) // 是, a lone character not in the decks
	m = press(m, tea.KeyEnter)
	if !strings.Contains(m.status, "not in your decks") {
		t.Errorf("status %q, want a note that the word isn't in the decks", m.status)
	}
}
