package model

import "time"

// CardType identifies the source category of a flashcard.
type CardType string

const (
	CardCharacter  CardType = "character"
	CardVocabulary CardType = "vocabulary"
	CardGrammar    CardType = "grammar"
)

// DrillMode determines how the card is presented.
type DrillMode string

const (
	ModeRecognition DrillMode = "recognition" // show Chinese → recall meaning
	ModeProduction  DrillMode = "production"  // show English → produce Chinese
	ModeSentence    DrillMode = "sentence"    // fill in missing word
	ModeListening   DrillMode = "listening"   // hear TTS → type Chinese
)

// DrillDifficulty represents the user's self-assessed difficulty after a drill.
type DrillDifficulty int

const (
	DiffAgain DrillDifficulty = iota // failed, reset interval
	DiffHard                         // correct but difficult
	DiffGood                         // correct, normal
	DiffEasy                         // effortless
)

// Card is a single flashcard derived from a wiki page.
type Card struct {
	ID       string   // file path relative to wiki/
	Type     CardType
	Chinese  string
	Pinyin   string
	English  string
	Examples []string // example sentences

	// SRS fields
	Introduced     time.Time
	LastReviewed    time.Time
	ReviewInterval int // days
	NextReview     time.Time
}

// IsDue reports whether the card is due for review on or before the given date.
func (c Card) IsDue(now time.Time) bool {
	if c.NextReview.IsZero() {
		return true
	}
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	due := time.Date(c.NextReview.Year(), c.NextReview.Month(), c.NextReview.Day(), 0, 0, 0, 0, c.NextReview.Location())
	return !due.After(today)
}

// DrillResult holds the Claude evaluation of a single drill answer.
type DrillResult struct {
	Correct    bool
	Feedback   string          // Claude's feedback (Chinese + English)
	Score      int             // 0-5 quality rating
	Difficulty DrillDifficulty // user-selected difficulty grade
}
