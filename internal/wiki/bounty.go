package wiki

import (
	"hash/fnv"
	"slices"
	"time"
)

// Daily bounties are three small tasks posted each day, mixing review,
// flashcards and battle. Progress counts only fresh recall (a word's first
// right answer of the day), so bounties can't be met by repeating easy
// words. Meeting all three earns a streak shield, which covers one missed
// day so a streak survives it.

// Bounty is a daily task: reach Target on a tally.
type Bounty struct {
	ID     string // the tally it counts
	Text   string
	Target int
	Pay    int // copper paid when met
}

var bountyPool = []Bounty{
	{"review_right", "Recall 20 due words in review (r)", 20, 30},
	{"drill_right", "Get 15 words right in flashcards (d)", 15, 25},
	{"battle_right", "Land 15 hits with fresh words in battle", 15, 25},
	{"battle_won", "Win 2 battles with fresh words", 2, 25},
}

const (
	bountiesPerDay = 3
	maxShields     = 2 // streak shields held at once
)

// BountiesFor picks the day's bounties, the same all day.
func BountiesFor(day string) []Bounty {
	h := fnv.New32a()
	h.Write([]byte(day)) //nolint:errcheck // hash writes never fail
	skip := int(h.Sum32() % uint32(len(bountyPool)))
	var b []Bounty
	for i := range bountyPool {
		if i != skip && len(b) < bountiesPerDay {
			b = append(b, bountyPool[i])
		}
	}
	return b
}

// BountyStatus is a bounty with today's progress.
type BountyStatus struct {
	Bounty
	Have int
	Done bool
}

// Bounties returns the bounties of now's day with progress, and the
// streak shields held.
func (s *Store) Bounties(now time.Time) ([]BountyStatus, int, error) {
	rec, err := s.readStudy()
	today := now.Format(dateFmt)
	var out []BountyStatus
	for _, b := range BountiesFor(today) {
		have := 0
		if rec.TallyDay == today {
			have = rec.Tally[b.ID]
		}
		out = append(out, BountyStatus{Bounty: b, Have: min(have, b.Target), Done: have >= b.Target})
	}
	return out, rec.Shields, err
}

// Tally counts fresh progress toward today's bounties and pays any met.
func (s *Store) Tally(id string, n int) error {
	study, err := s.readStudy()
	if err != nil {
		return err
	}
	rec, err := s.readJourney()
	if err != nil {
		return err
	}
	tally(&study, &rec, time.Now().Format(dateFmt), id, n)
	if err := s.writeStudy(study); err != nil {
		return err
	}
	return s.writeJourney(rec)
}

// tally adds n to a counter in the study record and pays bounties it
// completes into the journey's purse; a day with all bounties met earns a
// streak shield.
func tally(rec *studyRecord, game *journeyRecord, today, id string, n int) {
	if n <= 0 {
		return
	}
	if rec.TallyDay != today {
		rec.TallyDay, rec.Tally, rec.BountiesPaid = today, map[string]int{}, nil
	}
	if rec.Tally == nil {
		rec.Tally = map[string]int{}
	}
	rec.Tally[id] += n
	bounties := BountiesFor(today)
	for _, b := range bounties {
		if rec.Tally[b.ID] >= b.Target && !slices.Contains(rec.BountiesPaid, b.ID) {
			rec.BountiesPaid = append(rec.BountiesPaid, b.ID)
			game.Coins.Tongqian += b.Pay
			if len(rec.BountiesPaid) == len(bounties) {
				rec.Shields = min(rec.Shields+1, maxShields)
			}
		}
	}
}

// shieldGap spends streak shields on the days missed between the last day
// studied and today, when there are enough to cover them all.
func (s *Store) shieldGap(rec *studyRecord, today time.Time) {
	studied := map[string]bool{}
	for _, r := range s.readSessions() {
		studied[r.Date] = true
	}
	for _, d := range rec.Shielded {
		studied[d] = true
	}
	var gap []string
	for day := today.AddDate(0, 0, -1); len(gap) <= rec.Shields; day = day.AddDate(0, 0, -1) {
		if studied[day.Format(dateFmt)] {
			if len(gap) > 0 && len(gap) <= rec.Shields {
				rec.Shields -= len(gap)
				rec.Shielded = append(rec.Shielded, gap...)
			}
			return
		}
		gap = append(gap, day.Format(dateFmt))
	}
}
