package journey

import (
	"math/rand"
	"slices"
)

// Soldier ranks, weakest first; rank sets a soldier's strength and reward.
const (
	RankMinion  = "minion"
	RankElite   = "elite"
	RankCaptain = "captain"
)

var ranks = []string{RankMinion, RankElite, RankCaptain}

// RankOrder places a rank among the soldier ranks (0 is the weakest), or
// returns -1 for anything else.
func RankOrder(rank string) int { return slices.Index(ranks, rank) }

// Soldier is a rank-and-file demon met on stages From..To, before the
// stage's own demon and in practice battles.
type Soldier struct {
	BestiaryEntry `yaml:",inline"`
	Rank          string `yaml:"rank"`
	From          int    `yaml:"from"`
	To            int    `yaml:"to"`
}

// rankArt stands in for a soldier's portrait when images can't be shown.
var rankArt = map[string]string{
	RankMinion:  "   ,_,\n  (o.o)\n  /|=|\\\n   / \\",
	RankElite:   "  \\^^/\n  (o_o)\n --|#|--\n   / \\",
	RankCaptain: " /\\^^/\\\n ( >_< )\n/-|##|-\\\n  _/ \\_",
}

// Soldiers returns the whole roster.
func Soldiers() []Soldier { return content.Soldiers }

// SoldiersAt returns the soldiers serving at stage n, weakest rank first.
func SoldiersAt(n int) []Soldier {
	var at []Soldier
	for _, s := range content.Soldiers {
		if s.From <= n && n <= s.To {
			at = append(at, s)
		}
	}
	slices.SortStableFunc(at, func(a, b Soldier) int { return RankOrder(a.Rank) - RankOrder(b.Rank) })
	return at
}

// gauntletSize is how many soldiers stand before a stage's demon.
const gauntletSize = 2

// Gauntlet picks the soldiers fought before stage n's demon, weakest
// first: one of the lowest rank and one of a higher rank when the stage
// has them, otherwise two different soldiers. Stages without soldiers
// (the last, at Vulture Peak) go straight to their opponent.
func Gauntlet(n int, rng *rand.Rand) []Soldier {
	at := SoldiersAt(n)
	if len(at) <= gauntletSize {
		return at
	}
	// The patrol imp serves everywhere; prefer the stage's own soldiers.
	local := slices.DeleteFunc(slices.Clone(at), func(s Soldier) bool { return s.To-s.From > 20 })
	if len(local) >= gauntletSize {
		at = local
	}
	lowest := RankOrder(at[0].Rank)
	var weak, strong []Soldier
	for _, s := range at {
		if RankOrder(s.Rank) == lowest {
			weak = append(weak, s)
		} else {
			strong = append(strong, s)
		}
	}
	first := weak[rng.Intn(len(weak))]
	if len(strong) > 0 {
		return []Soldier{first, strong[rng.Intn(len(strong))]}
	}
	rest := slices.DeleteFunc(slices.Clone(weak), func(s Soldier) bool { return s.Name == first.Name })
	return []Soldier{first, rest[rng.Intn(len(rest))]}
}

// EnemyLookup finds a demon or soldier by Chinese name, or returns nil.
func EnemyLookup(name string) *BestiaryEntry {
	if d := BestiaryLookup(name); d != nil {
		return d
	}
	for i := range content.Soldiers {
		if content.Soldiers[i].Name == name {
			return &content.Soldiers[i].BestiaryEntry
		}
	}
	return nil
}

// Reinforce adds to a gauntlet one more of stage n's soldiers, of the
// highest rank not already in it, keeping the order weakest first. It
// returns the gauntlet unchanged when the stage has no one left to send.
func Reinforce(n int, gauntlet []Soldier, rng *rand.Rand) []Soldier {
	var spare []Soldier
	for _, s := range SoldiersAt(n) {
		if !slices.ContainsFunc(gauntlet, func(g Soldier) bool { return g.Name == s.Name }) {
			spare = append(spare, s)
		}
	}
	if len(spare) == 0 {
		return gauntlet
	}
	top := RankOrder(spare[len(spare)-1].Rank)
	strongest := slices.DeleteFunc(spare, func(s Soldier) bool { return RankOrder(s.Rank) != top })
	out := append(slices.Clone(gauntlet), strongest[rng.Intn(len(strongest))])
	slices.SortStableFunc(out, func(a, b Soldier) int { return RankOrder(a.Rank) - RankOrder(b.Rank) })
	return out
}
