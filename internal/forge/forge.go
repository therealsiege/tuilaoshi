// Package forge holds the rules of the character forge (字炉): build a
// character the learner is still learning from its written components,
// picked out of a tray that mixes in look-alike parts. Decompositions come
// from Claude once per character and are cached in the wiki (forge.yaml).
package forge

import (
	"math/rand"
	"strings"
	"unicode/utf8"

	"github.com/therealsiege/tuilaoshi/internal/model"
)

// TraySize is how many components the tray offers.
const TraySize = 8

// hardInterval: characters under this review interval (days) are forged.
const hardInterval = 7

// Targets picks up to n character cards still being learned (a single
// character, under a week's interval), shuffled.
func Targets(cards []model.Card, n int, rng *rand.Rand) []model.Card {
	var out []model.Card
	for _, c := range cards {
		if utf8.RuneCountInString(c.Chinese) == 1 && c.ReviewInterval < hardInterval && strings.HasPrefix(c.ID, "characters/") {
			out = append(out, c)
		}
	}
	rng.Shuffle(len(out), func(i, j int) { out[i], out[j] = out[j], out[i] })
	return out[:min(n, len(out))]
}

// Forgeable reports whether a decomposition makes a puzzle: two or more
// parts, none of them the character itself.
func Forgeable(ch string, parts []string) bool {
	if len(parts) < 2 {
		return false
	}
	for _, p := range parts {
		if p == ch {
			return false
		}
	}
	return true
}

// Tray mixes a character's parts with decoys from other characters'
// parts, the same radical family first so the eye has to work.
func Tray(parts []string, pool []string, rng *rand.Rand) []string {
	have := map[string]bool{}
	tray := []string{}
	for _, p := range parts {
		if !have[p] {
			have[p] = true
			tray = append(tray, p)
		}
	}
	decoys := append([]string(nil), pool...)
	rng.Shuffle(len(decoys), func(i, j int) { decoys[i], decoys[j] = decoys[j], decoys[i] })
	for _, d := range decoys {
		if len(tray) == TraySize {
			break
		}
		if !have[d] {
			have[d] = true
			tray = append(tray, d)
		}
	}
	rng.Shuffle(len(tray), func(i, j int) { tray[i], tray[j] = tray[j], tray[i] })
	return tray
}

// Pool gathers every part seen in the cache, for decoys.
func Pool(comps map[string][]string) []string {
	seen := map[string]bool{}
	var pool []string
	for _, parts := range comps {
		for _, p := range parts {
			if !seen[p] {
				seen[p] = true
				pool = append(pool, p)
			}
		}
	}
	return pool
}

// Forged reports whether the picked parts are the character's, in any
// order (a part used twice, as in 森, must be picked as often).
func Forged(parts, picked []string) bool {
	if len(parts) != len(picked) {
		return false
	}
	count := map[string]int{}
	for _, p := range parts {
		count[p]++
	}
	for _, p := range picked {
		count[p]--
		if count[p] < 0 {
			return false
		}
	}
	return true
}

// Material is the material a forged character yields, by its radical's
// element: water pearls from 氵, fire crystals from 火, herbs from plants,
// venom from 虫, feathers from birds, hide from flesh and beasts, a rare
// demon core from precious things (金 王 玉), else bone.
func Material(radical string) string {
	r := []rune(strings.TrimSpace(radical))
	if len(r) == 0 {
		return "bone"
	}
	switch string(r[0]) {
	case "氵", "水", "冫", "雨":
		return "pearl"
	case "火", "灬", "日":
		return "fire"
	case "艹", "木", "禾", "竹", "米":
		return "herb"
	case "虫":
		return "venom"
	case "鸟", "隹", "羽", "风":
		return "feather"
	case "月", "犭", "牛", "马", "羊":
		return "hide"
	case "王", "玉", "金", "钅":
		return "core"
	}
	return "bone"
}
