package wiki

import "time"

// Deed is something done in a town that the court gazette reports: an
// errand finished or a sack of food haggled for.
type Deed struct {
	Date  string `yaml:"date"` // YYYY-MM-DD
	Kind  string `yaml:"kind"` // DeedErrand, DeedHaggle or DeedUnmask
	Town  string `yaml:"town"` // town id
	What  string `yaml:"what"` // the errand's goal in Chinese, or what was bought
	Coins int    `yaml:"coins,omitempty"`
}

// The kinds of deed.
const (
	DeedErrand = "errand"
	DeedHaggle = "haggle"
	DeedUnmask = "unmask" // a demon in disguise unmasked with Fiery Eyes
)

// ErrandTaken reports whether a town's errand was taken today.
func (s *Store) ErrandTaken(town string, now time.Time) (bool, error) {
	rec, err := s.readJourney()
	return rec.Errands[town] == now.Format(dateFmt), err
}

// TakeErrand records taking a town's errand today; a town has one a day.
func (s *Store) TakeErrand(town string, now time.Time) error {
	rec, err := s.readJourney()
	if err != nil {
		return err
	}
	if rec.Errands == nil {
		rec.Errands = map[string]string{}
	}
	rec.Errands[town] = now.Format(dateFmt)
	return s.writeJourney(rec)
}

// FinishErrand pays an errand's reward and keeps it as a deed.
func (s *Store) FinishErrand(town, goal string, now time.Time, pay Wallet) error {
	rec, err := s.readJourney()
	if err != nil {
		return err
	}
	rec.Coins.Tongqian += pay.Tongqian
	rec.Coins.Yuanbao += pay.Yuanbao
	rec.Deeds = append(rec.Deeds, Deed{Date: now.Format(dateFmt), Kind: DeedErrand, Town: town, What: goal, Coins: pay.Tongqian})
	return s.writeJourney(rec)
}

// AddDeed keeps a deed for the gazette.
func (s *Store) AddDeed(d Deed) error {
	rec, err := s.readJourney()
	if err != nil {
		return err
	}
	rec.Deeds = append(rec.Deeds, d)
	return s.writeJourney(rec)
}

// Deeds returns the deeds done on the journey, oldest first.
func (s *Store) Deeds() ([]Deed, error) {
	rec, err := s.readJourney()
	return rec.Deeds, err
}

// MysteriesSettled returns the day each town's Fiery Eyes mystery was last
// settled, by town id.
func (s *Store) MysteriesSettled() (map[string]string, error) {
	rec, err := s.readJourney()
	return rec.Mysteries, err
}
