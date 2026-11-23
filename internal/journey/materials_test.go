package journey

import "testing"

func TestRecipesUseMaterialsThatDrop(t *testing.T) {
	dropped := map[string]bool{"core": true}
	for _, r := range content.Regions {
		for _, rank := range []string{RankMinion, RankElite, RankCaptain} {
			for _, e := range []BestiaryEntry{{}, {Abilities: []string{"burn"}}, {Abilities: []string{"venom"}}, {Abilities: []string{"blind"}}} {
				first, _ := Drop(e, rank, r.Terrain)
				dropped[first] = true
			}
		}
	}
	used := map[string]bool{}
	for _, g := range content.Gear {
		for id, n := range g.Recipe {
			used[id] = true
			if MaterialByID(id) == nil || n < 1 {
				t.Errorf("%s needs unknown material %q x%d", g.ID, id, n)
			}
		}
	}
	for _, m := range content.Materials {
		if !dropped[m.ID] {
			t.Errorf("material %s never drops", m.ID)
		}
		if !used[m.ID] {
			t.Errorf("material %s isn't used in any recipe", m.ID)
		}
		if m.Name == "" || m.English == "" || m.Desc == "" {
			t.Errorf("material %s is incomplete", m.ID)
		}
	}
	if first, core := Drop(BestiaryEntry{Abilities: []string{"burn"}}, "demon", "mountain"); first != "fire" || !core {
		t.Errorf("a burning demon drops %s, core %v", first, core)
	}
}
