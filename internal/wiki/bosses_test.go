package wiki

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestRecordBossVictoryKeepsFirstDate(t *testing.T) {
	store := NewStore(t.TempDir())
	if v, err := store.BossVictories(); err != nil || len(v) != 0 {
		t.Fatalf("fresh wiki: victories %v, err %v", v, err)
	}
	first := time.Date(2026, 11, 21, 9, 0, 0, 0, time.UTC)
	if err := store.RecordBossVictory("白骨精", first); err != nil {
		t.Fatal(err)
	}
	if err := store.RecordBossVictory("白骨精", first.AddDate(0, 0, 3)); err != nil {
		t.Fatal(err)
	}
	v, err := store.BossVictories()
	if err != nil {
		t.Fatal(err)
	}
	if v["白骨精"] != "2026-11-21" {
		t.Errorf("victory date = %q, want the first win 2026-11-21", v["白骨精"])
	}
}

func TestBossVictoriesReportsCorruptFile(t *testing.T) {
	store := NewStore(t.TempDir())
	if err := os.WriteFile(filepath.Join(store.Dir, bossFile), []byte("not: [valid"), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := store.BossVictories(); err == nil {
		t.Error("corrupt boss file returned no error")
	}
}
