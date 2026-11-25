package wiki

import (
	"hash/fnv"
	"time"

	"github.com/therealsiege/tuilaoshi/internal/model"
)

// After the level check, the words at or below the learner's level can be
// marked known: rather than all being due today, they get a month's
// interval and come back spread over the next two months, so they are
// kept fresh without burying the words still to learn.

// Known words come back between these many days from now.
const (
	knownInterval = 30
	knownFrom     = 14
	knownUntil    = 60
)

// MarkKnown gives cards never reviewed in the game a known word's
// schedule and returns how many it changed; reviewed cards keep theirs.
// It also records that the level check was taken.
func (s *Store) MarkKnown(cards []model.Card, now time.Time) (int, error) {
	seen, err := s.reviewed()
	if err != nil {
		return 0, err
	}
	db, err := s.progress()
	if err != nil {
		return 0, err
	}
	tx, err := db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback() //nolint:errcheck // a no-op after Commit
	n := 0
	for _, c := range cards {
		if seen[c.ID] {
			continue
		}
		h := fnv.New32a()
		_, _ = h.Write([]byte(c.ID)) //nolint:errcheck // hash writes don't fail
		next := now.AddDate(0, 0, knownFrom+int(h.Sum32()%(knownUntil-knownFrom+1)))
		if _, err := tx.Exec(`INSERT OR REPLACE INTO schedules (card_id, introduced, last_reviewed, review_interval, next_review) VALUES (?, ?, '', ?, ?)`,
			c.ID, FormatDate(now), knownInterval, FormatDate(next)); err != nil {
			return 0, err
		}
		n++
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return n, s.MarkAssessed(now)
}

// MarkAssessed records that the level check was taken.
func (s *Store) MarkAssessed(now time.Time) error {
	rec, err := s.readStudy()
	if err != nil {
		return err
	}
	rec.Assessed = now.Format(dateFmt)
	return s.writeStudy(rec)
}

// Assessed reports whether the level check has been taken.
func (s *Store) Assessed() (bool, error) {
	rec, err := s.readStudy()
	return rec.Assessed != "", err
}
