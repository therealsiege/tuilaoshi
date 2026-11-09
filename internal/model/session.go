package model

import "time"

// Session records a practice session (drill, chat, review).
type Session struct {
	StartedAt    time.Time
	EndedAt      time.Time
	Type         string // "drill", "chat", "review"
	CardsTotal   int
	CardsCorrect int
	Items        []SessionItem
}

// SessionItem records the result of a single drill interaction.
type SessionItem struct {
	CardID     string
	UserAnswer string
	Correct    bool
	Score      int             // 0-5 quality rating
	Difficulty DrillDifficulty // user-selected difficulty
	Feedback   string
}

// Accuracy returns the ratio of correct answers to total, or 0 if none.
func (s Session) Accuracy() float64 {
	if s.CardsTotal == 0 {
		return 0
	}
	return float64(s.CardsCorrect) / float64(s.CardsTotal)
}
