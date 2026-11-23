package wiki

import (
	"errors"
	"slices"
)

// ErrMissingMaterials is returned when crafting lacks materials.
var ErrMissingMaterials = errors.New("not enough materials")

// Materials returns the materials held, by id.
func (s *Store) Materials() (map[string]int, error) {
	rec, err := s.readJourney()
	if rec.Materials == nil {
		rec.Materials = map[string]int{}
	}
	return rec.Materials, err
}

// AddMaterials puts loot in the pack.
func (s *Store) AddMaterials(loot []string) error {
	if len(loot) == 0 {
		return nil
	}
	rec, err := s.readJourney()
	if err != nil {
		return err
	}
	if rec.Materials == nil {
		rec.Materials = map[string]int{}
	}
	for _, id := range loot {
		rec.Materials[id]++
	}
	return s.writeJourney(rec)
}

// Craft makes an item from materials and copper and adds it to the gear
// owned. Crafting what is already owned does nothing.
func (s *Store) Craft(id string, recipe map[string]int, price Wallet) error {
	rec, err := s.readJourney()
	if err != nil || slices.Contains(rec.Owned, id) {
		return err
	}
	for m, n := range recipe {
		if rec.Materials[m] < n {
			return ErrMissingMaterials
		}
	}
	if rec.Coins.Tongqian < price.Tongqian || rec.Coins.Yuanbao < price.Yuanbao {
		return ErrCantAfford
	}
	for m, n := range recipe {
		rec.Materials[m] -= n
	}
	rec.Coins.Tongqian -= price.Tongqian
	rec.Coins.Yuanbao -= price.Yuanbao
	rec.Owned = append(rec.Owned, id)
	return s.writeJourney(rec)
}
