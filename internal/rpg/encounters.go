package rpg

import (
	"math/rand"

	"github.com/therealsiege/tuilaoshi/internal/journey"
	"github.com/therealsiege/tuilaoshi/internal/model"
)

// openChapters of the novel's 100 are open before any progress, so a new
// learner's practice battles still meet a few demons.
const openChapters = 20

// ReachedDemons returns the demons met on stages up to current, plus those
// from the novel's open chapters, for practice battles.
func ReachedDemons(stages []journey.Stage, current int) []journey.BestiaryEntry {
	seen := map[string]bool{}
	var demons []journey.BestiaryEntry
	add := func(d journey.BestiaryEntry) {
		if !seen[d.Name] {
			seen[d.Name] = true
			demons = append(demons, d)
		}
	}
	for _, d := range journey.Bestiary() {
		if parseChapterNumber(d.Chapter) <= openChapters {
			add(d)
		}
	}
	for _, s := range stages {
		if s.Number <= current {
			add(s.Opponent)
		}
	}
	return demons
}

// NewRandomBattle creates a practice battle against one of demons.
func NewRandomBattle(cards []model.Card, masteredCount int, demons []journey.BestiaryEntry) *Battle {
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
