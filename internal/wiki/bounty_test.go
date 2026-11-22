package wiki

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestBountiesChangeDailyAndPayOnce(t *testing.T) {
	days := map[string]bool{}
	for d := 0; d < 10; d++ {
		day := time.Date(2026, 11, 1+d, 0, 0, 0, 0, time.UTC).Format(dateFmt)
		b := BountiesFor(day)
		if len(b) != bountiesPerDay {
			t.Fatalf("%s has %d bounties", day, len(b))
		}
		days[b[0].ID+b[1].ID+b[2].ID] = true
	}
	if len(days) < 2 {
		t.Error("the same bounties every day")
	}

	rec, game := studyRecord{}, journeyRecord{}
	today := "2026-11-22"
	bounties := BountiesFor(today)
	for i, b := range bounties {
		tally(&rec, &game, today, b.ID, b.Target)
		tally(&rec, &game, today, b.ID, b.Target) // met twice, paid once
		if want := i + 1; len(rec.BountiesPaid) != want {
			t.Fatalf("after %d bounties, %d paid", want, len(rec.BountiesPaid))
		}
	}
	pay := 0
	for _, b := range bounties {
		pay += b.Pay
	}
	if game.Coins.Tongqian != pay || rec.Shields != 1 {
		t.Errorf("all bounties met: %d copper (want %d), %d shields (want 1)", game.Coins.Tongqian, pay, rec.Shields)
	}
	tally(&rec, &game, "2026-11-23", bounties[0].ID, 1)
	if len(rec.BountiesPaid) != 0 || rec.Tally[bounties[0].ID] != 1 {
		t.Errorf("a new day should start fresh: %+v", rec)
	}
}

func TestShieldCoversAMissedDay(t *testing.T) {
	store := NewStore(t.TempDir())
	now := time.Now()
	// Studied two days ago, missed yesterday, holding one shield.
	if err := store.WriteSession(answers(now.AddDate(0, 0, -2), 1, 0)); err != nil {
		t.Fatal(err)
	}
	if err := store.writeStudy(studyRecord{Shields: 1}); err != nil {
		t.Fatal(err)
	}
	if err := store.WriteSession(answers(now, 1, 0)); err != nil {
		t.Fatal(err)
	}
	if got := store.Streak(); got != 3 {
		t.Errorf("streak %d, want 3 with yesterday shielded", got)
	}
	if _, shields, err := store.Bounties(time.Now()); err != nil || shields != 0 {
		t.Errorf("shields left %d %v, want 0", shields, err)
	}
	if err := os.Remove(filepath.Join(store.Dir, studyFile)); err != nil {
		t.Fatal(err)
	}
	if got := store.Streak(); got != 1 {
		t.Errorf("without the shield the streak is %d, want 1", got)
	}
}
