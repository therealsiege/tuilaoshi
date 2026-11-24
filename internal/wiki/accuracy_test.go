package wiki

import (
	"strings"
	"testing"
	"time"

	"github.com/therealsiege/tuilaoshi/internal/model"
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

func TestMissedSinceListsEachWrongCardOnce(t *testing.T) {
	store := NewStore(t.TempDir())
	now := time.Now()
	for i, s := range []struct {
		at    time.Time
		wrong []string
	}{
		{now.AddDate(0, 0, -10), []string{"old"}},
		{now.Add(-time.Hour), []string{"a", "b"}},
		{now, []string{"b", "c"}},
	} {
		sess := answers(s.at.Add(time.Duration(i)*time.Second), 1, 0)
		for _, id := range s.wrong {
			sess.Items = append(sess.Items, model.SessionItem{CardID: id})
		}
		if err := store.WriteSession(sess); err != nil {
			t.Fatal(err)
		}
	}
	got := store.MissedSince(now.AddDate(0, 0, -3))
	if strings.Join(got, ",") != "a,b,c" {
		t.Errorf("missed %v, want a,b,c", got)
	}
}

func TestSessionsInTheSameSecondKeepTheirOwnFiles(t *testing.T) {
	store := NewStore(t.TempDir())
	at := time.Now()
	for i := 0; i < 3; i++ {
		if err := store.WriteSession(answers(at, 1, 1)); err != nil {
			t.Fatal(err)
		}
	}
	if got := store.TodayStats().CardsReviewed; got != 6 {
		t.Errorf("three sessions in one second count %d answers, want 6", got)
	}
}
