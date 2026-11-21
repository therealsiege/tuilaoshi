package rpg

import (
	"math/rand"
	"strings"
	"time"

	"github.com/therealsiege/tuilaoshi/internal/model"
)

// clock is the time questions are picked at; tests replace it.
var clock = time.Now

// Study weights: a battle mostly asks about words that need practice.
const (
	weightDue     = 8 // reviewed before and due again
	weightShaky   = 4 // reviewed, but on a short interval
	weightNew     = 3 // never reviewed
	weightSettled = 1 // known well and not yet due
)

// studyWeight says how much a card needs practice now.
func studyWeight(c model.Card, now time.Time) int {
	switch {
	case c.LastReviewed.IsZero():
		return weightNew
	case c.IsDue(now):
		return weightDue
	case c.ReviewInterval < weakInterval:
		return weightShaky
	}
	return weightSettled
}

// pickTarget draws a card to ask about, weighted by how much it needs practice.
func pickTarget(cards []model.Card, now time.Time) model.Card {
	total := 0
	for _, c := range cards {
		total += studyWeight(c, now)
	}
	n := rand.Intn(total)
	for _, c := range cards {
		if n -= studyWeight(c, now); n < 0 {
			return c
		}
	}
	return cards[len(cards)-1]
}

// meaningHints explains a word through its characters' radicals, from the
// battle's radical index ("火 (fire)" style). A character that is its own
// radical adds nothing and is skipped.
func (b *Battle) meaningHints(word string) []string {
	var hints []string
	for _, r := range word {
		ch := string(r)
		rad := b.Radicals[ch]
		if rad == "" || strings.HasPrefix(rad, ch) {
			continue
		}
		hints = append(hints, ch+" ← "+rad)
	}
	return hints
}
