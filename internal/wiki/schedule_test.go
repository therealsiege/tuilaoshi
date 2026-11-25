package wiki

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/therealsiege/tuilaoshi/internal/model"
)

const vocabPage = `---
type: vocabulary
review_interval: 4
next_review: 2026-11-20
---

# Food

| Word | Pinyin | English | Example |
|------|--------|---------|---------|
| 米饭 | mǐfàn | rice | 我吃米饭。 |
| 面条 | miàntiáo | noodles | 我喜欢面条。 |
`

const characterPage = `---
type: character
review_interval: 3
---

# 书

- **Character**: 书
- **Pinyin**: shū
- **Meaning**: book
`

func newTestStore(t *testing.T) *Store {
	t.Helper()
	dir := t.TempDir()
	for path, body := range map[string]string{
		"vocabulary/food.md": vocabPage,
		"characters/shu1.md": characterPage,
	} {
		full := filepath.Join(dir, path)
		if err := os.MkdirAll(filepath.Dir(full), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(body), 0644); err != nil {
			t.Fatal(err)
		}
	}
	return NewStore(dir)
}

func findCard(t *testing.T, cards []model.Card, id string) model.Card {
	t.Helper()
	for _, c := range cards {
		if c.ID == id {
			return c
		}
	}
	t.Fatalf("no card %q in %d cards", id, len(cards))
	return model.Card{}
}

func TestRecordReviewSchedulesOneVocabularyWord(t *testing.T) {
	store := newTestStore(t)
	cards, err := store.DrillableCards()
	if err != nil {
		t.Fatal(err)
	}
	rice := findCard(t, cards, "vocabulary/food.md#米饭")
	now := time.Date(2026, 11, 18, 0, 0, 0, 0, time.UTC)

	if err := store.RecordReview(&rice, model.DiffAgain, now); err != nil {
		t.Fatalf("RecordReview: %v", err)
	}

	cards, err = store.DrillableCards()
	if err != nil {
		t.Fatal(err)
	}
	if got := findCard(t, cards, "vocabulary/food.md#米饭"); got.ReviewInterval != 1 || FormatDate(got.NextReview) != "2026-11-19" {
		t.Errorf("missed word: interval %d next %s, want 1 and 2026-11-19", got.ReviewInterval, FormatDate(got.NextReview))
	}
	if got := findCard(t, cards, "vocabulary/food.md#面条"); got.ReviewInterval != 4 {
		t.Errorf("other word on the page: interval %d, want page's 4", got.ReviewInterval)
	}

	page, err := store.ReadPage("vocabulary/food.md")
	if err != nil {
		t.Fatal(err)
	}
	if page.ReviewInterval != 4 {
		t.Errorf("page interval changed to %d, want 4", page.ReviewInterval)
	}
}

func TestRecordReviewKeepsThePageUnchanged(t *testing.T) {
	store := newTestStore(t)
	cards, err := store.DrillableCards()
	if err != nil {
		t.Fatal(err)
	}
	book := findCard(t, cards, "characters/shu1.md")
	if err := store.RecordReview(&book, model.DiffGood, time.Now()); err != nil {
		t.Fatalf("RecordReview: %v", err)
	}
	cards, err = store.DrillableCards()
	if err != nil {
		t.Fatal(err)
	}
	if got := findCard(t, cards, "characters/shu1.md"); got.ReviewInterval != 7 {
		t.Errorf("saved interval = %d, want 7 (3 × 2.5)", got.ReviewInterval)
	}
	page, err := store.ReadPage("characters/shu1.md")
	if err != nil {
		t.Fatal(err)
	}
	if page.ReviewInterval != 3 {
		t.Errorf("the page changed to interval %d; pages are content only", page.ReviewInterval)
	}
}

func TestDrillableCardsReportsCorruptSchedule(t *testing.T) {
	store := newTestStore(t)
	if err := os.WriteFile(filepath.Join(store.Dir, wordScheduleFile), []byte("not: [valid"), 0644); err != nil {
		t.Fatal(err)
	}

	if _, err := store.DrillableCards(); err == nil {
		t.Error("expected an error for a corrupt schedule file")
	}
}

func TestUpdateSRSPassingReviewAlwaysGrows(t *testing.T) {
	tests := []struct {
		name     string
		interval int
		diff     model.DrillDifficulty
		want     int
	}{
		{"new card starts at one day", 0, model.DiffHard, 1},
		{"hard grows a one-day interval", 1, model.DiffHard, 2},
		{"hard grows a two-day interval", 2, model.DiffHard, 3},
		{"hard multiplies longer intervals", 10, model.DiffHard, 12},
		{"good multiplies", 4, model.DiffGood, 10},
		{"again resets", 30, model.DiffAgain, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			card := model.Card{ReviewInterval: tt.interval}
			updateSRS(&card, tt.diff, time.Now())
			if card.ReviewInterval != tt.want {
				t.Errorf("interval = %d, want %d", card.ReviewInterval, tt.want)
			}
		})
	}
}

func TestTotalXPSumsBattleSessions(t *testing.T) {
	store := newTestStore(t)
	start := time.Date(2026, 11, 18, 9, 0, 0, 0, time.UTC)
	for i, xp := range []int{40, 0, 25} {
		sess := model.Session{
			StartedAt: start.Add(time.Duration(i) * time.Minute),
			EndedAt:   start.Add(time.Duration(i)*time.Minute + 30*time.Second),
			Type:      "battle",
			XP:        xp,
		}
		if err := store.WriteSession(sess); err != nil {
			t.Fatal(err)
		}
	}

	if got := store.TotalXP(); got != 65 {
		t.Errorf("TotalXP() = %d, want 65", got)
	}

	data, err := os.ReadFile(filepath.Join(store.Dir, "sessions", "2026-11-18-090000.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "- **XP**: 40") {
		t.Errorf("session file is missing its XP line:\n%s", data)
	}
}
