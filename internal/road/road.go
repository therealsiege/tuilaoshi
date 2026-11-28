// Package road holds the rules of walking a region's tile map, shared by
// the Mac's map (internal/ui/roam) and the phone (mobile): what a step
// meets, the doors between regions, the tribulations and crossings that
// bar the way, towns, and the encounters of wild ground. Where the pilgrim
// stands is saved with the journey.
//
// A step is two halves: Move goes onto the next cell (or is stopped, or
// travels through a door), and Arrive says what the cell holds. The Mac
// passes the trail's day between them (see roam's trail.go); Step does
// both at once.
package road

import (
	"errors"
	"fmt"
	"math/rand"
	"slices"

	"github.com/therealsiege/tuilaoshi/internal/journey"
	"github.com/therealsiege/tuilaoshi/internal/wiki"
)

// EncounterChance is the chance, in percent, that a step onto wild ground
// meets one of the region's clans; the road itself is safe.
const EncounterChance = 12

// Rate is the encounter chance in effect for the learner's setting
// (wiki.Settings.Encounters: normal, fewer or off).
func Rate(setting string) int {
	switch setting {
	case "fewer":
		return EncounterChance / 2
	case "off":
		return 0
	}
	return EncounterChance
}

// Kind is what a step came to.
type Kind string

const (
	Moved       Kind = "moved"       // onto the next cell
	Blocked     Kind = "blocked"     // ground that can't be walked
	Travelled   Kind = "region"      // through a door into the next region
	Tribulation Kind = "tribulation" // a tribulation not yet won bars the way
	Crossing    Kind = "crossing"    // the region's obstacle bars the way until crossed
	Town        Kind = "town"        // into a town
	Encounter   Kind = "encounter"   // a soldier of the region's clans
)

// Outcome is what a step came to, with the Mac's words for it in Note.
type Outcome struct {
	Kind Kind
	Note string

	Stage int  // Tribulation: the stage met
	Fight bool // Tribulation: it is the open stage, so its battle starts
	Gate  bool // Tribulation: the lair, behind a grammar gate not yet passed

	Event   *journey.Event   // Crossing: today's conditions and the ways across
	Town    *journey.Town    // Town
	Soldier *journey.Soldier // Encounter
}

// Walk is the pilgrim on a region's map, with what bars the road.
type Walk struct {
	Store   *wiki.Store
	Region  journey.RegionMap
	Pos     journey.Point
	Cleared map[int]string // the stages won
	Counts  journey.Counts
	Stage   *journey.Stage // the stage being worked on; nil once all are cleared
	Crossed []string       // regions whose obstacle is behind the party
	Gates   []string       // regions whose grammar gate is passed
}

// Load reads the journey and puts the pilgrim where they last stood, or
// at the start of the region being worked on (saving it there). The walk
// is usable even with an error, as far as it was read.
func Load(store *wiki.Store) (Walk, error) {
	pages, pagesErr := store.ListPages()
	cleared, clearedErr := store.ClearedStages()
	place, placeErr := store.Place()
	gates, gatesErr := store.GrammarGates()
	trail, _, trailErr := journey.LoadTrail(store, cleared)
	w := Walk{Store: store, Cleared: cleared, Gates: gates, Crossed: trail.Crossed}
	err := errors.Join(pagesErr, clearedErr, placeErr, gatesErr, trailErr)
	w.Counts = journey.CountProgress(pages)
	w.Stage = journey.CurrentStage(journey.Stages(w.Counts, cleared))
	if i := journey.RegionByID(place.Region); i >= 0 {
		w.Region, w.Pos = journey.BuildMap(i), journey.Point{X: place.X, Y: place.Y}
		return w, err
	}
	n := 81 // the pilgrimage is complete: stand at its end
	if w.Stage != nil {
		n = w.Stage.Number
	}
	w.Region = journey.BuildMap(journey.RegionOf(n))
	w.Pos = w.Region.Start
	return w, errors.Join(err, w.Save())
}

// Save records where the pilgrim stands.
func (w Walk) Save() error {
	return w.Store.SetPlace(wiki.Place{Region: w.Region.Region.ID, X: w.Pos.X, Y: w.Pos.Y})
}

// Step moves the pilgrim one cell and says what is there: Move, then
// Arrive when the step was taken. rate is the encounter chance (see Rate)
// and roll draws 0-99 against it; nil rolls at random.
func (w *Walk) Step(dx, dy, rate int, roll func() int) (Outcome, error) {
	o, err := w.Move(dx, dy)
	if o.Kind != Moved {
		return o, err
	}
	return w.Arrive(w.Pos, rate, roll), err
}

// Move steps one cell, if the ground allows: a door travels to the
// neighbouring region, and an uncleared tribulation or an uncrossed
// obstacle bars the way. The place is saved whenever it changes; the
// error is the save's.
func (w *Walk) Move(dx, dy int) (Outcome, error) {
	to := journey.Point{X: w.Pos.X + dx, Y: w.Pos.Y + dy}
	switch w.Region.At(to) {
	case journey.TileBlock:
		return Outcome{Kind: Blocked}, nil
	case journey.TileEast:
		return w.travel(w.Region.Index-1, false)
	case journey.TileWest:
		return w.travel(w.Region.Index+1, true)
	case journey.TileStage:
		if !w.StageCleared(to) {
			return w.Face(to), nil // the tribulation bars the road: face it, don't pass
		}
	case journey.TileCrossing:
		if !w.HasCrossed() {
			e := w.Region.Crossing.Event(rand.New(rand.NewSource(rand.Int63())))
			return Outcome{Kind: Crossing, Event: &e}, nil // the obstacle bars the road until crossed
		}
	}
	w.Pos = to
	return Outcome{Kind: Moved}, w.Save()
}

// travel walks through a door into the neighbouring region.
func (w *Walk) travel(i int, west bool) (Outcome, error) {
	w.Region = journey.BuildMap(i)
	w.Pos = w.Region.End
	if west {
		w.Pos = w.Region.Start
	}
	return Outcome{Kind: Travelled, Note: "You come to " + w.Region.Region.Name + " " + w.Region.Region.English + "."}, w.Save()
}

// Arrive says what the cell at p holds once stepped on: a town is entered,
// and wild ground may meet a soldier (rate and roll as for Step).
func (w Walk) Arrive(p journey.Point, rate int, roll func() int) Outcome {
	if t, ok := w.Region.Towns[p]; ok {
		return Outcome{Kind: Town, Town: &t}
	}
	if roll == nil {
		roll = func() int { return rand.Intn(100) }
	}
	if w.Region.At(p) == journey.TileWild && roll() < rate {
		if s, ok := w.Region.Region.Encounter(rand.New(rand.NewSource(rand.Int63()))); ok {
			return Outcome{Kind: Encounter, Soldier: &s, Note: s.Name + " " + s.English + " leaps out of the undergrowth!"}
		}
	}
	return Outcome{Kind: Moved}
}

// Face meets the tribulation at p: its battle when it is the open stage
// (for the demon's lair, once its grammar gate is passed), otherwise what
// it still needs.
func (w Walk) Face(p journey.Point) Outcome {
	n := w.Region.Stages[p]
	o := Outcome{Kind: Tribulation, Stage: n}
	s := w.Stage
	switch {
	case s == nil || s.Number != n:
		o.Note = "An earlier tribulation still waits on the road behind you."
	case s.Status != journey.StageOpen:
		o.Note = fmt.Sprintf("Tribulation %d is sealed: %s. %s", n, journey.NeedText(*s, w.Counts), journey.HowToOpen(*s))
	case n == w.Region.Region.To && !w.GatePassed():
		o.Gate = true // the gate stands before the lair
	default:
		o.Fight = true
	}
	return o
}

// StageCleared reports whether the tribulation at p is won.
func (w Walk) StageCleared(p journey.Point) bool {
	_, ok := w.Cleared[w.Region.Stages[p]]
	return ok
}

// HasCrossed reports whether this region's obstacle is behind the party.
func (w Walk) HasCrossed() bool { return slices.Contains(w.Crossed, w.Region.Region.ID) }

// GatePassed reports whether this region's lair is free to fight: its
// gate is passed, or it has none.
func (w Walk) GatePassed() bool {
	_, ok := journey.GrammarGateFor(w.Region.Region.ID)
	return !ok || slices.Contains(w.Gates, w.Region.Region.ID)
}
