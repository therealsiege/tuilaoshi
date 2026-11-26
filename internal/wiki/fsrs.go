package wiki

import (
	"database/sql"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/therealsiege/tuilaoshi/internal/model"
)

// FSRS scheduling: a compact FSRS-4.5 memory model. Each card has a
// stability (the days until recall falls to 90%) and a difficulty (1 to
// 10). A review moves both, and the next interval is the day recall is
// expected to reach the desired retention. The model runs for every
// review, whichever scheduler is chosen, so switching to FSRS finds each
// card's memory up to date.

// The schedulers the learner can choose (Settings.Scheduler).
const (
	SchedulerClassic = "classic"
	SchedulerFSRS    = "fsrs"
)

// fsrsWeights are the published FSRS-4.5 default weights.
var fsrsWeights = [17]float64{
	0.4872, 1.4003, 3.7145, 13.8206, 5.1618, 1.2298, 0.8975, 0.031,
	1.6474, 0.1367, 1.0461, 2.1072, 0.0793, 0.3246, 1.587, 0.2272, 2.8755,
}

// FSRS constants: the forgetting curve's decay and factor (chosen so
// recall is 90% when the elapsed days equal the stability), and the
// recall the intervals aim for.
const (
	fsrsDecay       = -0.5
	fsrsFactor      = 19.0 / 81
	desiredRetained = 0.9
)

// fsrsMemory is a card's FSRS state. A zero stability means the card has
// not been through the model yet.
type fsrsMemory struct {
	Stability  float64
	Difficulty float64
}

// fsrsGrade turns a drill rating into the FSRS grade: Again 1, Hard 2,
// Good 3, Easy 4.
func fsrsGrade(d model.DrillDifficulty) float64 {
	return float64(min(max(int(d), 0), 3) + 1)
}

// retrievability is the chance of recall after some days at a stability.
func retrievability(days, stability float64) float64 {
	return math.Pow(1+fsrsFactor*days/stability, fsrsDecay)
}

// initialDifficulty is a new card's difficulty after its first grade.
func initialDifficulty(g float64) float64 {
	w := fsrsWeights
	return clampDifficulty(w[4] - (g-3)*w[5])
}

func clampDifficulty(d float64) float64 { return min(max(d, 1), 10) }

// next is the memory after a review graded d, days after the last one.
// A card new to the model starts from its first grade when it has no
// interval yet, or from its interval (taken as its stability) when the
// classic scheduler has already been spacing it.
func (m fsrsMemory) next(d model.DrillDifficulty, interval, days int) fsrsMemory {
	w := fsrsWeights
	g := fsrsGrade(d)
	if m.Stability <= 0 {
		if interval <= 0 {
			return fsrsMemory{Stability: w[int(g)-1], Difficulty: initialDifficulty(g)}
		}
		m = fsrsMemory{Stability: float64(interval), Difficulty: w[4]}
	}
	r := retrievability(float64(max(days, 0)), m.Stability)
	var s float64
	if g == 1 {
		s = w[11] * math.Pow(m.Difficulty, -w[12]) * (math.Pow(m.Stability+1, w[13]) - 1) * math.Exp(w[14]*(1-r))
		s = min(s, m.Stability) // forgetting never raises stability
	} else {
		bonus := 1.0
		switch g {
		case 2:
			bonus = w[15]
		case 4:
			bonus = w[16]
		}
		s = m.Stability * (math.Exp(w[8])*(11-m.Difficulty)*math.Pow(m.Stability, -w[9])*(math.Exp(w[10]*(1-r))-1)*bonus + 1)
	}
	// Difficulty moves with the grade, then reverts slightly toward a
	// new card's Good difficulty.
	dn := m.Difficulty - w[6]*(g-3)
	dn = clampDifficulty(w[7]*w[4] + (1-w[7])*dn)
	return fsrsMemory{Stability: s, Difficulty: dn}
}

// interval is the whole days until recall falls to the desired retention,
// at least one.
func (m fsrsMemory) interval() int {
	days := m.Stability / fsrsFactor * (math.Pow(desiredRetained, 1/fsrsDecay) - 1)
	return max(int(math.Round(days)), initialInterval)
}

// daysBetween counts the calendar days from one review to the next; zero
// when there was none before.
func daysBetween(from, to time.Time) int {
	if from.IsZero() {
		return 0
	}
	a := time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, time.UTC)
	b := time.Date(to.Year(), to.Month(), to.Day(), 0, 0, 0, 0, time.UTC)
	return int(b.Sub(a).Hours() / 24)
}

// addMemoryColumns gives the schedules table its FSRS columns, so a
// progress file made before them upgrades in place.
func addMemoryColumns(db *sql.DB) error {
	rows, err := db.Query(`SELECT name FROM pragma_table_info('schedules')`)
	if err != nil {
		return fmt.Errorf("read schedules columns: %w", err)
	}
	defer rows.Close() //nolint:errcheck // read-only
	have := map[string]bool{}
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return err
		}
		have[name] = true
	}
	if err := rows.Err(); err != nil {
		return err
	}
	for _, col := range []string{"stability", "difficulty"} {
		if have[col] {
			continue
		}
		if _, err := db.Exec(`ALTER TABLE schedules ADD COLUMN ` + col + ` REAL NOT NULL DEFAULT 0`); err != nil {
			return fmt.Errorf("add schedules.%s: %w", col, err)
		}
	}
	return nil
}

// memory reads a card's saved FSRS state; zero when it has none.
func (s *Store) memory(id string) (fsrsMemory, error) {
	db, err := s.progress()
	if err != nil {
		return fsrsMemory{}, err
	}
	var m fsrsMemory
	err = db.QueryRow(`SELECT stability, difficulty FROM schedules WHERE card_id = ?`, id).Scan(&m.Stability, &m.Difficulty)
	if errors.Is(err, sql.ErrNoRows) {
		return fsrsMemory{}, nil
	}
	if err != nil {
		return fsrsMemory{}, fmt.Errorf("read memory %s: %w", id, err)
	}
	return m, nil
}

// useFSRS reports whether the store schedules with FSRS.
func (s *Store) useFSRS() bool { return s.Scheduler == SchedulerFSRS }

// CompareSchedulers replays the whole review log through both schedulers
// and counts, for each, the reviewed words that would be due in the next
// days (today, with anything overdue, and the days after).
func (s *Store) CompareSchedulers(now time.Time, days int) (classic, fsrs int, err error) {
	db, err := s.progress()
	if err != nil {
		return 0, 0, err
	}
	rows, err := db.Query(`SELECT card_id, at, rating FROM reviews ORDER BY card_id, at, id`)
	if err != nil {
		return 0, 0, fmt.Errorf("replay reviews: %w", err)
	}
	defer rows.Close() //nolint:errcheck // read-only
	type replay struct {
		classic, fsrs model.Card
		mem           fsrsMemory
	}
	cards := map[string]*replay{}
	for rows.Next() {
		var id, at string
		var rating int
		if err := rows.Scan(&id, &at, &rating); err != nil {
			return 0, 0, err
		}
		when, err := time.Parse(time.RFC3339, at)
		if err != nil {
			continue // a bad stamp can't be placed in time
		}
		r := cards[id]
		if r == nil {
			r = &replay{classic: model.Card{ID: id}, fsrs: model.Card{ID: id}}
			cards[id] = r
		}
		d := model.DrillDifficulty(rating)
		updateSRS(&r.classic, d, when, nil, false)
		updateSRS(&r.fsrs, d, when, &r.mem, true)
	}
	if err := rows.Err(); err != nil {
		return 0, 0, err
	}
	last := now.AddDate(0, 0, days-1).Format(dateFmt)
	for _, r := range cards {
		if r.classic.NextReview.Format(dateFmt) <= last {
			classic++
		}
		if r.fsrs.NextReview.Format(dateFmt) <= last {
			fsrs++
		}
	}
	return classic, fsrs, nil
}
