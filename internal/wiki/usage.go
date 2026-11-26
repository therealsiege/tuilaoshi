package wiki

import (
	"fmt"
	"time"
)

// Claude's use is counted by day and model in the usage table, so My
// progress can show it and settings can cap it.

// Usage is a day's calls to Claude and their tokens.
type Usage struct {
	Calls         int
	Input, Output int64
}

// RecordUsage counts one call to Claude.
func (s *Store) RecordUsage(model string, input, output int64, now time.Time) error {
	db, err := s.progress()
	if err != nil {
		return err
	}
	_, err = db.Exec(`INSERT INTO usage (day, model, calls, input, output) VALUES (?, ?, 1, ?, ?)
		ON CONFLICT(day, model) DO UPDATE SET calls = calls + 1, input = input + excluded.input, output = output + excluded.output`,
		now.Format(dateFmt), model, input, output)
	if err != nil {
		return fmt.Errorf("record usage: %w", err)
	}
	return nil
}

// UsageOn is a day's use of Claude, all models together.
func (s *Store) UsageOn(now time.Time) (Usage, error) {
	db, err := s.progress()
	if err != nil {
		return Usage{}, err
	}
	var u Usage
	err = db.QueryRow(`SELECT coalesce(sum(calls), 0), coalesce(sum(input), 0), coalesce(sum(output), 0) FROM usage WHERE day = ?`,
		now.Format(dateFmt)).Scan(&u.Calls, &u.Input, &u.Output)
	return u, err
}
