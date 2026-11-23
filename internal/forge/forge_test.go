package forge

import (
	"math/rand"
	"slices"
	"testing"

	"github.com/therealsiege/tuilaoshi/internal/model"
)

func TestTrayHoldsThePartsAndDecoys(t *testing.T) {
	rng := rand.New(rand.NewSource(1))
	tray := Tray([]string{"氵", "可"}, []string{"氵", "木", "口", "可", "女", "子", "心", "相", "艹", "早"}, rng)
	if len(tray) != TraySize || !slices.Contains(tray, "氵") || !slices.Contains(tray, "可") {
		t.Fatalf("tray %v", tray)
	}
	seen := map[string]bool{}
	for _, p := range tray {
		if seen[p] {
			t.Errorf("tray repeats %s", p)
		}
		seen[p] = true
	}
}

func TestForgedIgnoresOrderButCountsRepeats(t *testing.T) {
	if !Forged([]string{"氵", "可"}, []string{"可", "氵"}) {
		t.Error("order shouldn't matter")
	}
	if Forged([]string{"木", "木", "木"}, []string{"木", "木"}) || !Forged([]string{"木", "木", "木"}, []string{"木", "木", "木"}) {
		t.Error("森 needs three 木")
	}
	if Forged([]string{"氵", "可"}, []string{"氵", "口"}) {
		t.Error("a wrong part forged")
	}
}

func TestTargetsAreCharactersStillBeingLearned(t *testing.T) {
	cards := []model.Card{
		{ID: "characters/he.md", Chinese: "河", ReviewInterval: 1},
		{ID: "characters/shan.md", Chinese: "山", ReviewInterval: 30},
		{ID: "vocabulary/a.md#河流", Chinese: "河流"},
	}
	got := Targets(cards, 5, rand.New(rand.NewSource(1)))
	if len(got) != 1 || got[0].Chinese != "河" {
		t.Errorf("targets %v", got)
	}
	if Forgeable("人", []string{"人"}) || !Forgeable("河", []string{"氵", "可"}) {
		t.Error("forgeable")
	}
	if Material("氵(water)") != "pearl" || Material("金") != "core" || Material("") != "bone" {
		t.Error("materials by radical")
	}
}
