package wiki

import (
	"testing"
	"time"
)

func TestTalkingRaisesABondUpToADailyCap(t *testing.T) {
	store := NewStore(t.TempDir())
	now := time.Now()
	total := 0
	for i := 0; i < 5; i++ {
		n, err := store.AddTalkBond("猪八戒", 2, now)
		if err != nil {
			t.Fatal(err)
		}
		total += n
	}
	if total != TalkCap {
		t.Errorf("a day of talk added %d, want the cap %d", total, TalkCap)
	}
	if n, err := store.AddTalkBond("猪八戒", 2, now.AddDate(0, 0, 1)); err != nil || n != 2 {
		t.Errorf("next day added %d %v, want 2", n, err)
	}
	if err := store.AddBond("猪八戒", 1); err != nil {
		t.Fatal(err)
	}
	if b, err := store.BondPoints(); err != nil || b["猪八戒"] != TalkCap+3 {
		t.Errorf("bond points %v %v", b, err)
	}
}
