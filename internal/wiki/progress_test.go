package wiki

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/therealsiege/tuilaoshi/internal/model"
)

// oldWiki is a wiki kept the way it was before the progress file.
func oldWiki(t *testing.T, files map[string]string) *Store {
	t.Helper()
	dir := t.TempDir()
	for name, body := range files {
		path := filepath.Join(dir, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(body), 0644); err != nil {
			t.Fatal(err)
		}
	}
	store := NewStore(dir)
	t.Cleanup(func() { _ = store.Close() }) //nolint:errcheck // test cleanup
	return store
}

func TestOldFilesAreImportedOnce(t *testing.T) {
	store := oldWiki(t, map[string]string{
		"study.yaml":           "shields: 2\n",
		"settings.yaml":        "new_words: 4\n",
		"saves/journey-1.yaml": "name: 西行\ncleared:\n    1: \"2026-11-01\"\ncoins:\n    tongqian: 55\n",
		"word-schedule.yaml":   "vocabulary/food.md#米饭:\n    review_interval: 9\n    next_review: \"2026-12-01\"\n",
	})
	saves, err := store.Saves()
	if err != nil || len(saves) != 1 || saves[0].Coins.Tongqian != 55 || saves[0].Cleared != 1 {
		t.Fatalf("saves %+v %v", saves, err)
	}
	if set, err := store.Settings(); err != nil || set.NewWords != 4 {
		t.Errorf("settings %+v %v", set, err)
	}
	if rec, err := store.readStudy(); err != nil || rec.Shields != 2 {
		t.Errorf("study %+v %v", rec, err)
	}
	sch, err := store.readSchedules()
	if err != nil || sch["vocabulary/food.md#米饭"].ReviewInterval != 9 {
		t.Errorf("schedules %+v %v", sch, err)
	}
	for _, name := range []string{"study.yaml", "settings.yaml", "saves/journey-1.yaml", "word-schedule.yaml"} {
		if _, err := os.Stat(filepath.Join(store.Dir, filepath.FromSlash(name)+".imported")); err != nil {
			t.Errorf("%s was not kept as .imported: %v", name, err)
		}
	}
	// A file appearing later is not imported again over the progress.
	if err := os.WriteFile(filepath.Join(store.Dir, "settings.yaml"), []byte("new_words: 9\n"), 0644); err != nil {
		t.Fatal(err)
	}
	reopened := NewStore(store.Dir)
	defer reopened.Close() //nolint:errcheck // test
	if set, err := reopened.Settings(); err != nil || set.NewWords != 4 {
		t.Errorf("reopened settings %+v %v, want the saved 4", set, err)
	}
}

func TestReviewsAreLoggedWithLapses(t *testing.T) {
	store := newTestStore(t)
	cards, err := store.DrillableCards()
	if err != nil {
		t.Fatal(err)
	}
	book := findCard(t, cards, "characters/shu1.md")
	now := time.Date(2026, 11, 25, 9, 0, 0, 0, time.Local)
	for _, d := range []model.DrillDifficulty{model.DiffGood, model.DiffAgain, model.DiffGood} {
		if err := store.RecordReview(&book, d, now); err != nil {
			t.Fatal(err)
		}
	}
	db, err := store.progress()
	if err != nil {
		t.Fatal(err)
	}
	var reps, lapses, logged int
	if err := db.QueryRow(`SELECT reps, lapses FROM schedules WHERE card_id = ?`, book.ID).Scan(&reps, &lapses); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`SELECT count(*) FROM reviews WHERE card_id = ?`, book.ID).Scan(&logged); err != nil {
		t.Fatal(err)
	}
	if reps != 3 || lapses != 1 || logged != 3 {
		t.Errorf("reps %d lapses %d logged %d, want 3 1 3", reps, lapses, logged)
	}
}

func TestGeneratedWordPagesMoveToPrivate(t *testing.T) {
	store := oldWiki(t, map[string]string{
		"vocabulary/festivals.md": "---\ntype: vocabulary\n---\n\n| 词 | 拼音 | 意思 |\n|---|---|---|\n| 月饼 | yuèbing | mooncake |\n",
		"word-schedule.yaml":      "vocabulary/festivals.md#月饼:\n    review_interval: 6\n",
	})
	sch, err := store.readSchedules()
	if err != nil {
		t.Fatal(err)
	}
	if sch["private/vocabulary/festivals.md#月饼"].ReviewInterval != 6 {
		t.Errorf("the word's schedule didn't follow the page: %+v", sch)
	}
	if _, err := os.Stat(filepath.Join(store.Dir, festivalPage)); err != nil {
		t.Errorf("page not moved: %v", err)
	}
}

func TestBackupsAreDailyAndPruned(t *testing.T) {
	store := newTestStore(t)
	db, err := store.progress()
	if err != nil {
		t.Fatal(err)
	}
	day := time.Date(2026, 11, 1, 9, 0, 0, 0, time.Local)
	for i := range 10 {
		if err := store.backupDaily(db, day.AddDate(0, 0, i)); err != nil {
			t.Fatal(err)
		}
	}
	if err := store.backupDaily(db, day.AddDate(0, 0, 9)); err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(filepath.Join(store.Dir, backupsDir))
	if err != nil || len(entries) != keepBackups || entries[0].Name() != "progress-2026-11-04.db" {
		t.Errorf("backups %v %v, want the newest %d from 11-04", entries, err, keepBackups)
	}
}

func TestExportThenImportOnAnotherMac(t *testing.T) {
	here := newTestStore(t)
	if err := here.SaveSettings(Settings{NewWords: 3, ReviewChunk: 50}); err != nil {
		t.Fatal(err)
	}
	cards, err := here.DrillableCards()
	if err != nil {
		t.Fatal(err)
	}
	book := findCard(t, cards, "characters/shu1.md")
	if err := here.RecordReview(&book, model.DiffAgain, time.Now()); err != nil {
		t.Fatal(err)
	}
	out := t.TempDir()
	if err := here.Export(out); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"settings.yaml", exportSchedules, exportReviews} {
		if _, err := os.Stat(filepath.Join(out, name)); err != nil {
			t.Errorf("export lacks %s: %v", name, err)
		}
	}
	there := newTestStore(t)
	for range 2 { // importing twice adds no review twice
		if err := there.Import(out); err != nil {
			t.Fatal(err)
		}
	}
	if set, err := there.Settings(); err != nil || set.NewWords != 3 {
		t.Errorf("settings %+v %v", set, err)
	}
	db, err := there.progress()
	if err != nil {
		t.Fatal(err)
	}
	var lapses, logged int
	if err := db.QueryRow(`SELECT lapses FROM schedules WHERE card_id = ?`, book.ID).Scan(&lapses); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`SELECT count(*) FROM reviews`).Scan(&logged); err != nil {
		t.Fatal(err)
	}
	if lapses != 1 || logged != 1 {
		t.Errorf("lapses %d, reviews %d; want 1 and 1", lapses, logged)
	}
	var stability float64
	if err := db.QueryRow(`SELECT stability FROM schedules WHERE card_id = ?`, book.ID).Scan(&stability); err != nil || stability == 0 {
		t.Errorf("FSRS stability didn't travel: %v %v", stability, err)
	}
}
