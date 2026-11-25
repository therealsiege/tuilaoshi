package wiki

import (
	"testing"
	"time"
)

func TestWoundsHealByStudyAndAtTheInn(t *testing.T) {
	store := NewStore(t.TempDir())
	if err := store.SetWounds(30); err != nil {
		t.Fatal(err)
	}
	if err := store.SetPlace(Place{Region: "r", X: 3, Y: 4}); err != nil {
		t.Fatal(err)
	}
	if p, err := store.Place(); err != nil || p.Wounds != 30 || p.X != 3 {
		t.Fatalf("moving lost the wounds: %+v %v", p, err)
	}
	if err := store.WriteSession(answers(time.Now(), 5, 2)); err != nil { // 5 fresh right answers
		t.Fatal(err)
	}
	if p, err := store.Place(); err != nil || p.Wounds != 30-5*studyHeal {
		t.Fatalf("after 5 fresh answers, wounds %d, want %d", p.Wounds, 30-5*studyHeal)
	}
	if ok, err := store.RestAtInn(10); err != nil || ok {
		t.Fatalf("rested without the price: %v %v", ok, err)
	}
	if err := store.Earn(Wallet{Tongqian: 10}); err != nil {
		t.Fatal(err)
	}
	ok, err := store.RestAtInn(10)
	p, perr := store.Place()
	w, werr := store.Wallet()
	if err != nil || perr != nil || werr != nil || !ok || p.Wounds != 0 || w.Tongqian != 5 {
		t.Errorf("rest at the inn: %v, wounds %d, purse %+v (%v %v %v)", ok, p.Wounds, w, err, perr, werr)
	}
}

func TestTrailRoundTrips(t *testing.T) {
	store := NewStore(t.TempDir())
	var got struct{ Day, Food int }
	if ok, err := store.Trail(&got); ok || err != nil {
		t.Fatalf("a new journey has a trail: %v %v", ok, err)
	}
	if err := store.SaveTrail(struct{ Day, Food int }{3, 17}); err != nil {
		t.Fatal(err)
	}
	if ok, err := store.Trail(&got); !ok || err != nil || got.Day != 3 || got.Food != 17 {
		t.Errorf("trail %+v %v %v", got, ok, err)
	}
	if err := store.SetPlace(Place{Region: "r", X: 1}); err != nil {
		t.Fatal(err)
	}
	if ok, err := store.Trail(&got); !ok || err != nil || got.Food != 17 {
		t.Error("moving lost the trail")
	}
}
