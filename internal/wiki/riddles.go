package wiki

import (
	"strings"
	"time"
)

// RiddlesDone is how many of a town's lantern riddles were finished today
// (solved, or the answer shown).
func (s *Store) RiddlesDone(town string, now time.Time) (int, error) {
	rec, err := s.readJourney()
	return rec.Riddles[town+"@"+now.Format(dateFmt)], err
}

// FinishRiddle counts one of a town's riddles finished today and pays for
// it; earlier days' counts are dropped.
func (s *Store) FinishRiddle(town string, now time.Time, pay Wallet) error {
	rec, err := s.readJourney()
	if err != nil {
		return err
	}
	today := "@" + now.Format(dateFmt)
	for k := range rec.Riddles {
		if !strings.HasSuffix(k, today) {
			delete(rec.Riddles, k)
		}
	}
	if rec.Riddles == nil {
		rec.Riddles = map[string]int{}
	}
	rec.Riddles[town+today]++
	rec.Coins.Tongqian += pay.Tongqian
	rec.Coins.Yuanbao += pay.Yuanbao
	return s.writeJourney(rec)
}
