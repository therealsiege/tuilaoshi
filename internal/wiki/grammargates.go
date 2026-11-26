package wiki

import "slices"

// GrammarGates returns the regions whose grammar gate (语法关) the
// pilgrim has passed on this journey, in the order passed.
func (s *Store) GrammarGates() ([]string, error) {
	rec, err := s.readJourney()
	return rec.GrammarGates, err
}

// PassGrammarGate records that a region's grammar gate is passed, once.
func (s *Store) PassGrammarGate(region string) error {
	rec, err := s.readJourney()
	if err != nil || slices.Contains(rec.GrammarGates, region) {
		return err
	}
	rec.GrammarGates = append(rec.GrammarGates, region)
	return s.writeJourney(rec)
}
