package journey

import (
	"math/rand"
	"testing"
)

func TestRegionsCoverTheRoad(t *testing.T) {
	next := 1
	terrains := map[string]bool{"mountain": true, "forest": true, "river": true, "desert": true, "fire": true, "palace": true, "cave": true}
	for _, r := range content.Regions {
		if r.From != next || r.To < r.From {
			t.Errorf("region %s covers %d-%d, want it to start at %d", r.ID, r.From, r.To, next)
		}
		next = r.To + 1
		if r.Name == "" || r.English == "" || !terrains[r.Terrain] {
			t.Errorf("region %s is incomplete or has terrain %q", r.ID, r.Terrain)
		}
	}
	if next != len(content.Tribulations)+1 {
		t.Errorf("regions end at %d, want 81", next-1)
	}
	towns := 0
	for _, r := range content.Regions {
		towns += len(r.TownsIn())
	}
	if towns != len(content.Towns) {
		t.Errorf("%d towns placed in regions, want all %d", towns, len(content.Towns))
	}
}

func TestEncountersFavourMinionsOfTheRegion(t *testing.T) {
	rng := rand.New(rand.NewSource(3))
	r := content.Regions[RegionByID("heifeng")]
	ranks := map[string]int{}
	for i := 0; i < 2000; i++ {
		s, ok := r.Encounter(rng)
		if !ok || s.To < r.From || s.From > r.To {
			t.Fatalf("encountered %s (%d-%d) in %s", s.Name, s.From, s.To, r.ID)
		}
		ranks[s.Rank]++
	}
	if ranks[RankMinion] <= ranks[RankElite] || ranks[RankElite] <= ranks[RankCaptain] || ranks[RankCaptain] == 0 {
		t.Errorf("rank mix %v", ranks)
	}
	if _, ok := content.Regions[len(content.Regions)-1].Encounter(rng); ok {
		t.Error("Vulture Peak should be free of soldiers")
	}
}
