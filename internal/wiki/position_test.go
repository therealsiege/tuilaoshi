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
	if ok, err := store.RestAtInn(); err != nil || ok {
		t.Fatalf("rested without the price: %v %v", ok, err)
	}
	if err := store.Earn(Wallet{Tongqian: InnPrice}); err != nil {
		t.Fatal(err)
	}
	ok, err := store.RestAtInn()
	p, perr := store.Place()
	w, werr := store.Wallet()
	if err != nil || perr != nil || werr != nil || !ok || p.Wounds != 0 || w.Tongqian != 5 {
		t.Errorf("rest at the inn: %v, wounds %d, purse %+v (%v %v %v)", ok, p.Wounds, w, err, perr, werr)
	}
}
