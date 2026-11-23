package wiki

import (
	"errors"
	"testing"
)

func TestCraftingSpendsMaterialsAndCopper(t *testing.T) {
	store := NewStore(t.TempDir())
	recipe := map[string]int{"hide": 3}
	if err := store.AddMaterials([]string{"hide", "hide"}); err != nil {
		t.Fatal(err)
	}
	if err := store.Earn(Wallet{Tongqian: 25}); err != nil {
		t.Fatal(err)
	}
	if err := store.Craft("vest", recipe, Wallet{Tongqian: 20}); !errors.Is(err, ErrMissingMaterials) {
		t.Fatalf("crafting with 2 of 3 hides: %v", err)
	}
	if err := store.AddMaterials([]string{"hide"}); err != nil {
		t.Fatal(err)
	}
	if err := store.Craft("vest", recipe, Wallet{Tongqian: 20}); err != nil {
		t.Fatal(err)
	}
	mats, err := store.Materials()
	w, werr := store.Wallet()
	kit, kerr := store.Kit()
	if err != nil || werr != nil || kerr != nil || mats["hide"] != 0 || w.Tongqian != 5 || len(kit.Owned) != 1 {
		t.Errorf("after crafting: materials %v, purse %+v, owned %v (%v %v %v)", mats, w, kit.Owned, err, werr, kerr)
	}
}
