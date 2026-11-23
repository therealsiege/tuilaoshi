package wiki

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"time"

	"gopkg.in/yaml.v3"
)

// forgeFile caches each character's main written components for the
// character forge, asked of Claude once per character. Like the word
// schedules it belongs to the learner's study, not to one journey.
const forgeFile = "forge.yaml"

// Components returns the cached decompositions, character -> components.
func (s *Store) Components() (map[string][]string, error) {
	comps := map[string][]string{}
	data, err := os.ReadFile(filepath.Join(s.Dir, forgeFile))
	if os.IsNotExist(err) {
		return comps, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", forgeFile, err)
	}
	if err := yaml.Unmarshal(data, &comps); err != nil {
		return nil, fmt.Errorf("parse %s: %w", forgeFile, err)
	}
	return comps, nil
}

// SaveComponents adds decompositions to the cache.
func (s *Store) SaveComponents(add map[string][]string) error {
	comps, err := s.Components()
	if err != nil {
		return err
	}
	for ch, parts := range add {
		comps[ch] = parts
	}
	data, err := yaml.Marshal(comps)
	if err != nil {
		return fmt.Errorf("encode %s: %w", forgeFile, err)
	}
	return os.WriteFile(filepath.Join(s.Dir, forgeFile), data, 0644)
}

// PaidToday reports whether a card has already paid its coin today, so a
// caller can tell fresh recall from a repeat before saving a session.
func (s *Store) PaidToday(cardID string, now time.Time) (bool, error) {
	rec, err := s.readStudy()
	return rec.PaidDay == now.Format(dateFmt) && slices.Contains(rec.PaidCards, cardID), err
}
