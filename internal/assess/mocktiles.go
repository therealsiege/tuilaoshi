package assess

import (
	"math/rand"
	"strings"
	"unicode/utf8"

	"github.com/therealsiege/tuilaoshi/internal/segment"
)

// particles join the word before them as one tile, as 去了 or 我的.
const particles = "了的着过吗呢吧啊"

// tiles splits a short sentence into word tiles to put back in order. Only
// sentences the wiki's words split cleanly are used: three to six tiles,
// with few lone characters and never two side by side.
func (p *pools) tiles(s sentence, rng *rand.Rand) (Item, bool) {
	core := strings.TrimRightFunc(s.zh, func(r rune) bool { return strings.ContainsRune("。！？!?.", r) })
	if p.used[s.zh] || !allHan(core) {
		return Item{}, false
	}
	if n := utf8.RuneCountInString(core); n < 5 || n > 14 {
		return Item{}, false
	}
	parts := splitTiles(core, p.known)
	if len(parts) < 3 || len(parts) > maxTiles {
		return Item{}, false
	}
	singles := 0
	for i, t := range parts {
		if utf8.RuneCountInString(t) == 1 {
			singles++
			if i > 0 && utf8.RuneCountInString(parts[i-1]) == 1 {
				return Item{}, false
			}
		}
	}
	if singles > 2 {
		return Item{}, false
	}
	shuffled := append([]string{}, parts...)
	for try := 0; try < 10 && strings.Join(shuffled, "") == core; try++ {
		shuffle(rng, shuffled)
	}
	if strings.Join(shuffled, "") == core {
		return Item{}, false
	}
	p.used[s.zh] = true
	return Item{Kind: Tiles, Tiles: shuffled, Answer: core, English: s.en,
		Prompt: "Put the words in order to make a sentence."}, true
}

// splitTiles segments a sentence by the known words, with particles
// joined to the word before them.
func splitTiles(core string, known map[string]bool) []string {
	var parts []string
	for _, t := range segment.Split(core, known) {
		if len(parts) > 0 && utf8.RuneCountInString(t.Text) == 1 && strings.Contains(particles, t.Text) {
			parts[len(parts)-1] += t.Text
			continue
		}
		parts = append(parts, t.Text)
	}
	return parts
}
