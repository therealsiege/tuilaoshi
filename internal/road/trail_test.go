package road

import (
	"strings"
	"testing"

	"github.com/therealsiege/tuilaoshi/internal/journey"
)

func always() int { return 0 }

// endOfDay readies the walk so its next step ends a steady day.
func endOfDay(w *Walk) { w.Trail.Pace, w.Trail.Steps = "steady", 5 }

func walked(t *testing.T, w *Walk, eventRoll, plotRoll func() int) Day {
	t.Helper()
	d, err := w.Walked("", eventRoll, plotRoll)
	if err != nil {
		t.Fatal(err)
	}
	return d
}

func reloaded(t *testing.T, w Walk) Walk {
	t.Helper()
	again, err := Load(w.Store)
	if err != nil {
		t.Fatal(err)
	}
	return again
}

func TestAStepSpendsPartOfADay(t *testing.T) {
	w := open(t, 0)
	if d, err := w.Walked("", never, never); err != nil || d.Ended || d.Note != "" || w.Trail.Steps != 1 {
		t.Fatalf("one step: %+v %v, steps %d", d, err, w.Trail.Steps)
	}
	endOfDay(&w)
	food := w.Trail.Food
	d, err := w.Walked("", never, never)
	if err != nil || !d.Ended || !strings.HasPrefix(d.Note, "Day 2.") || d.Event != nil || d.Plot != nil {
		t.Fatalf("the day's last step: %+v %v", d, err)
	}
	if w.Trail.Food != food-3*len(w.Members) {
		t.Errorf("food %d -> %d over a filling day", food, w.Trail.Food)
	}
	if saved := reloaded(t, w).Trail; saved.Day != 2 || saved.Food != w.Trail.Food {
		t.Errorf("the day wasn't saved: %+v", saved)
	}
}

func TestATownRefillsTheWater(t *testing.T) {
	w := open(t, 4)
	w.Region = journey.BuildMap(1)
	for p := range w.Region.Towns {
		w.Pos = p
	}
	w.Trail.Water = 0
	if _, err := w.Walked("", never, never); err != nil || w.Trail.Water != journey.MaxWater {
		t.Errorf("water in a town: %d %v", w.Trail.Water, err)
	}
}

func TestACollapsedMasterIsCarriedBack(t *testing.T) {
	w := open(t, 0)
	w.Pos = journey.Point{X: w.Region.Start.X - 3, Y: w.Region.Start.Y}
	w.Trail.Health[journey.Master], w.Trail.Food, w.Trail.Water = 5, 0, 0
	endOfDay(&w)
	d, err := w.Walked("", always, always)
	if err != nil || !d.Collapsed || !strings.Contains(d.Note, "collapses") || d.Event != nil || d.Plot != nil {
		t.Fatalf("a starving, thirsty master: %+v %v", d, err)
	}
	if w.Pos != w.Region.Start || w.Trail.Health[journey.Master] != 30 {
		t.Errorf("carried back to %v with health %d", w.Pos, w.Trail.Health[journey.Master])
	}
	if again := reloaded(t, w); again.Pos != w.Region.Start {
		t.Errorf("the carry back wasn't saved: %v", again.Pos)
	}
}

func TestADaysEndBringsAnEventOrDemons(t *testing.T) {
	w := open(t, 0)
	endOfDay(&w)
	if d := walked(t, &w, always, always); d.Event == nil || d.Plot != nil {
		t.Errorf("an event night: %+v", d)
	}
	endOfDay(&w)
	d := walked(t, &w, never, always)
	if d.Event != nil || d.Plot == nil || len(d.Order) != len(d.Plot.Choices) {
		t.Fatalf("a night of plotting: %+v", d)
	}
	w.Trail.Forewarned = true
	endOfDay(&w)
	if d := walked(t, &w, never, always); d.Plot != nil {
		t.Error("a forewarned party overheard more demons")
	}
}

func TestPaceRationsAndRest(t *testing.T) {
	w := open(t, 0)
	if note, err := w.NextPace(); err != nil || w.Trail.Pace != "brisk" || note != "Pace: brisk"+PaceNote["brisk"] {
		t.Errorf("pace: %q %v", note, err)
	}
	if note, err := w.NextRations(); err != nil || w.Trail.Rations != "meager" || !strings.Contains(note, "meager") {
		t.Errorf("rations: %q %v", note, err)
	}
	note, err := w.Rest("rain")
	if err != nil || !strings.HasPrefix(note, "Day 2. The party rests.") {
		t.Errorf("rest: %q %v", note, err)
	}
	saved := reloaded(t, w).Trail
	if saved.Pace != "brisk" || saved.Rations != "meager" || saved.Day != 2 || saved.Food != journey.StartFood-2*len(w.Members) {
		t.Errorf("after resting: %+v", saved)
	}
}

func TestTheTrailsWarnings(t *testing.T) {
	w := open(t, 0)
	if w.FoodLow() || w.WaterLow() || w.SandalDaysLeft() != journey.SandalLife {
		t.Errorf("a fresh trail: food low %v, water low %v, sandals %d", w.FoodLow(), w.WaterLow(), w.SandalDaysLeft())
	}
	w.Trail.Food, w.Trail.Water = 1, 1
	if !w.FoodLow() || !w.WaterLow() {
		t.Error("one portion and one gourd should be low")
	}
}

func TestThePilgrimsHP(t *testing.T) {
	w := open(t, 0)
	hp, most, err := PilgrimHP(w.Store)
	if err != nil || most <= 0 || hp != most {
		t.Errorf("a fresh pilgrim: %d/%d %v", hp, most, err)
	}
}
