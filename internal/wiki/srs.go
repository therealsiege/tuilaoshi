package wiki

import (
	"time"

	"github.com/cj/mandarin-wiki/internal/model"
)

const initialInterval = 1 // days

// Difficulty multipliers (SM-2 inspired).
var difficultyMultipliers = map[model.DrillDifficulty]float64{
	model.DiffAgain: 0,   // reset to 1 day
	model.DiffHard:  1.2,
	model.DiffGood:  2.5,
	model.DiffEasy:  3.5,
}

// UpdateSRS updates a card's SRS fields after a drill response.
// Uses difficulty-graded intervals: Again resets, Hard/Good/Easy multiply.
func UpdateSRS(card *model.Card, difficulty model.DrillDifficulty, now time.Time) {
	card.LastReviewed = now

	mult := difficultyMultipliers[difficulty]
	if difficulty == model.DiffAgain || mult == 0 {
		card.ReviewInterval = initialInterval
	} else {
		if card.ReviewInterval == 0 {
			card.ReviewInterval = initialInterval
		} else {
			card.ReviewInterval = int(float64(card.ReviewInterval) * mult)
		}
		if card.ReviewInterval < 1 {
			card.ReviewInterval = 1
		}
	}

	card.NextReview = now.AddDate(0, 0, card.ReviewInterval)

	if card.Introduced.IsZero() {
		card.Introduced = now
	}
}

// UpdateSRSBool is a convenience wrapper for boolean correct/incorrect.
func UpdateSRSBool(card *model.Card, correct bool, now time.Time) {
	if correct {
		UpdateSRS(card, model.DiffGood, now)
	} else {
		UpdateSRS(card, model.DiffAgain, now)
	}
}

// DueItems scans all drillable pages and returns cards due for review.
func DueItems(store *Store, now time.Time) ([]model.Card, error) {
	pages, err := store.ListPages()
	if err != nil {
		return nil, err
	}

	var drillable []model.Page
	for _, p := range pages {
		switch p.Type {
		case "character", "vocabulary", "grammar":
			drillable = append(drillable, p)
		}
	}

	allCards := CardsFromPages(drillable)
	var due []model.Card
	for _, c := range allCards {
		if c.IsDue(now) {
			due = append(due, c)
		}
	}
	return due, nil
}

// ApplySRSToPage updates a page's SRS frontmatter fields from a card.
func ApplySRSToPage(page *model.Page, card model.Card) {
	page.LastReviewed = card.LastReviewed
	page.ReviewInterval = card.ReviewInterval
	page.NextReview = card.NextReview
	if page.Introduced.IsZero() {
		page.Introduced = card.Introduced
	}
}
