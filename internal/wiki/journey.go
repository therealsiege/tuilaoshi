package wiki

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"time"

	"gopkg.in/yaml.v3"
)

// journeyFile records the date each journey stage (tribulation 1-81) was
// cleared by winning its battle.
const journeyFile = "journey.yaml"

type journeyRecord struct {
	Cleared      map[int]string    `yaml:"cleared"`
	Seen         []int             `yaml:"seen,omitempty"` // stages whose story scene has been shown
	Coins        Wallet            `yaml:"coins,omitempty"`
	DailyPaid    string            `yaml:"daily_paid,omitempty"`    // the day the daily bonus was last paid
	PaidDay      string            `yaml:"paid_day,omitempty"`      // the day PaidCards covers
	PaidCards    []string          `yaml:"paid_cards,omitempty"`    // cards already paid for that day
	Owned        []string          `yaml:"owned,omitempty"`         // gear bought at the shrine, by id
	Worn         map[string]string `yaml:"worn,omitempty"`          // slot -> gear id
	Defeated     []string          `yaml:"defeated,omitempty"`      // enemies beaten at least once, for the collection
	Claimed      []string          `yaml:"claimed,omitempty"`       // collection sets whose reward has been paid
	TallyDay     string            `yaml:"tally_day,omitempty"`     // the day Tally counts
	Tally        map[string]int    `yaml:"tally,omitempty"`         // fresh progress toward today's bounties
	BountiesPaid []string          `yaml:"bounties_paid,omitempty"` // today's bounties already paid
	Shields      int               `yaml:"shields,omitempty"`       // streak shields held
	Shielded     []string          `yaml:"shielded,omitempty"`      // missed days a shield covered
}

func (s *Store) readJourney() (journeyRecord, error) {
	rec := journeyRecord{}
	data, err := os.ReadFile(filepath.Join(s.Dir, journeyFile))
	if err != nil && !os.IsNotExist(err) {
		return rec, fmt.Errorf("read journey progress: %w", err)
	}
	if err == nil {
		if err := yaml.Unmarshal(data, &rec); err != nil {
			return rec, fmt.Errorf("parse %s: %w", journeyFile, err)
		}
	}
	if rec.Cleared == nil {
		rec.Cleared = map[int]string{}
	}
	return rec, nil
}

func (s *Store) writeJourney(rec journeyRecord) error {
	data, err := yaml.Marshal(rec)
	if err != nil {
		return fmt.Errorf("encode journey progress: %w", err)
	}
	return os.WriteFile(filepath.Join(s.Dir, journeyFile), data, 0644)
}

// ClearedStages returns the date each stage was cleared, by tribulation
// number. A missing file means none has been cleared yet.
func (s *Store) ClearedStages() (map[int]string, error) {
	rec, err := s.readJourney()
	return rec.Cleared, err
}

// RecordStageCleared saves a win that clears a stage. Winning again keeps
// the date of the first clear.
func (s *Store) RecordStageCleared(stage int, when time.Time) error {
	rec, err := s.readJourney()
	if err != nil {
		return err
	}
	if _, ok := rec.Cleared[stage]; ok {
		return nil
	}
	rec.Cleared[stage] = when.Format(dateFmt)
	return s.writeJourney(rec)
}

// SceneSeen reports whether a stage's story scene has been shown.
func (s *Store) SceneSeen(stage int) (bool, error) {
	rec, err := s.readJourney()
	return slices.Contains(rec.Seen, stage), err
}

// MarkSceneSeen records that a stage's story scene has been shown.
func (s *Store) MarkSceneSeen(stage int) error {
	rec, err := s.readJourney()
	if err != nil || slices.Contains(rec.Seen, stage) {
		return err
	}
	rec.Seen = append(rec.Seen, stage)
	return s.writeJourney(rec)
}
