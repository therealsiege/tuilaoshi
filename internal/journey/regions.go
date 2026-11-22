package journey

import "math/rand"

// Region is a stretch of the road walked as one tile map (see regionmap.go).
type Region struct {
	ID      string `yaml:"id"`
	Name    string `yaml:"name"`
	English string `yaml:"english"`
	From    int    `yaml:"from"`
	To      int    `yaml:"to"`
	Terrain string `yaml:"terrain"`
}

// Regions returns the regions in road order.
func Regions() []Region { return content.Regions }

// RegionOf returns the index of the region holding stage n.
func RegionOf(n int) int {
	for i, r := range content.Regions {
		if r.From <= n && n <= r.To {
			return i
		}
	}
	return len(content.Regions) - 1
}

// RegionByID returns a region's index, or -1.
func RegionByID(id string) int {
	for i, r := range content.Regions {
		if r.ID == id {
			return i
		}
	}
	return -1
}

// TownsIn returns the towns reached on a region's stages, in road order.
func (r Region) TownsIn() []Town {
	var towns []Town
	for _, t := range content.Towns {
		if r.From <= t.Stage && t.Stage <= r.To {
			towns = append(towns, t)
		}
	}
	return towns
}

// rankWeight is how often each rank turns up on wild ground.
var rankWeight = map[string]int{RankMinion: 6, RankElite: 3, RankCaptain: 1}

// Encounter picks a soldier met on the region's wild ground: one serving
// its stages, minions most often and captains rarely. It reports false
// where no soldiers serve.
func (r Region) Encounter(rng *rand.Rand) (Soldier, bool) {
	var pool []Soldier
	total := 0
	for _, s := range content.Soldiers {
		if s.From <= r.To && r.From <= s.To {
			pool = append(pool, s)
			total += rankWeight[s.Rank]
		}
	}
	if total == 0 {
		return Soldier{}, false
	}
	n := rng.Intn(total)
	for _, s := range pool {
		if n -= rankWeight[s.Rank]; n < 0 {
			return s, true
		}
	}
	return pool[len(pool)-1], true
}
