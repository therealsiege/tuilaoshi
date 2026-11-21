package reader

import (
	"fmt"
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
	if m.loadErr != nil || len(m.items) != 1 {
		t.Fatalf("items %d, err %v", len(m.items), m.loadErr)
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

func TestPrivateEpisodeReadsAsDialogue(t *testing.T) {
	srt := "1\n00:00:01,000 --> 00:00:02,000\n孙悟空：师父，有妖怪！\nMaster, a demon!\n\n2\n00:00:03,000 --> 00:00:04,000\n菜单在这里。\n"
	store := uitest.Wiki(t, map[string]string{"private/episodes/ep01.srt": srt, "vocabulary/food.md": food})
	m := New(store).SetSize(80, 30).Refresh()
	if len(m.items) != 1 || m.items[0].tag != "your episode" {
		t.Fatalf("items %+v; want the private episode", m.items)
	}
	m = press(m, tea.KeyEnter)
	if !m.Open() || m.current() != "孙悟空" {
		t.Fatalf("opened %v at %q; want the speaker, glossed from the story glossary", m.Open(), m.current())
	}
	if v := m.View(); !strings.Contains(v, "Sūn Wùkōng") || !strings.Contains(v, "e: show this line in English") {
		t.Errorf("episode view:\n%s", v)
	}
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("e")})
	if !strings.Contains(m.View(), "Master, a demon!") {
		t.Errorf("e didn't show the line's English:\n%s", m.View())
	}
	m = press(m, tea.KeyDown) // the next line has no translation
	if strings.Contains(m.View(), "Master, a demon!") || m.current() != "菜单" {
		t.Errorf("line 2: at %q\n%s", m.current(), m.View())
	}
}

func TestJourneyEpisodesUnlockWithTheStage(t *testing.T) {
	titles := func(m Model) string {
		var out []string
		for _, it := range m.items {
			out = append(out, it.title)
		}
		return strings.Join(out, " | ")
	}
	fresh := New(uitest.Wiki(t, map[string]string{"vocabulary/food.md": food})).Refresh()
	if strings.Contains(titles(fresh), "鹰愁涧") {
		t.Errorf("stage 9's episode listed at the start: %s", titles(fresh))
	}
	cleared := "cleared:\n"
	for i := 1; i <= 8; i++ {
		cleared += fmt.Sprintf("    %d: \"2026-11-21\"\n", i)
	}
	at9 := New(uitest.Wiki(t, map[string]string{"vocabulary/food.md": food, "journey.yaml": cleared})).Refresh()
	if !strings.Contains(titles(at9), "鹰愁涧") || strings.Contains(titles(at9), "黑风山") {
		t.Errorf("at stage 9: %s; want stage 9's episode and not stage 11's", titles(at9))
	}
}
