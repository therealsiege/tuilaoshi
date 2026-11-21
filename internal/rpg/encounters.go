package rpg

import (
	"math/rand"

	"github.com/therealsiege/tuilaoshi/internal/journey"
	"github.com/therealsiege/tuilaoshi/internal/model"
)

// openChapters of the novel's 100 are open before any progress, so a new
// learner still meets a few demons.
const openChapters = 20

// ReachedDemons returns the demons whose chapter the pilgrimage has reached:
// the open chapters plus the rest of the novel in proportion to journey
// progress (percent, 0-100).
func ReachedDemons(percent float64) []journey.BestiaryEntry {
	reached := openChapters + int(percent*(100-openChapters)/100)
	var demons []journey.BestiaryEntry
	for _, d := range journey.Bestiary() {
		if parseChapterNumber(d.Chapter) <= reached {
			demons = append(demons, d)
		}
	}
	return demons
}

// NewRandomBattle creates a battle against a demon the pilgrimage has reached.
func NewRandomBattle(cards []model.Card, masteredCount int, percent float64) *Battle {
	demons := ReachedDemons(percent)
	if len(demons) == 0 {
		return nil
	}
	demon := demons[rand.Intn(len(demons))]
	return NewBattle(cards, demon, "", masteredCount)
}

const (
	// Bosses last longer and hit harder than demons met on the road.
	bossHPMultiplier = 2
	bossAttackBonus  = 5
	// weakInterval: words on a review interval shorter than this are still shaky.
	weakInterval = 7
	// minBossDeck is the fewest weak words a boss quizzes before it falls
	// back to the whole deck.
	minBossDeck = 8
)

// NewBossBattle creates a fight against a journey boss.
func NewBossBattle(cards []model.Card, boss journey.BestiaryEntry, masteredCount int) *Battle {
	b := NewBattle(cards, boss, "", masteredCount)
	b.Enemy.HP *= bossHPMultiplier
	b.Enemy.MaxHP = b.Enemy.HP
	b.Enemy.Attack += bossAttackBonus
	return b
}

// WeakCards returns the usable cards still being learned, so a boss tests
// what the learner finds hardest. With too few of those it returns every
// usable card.
func WeakCards(cards []model.Card) []model.Card {
	usable := filterUsableCards(cards)
	var weak []model.Card
	for _, c := range usable {
		if c.ReviewInterval < weakInterval {
			weak = append(weak, c)
		}
	}
	if len(weak) < minBossDeck {
		return usable
	}
	return weak
}
