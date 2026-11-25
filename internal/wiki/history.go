package wiki

import (
	"fmt"
	"time"
)

// What the review history tells: the words that keep slipping, what is
// coming due, how much is remembered, and when to hold back new words.

// HardWordLapses is how many misses make a word a hard word (a leech).
const HardWordLapses = 6

// Backlog limits: new frequency words wait while more than PauseDue
// studied words are due, or while the last few days' new words are still
// unseen (PauseUnseenDays of them).
const (
	PauseDue        = 100
	PauseUnseenDays = 3
)

// HardWord is a card missed HardWordLapses times or more.
type HardWord struct {
	CardID string
	Lapses int
}

// HardWords returns the hard words, most missed first.
func (s *Store) HardWords() ([]HardWord, error) {
	db, err := s.progress()
	if err != nil {
		return nil, err
	}
	rows, err := db.Query(`SELECT card_id, lapses FROM schedules WHERE lapses >= ? ORDER BY lapses DESC, card_id`, HardWordLapses)
	if err != nil {
		return nil, fmt.Errorf("hard words: %w", err)
	}
	defer rows.Close() //nolint:errcheck // read-only
	var words []HardWord
	for rows.Next() {
		var w HardWord
		if err := rows.Scan(&w.CardID, &w.Lapses); err != nil {
			return nil, err
		}
		words = append(words, w)
	}
	return words, rows.Err()
}

// DueForecast counts the studied words (reviewed in the game at least
// once) due on each of the next days: [0] is today with anything overdue.
func (s *Store) DueForecast(now time.Time, days int) ([]int, error) {
	db, err := s.progress()
	if err != nil {
		return nil, err
	}
	out := make([]int, days)
	rows, err := db.Query(`SELECT next_review FROM schedules WHERE next_review != ''`)
	if err != nil {
		return nil, fmt.Errorf("forecast: %w", err)
	}
	defer rows.Close() //nolint:errcheck // read-only
	today := now.Format(dateFmt)
	for rows.Next() {
		var next string
		if err := rows.Scan(&next); err != nil {
			return nil, err
		}
		if next <= today {
			out[0]++
			continue
		}
		for d := 1; d < days; d++ {
			if next == now.AddDate(0, 0, d).Format(dateFmt) {
				out[d]++
				break
			}
		}
	}
	return out, rows.Err()
}

// Retention is how many reviews in the last days were remembered.
func (s *Store) Retention(now time.Time, days int) (right, total int, err error) {
	db, err := s.progress()
	if err != nil {
		return 0, 0, err
	}
	since := now.AddDate(0, 0, -days).Format(time.RFC3339)
	err = db.QueryRow(`SELECT count(*), coalesce(sum(rating > 0), 0) FROM reviews WHERE at >= ?`, since).Scan(&total, &right)
	return right, total, err
}

// reviewed reports which cards have a saved schedule.
func (s *Store) reviewed() (map[string]bool, error) {
	schedules, err := s.readSchedules()
	seen := make(map[string]bool, len(schedules))
	for id, sch := range schedules {
		seen[id] = sch.LastReviewed != ""
	}
	return seen, err
}

// NewWordsPaused says why new frequency words are waiting, or "" when
// they aren't: too many studied words due, or too many new words unseen.
func (s *Store) NewWordsPaused(perDay int, now time.Time) (string, error) {
	due, err := s.DueForecast(now, 1)
	if err != nil {
		return "", err
	}
	if due[0] > PauseDue {
		return fmt.Sprintf("%d studied words are due", due[0]), nil
	}
	cards, err := s.DrillableCards()
	if err != nil {
		return "", err
	}
	seen, err := s.reviewed()
	if err != nil {
		return "", err
	}
	unseen := 0
	for _, c := range cards {
		if FrequencyPage(c.ID) && !seen[c.ID] {
			unseen++
		}
	}
	if perDay > 0 && unseen > PauseUnseenDays*perDay {
		return fmt.Sprintf("%d new words are still unseen", unseen), nil
	}
	return "", nil
}
