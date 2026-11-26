package wiki

import (
	"fmt"

	"github.com/therealsiege/tuilaoshi/internal/model"
)

// The spider's web (盘丝洞) on the progress screen maps the learner's
// studied words; this file reads them with how often each has slipped.

// StudiedWord is a word the learner has reviewed at least once, with its
// card (and schedule) and how many times it was answered "again".
type StudiedWord struct {
	Card   model.Card
	Lapses int
}

// StudiedWords returns the character and vocabulary cards reviewed in the
// game at least once (a review saved in the progress file; a page's
// starting review dates don't count). Grammar cards are left out: they are
// patterns, not words.
func (s *Store) StudiedWords() ([]StudiedWord, error) {
	cards, err := s.DrillableCards()
	if err != nil {
		return nil, err
	}
	seen, err := s.reviewed()
	if err != nil {
		return nil, err
	}
	lapses, err := s.lapses()
	if err != nil {
		return nil, err
	}
	var words []StudiedWord
	for _, c := range cards {
		if c.Type == model.CardGrammar || !seen[c.ID] || c.Chinese == "" {
			continue
		}
		words = append(words, StudiedWord{Card: c, Lapses: lapses[c.ID]})
	}
	return words, nil
}

// lapses reads how many times each card was missed, for cards missed at all.
func (s *Store) lapses() (map[string]int, error) {
	db, err := s.progress()
	if err != nil {
		return nil, err
	}
	rows, err := db.Query(`SELECT card_id, lapses FROM schedules WHERE lapses > 0`)
	if err != nil {
		return nil, fmt.Errorf("read lapses: %w", err)
	}
	defer rows.Close() //nolint:errcheck // read-only
	out := map[string]int{}
	for rows.Next() {
		var id string
		var n int
		if err := rows.Scan(&id, &n); err != nil {
			return nil, err
		}
		out[id] = n
	}
	return out, rows.Err()
}
