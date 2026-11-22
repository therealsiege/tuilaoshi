package journey

import "testing"

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
