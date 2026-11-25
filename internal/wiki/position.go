package wiki

import "fmt"

// Place is where the pilgrim stands on the road, and the HP they have lost.
// Wounds rather than HP are kept so gear that raises max HP still counts.
type Place struct {
	Region string // region id; empty before the journey's first step
	X, Y   int
	Wounds int
}

// Healing: each fresh right answer heals study points, and a night at an
// inn heals everything (the town sets the price).
const studyHeal = 2

// Place returns where the journey being played stands.
func (s *Store) Place() (Place, error) {
	rec, err := s.readJourney()
	return Place{Region: rec.Region, X: rec.X, Y: rec.Y, Wounds: rec.Wounds}, err
}

// SetPlace saves where the pilgrim stands; wounds are kept (see SetWounds).
func (s *Store) SetPlace(p Place) error {
	rec, err := s.readJourney()
	if err != nil {
		return err
	}
	rec.Region, rec.X, rec.Y = p.Region, p.X, p.Y
	return s.writeJourney(rec)
}

// SetWounds saves the HP lost, leaving the place as it is.
func (s *Store) SetWounds(wounds int) error {
	rec, err := s.readJourney()
	if err != nil {
		return err
	}
	rec.Wounds = max(wounds, 0)
	return s.writeJourney(rec)
}

// RestAtInn spends a night at an inn for price copper, healing every
// wound. It reports false when the purse is short.
func (s *Store) RestAtInn(price int) (bool, error) {
	rec, err := s.readJourney()
	if err != nil || rec.Coins.Tongqian < price {
		return false, err
	}
	rec.Coins.Tongqian -= price
	rec.Wounds = 0
	return true, s.writeJourney(rec)
}

// Trail decodes the journey's trail state into out; ok is false when the
// journey has none yet.
func (s *Store) Trail(out any) (ok bool, err error) {
	rec, err := s.readJourney()
	if err != nil || rec.Trail.Kind == 0 {
		return false, err
	}
	return true, rec.Trail.Decode(out)
}

// SaveTrail stores the journey's trail state.
func (s *Store) SaveTrail(v any) error {
	rec, err := s.readJourney()
	if err != nil {
		return err
	}
	if err := rec.Trail.Encode(v); err != nil {
		return fmt.Errorf("encode trail: %w", err)
	}
	return s.writeJourney(rec)
}
