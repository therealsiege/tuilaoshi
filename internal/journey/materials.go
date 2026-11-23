package journey

import "slices"

// Material is something an enemy drops, crafted into gear at the shrine.
type Material struct {
	ID      string `yaml:"id"`
	Name    string `yaml:"name"`
	English string `yaml:"english"`
	Desc    string `yaml:"desc"`
}

// Materials returns every material.
func Materials() []Material { return content.Materials }

// MaterialByID finds a material, or returns nil.
func MaterialByID(id string) *Material {
	for i := range content.Materials {
		if content.Materials[i].ID == id {
			return &content.Materials[i]
		}
	}
	return nil
}

// Drop is what an enemy leaves when beaten: by its nature first (a burning
// foe a fire crystal, a stinging one a venom sac, a blinding one a spirit
// feather), else by where it is met (river pearls, palace and forest herbs),
// else hide from minions and bone from the rest. Strong is true for a
// captain, demon or boss, which can leave a demon core besides.
func Drop(e BestiaryEntry, rank, terrain string) (first string, core bool) {
	switch {
	case slices.Contains(e.Abilities, "burn"):
		first = "fire"
	case slices.Contains(e.Abilities, "venom"):
		first = "venom"
	case slices.Contains(e.Abilities, "blind"):
		first = "feather"
	case terrain == "river":
		first = "pearl"
	case terrain == "palace" || terrain == "forest":
		first = "herb"
	case rank == RankMinion:
		first = "hide"
	default:
		first = "bone"
	}
	return first, rank != RankMinion && rank != RankElite
}
