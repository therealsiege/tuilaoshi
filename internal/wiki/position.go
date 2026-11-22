package wiki

// Place is where the pilgrim stands on the road, and their health there.
type Place struct {
	Region string // region id; empty before the journey's first step
	X, Y   int
	HP     int // 0 means full health
}

// Place returns where the journey being played stands.
func (s *Store) Place() (Place, error) {
	rec, err := s.readJourney()
	return Place{Region: rec.Region, X: rec.X, Y: rec.Y, HP: rec.HP}, err
}

// SetPlace saves where the pilgrim stands.
func (s *Store) SetPlace(p Place) error {
	rec, err := s.readJourney()
	if err != nil {
		return err
	}
	rec.Region, rec.X, rec.Y, rec.HP = p.Region, p.X, p.Y, p.HP
	return s.writeJourney(rec)
}
