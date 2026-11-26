package wiki

import (
	"fmt"
	"time"

	"github.com/therealsiege/tuilaoshi/internal/model"
)

// wordScheduleFile held review schedules for vocabulary words before the
// progress file; it is read once, when the progress file is made.
const wordScheduleFile = "word-schedule.yaml"

// wordSchedule is one card's schedule: an entry of the old schedule file,
// and a row of the schedules table.
type wordSchedule struct {
	Introduced     string `yaml:"introduced,omitempty"`
	LastReviewed   string `yaml:"last_reviewed,omitempty"`
	ReviewInterval int    `yaml:"review_interval"`
	NextReview     string `yaml:"next_review,omitempty"`
}

// readSchedules loads every card's saved schedule, by card id.
func (s *Store) readSchedules() (map[string]wordSchedule, error) {
	db, err := s.progress()
	if err != nil {
		return nil, err
	}
	rows, err := db.Query(`SELECT card_id, introduced, last_reviewed, review_interval, next_review FROM schedules`)
	if err != nil {
		return nil, fmt.Errorf("read schedules: %w", err)
	}
	defer rows.Close() //nolint:errcheck // read-only
	schedules := map[string]wordSchedule{}
	for rows.Next() {
		var id string
		var sch wordSchedule
		if err := rows.Scan(&id, &sch.Introduced, &sch.LastReviewed, &sch.ReviewInterval, &sch.NextReview); err != nil {
			return nil, err
		}
		schedules[id] = sch
	}
	return schedules, rows.Err()
}

// DrillableCards returns flashcards for every character, vocabulary, and
// grammar page. A card reviewed in the game carries its saved schedule;
// the rest start from their page's frontmatter.
func (s *Store) DrillableCards() ([]model.Card, error) {
	pages, err := s.ListPages()
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
	cards := CardsFromPages(drillable)

	schedules, err := s.readSchedules()
	if err != nil {
		return nil, err
	}
	for i := range cards {
		sch, ok := schedules[cards[i].ID]
		if !ok {
			continue
		}
		cards[i].Introduced = parseDate(sch.Introduced)
		cards[i].LastReviewed = parseDate(sch.LastReviewed)
		cards[i].ReviewInterval = sch.ReviewInterval
		cards[i].NextReview = parseDate(sch.NextReview)
	}
	return cards, nil
}

// RecordReview reschedules a card after a review and saves its schedule
// and the answer to the progress file.
func (s *Store) RecordReview(card *model.Card, difficulty model.DrillDifficulty, now time.Time) error {
	mem, err := s.memory(card.ID)
	if err != nil {
		return err
	}
	was := card.ReviewInterval
	updateSRS(card, difficulty, now, &mem, s.useFSRS())
	return s.saveReview(card, difficulty, was, mem, now)
}

// RecordForgotten starts a card over, due tomorrow: the learner marked it
// unknown (while reading, say) rather than missing it in a review. It is
// logged and counted as a miss, and its FSRS memory starts over too.
func (s *Store) RecordForgotten(card *model.Card, now time.Time) error {
	was := card.ReviewInterval
	card.ReviewInterval = 0
	var mem fsrsMemory
	updateSRS(card, model.DiffAgain, now, &mem, s.useFSRS())
	return s.saveReview(card, model.DiffAgain, was, mem, now)
}

// saveReview saves a card's new schedule and FSRS memory and logs the answer.
func (s *Store) saveReview(card *model.Card, difficulty model.DrillDifficulty, was int, mem fsrsMemory, now time.Time) error {
	db, err := s.progress()
	if err != nil {
		return err
	}
	lapse := 0
	if difficulty == model.DiffAgain {
		lapse = 1
	}
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck // a no-op after Commit
	if _, err := tx.Exec(`INSERT INTO schedules (card_id, introduced, last_reviewed, review_interval, next_review, reps, lapses, stability, difficulty)
		VALUES (?, ?, ?, ?, ?, 1, ?, ?, ?)
		ON CONFLICT(card_id) DO UPDATE SET introduced = excluded.introduced, last_reviewed = excluded.last_reviewed,
			review_interval = excluded.review_interval, next_review = excluded.next_review,
			reps = reps + 1, lapses = lapses + excluded.lapses,
			stability = excluded.stability, difficulty = excluded.difficulty`,
		card.ID, FormatDate(card.Introduced), FormatDate(card.LastReviewed), card.ReviewInterval, FormatDate(card.NextReview), lapse,
		mem.Stability, mem.Difficulty); err != nil {
		return fmt.Errorf("save schedule %s: %w", card.ID, err)
	}
	if _, err := tx.Exec(`INSERT INTO reviews (card_id, at, rating, interval_was, interval_now) VALUES (?, ?, ?, ?, ?)`,
		card.ID, now.Format(time.RFC3339), int(difficulty), was, card.ReviewInterval); err != nil {
		return fmt.Errorf("log review %s: %w", card.ID, err)
	}
	return tx.Commit()
}
