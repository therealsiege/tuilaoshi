package wiki

import (
	"testing"
	"time"
)

func TestQuestsLockForTheDayAndPayOnce(t *testing.T) {
	store := NewStore(t.TempDir())
	now := time.Now()
	if err := store.FailQuest("q", now); err != nil {
		t.Fatal(err)
	}
	if st, err := store.Quests(now); err != nil || st["q"] != QuestLocked {
		t.Fatalf("after a wrong answer: %v %v", st, err)
	}
	if st, err := store.Quests(now.AddDate(0, 0, 1)); err != nil || st["q"] != QuestOpen {
		t.Fatalf("the next day: %v %v", st, err)
	}
	for i := 0; i < 2; i++ {
		if err := store.FinishQuest("q", Wallet{Tongqian: 20}); err != nil {
			t.Fatal(err)
		}
	}
	st, err := store.Quests(now)
	w, werr := store.Wallet()
	if err != nil || werr != nil || st["q"] != QuestDone || w.Tongqian != 20 {
		t.Errorf("done: %v, purse %+v (%v %v)", st, w, err, werr)
	}
}
