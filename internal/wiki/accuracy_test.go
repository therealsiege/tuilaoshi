package wiki

import (
	"testing"
	"time"
)

func TestRecentAccuracyReadsTheNewestSessions(t *testing.T) {
	store := NewStore(t.TempDir())
	if pct, n := store.RecentAccuracy(60); pct != 0 || n != 0 {
		t.Errorf("no sessions: %d%% of %d", pct, n)
	}
	now := time.Now()
	sessions := []struct{ right, wrong int }{{0, 40}, {30, 0}, {27, 3}} // oldest first
	for i, s := range sessions {
		if err := store.WriteSession(answers(now.Add(time.Duration(i)*time.Minute), s.right, s.wrong)); err != nil {
			t.Fatal(err)
		}
	}
	if pct, n := store.RecentAccuracy(60); pct != 95 || n != 60 {
		t.Errorf("last 60 answers: %d%% of %d, want 95%% of 60", pct, n)
	}
}
