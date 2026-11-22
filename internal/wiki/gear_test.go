package wiki

import (
	"errors"
	"testing"
)

func TestBuyAndWearGear(t *testing.T) {
	store := NewStore(t.TempDir())
	if err := store.Earn(Wallet{Tongqian: 100}); err != nil {
		t.Fatal(err)
	}
	if err := store.Buy("staff", Wallet{Tongqian: 150}); !errors.Is(err, ErrCantAfford) {
		t.Fatalf("buying beyond the purse: %v", err)
	}
	if err := store.Wear("weapon", "staff"); err == nil {
		t.Fatal("wore gear not owned")
	}
	for i := 0; i < 2; i++ { // the second buy is a no-op
		if err := store.Buy("staff", Wallet{Tongqian: 80}); err != nil {
			t.Fatal(err)
		}
	}
	if err := store.Wear("weapon", "staff"); err != nil {
		t.Fatal(err)
	}
	w, err := store.Wallet()
	if err != nil || w.Tongqian != 20 {
		t.Fatalf("purse %+v %v, want 20 copper left", w, err)
	}
	kit, err := store.Kit()
	if err != nil || len(kit.Owned) != 1 || kit.Worn["weapon"] != "staff" {
		t.Fatalf("kit %+v %v", kit, err)
	}
	if err := store.Wear("weapon", ""); err != nil {
		t.Fatal(err)
	}
	if kit, err := store.Kit(); err != nil || kit.Worn["weapon"] != "" {
		t.Fatalf("took off the staff: %+v %v", kit, err)
	}
}
