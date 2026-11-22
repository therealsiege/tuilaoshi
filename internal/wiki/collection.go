package wiki

import "slices"

// Collection is what the learner has collected: enemies beaten, and the
// sets whose reward has been claimed.
type Collection struct {
	Defeated, Claimed []string
}

// Collection returns the enemies beaten and the sets claimed.
func (s *Store) Collection() (Collection, error) {
	rec, err := s.readJourney()
	return Collection{Defeated: rec.Defeated, Claimed: rec.Claimed}, err
}

// RecordDefeat adds an enemy to the collection the first time it is beaten.
func (s *Store) RecordDefeat(name string) error {
	rec, err := s.readJourney()
	if err != nil || slices.Contains(rec.Defeated, name) {
		return err
	}
	rec.Defeated = append(rec.Defeated, name)
	return s.writeJourney(rec)
}

// Claim pays a completed set's reward once; claiming again pays nothing.
// It reports whether it paid.
func (s *Store) Claim(set string, reward Wallet) (bool, error) {
	rec, err := s.readJourney()
	if err != nil || slices.Contains(rec.Claimed, set) {
		return false, err
	}
	rec.Claimed = append(rec.Claimed, set)
	rec.Coins.Tongqian += reward.Tongqian
	rec.Coins.Yuanbao += reward.Yuanbao
	return true, s.writeJourney(rec)
}
