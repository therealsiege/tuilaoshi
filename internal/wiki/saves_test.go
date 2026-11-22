package wiki

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestJourneysKeepTheirOwnGameButShareStudy(t *testing.T) {
	store := NewStore(t.TempDir())
	t0 := time.Now()
	first, err := store.NewSave("first", t0)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.RecordStageCleared(1, t0); err != nil {
		t.Fatal(err)
	}
	if err := store.WriteSession(answers(t0, 3, 0)); err != nil { // pays 3 into "first"
		t.Fatal(err)
	}
	second, err := store.NewSave("second", t0.Add(time.Second))
	if err != nil {
		t.Fatal(err)
	}
	if cleared, err := store.ClearedStages(); err != nil || len(cleared) != 0 {
		t.Fatalf("a new journey starts with cleared stages %v %v", cleared, err)
	}
	same := answers(t0.Add(time.Minute), 3, 0)
	same.Items = answers(t0, 3, 0).Items // the words already paid today
	if err := store.WriteSession(same); err != nil {
		t.Fatal(err)
	}
	if w, err := store.Wallet(); err != nil || w.Tongqian != 0 {
		t.Errorf("words paid in one journey paid again in another: %+v %v", w, err)
	}

	if err := store.UseSave(first, t0.Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	if w, err := store.Wallet(); err != nil || w.Tongqian != 3 {
		t.Errorf("first journey's purse %+v %v, want 3", w, err)
	}
	saves, err := store.Saves()
	if err != nil || len(saves) != 2 || saves[0].ID != first || saves[0].Cleared != 1 || saves[1].ID != second {
		t.Fatalf("saves %+v %v; want first (just played, 1 cleared), then second", saves, err)
	}
	if err := store.DeleteSave(second); err != nil {
		t.Fatal(err)
	}
	if saves, err := store.Saves(); err != nil || len(saves) != 1 {
		t.Errorf("after deleting: %+v", saves)
	}
	if err := store.UseSave("../study", t0); err == nil {
		t.Error("UseSave accepted a path")
	}
}

func TestOldJourneyIsAdoptedAsASave(t *testing.T) {
	dir := t.TempDir()
	old := "cleared: {1: \"2026-11-01\"}\ncoins: {tongqian: 40}\nshields: 1\n"
	if err := os.WriteFile(filepath.Join(dir, journeyFile), []byte(old), 0644); err != nil {
		t.Fatal(err)
	}
	store := NewStore(dir)
	saves, err := store.Saves()
	if err != nil || len(saves) != 1 || saves[0].Cleared != 1 || saves[0].Coins.Tongqian != 40 {
		t.Fatalf("adopted saves %+v %v", saves, err)
	}
	if _, err := os.Stat(filepath.Join(dir, journeyFile)); !os.IsNotExist(err) {
		t.Error("the old journey file is still in place")
	}
	if _, shields, err := store.Bounties(time.Now()); err != nil || shields != 1 {
		t.Errorf("study bookkeeping not carried over: %d shields, %v", shields, err)
	}
}
