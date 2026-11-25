package wiki

import "time"

// MysterySettled reports whether a town's Fiery Eyes mystery was settled
// (a villager accused) today.
func (s *Store) MysterySettled(town string, now time.Time) (bool, error) {
	rec, err := s.readJourney()
	return rec.Mysteries[town] == now.Format(dateFmt), err
}

// SettleMystery records today's accusation in a town and pays the bounty
// (nothing when the demon slipped away).
func (s *Store) SettleMystery(town string, now time.Time, pay Wallet) error {
	rec, err := s.readJourney()
	if err != nil {
		return err
	}
	if rec.Mysteries == nil {
		rec.Mysteries = map[string]string{}
	}
	rec.Mysteries[town] = now.Format(dateFmt)
	rec.Coins.Tongqian += pay.Tongqian
	rec.Coins.Yuanbao += pay.Yuanbao
	return s.writeJourney(rec)
}
