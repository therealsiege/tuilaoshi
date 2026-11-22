package wiki

// BondPoints returns each companion's bond points.
func (s *Store) BondPoints() (map[string]int, error) {
	rec, err := s.readJourney()
	if rec.Bonds == nil {
		rec.Bonds = map[string]int{}
	}
	return rec.Bonds, err
}

// AddBond adds points to a companion's bond.
func (s *Store) AddBond(companion string, points int) error {
	if points <= 0 {
		return nil
	}
	rec, err := s.readJourney()
	if err != nil {
		return err
	}
	if rec.Bonds == nil {
		rec.Bonds = map[string]int{}
	}
	rec.Bonds[companion] += points
	return s.writeJourney(rec)
}
