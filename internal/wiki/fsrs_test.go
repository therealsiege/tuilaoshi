package wiki

import (
	"database/sql"
	"math"
	"path/filepath"
	"testing"
	"time"

	"github.com/therealsiege/tuilaoshi/internal/model"
)

// The expected values were worked separately from the FSRS-4.5 formulas
// and default weights, each review on the day the previous one scheduled.
func TestFSRSGoodGoodAgainGood(t *testing.T) {
	steps := []struct {
		grade    model.DrillDifficulty
		days     int
		s, d     float64
		interval int
	}{
		{model.DiffGood, 0, 3.7145, 5.1618, 4},
		{model.DiffGood, 4, 14.808101, 5.1618, 15},
		{model.DiffAgain, 15, 3.149321, 6.901155, 3},
		{model.DiffGood, 3, 9.198208, 6.847235, 9},
	}
	var m fsrsMemory
	for i, st := range steps {
		m = m.next(st.grade, 0, st.days)
		if math.Abs(m.Stability-st.s) > 1e-5 || math.Abs(m.Difficulty-st.d) > 1e-5 {
			t.Errorf("review %d: stability %.6f difficulty %.6f, want %.6f and %.6f", i+1, m.Stability, m.Difficulty, st.s, st.d)
		}
		if got := m.interval(); got != st.interval {
			t.Errorf("review %d: interval %d, want %d", i+1, got, st.interval)
		}
	}
	if r := retrievability(10, 10); math.Abs(r-0.9) > 1e-9 {
		t.Errorf("recall after its stability = %.6f, want 0.9", r)
	}
}

func TestFSRSSeedsFromAClassicInterval(t *testing.T) {
	m := fsrsMemory{}.next(model.DiffGood, 10, 10) // due today after 10 days
	if m.Stability <= 10 || m.Difficulty != fsrsWeights[4] {
		t.Errorf("seeded memory %+v, want stability above 10 and difficulty %.4f", m, fsrsWeights[4])
	}
}

func TestSchedulerSettingChoosesTheInterval(t *testing.T) {
	now := time.Date(2026, 11, 18, 0, 0, 0, 0, time.UTC)
	for _, tt := range []struct {
		scheduler string
		want      int
	}{
		{SchedulerClassic, 1}, // a new card's first day
		{SchedulerFSRS, 4},    // round(w2 = 3.7145)
	} {
		store := newTestStore(t)
		store.Scheduler = tt.scheduler
		cards, err := store.DrillableCards()
		if err != nil {
			t.Fatal(err)
		}
		rice := findCard(t, cards, "vocabulary/food.md#米饭")
		rice.ReviewInterval = 0 // new to reviews
		if err := store.RecordReview(&rice, model.DiffGood, now); err != nil {
			t.Fatal(err)
		}
		if rice.ReviewInterval != tt.want {
			t.Errorf("%s: interval %d, want %d", tt.scheduler, rice.ReviewInterval, tt.want)
		}
		mem, err := store.memory(rice.ID)
		if err != nil || mem.Stability != fsrsWeights[2] {
			t.Errorf("%s: saved memory %+v (%v), want stability %.4f either way", tt.scheduler, mem, err, fsrsWeights[2])
		}
	}

	store := newTestStore(t)
	if set, err := store.Settings(); err != nil || set.Scheduler != SchedulerClassic {
		t.Errorf("default scheduler %q (%v), want classic", set.Scheduler, err)
	}
	set := DefaultSettings()
	set.Scheduler = SchedulerFSRS
	if err := store.SaveSettings(set); err != nil {
		t.Fatal(err)
	}
	if got, err := store.Settings(); err != nil || got.Scheduler != SchedulerFSRS {
		t.Errorf("saved scheduler %q (%v), want fsrs", got.Scheduler, err)
	}
	set.Scheduler = "sm2"
	if err := store.SaveSettings(set); err != nil {
		t.Fatal(err)
	}
	if got, err := store.Settings(); err != nil || got.Scheduler != SchedulerClassic {
		t.Errorf("unknown scheduler read as %q (%v), want classic", got.Scheduler, err)
	}
}

func TestOldProgressFileGainsMemoryColumns(t *testing.T) {
	store := newTestStore(t)
	db, err := sql.Open("sqlite", "file:"+filepath.Join(store.Dir, progressFile))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(schema); err != nil { // the tables as they were
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO schedules (card_id, introduced, last_reviewed, review_interval, next_review, reps)
		VALUES ('characters/shu1.md', '2026-11-01', '2026-11-08', 10, '2026-11-18', 3)`); err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}

	store.Scheduler = SchedulerFSRS
	if mem, err := store.memory("characters/shu1.md"); err != nil || mem != (fsrsMemory{}) {
		t.Fatalf("upgraded row memory %+v (%v), want zero", mem, err)
	}
	cards, err := store.DrillableCards()
	if err != nil {
		t.Fatal(err)
	}
	book := findCard(t, cards, "characters/shu1.md")
	if err := store.RecordReview(&book, model.DiffGood, time.Date(2026, 11, 18, 9, 0, 0, 0, time.UTC)); err != nil {
		t.Fatal(err)
	}
	mem, err := store.memory(book.ID)
	if err != nil || mem.Stability <= 10 {
		t.Errorf("memory after review %+v (%v), want seeded from the 10-day interval and grown", mem, err)
	}
	if book.ReviewInterval != mem.interval() {
		t.Errorf("interval %d, want FSRS's %d", book.ReviewInterval, mem.interval())
	}
	// Opening again leaves the columns as they are.
	again := NewStore(store.Dir)
	defer again.Close() //nolint:errcheck // test cleanup
	if got, err := again.memory(book.ID); err != nil || got != mem {
		t.Errorf("reopened memory %+v (%v), want %+v", got, err, mem)
	}
}

func TestCompareSchedulersReplaysTheLog(t *testing.T) {
	store := newTestStore(t)
	cards, err := store.DrillableCards()
	if err != nil {
		t.Fatal(err)
	}
	day := time.Date(2026, 11, 1, 9, 0, 0, 0, time.UTC)
	rice := findCard(t, cards, "vocabulary/food.md#米饭")
	rice.ReviewInterval = 0
	// Good on the 1st, then Good on the 2nd: classic gives 1 then 2 days
	// (due the 4th); FSRS gives 4 days and more (due after the 4th).
	for _, at := range []time.Time{day, day.AddDate(0, 0, 1)} {
		if err := store.RecordReview(&rice, model.DiffGood, at); err != nil {
			t.Fatal(err)
		}
	}
	classic, fsrs, err := store.CompareSchedulers(day.AddDate(0, 0, 2), 2) // the 3rd and 4th
	if err != nil {
		t.Fatal(err)
	}
	if classic != 1 || fsrs != 0 {
		t.Errorf("due in 2 days: classic %d, fsrs %d; want 1 and 0", classic, fsrs)
	}
}
