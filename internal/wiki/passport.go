package wiki

import "slices"

// Stamps returns the towns whose kingdoms have stamped the travel
// rescript, in the order stamped.
func (s *Store) Stamps() ([]string, error) {
	rec, err := s.readJourney()
	return rec.Stamps, err
}

// Stamp records a kingdom's stamp on the rescript and pays the court's
// gift, once.
func (s *Store) Stamp(town string, pay Wallet) error {
	rec, err := s.readJourney()
	if err != nil || slices.Contains(rec.Stamps, town) {
		return err
	}
	rec.Stamps = append(rec.Stamps, town)
	rec.Coins.Tongqian += pay.Tongqian
	rec.Coins.Yuanbao += pay.Yuanbao
	return s.writeJourney(rec)
}
