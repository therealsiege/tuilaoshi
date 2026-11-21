package wiki

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// bossFile records the date each journey boss was first beaten in battle,
// keyed by the boss's Chinese name.
const bossFile = "bosses.yaml"

// BossVictories returns the date each boss was first beaten. A missing file
// means no boss has been beaten yet.
func (s *Store) BossVictories() (map[string]string, error) {
	victories := map[string]string{}
	data, err := os.ReadFile(filepath.Join(s.Dir, bossFile))
	if os.IsNotExist(err) {
		return victories, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read boss victories: %w", err)
	}
	if err := yaml.Unmarshal(data, &victories); err != nil {
		return nil, fmt.Errorf("parse %s: %w", bossFile, err)
	}
	return victories, nil
}

// RecordBossVictory saves a win against a boss. A rematch win keeps the
// date of the first.
func (s *Store) RecordBossVictory(name string, when time.Time) error {
	victories, err := s.BossVictories()
	if err != nil {
		return err
	}
	if _, ok := victories[name]; ok {
		return nil
	}
	victories[name] = when.Format(dateFmt)
	data, err := yaml.Marshal(victories)
	if err != nil {
		return fmt.Errorf("encode boss victories: %w", err)
	}
	return os.WriteFile(filepath.Join(s.Dir, bossFile), data, 0644)
}
