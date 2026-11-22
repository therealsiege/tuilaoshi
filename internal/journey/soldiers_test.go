package journey

import (
	"math/rand"
	"strconv"
	"testing"
)

func TestSoldiersAreCompleteAndServeTheRoad(t *testing.T) {
	seen := map[string]bool{}
	for _, d := range content.Bestiary {
		seen[d.Name] = true
	}
	for _, s := range content.Soldiers {
		if seen[s.Name] {
			t.Errorf("soldier %s is listed twice or is also a demon", s.Name)
		}
		seen[s.Name] = true
		if s.English == "" || s.Origin == "" || s.Powers == "" || s.Art == "" {
			t.Errorf("soldier %s is missing a field", s.Name)
		}
		if RankOrder(s.Rank) < 0 {
			t.Errorf("soldier %s has rank %q", s.Name, s.Rank)
		}
		if s.From < 1 || s.To < s.From || s.To > len(content.Tribulations) {
			t.Errorf("soldier %s serves stages %d-%d", s.Name, s.From, s.To)
		}
		if n, err := strconv.Atoi(firstNumber.FindString(s.Chapter)); err != nil || n < 1 || n > 100 {
			t.Errorf("soldier %s has chapter %q", s.Name, s.Chapter)
		}
	}
	// Every stage on the road has soldiers; Vulture Peak and the way home don't.
	for n := 1; n <= 78; n++ {
		if len(SoldiersAt(n)) == 0 {
			t.Errorf("no soldiers serve stage %d", n)
		}
	}
}

func TestGauntletRisesInRank(t *testing.T) {
	rng := rand.New(rand.NewSource(1))
	for n := 1; n <= 81; n++ {
		g := Gauntlet(n, rng)
		if n > 78 {
			if len(g) != 0 {
				t.Errorf("stage %d has a gauntlet: %v", n, g)
			}
			continue
		}
		if len(g) == 0 || len(g) > gauntletSize {
			t.Fatalf("stage %d gauntlet has %d soldiers", n, len(g))
		}
		for i := 1; i < len(g); i++ {
			if RankOrder(g[i].Rank) < RankOrder(g[i-1].Rank) || g[i].Name == g[i-1].Name {
				t.Errorf("stage %d gauntlet out of order or repeated: %s then %s", n, g[i-1].Name, g[i].Name)
			}
		}
	}
}

func TestReinforceAddsTheStrongestSpare(t *testing.T) {
	rng := rand.New(rand.NewSource(2))
	g := Reinforce(26, Gauntlet(26, rng), rng) // Flat-Top: imps, elites and the old fox
	if len(g) != gauntletSize+1 {
		t.Fatalf("reinforced gauntlet has %d soldiers", len(g))
	}
	for i := 1; i < len(g); i++ {
		if RankOrder(g[i].Rank) < RankOrder(g[i-1].Rank) {
			t.Errorf("out of rank order: %s then %s", g[i-1].Name, g[i].Name)
		}
	}
	if g := Reinforce(80, nil, rng); len(g) != 0 {
		t.Errorf("a stage without soldiers was reinforced: %v", g)
	}
}
