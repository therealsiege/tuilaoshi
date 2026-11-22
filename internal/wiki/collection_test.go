package wiki

import "testing"

func TestDefeatsAndClaimsCountOnce(t *testing.T) {
	store := NewStore(t.TempDir())
	for i := 0; i < 2; i++ {
		if err := store.RecordDefeat("虎先锋"); err != nil {
			t.Fatal(err)
		}
	}
	paid := 0
	for i := 0; i < 2; i++ {
		ok, err := store.Claim("enemies-13", Wallet{Tongqian: 50, Yuanbao: 1})
		if err != nil {
			t.Fatal(err)
		}
		if ok {
			paid++
		}
	}
	c, err := store.Collection()
	if err != nil || len(c.Defeated) != 1 || len(c.Claimed) != 1 || paid != 1 {
		t.Fatalf("collection %+v, paid %d times, %v", c, paid, err)
	}
	if w, err := store.Wallet(); err != nil || w != (Wallet{Tongqian: 50, Yuanbao: 1}) {
		t.Errorf("purse %+v %v", w, err)
	}
}
