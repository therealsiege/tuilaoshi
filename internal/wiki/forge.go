package wiki

import (
	"slices"
	"time"
)

// forgeFile caches each character's main written components for the
// character forge, asked of Claude once per character. Like the word
// schedules it belongs to the learner's study, not to one journey.
const forgeFile = "forge.yaml"

// Components returns the cached decompositions, character -> components.
func (s *Store) Components() (map[string][]string, error) {
	comps := map[string][]string{}
	if err := s.readYAML(forgeFile, &comps); err != nil {
		return nil, err
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
	return s.writeYAML(forgeFile, comps)
}

// PaidToday reports whether a card has already paid its coin today, so a
// caller can tell fresh recall from a repeat before saving a session.
func (s *Store) PaidToday(cardID string, now time.Time) (bool, error) {
	rec, err := s.readStudy()
	return rec.PaidDay == now.Format(dateFmt) && slices.Contains(rec.PaidCards, cardID), err
}
