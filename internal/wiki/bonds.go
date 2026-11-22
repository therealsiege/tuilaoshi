package wiki

import "time"

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

// TalkCap is the most bond points talking with one companion earns a day.
const TalkCap = 6

// AddTalkBond adds bond points earned by talking, up to TalkCap per
// companion per day, and returns how many were added.
func (s *Store) AddTalkBond(companion string, points int, now time.Time) (int, error) {
	rec, err := s.readJourney()
	if err != nil || points <= 0 {
		return 0, err
	}
	today := now.Format(dateFmt)
	if rec.TalkDay != today {
		rec.TalkDay, rec.TalkPoints = today, map[string]int{}
	}
	if rec.TalkPoints == nil {
		rec.TalkPoints = map[string]int{}
	}
	if rec.Bonds == nil {
		rec.Bonds = map[string]int{}
	}
	add := min(points, TalkCap-rec.TalkPoints[companion])
	if add <= 0 {
		return 0, nil
	}
	rec.TalkPoints[companion] += add
	rec.Bonds[companion] += add
	return add, s.writeJourney(rec)
}
