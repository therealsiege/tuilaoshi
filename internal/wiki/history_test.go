package wiki

import (
	"testing"
	"time"

	"github.com/therealsiege/tuilaoshi/internal/model"
)

func review(t *testing.T, store *Store, id string, d model.DrillDifficulty, at time.Time) {
	t.Helper()
	cards, err := store.DrillableCards()
	if err != nil {
		t.Fatal(err)
	}
	c := findCard(t, cards, id)
	if err := store.RecordReview(&c, d, at); err != nil {
		t.Fatal(err)
	}
}

func TestHardWordsAreTheOnesMissedSixTimes(t *testing.T) {
	store := newTestStore(t)
	now := time.Date(2026, 11, 25, 9, 0, 0, 0, time.Local)
	for range HardWordLapses {
		review(t, store, "characters/shu1.md", model.DiffAgain, now)
	}
	review(t, store, "vocabulary/food.md#米饭", model.DiffAgain, now)
	hard, err := store.HardWords()
	if err != nil || len(hard) != 1 || hard[0].CardID != "characters/shu1.md" || hard[0].Lapses != HardWordLapses {
		t.Errorf("hard words %+v %v", hard, err)
	}
}

func TestForecastAndRetention(t *testing.T) {
	store := newTestStore(t)
	now := time.Date(2026, 11, 25, 9, 0, 0, 0, time.Local)
	review(t, store, "vocabulary/food.md#米饭", model.DiffAgain, now) // due tomorrow
	review(t, store, "characters/shu1.md", model.DiffGood, now)     // 3 × 2.5: due in 7 days
	f, err := store.DueForecast(now, 7)
	if err != nil || f[0] != 0 || f[1] != 1 {
		t.Errorf("forecast %v %v, want nothing today and 1 tomorrow", f, err)
	}
	if f, _ = store.DueForecast(now.AddDate(0, 0, 8), 1); f[0] != 2 { //nolint:errcheck // checked above
		t.Errorf("a week later both are due: %v", f)
	}
	right, total, err := store.Retention(now, 30)
	if err != nil || right != 1 || total != 2 {
		t.Errorf("retention %d/%d %v, want 1/2", right, total, err)
	}
}

func TestNewWordsWaitWhileTooManyAreDue(t *testing.T) {
	store := newTestStore(t)
	now := time.Date(2026, 11, 25, 9, 0, 0, 0, time.Local)
	if why, err := store.NewWordsPaused(10, now); err != nil || why != "" {
		t.Fatalf("paused with nothing due: %q %v", why, err)
	}
	db, err := store.progress()
	if err != nil {
		t.Fatal(err)
	}
	for i := range PauseDue + 1 {
		if _, err := db.Exec(`INSERT INTO schedules (card_id, last_reviewed, review_interval, next_review) VALUES (?, '2026-11-20', 1, '2026-11-21')`,
			"characters/x"+string(rune('a'+i%26))+string(rune('a'+i/26))+".md"); err != nil {
			t.Fatal(err)
		}
	}
	if why, err := store.NewWordsPaused(10, now); err != nil || why == "" {
		t.Errorf("not paused with %d due: %q %v", PauseDue+1, why, err)
	}
}
