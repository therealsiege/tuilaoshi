package wiki

import (
	"time"

	"github.com/therealsiege/tuilaoshi/internal/model"
)

const initialInterval = 1 // days

// Difficulty multipliers (SM-2 inspired).
var difficultyMultipliers = map[model.DrillDifficulty]float64{
	model.DiffAgain: 0, // reset to 1 day
	model.DiffHard:  1.2,
	model.DiffGood:  2.5,
	model.DiffEasy:  3.5,
}

// updateSRS updates a card's SRS fields after a review. Callers outside this
// package use Store.RecordReview, which also saves the result.
// Uses difficulty-graded intervals: Again resets, Hard/Good/Easy multiply,
// and any passing review grows the interval by at least a day.
func updateSRS(card *model.Card, difficulty model.DrillDifficulty, now time.Time) {
	card.LastReviewed = now

	mult := difficultyMultipliers[difficulty]
	if difficulty == model.DiffAgain || mult == 0 {
		card.ReviewInterval = initialInterval
	} else if card.ReviewInterval == 0 {
		card.ReviewInterval = initialInterval
	} else {
		next := int(float64(card.ReviewInterval) * mult)
		if next <= card.ReviewInterval {
			next = card.ReviewInterval + 1
		}
		card.ReviewInterval = next
	}

	card.NextReview = now.AddDate(0, 0, card.ReviewInterval)

	if card.Introduced.IsZero() {
		card.Introduced = now
	}
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

// ApplySRSToPage updates a page's SRS frontmatter fields from a card.
func applySRSToPage(page *model.Page, card model.Card) {
	page.LastReviewed = card.LastReviewed
	page.ReviewInterval = card.ReviewInterval
	page.NextReview = card.NextReview
	if page.Introduced.IsZero() {
		page.Introduced = card.Introduced
	}
}
