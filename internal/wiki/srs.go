package wiki

import (
	"hash/fnv"
	"time"

	"github.com/therealsiege/tuilaoshi/internal/model"
)

const initialInterval = 1 // days

// Interval limits: a missed card keeps a quarter of its interval rather
// than starting over, no interval passes a year, and intervals of three
// weeks or more spread by up to 5% either way, so a big day's reviews
// don't all come due on one later day.
const (
	lapseDivisor = 4
	maxInterval  = 365
	spreadFrom   = 20
)

// Difficulty multipliers (SM-2 inspired).
var difficultyMultipliers = map[model.DrillDifficulty]float64{
	model.DiffAgain: 0, // keep a quarter of the interval
	model.DiffHard:  1.2,
	model.DiffGood:  2.5,
	model.DiffEasy:  3.5,
}

// updateSRS updates a card's SRS fields after a review. Callers outside this
// package use Store.RecordReview, which also saves the result.
// Again keeps a quarter of the interval (at least a day), Hard/Good/Easy
// multiply it, and any passing review grows it by at least a day.
func updateSRS(card *model.Card, difficulty model.DrillDifficulty, now time.Time) {
	card.LastReviewed = now

	mult := difficultyMultipliers[difficulty]
	switch {
	case difficulty == model.DiffAgain || mult == 0:
		card.ReviewInterval = max(card.ReviewInterval/lapseDivisor, initialInterval)
	case card.ReviewInterval == 0:
		card.ReviewInterval = initialInterval
	default:
		next := int(float64(card.ReviewInterval) * mult)
		if next <= card.ReviewInterval {
			next = card.ReviewInterval + 1
		}
		card.ReviewInterval = spread(min(next, maxInterval), card.ID, now)
	}

	card.NextReview = now.AddDate(0, 0, card.ReviewInterval)

	if card.Introduced.IsZero() {
		card.Introduced = now
	}
}

// spread moves a long interval by up to 5% either way, the same for a
// card on a given day.
func spread(interval int, id string, now time.Time) int {
	if interval < spreadFrom {
		return interval
	}
	width := interval / spreadFrom
	h := fnv.New32a()
	_, _ = h.Write([]byte(id + now.Format(dateFmt))) //nolint:errcheck // hash writes don't fail
	return min(interval+int(h.Sum32()%uint32(2*width+1))-width, maxInterval)
}

// DueItems scans all drillable pages and returns cards due for review.
func DueItems(store *Store, now time.Time) ([]model.Card, error) {
	allCards, err := store.DrillableCards()
	if err != nil {
		return nil, err
	}

	var due []model.Card
	for _, c := range allCards {
		if c.IsDue(now) {
			due = append(due, c)
		}
	}
	return due, nil
}
