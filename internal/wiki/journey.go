package wiki

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// journeyFile records the date each journey stage (tribulation 1-81) was
// cleared by winning its battle.
const journeyFile = "journey.yaml"

type journeyRecord struct {
	Cleared map[int]string `yaml:"cleared"`
}

// ClearedStages returns the date each stage was cleared, by tribulation
// number. A missing file means none has been cleared yet.
func (s *Store) ClearedStages() (map[int]string, error) {
	rec := journeyRecord{Cleared: map[int]string{}}
	data, err := os.ReadFile(filepath.Join(s.Dir, journeyFile))
	if os.IsNotExist(err) {
		return rec.Cleared, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read journey progress: %w", err)
	}
	if err := yaml.Unmarshal(data, &rec); err != nil {
		return nil, fmt.Errorf("parse %s: %w", journeyFile, err)
	}
	if rec.Cleared == nil {
		rec.Cleared = map[int]string{}
	}
	return rec.Cleared, nil
}

// RecordStageCleared saves a win that clears a stage. Winning again keeps
// the date of the first clear.
func (s *Store) RecordStageCleared(stage int, when time.Time) error {
	cleared, err := s.ClearedStages()
	if err != nil {
		return err
	}
	if _, ok := cleared[stage]; ok {
		return nil
	}
	cleared[stage] = when.Format(dateFmt)
	data, err := yaml.Marshal(journeyRecord{Cleared: cleared})
	if err != nil {
		return fmt.Errorf("encode journey progress: %w", err)
	}
	return os.WriteFile(filepath.Join(s.Dir, journeyFile), data, 0644)
}
