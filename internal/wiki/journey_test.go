package wiki

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestRecordStageClearedKeepsFirstDate(t *testing.T) {
	store := NewStore(t.TempDir())
	if c, err := store.ClearedStages(); err != nil || len(c) != 0 {
		t.Fatalf("fresh wiki: cleared %v, err %v", c, err)
	}
	first := time.Date(2026, 11, 21, 9, 0, 0, 0, time.UTC)
	for _, when := range []time.Time{first, first.AddDate(0, 0, 3)} {
		if err := store.RecordStageCleared(8, when); err != nil {
			t.Fatal(err)
		}
	}
	c, err := store.ClearedStages()
	if err != nil {
		t.Fatal(err)
	}
	if len(c) != 1 || c[8] != "2026-11-21" {
		t.Errorf("cleared %v, want stage 8 on the first win's date", c)
	}
}

func TestClearedStagesReportsCorruptFile(t *testing.T) {
	store := NewStore(t.TempDir())
	if err := os.WriteFile(filepath.Join(store.Dir, journeyFile), []byte("cleared: [nope"), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := store.ClearedStages(); err == nil {
		t.Error("corrupt journey file returned no error")
	}
}
