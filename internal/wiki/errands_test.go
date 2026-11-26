package wiki

import (
	"testing"
	"time"
)

func TestAnErrandADayPerTownPaysAndIsKept(t *testing.T) {
	store := NewStore(t.TempDir())
	day := time.Date(2026, 11, 22, 9, 0, 0, 0, time.Local)
	if taken, err := store.ErrandTaken("changan", day); err != nil || taken {
		t.Fatalf("taken before taking: %v %v", taken, err)
	}
	if err := store.TakeErrand("changan", day); err != nil {
		t.Fatal(err)
	}
	if taken, err := store.ErrandTaken("changan", day); err != nil || !taken {
		t.Errorf("not taken today: %v %v", taken, err)
	}
	if taken, err := store.ErrandTaken("gao-village", day); err != nil || taken {
		t.Error("another town's errand was taken")
	}
	if taken, err := store.ErrandTaken("changan", day.AddDate(0, 0, 1)); err != nil || taken {
		t.Error("still taken tomorrow")
	}
	if err := store.FinishErrand("changan", "订一间房", day, Wallet{Tongqian: 30}); err != nil {
		t.Fatal(err)
	}
	if err := store.AddDeed(Deed{Date: "2026-11-22", Kind: DeedHaggle, Town: "changan", What: "20份斋饭", Coins: 25}); err != nil {
		t.Fatal(err)
	}
	w, werr := store.Wallet()
	deeds, err := store.Deeds()
	if werr != nil || err != nil || w.Tongqian != 30 || len(deeds) != 2 || deeds[0].Kind != DeedErrand || deeds[0].Date != "2026-11-22" {
		t.Errorf("purse %+v %v, deeds %+v %v", w, werr, deeds, err)
	}
}
