package rpg

import (
	"math/rand"
	"strings"

	"github.com/therealsiege/tuilaoshi/internal/model"
	"github.com/therealsiege/tuilaoshi/internal/segment"
	"github.com/therealsiege/tuilaoshi/internal/wiki"
)

// An "order" sentence must split into this many words: enough to be a
// puzzle, few enough for the 1-9 keys.
const (
	minPieces = 3
	maxPieces = 9
)

// orderQuestion asks the player to rebuild one of card's example sentences
// from its shuffled words, or returns nil when none fits. The sentence is
// split by the offline segmenter; grading compares the rebuilt text, so a
// word split in two only makes one more piece.
func (b *Battle) orderQuestion(card model.Card, d deck) *CombatQuestion {
	for _, i := range rand.Perm(len(card.Examples)) {
		zh, en, ok := wiki.ParseExample(card.Examples[i])
		if !ok {
			continue
		}
		var pieces []string
		for _, t := range segment.Split(zh, d.known) {
			if t.Word {
				pieces = append(pieces, t.Text)
			}
		}
		if len(pieces) < minPieces || len(pieces) > maxPieces {
			continue
		}
		target := strings.Join(pieces, "")
		shuffled := append([]string{}, pieces...)
		for try := 0; try < 10 && strings.Join(shuffled, "") == target; try++ {
			rand.Shuffle(len(shuffled), func(i, j int) { shuffled[i], shuffled[j] = shuffled[j], shuffled[i] })
		}
		if strings.Join(shuffled, "") == target {
			continue // every order reads the same, e.g. repeated words
		}
		return &CombatQuestion{
			Prompt:   "Put the words in order: " + en,
			Category: "order",
			Damage:   b.damage() + writeBonus,
			Chinese:  target,
			English:  en,
			Card:     card,
			Accepted: []string{target},
			Pieces:   shuffled,
		}
	}
	return nil
}
