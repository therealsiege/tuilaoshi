package wiki

import (
	"os"
	"testing"
	"time"

	"github.com/therealsiege/tuilaoshi/internal/model"
)

func TestTwoMacsSync(t *testing.T) {
	root := t.TempDir()
	studio, laptop := newTestStore(t), newTestStore(t)
	now := time.Date(2026, 11, 26, 9, 0, 0, 0, time.Local)

	if err := studio.SaveSettings(Settings{NewWords: 5, ReviewChunk: 30}); err != nil {
		t.Fatal(err)
	}
	review(t, studio, "characters/shu1.md", model.DiffGood, now)
	if err := studio.Sync(root, "studio"); err != nil {
		t.Fatal(err)
	}
	if err := laptop.Sync(root, "laptop"); err != nil {
		t.Fatal(err)
	}
	if set, err := laptop.Settings(); err != nil || set.NewWords != 5 {
		t.Fatalf("the laptop didn't get the studio's settings: %+v %v", set, err)
	}
	if sch, err := laptop.readSchedules(); err != nil || sch["characters/shu1.md"].LastReviewed == "" {
		t.Fatalf("the laptop didn't get the review: %+v %v", sch, err)
	}

	// The laptop changes settings later and reviews again the next day.
	time.Sleep(1100 * time.Millisecond) // docs are stamped to the second
	if err := laptop.SaveSettings(Settings{NewWords: 20, ReviewChunk: 30}); err != nil {
		t.Fatal(err)
	}
	review(t, laptop, "characters/shu1.md", model.DiffGood, now.AddDate(0, 0, 1))
	if err := laptop.Sync(root, "laptop"); err != nil {
		t.Fatal(err)
	}
	if err := studio.Sync(root, "studio"); err != nil {
		t.Fatal(err)
	}
	if set, err := studio.Settings(); err != nil || set.NewWords != 20 {
		t.Errorf("the newer settings should win: %+v %v", set, err)
	}
	sch, err := studio.readSchedules()
	if err != nil || sch["characters/shu1.md"].LastReviewed != FormatDate(now.AddDate(0, 0, 1)) {
		t.Errorf("the later review's schedule should win: %+v %v", sch["characters/shu1.md"], err)
	}
	db, err := studio.progress()
	if err != nil {
		t.Fatal(err)
	}
	var n int
	if err := db.QueryRow(`SELECT count(*) FROM reviews`).Scan(&n); err != nil || n != 2 {
		t.Errorf("both reviews, once each: %d %v", n, err)
	}
	// An older copy doesn't overwrite a newer one.
	if err := laptop.Sync(root, "laptop"); err != nil {
		t.Fatal(err)
	}
	if set, err := laptop.Settings(); err != nil || set.NewWords != 20 {
		t.Errorf("the laptop's newer settings were overwritten: %+v %v", set, err)
	}
	if entries, err := os.ReadDir(root); err != nil || len(entries) != 2 {
		t.Errorf("the shared folder holds one folder per Mac: %v %v", entries, err)
	}
}
