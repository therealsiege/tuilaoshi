package journey

import (
	"fmt"
	"sort"
	"strings"

	"github.com/therealsiege/tuilaoshi/internal/model"
)

// CardSet is a group of collectible cards: the enemies of one stretch of
// the road, or the characters sharing a radical. Completing a set pays its
// reward once, when claimed.
type CardSet struct {
	ID      string // "enemies-11" or "radical-氵", for recording the claim
	Name    string
	Members []string // enemy names or characters
	Reward  Price
}

// wideRange marks soldiers who serve the whole road; they belong to no region.
const wideRange = 20

// EnemySets groups the enemies by where they are met: the soldiers serving
// the same stretch of stages, with the demons first fought there.
func EnemySets() []CardSet {
	var sets []CardSet
	ranges := map[string]int{} // "from-to" -> index in sets
	type span struct{ from, to int }
	var spans []span
	for _, s := range content.Soldiers {
		if s.To-s.From > wideRange {
			continue
		}
		k := fmt.Sprintf("%d-%d", s.From, s.To)
		i, ok := ranges[k]
		if !ok {
			i = len(sets)
			ranges[k] = i
			spans = append(spans, span{s.From, s.To})
			name := fmt.Sprintf("Tribulations %d–%d", s.From, s.To)
			if t := TribulationAt(s.From); t != nil {
				name += " · " + t.Location
			}
			sets = append(sets, CardSet{ID: fmt.Sprintf("enemies-%d", s.From), Name: name, Reward: Price{Tongqian: 50, Yuanbao: 1}})
		}
		sets[i].Members = append(sets[i].Members, s.Name)
	}
	placed := map[string]bool{}
	for _, t := range content.Tribulations {
		if t.Demon == "" || placed[t.Demon] {
			continue
		}
		for i, sp := range spans {
			if sp.from <= t.Number && t.Number <= sp.to {
				sets[i].Members = append(sets[i].Members, t.Demon)
				placed[t.Demon] = true
				break
			}
		}
	}
	// Demons met where no soldiers serve form a set of their own.
	lone := CardSet{ID: "enemies-lone", Name: "Lone demons of the road", Reward: Price{Tongqian: 50, Yuanbao: 1}}
	for _, t := range content.Tribulations {
		if t.Demon != "" && !placed[t.Demon] && t.Number < len(content.Tribulations)-2 {
			lone.Members = append(lone.Members, t.Demon)
			placed[t.Demon] = true
		}
	}
	if len(lone.Members) > 0 {
		sets = append(sets, lone)
	}
	return sets
}

// RadicalSets groups the wiki's characters by radical, largest family first.
// A family pays a 元宝 and ten 铜钱 a character when all are mastered.
func RadicalSets(pages []model.Page) []CardSet {
	byRadical := map[string]*CardSet{}
	for ch, rad := range RadicalIndex(pages) {
		key := strings.Fields(rad)[0]
		s := byRadical[key]
		if s == nil {
			s = &CardSet{ID: "radical-" + key, Name: rad}
			byRadical[key] = s
		}
		s.Members = append(s.Members, ch)
	}
	sets := make([]CardSet, 0, len(byRadical))
	for _, s := range byRadical {
		sort.Strings(s.Members)
		s.Reward = Price{Tongqian: 10 * len(s.Members), Yuanbao: 1}
		sets = append(sets, *s)
	}
	sort.Slice(sets, func(i, j int) bool {
		if len(sets[i].Members) != len(sets[j].Members) {
			return len(sets[i].Members) > len(sets[j].Members)
		}
		return sets[i].ID < sets[j].ID
	})
	return sets
}

// MasteredCharacters is the set of characters whose pages count as mastered.
func MasteredCharacters(pages []model.Page) map[string]bool {
	got := map[string]bool{}
	for _, p := range pages {
		if p.Type == "character" && p.Character != "" && IsMastered(p) {
			got[p.Character] = true
		}
	}
	return got
}

// Have counts the members of a set found in got.
func (s CardSet) Have(got map[string]bool) int {
	n := 0
	for _, m := range s.Members {
		if got[m] {
			n++
		}
	}
	return n
}
