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
// mem, when given, is the card's FSRS memory: it is brought up to date on
// every review, and with useFSRS its interval schedules the card. Without
// it, Again keeps a quarter of the interval (at least a day), Hard/Good/Easy
// multiply it, and any passing review grows it by at least a day.
// Either way no interval passes a year and long ones spread.
func updateSRS(card *model.Card, difficulty model.DrillDifficulty, now time.Time, mem *fsrsMemory, useFSRS bool) {
	if mem != nil {
		*mem = mem.next(difficulty, card.ReviewInterval, daysBetween(card.LastReviewed, now))
	}
	card.LastReviewed = now

	if useFSRS && mem != nil {
		card.ReviewInterval = spread(min(mem.interval(), maxInterval), card.ID, now)
	} else {
		card.ReviewInterval = classicInterval(card.ReviewInterval, difficulty, card.ID, now)
	}

	card.NextReview = now.AddDate(0, 0, card.ReviewInterval)

	if card.Introduced.IsZero() {
		card.Introduced = now
	}
}

// classicInterval is the next interval under the fixed multipliers.
func classicInterval(interval int, difficulty model.DrillDifficulty, id string, now time.Time) int {
	mult := difficultyMultipliers[difficulty]
	switch {
	case difficulty == model.DiffAgain || mult == 0:
		return max(interval/lapseDivisor, initialInterval)
	case interval == 0:
		return initialInterval
	}
	next := int(float64(interval) * mult)
	if next <= interval {
		next = interval + 1
	}
	return spread(min(next, maxInterval), id, now)
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
