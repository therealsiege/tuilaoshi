package mobile

import (
	"encoding/json"
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/therealsiege/tuilaoshi/internal/journey"
)

func always() int { return 0 }
func never() int  { return 99 }

// fixRolls fixes the trail's chances for a test: events, plots and wild
// encounters each always or never, and no season's weather.
func fixRolls(t *testing.T, event, plot func() int) {
	t.Helper()
	oldEvent, oldPlot, oldEnc, oldWeather := eventRoll, plotRoll, encounterRoll, termWeather
	eventRoll, plotRoll, encounterRoll, termWeather = event, plot, never, func() string { return "" }
	t.Cleanup(func() { eventRoll, plotRoll, encounterRoll, termWeather = oldEvent, oldPlot, oldEnc, oldWeather })
}

// must decodes a facade call's JSON, failing the test on its error.
func must[T any](t *testing.T) func(string, error) T {
	return func(data string, err error) T {
		t.Helper()
		var v T
		if err != nil {
			t.Fatalf("%s %v", data, err)
		}
		if err := json.Unmarshal([]byte(data), &v); err != nil {
			t.Fatalf("%s: %v", data, err)
		}
		return v
	}
}

// nightfall steps west from the first region's start on the step that
// ends a steady day.
func nightfall(t *testing.T, c *Core) stepOutcome {
	t.Helper()
	c.walk.Pos, c.walk.Trail.Pace, c.walk.Trail.Steps = c.walk.Region.Start, "steady", 5
	return step(t, c, -1, 0)
}

func TestTheTrailIsTheDayAndSupplies(t *testing.T) {
	c, _ := openRoad(t)
	tr := must[trailState](t)(c.Trail())
	if tr.Day != 1 || tr.Food != journey.StartFood || tr.Water != journey.StartWater || tr.Pace != "steady" || tr.Rations != "filling" || tr.PaceNote == "" {
		t.Errorf("a new journey's trail: %+v", tr)
	}
	if len(tr.Party) != 1 || tr.Party[0].Name != journey.Master || tr.Party[0].Health != 100 || tr.HP <= 0 || tr.HP != tr.MaxHP || tr.SandalDays != journey.SandalLife {
		t.Errorf("the party: %+v, HP %d/%d, sandals %d", tr.Party, tr.HP, tr.MaxHP, tr.SandalDays)
	}
	if tr = must[trailState](t)(c.NextPace()); tr.Pace != "brisk" || !strings.HasPrefix(tr.Note, "Pace: brisk") {
		t.Errorf("next pace: %+v", tr)
	}
	if tr = must[trailState](t)(c.NextRations()); tr.Rations != "meager" || !strings.HasPrefix(tr.Note, "Rations: meager") {
		t.Errorf("next rations: %+v", tr)
	}
	if tr = must[trailState](t)(c.Rest()); tr.Day != 2 || !strings.Contains(tr.Note, "The party rests.") {
		t.Errorf("a day's rest: %+v", tr)
	}
	c.forgetRoad()
	if again := must[trailState](t)(c.Trail()); again.Day != 2 || again.Pace != "brisk" {
		t.Errorf("the trail wasn't saved: %+v", again)
	}
}

func TestAStepSpendsTheDay(t *testing.T) {
	fixRolls(t, never, never)
	c, _ := openRoad(t)
	if o := nightfall(t, c); o.Kind != "moved" || !strings.HasPrefix(o.Day, "Day 2.") {
		t.Errorf("the day's last step: %+v", o)
	}
	if o := step(t, c, 1, 0); o.Day != "" {
		t.Errorf("the next step ended a day: %+v", o)
	}
	c.walk.Trail.Health[journey.Master], c.walk.Trail.Food, c.walk.Trail.Water = 5, 0, 0
	o := nightfall(t, c)
	if !o.Collapsed || !strings.Contains(o.Day, "collapses") || o.Pos != (point{c.walk.Region.Start.X, c.walk.Region.Start.Y}) {
		t.Errorf("a starving master: %+v", o)
	}
}

func TestADaysEndBringsAnEventToChoose(t *testing.T) {
	fixRolls(t, always, never)
	c, _ := openRoad(t)
	o := nightfall(t, c)
	e := o.Event
	if o.Kind != "event" || e == nil || e.Chose != -1 || e.Text == "" || e.English != "" || len(e.Choices) < 2 || e.Choices[0].English != "" || o.Chinese != e.Text {
		t.Fatalf("the day's event: %+v %+v", o, e)
	}
	if _, err := c.Step(-1, 0); !errors.Is(err, errOpen) {
		t.Errorf("a step with the event open: %v", err)
	}
	if _, err := c.WalkOn(); err == nil {
		t.Error("walked on before choosing")
	}
	if _, err := c.Choose(len(e.Choices)); err == nil {
		t.Error("a choice that isn't there")
	}
	*e = must[roadEvent](t)(c.Choose(0))
	if e.Chose != 0 || e.English == "" || e.Choices[0].English == "" || e.Result == "" || e.ResultEnglish == "" || e.Effect == "" {
		t.Fatalf("after choosing: %+v", e)
	}
	if _, err := c.Choose(1); err == nil {
		t.Error("chose twice")
	}
	if on := must[stepOutcome](t)(c.WalkOn()); on.Kind != "moved" || on.Pos != o.Pos {
		t.Errorf("walking on: %+v", on)
	}
	if _, err := c.Step(-1, 0); err != nil {
		t.Errorf("a step after walking on: %v", err)
	}
}

func TestOverheardDemonsForewarnTheParty(t *testing.T) {
	fixRolls(t, never, always)
	c, _ := openRoad(t)
	o := nightfall(t, c)
	p := o.Plot
	if o.Kind != "plot" || p == nil || len(p.Lines) < 2 || p.Lines[0].English != "" || p.Speech == "" || p.Question == "" || p.Chose != -1 {
		t.Fatalf("overheard: %+v %+v", o, p)
	}
	if _, err := c.WalkOn(); err == nil {
		t.Error("walked on before answering")
	}
	right := slices.Index(c.plot.order, 0)
	*p = must[roadPlot](t)(c.Answer(right))
	if !p.Forewarned || p.Warning == "" || p.Lines[0].English == "" || p.QuestionEnglish == "" || !p.Choices[right].Right || p.Choices[right].English == "" {
		t.Fatalf("the right answer: %+v", p)
	}
	if tr := must[trailState](t)(c.Trail()); !tr.Forewarned {
		t.Error("the party isn't forewarned")
	}
	must[stepOutcome](t)(c.WalkOn())
	if o := nightfall(t, c); o.Plot != nil {
		t.Error("a forewarned party overheard more demons")
	}
}

func TestAWrongAnswerLeavesTheRoadAsItWas(t *testing.T) {
	fixRolls(t, never, always)
	c, _ := openRoad(t)
	nightfall(t, c)
	wrong := (slices.Index(c.plot.order, 0) + 1) % len(c.plot.order)
	if p := must[roadPlot](t)(c.Answer(wrong)); p.Forewarned || p.Warning != "" || p.Choices[wrong].English == "" || p.Choices[wrong].Right {
		t.Errorf("a wrong answer: %+v", p)
	}
}

func TestTheFirstRegionsCrossingIsTaken(t *testing.T) {
	fixRolls(t, never, never)
	c, rm := openRoad(t)
	cr := rm.Crossing
	c.walk.Pos.X, c.walk.Pos.Y = cr.X+1, cr.Y
	o := step(t, c, -1, 0)
	if o.Kind != "crossing" || o.Event == nil || !o.Event.Crossing || !strings.Contains(o.Chinese, cr.Name) || o.Pos.X != cr.X+1 {
		t.Fatalf("the river: %+v", o)
	}
	must[roadEvent](t)(c.Choose(2)) // wait a day
	on := must[stepOutcome](t)(c.WalkOn())
	if on.Pos != (point{cr.X, cr.Y}) || !strings.Contains(on.Text, "across") {
		t.Fatalf("after crossing: %+v", on)
	}
	c.forgetRoad()
	data, err := c.Road()
	if again := must[roadMap](t)(data, err); !again.Crossing.Crossed || again.Pos != on.Pos {
		t.Errorf("the crossing wasn't saved: %+v at %+v", again.Crossing, again.Pos)
	}
	if o := step(t, c, -1, 0); o.Kind != "moved" {
		t.Errorf("on past the river: %+v", o)
	}
	if o := step(t, c, 1, 0); o.Kind != "moved" || o.Event != nil {
		t.Errorf("back onto the crossed river: %+v", o)
	}
}
