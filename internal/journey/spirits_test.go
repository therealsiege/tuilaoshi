package journey

import (
	"fmt"
	"testing"

	"github.com/therealsiege/tuilaoshi/internal/model"
)

func TestSpiritsAreCompleteAndUnique(t *testing.T) {
	names, radicals := map[string]bool{}, map[string]bool{}
	for _, d := range content.Bestiary {
		names[d.Name] = true
	}
	for _, s := range content.Soldiers {
		names[s.Name] = true
	}
	for _, s := range content.Spirits {
		if names[s.Name] || radicals[s.Radical] {
			t.Errorf("spirit %s (%s) repeats a name or radical", s.Name, s.Radical)
		}
		names[s.Name], radicals[s.Radical] = true, true
		if s.English == "" || s.Origin == "" || s.Powers == "" || len([]rune(s.Radical)) != 1 {
			t.Errorf("spirit %s is incomplete", s.Name)
		}
		if EnemyLookup(s.Name) == nil {
			t.Errorf("spirit %s can't be looked up for its abilities", s.Name)
		}
	}
}

func TestSpiritWakesWithItsFamily(t *testing.T) {
	var pages []model.Page
	var cards []model.Card
	for i := 0; i < SpiritWakes; i++ {
		ch := string(rune(0x6C5F + i)) // stand-ins, all given the water radical
		pages = append(pages, model.Page{Type: "character", Character: ch, Radical: "氵(water)", ReviewInterval: 7})
		cards = append(cards, model.Card{Chinese: ch + "边"}, model.Card{Chinese: fmt.Sprint(i)})
	}
	awake := AwakeSpirits(pages, nil)
	if len(awake) != 1 || awake[0].Name != "河伯" {
		t.Fatalf("awake %v, want 河伯", awake)
	}
	if len(AwakeSpirits(pages[1:], nil)) != 0 || len(AwakeSpirits(pages, map[string]bool{"河伯": true})) != 0 {
		t.Error("a spirit woke too early, or again after being beaten")
	}
	if got := FamilyCards(cards, RadicalIndex(pages), "氵"); len(got) != SpiritWakes {
		t.Errorf("family cards %d, want %d", len(got), SpiritWakes)
	}
}
