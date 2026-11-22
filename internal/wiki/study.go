package wiki

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// studyFile keeps the day's study bookkeeping, shared by every journey:
// which words have already paid today, the daily bonus, bounty progress
// and streak shields. Study belongs to the learner, not to one journey.
const studyFile = "study.yaml"

type studyRecord struct {
	DailyPaid    string         `yaml:"daily_paid,omitempty"`    // the day the daily bonus was last paid
	PaidDay      string         `yaml:"paid_day,omitempty"`      // the day PaidCards covers
	PaidCards    []string       `yaml:"paid_cards,omitempty"`    // cards already paid for that day
	TallyDay     string         `yaml:"tally_day,omitempty"`     // the day Tally counts
	Tally        map[string]int `yaml:"tally,omitempty"`         // fresh progress toward today's bounties
	BountiesPaid []string       `yaml:"bounties_paid,omitempty"` // today's bounties already paid
	Shields      int            `yaml:"shields,omitempty"`       // streak shields held
	Shielded     []string       `yaml:"shielded,omitempty"`      // missed days a shield covered
}

func (s *Store) readStudy() (studyRecord, error) {
	var rec studyRecord
	data, err := os.ReadFile(filepath.Join(s.Dir, studyFile))
	if os.IsNotExist(err) {
		return rec, nil
	}
	if err != nil {
		return rec, fmt.Errorf("read study progress: %w", err)
	}
	if err := yaml.Unmarshal(data, &rec); err != nil {
		return rec, fmt.Errorf("parse %s: %w", studyFile, err)
	}
	return rec, nil
}

func (s *Store) writeStudy(rec studyRecord) error {
	data, err := yaml.Marshal(rec)
	if err != nil {
		return fmt.Errorf("encode study progress: %w", err)
	}
	return os.WriteFile(filepath.Join(s.Dir, studyFile), data, 0644)
}
