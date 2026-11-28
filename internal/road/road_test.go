package road

import (
	"strings"
	"testing"
	"time"

	"github.com/therealsiege/tuilaoshi/internal/journey"
	"github.com/therealsiege/tuilaoshi/internal/wiki"
)

// open is a new journey with stages 1..cleared won and nothing learned.
func open(t *testing.T, cleared int) Walk {
	t.Helper()
	store := wiki.NewStore(t.TempDir())
	t.Cleanup(func() { store.Close() }) //nolint:errcheck // test
	for n := 1; n <= cleared; n++ {
		if err := store.RecordStageCleared(n, time.Now()); err != nil {
			t.Fatal(err)
		}
	}
	w, err := Load(store)
	if err != nil {
		t.Fatal(err)
	}
	return w
}

func never() int { return 99 }

// stageAt is where tribulation n stands on the walk's map.
func stageAt(w Walk, n int) journey.Point {
	for p, s := range w.Region.Stages {
		if s == n {
			return p
		}
	}
	return journey.Point{}
}

// from stands the pilgrim just east of p and steps west onto it.
func from(t *testing.T, w Walk, p journey.Point) (Walk, Outcome) {
	t.Helper()
	w.Pos = journey.Point{X: p.X + 1, Y: p.Y}
	o, err := w.Step(-1, 0, EncounterChance, never)
	if err != nil {
		t.Fatal(err)
	}
	return w, o
}

func TestANewJourneyStartsAtTheFirstRegion(t *testing.T) {
	w := open(t, 0)
	if w.Region.Index != 0 || w.Pos != w.Region.Start {
		t.Fatalf("new journey at region %d %v", w.Region.Index, w.Pos)
	}
	if place, err := w.Store.Place(); err != nil || place.Region != w.Region.Region.ID || place.X != w.Pos.X {
		t.Errorf("start not saved: %+v %v", place, err)
	}
}

func TestBlockedGroundStopsTheStep(t *testing.T) {
	w := open(t, 0)
	start := w.Pos
	o, err := w.Step(1, 0, EncounterChance, never) // the first region has no door east
	if err != nil || o.Kind != Blocked || w.Pos != start {
		t.Fatalf("east of the first region's start: %+v %v at %v", o, err, w.Pos)
	}
	if o, _ = w.Step(-1, 0, EncounterChance, never); o.Kind != Moved || w.Pos.X != start.X-1 { //nolint:errcheck // checked by the place
		t.Fatalf("west along the road: %+v at %v", o, w.Pos)
	}
}

func TestDoorsLeadBetweenRegions(t *testing.T) {
	w := open(t, 4)
	w.Region = journey.BuildMap(0)
	w.Pos = w.Region.End
	o, err := w.Step(-1, 0, EncounterChance, never)
	if err != nil || o.Kind != Travelled || w.Region.Index != 1 || w.Pos != w.Region.Start {
		t.Fatalf("west door: %+v %v, region %d %v", o, err, w.Region.Index, w.Pos)
	}
	if want := "You come to " + w.Region.Region.Name; !strings.HasPrefix(o.Note, want) {
		t.Errorf("note %q, want %q…", o.Note, want)
	}
	if o, _ = w.Step(1, 0, EncounterChance, never); o.Kind != Travelled || w.Region.Index != 0 || w.Pos != w.Region.End { //nolint:errcheck // checked by the region
		t.Errorf("east door: %+v, region %d %v", o, w.Region.Index, w.Pos)
	}
}

func TestATribulationBarsTheRoad(t *testing.T) {
	w := open(t, 0) // stage 1 opens with no characters
	first := stageAt(w, 1)
	w, o := from(t, w, first)
	if o.Kind != Tribulation || !o.Fight || o.Stage != 1 || w.Pos == first {
		t.Fatalf("the open tribulation: %+v, at %v", o, w.Pos)
	}
	w, o = from(t, w, stageAt(w, 3))
	if o.Kind != Tribulation || o.Fight || o.Note != "An earlier tribulation still waits on the road behind you." {
		t.Errorf("a later tribulation: %+v", o)
	}

	w = open(t, 1)
	if w, o = from(t, w, first); o.Kind != Moved || w.Pos != first {
		t.Fatalf("a won tribulation should let the pilgrim by: %+v at %v", o, w.Pos)
	}
	_, o = from(t, w, stageAt(w, 2))
	if o.Kind != Tribulation || o.Fight || !strings.HasPrefix(o.Note, "Tribulation 2 is sealed: learn ") {
		t.Errorf("a sealed tribulation: %+v", o)
	}
}

func TestTheLairWaitsBehindItsGrammarGate(t *testing.T) {
	for i := range journey.Regions() {
		w := open(t, 0)
		w.Region = journey.BuildMap(i)
		if _, ok := journey.GrammarGateFor(w.Region.Region.ID); !ok {
			continue
		}
		stages := journey.Stages(w.Counts, w.Cleared)
		w.Stage = &stages[w.Region.Region.To-1]
		w.Stage.Status = journey.StageOpen // as if enough were learned
		if o := w.Face(w.Region.Lair()); !o.Gate || o.Fight {
			t.Fatalf("the lair with the gate unpassed: %+v", o)
		}
		w.Gates = []string{w.Region.Region.ID}
		if o := w.Face(w.Region.Lair()); o.Gate || !o.Fight {
			t.Errorf("the lair with the gate passed: %+v", o)
		}
		return
	}
	t.Skip("no region has a grammar gate")
}

func TestTheCrossingBarsTheRoadUntilCrossed(t *testing.T) {
	w := open(t, 0) // the first region is a river region
	if w.Region.Crossing == nil {
		t.Fatal("no crossing in the first region")
	}
	w, o := from(t, w, w.Region.CrossAt)
	if o.Kind != Crossing || o.Event == nil || w.Pos == w.Region.CrossAt {
		t.Fatalf("stepping onto the river: %+v at %v", o, w.Pos)
	}
	if !strings.HasPrefix(o.Event.Text, w.Region.Crossing.Name+"：") || len(o.Event.Choices) == 0 {
		t.Errorf("the crossing's event: %+v", o.Event)
	}
	w.Trail.Crossed = []string{w.Region.Region.ID}
	if w, o = from(t, w, w.Region.CrossAt); o.Kind != Moved || w.Pos != w.Region.CrossAt {
		t.Errorf("a crossed river: %+v at %v", o, w.Pos)
	}
}

func TestWalkingIntoATownEntersIt(t *testing.T) {
	w := open(t, 4)
	w.Region = journey.BuildMap(1) // Chang'an's region
	for p, town := range w.Region.Towns {
		_, o := from(t, w, p)
		if o.Kind != Town || o.Town == nil || o.Town.ID != town.ID {
			t.Errorf("into %s: %+v", town.ID, o)
		}
	}
	if len(w.Region.Towns) == 0 {
		t.Fatal("no towns in Chang'an's region")
	}
}

func TestWildGroundMeetsTheRegionsClans(t *testing.T) {
	w := open(t, 0)
	var wild, path journey.Point
	for y := 0; y < w.Region.H; y++ {
		for x := 0; x < w.Region.W; x++ {
			switch p := (journey.Point{X: x, Y: y}); w.Region.At(p) {
			case journey.TileWild:
				wild = p
			case journey.TilePath:
				path = p
			}
		}
	}
	always := func() int { return 0 }
	if o := w.Arrive(path, EncounterChance, always); o.Kind != Moved {
		t.Errorf("the road itself drew %+v", o)
	}
	o := w.Arrive(wild, EncounterChance, always)
	if o.Kind != Encounter || o.Soldier == nil || o.Soldier.From > w.Region.Region.To || !strings.HasSuffix(o.Note, " leaps out of the undergrowth!") {
		t.Fatalf("wild ground: %+v", o)
	}
	if o := w.Arrive(wild, Rate("off"), always); o.Kind != Moved {
		t.Errorf("encounters off still drew %+v", o)
	}
	for setting, want := range map[string]int{"normal": EncounterChance, "fewer": EncounterChance / 2, "off": 0} {
		if got := Rate(setting); got != want {
			t.Errorf("%s: rate %d, want %d", setting, got, want)
		}
	}
}

func TestThePlaceSurvivesAReopen(t *testing.T) {
	w := open(t, 4)
	w.Region = journey.BuildMap(0)
	w.Pos = w.Region.End
	if _, err := w.Step(-1, 0, EncounterChance, never); err != nil {
		t.Fatal(err)
	}
	if _, err := w.Step(-1, 0, EncounterChance, never); err != nil {
		t.Fatal(err)
	}
	again, err := Load(w.Store)
	if err != nil || again.Region.Index != w.Region.Index || again.Pos != w.Pos {
		t.Errorf("reopened at region %d %v (%v), want %d %v", again.Region.Index, again.Pos, err, w.Region.Index, w.Pos)
	}
}
