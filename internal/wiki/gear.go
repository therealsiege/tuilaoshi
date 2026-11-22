package wiki

import (
	"errors"
	"slices"
)

// ErrCantAfford is returned when the purse is short of an item's price.
var ErrCantAfford = errors.New("not enough coins")

// Kit is the gear the learner owns and which piece is worn in each slot.
type Kit struct {
	Owned []string
	Worn  map[string]string // slot -> gear id
}

// Kit returns the gear owned and worn.
func (s *Store) Kit() (Kit, error) {
	rec, err := s.readJourney()
	if rec.Worn == nil {
		rec.Worn = map[string]string{}
	}
	return Kit{Owned: rec.Owned, Worn: rec.Worn}, err
}

// Buy pays for an item and adds it to the owned gear. Buying what is
// already owned does nothing.
func (s *Store) Buy(id string, price Wallet) error {
	rec, err := s.readJourney()
	if err != nil || slices.Contains(rec.Owned, id) {
		return err
	}
	if rec.Coins.Tongqian < price.Tongqian || rec.Coins.Yuanbao < price.Yuanbao {
		return ErrCantAfford
	}
	rec.Coins.Tongqian -= price.Tongqian
	rec.Coins.Yuanbao -= price.Yuanbao
	rec.Owned = append(rec.Owned, id)
	return s.writeJourney(rec)
}

// Wear puts an owned item in its slot, or takes it off when id is "".
func (s *Store) Wear(slot, id string) error {
	rec, err := s.readJourney()
	if err != nil {
		return err
	}
	if id != "" && !slices.Contains(rec.Owned, id) {
		return errors.New("not owned: " + id)
	}
	if rec.Worn == nil {
		rec.Worn = map[string]string{}
	}
	if id == "" {
		delete(rec.Worn, slot)
	} else {
		rec.Worn[slot] = id
	}
	return s.writeJourney(rec)
}
