package wiki

import (
	"fmt"
	"testing"
	"time"

	"github.com/therealsiege/tuilaoshi/internal/model"
)

// answers makes a session with right answers and misses, started at.
func answers(at time.Time, right, wrong int) model.Session {
	sess := model.Session{StartedAt: at, EndedAt: at, Type: "practice", CardsTotal: right + wrong, CardsCorrect: right}
	for i := 0; i < right+wrong; i++ {
		sess.Items = append(sess.Items, model.SessionItem{CardID: fmt.Sprintf("c%d-%d", at.UnixNano(), i), Correct: i < right})
	}
	return sess
}

func TestSessionsPayCopperAndTheDailyBonusOnce(t *testing.T) {
	store := NewStore(t.TempDir())
	now := time.Now()
	if err := store.WriteSession(answers(now, 8, 4)); err != nil {
		t.Fatal(err)
	}
	if w, err := store.Wallet(); err != nil || w.Tongqian != 8 {
		t.Fatalf("after 8 right of 12: %+v %v, want 8 copper", w, err)
	}
	// The second session takes the day past DailyGoal answers: bonus plus a
	// one-day streak.
	if err := store.WriteSession(answers(now.Add(time.Second), 10, 0)); err != nil {
		t.Fatal(err)
	}
	want := 8 + 10 + dailyBonus + 1
	if w, err := store.Wallet(); err != nil || w.Tongqian != want {
		t.Fatalf("after reaching the goal: %d copper, want %d", w.Tongqian, want)
	}
	if err := store.WriteSession(answers(now.Add(2*time.Second), 5, 0)); err != nil {
		t.Fatal(err)
	}
	if w, err := store.Wallet(); err != nil || w.Tongqian != want+5 {
		t.Fatalf("bonus paid twice: %d copper, want %d", w.Tongqian, want+5)
	}
}

func TestACardPaysOncePerDay(t *testing.T) {
	store := NewStore(t.TempDir())
	now := time.Now()
	same := answers(now, 3, 0)
	for i := range same.Items {
		same.Items[i].CardID = "书"
	}
	for i := 0; i < 2; i++ {
		same.StartedAt = now.Add(time.Duration(i) * time.Second)
		if err := store.WriteSession(same); err != nil {
			t.Fatal(err)
		}
	}
	if w, err := store.Wallet(); err != nil || w.Tongqian != 1 {
		t.Fatalf("one card answered right six times paid %+v %v, want 1 copper", w, err)
	}
}

func TestEarnAddsIngotsAndKeepsProgress(t *testing.T) {
	store := NewStore(t.TempDir())
	if err := store.RecordStageCleared(1, time.Now()); err != nil {
		t.Fatal(err)
	}
	if err := store.Earn(Wallet{Tongqian: 50, Yuanbao: 1}); err != nil {
		t.Fatal(err)
	}
	w, err := store.Wallet()
	if err != nil || w != (Wallet{Tongqian: 50, Yuanbao: 1}) {
		t.Fatalf("wallet %+v %v", w, err)
	}
	if cleared, err := store.ClearedStages(); err != nil || cleared[1] == "" {
		t.Error("earning lost the cleared stage")
	}
}
