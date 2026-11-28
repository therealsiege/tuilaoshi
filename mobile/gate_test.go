package mobile

import (
	"math/rand"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/therealsiege/tuilaoshi/internal/journey"
)

// beforeTheLair is the first region with its first three tribulations
// won and its lair open to fight, the pilgrim standing just east of it.
func beforeTheLair(t *testing.T) *Core {
	t.Helper()
	fixRolls(t, never, never)
	old := gateRand
	gateRand = func() *rand.Rand { return rand.New(rand.NewSource(1)) }
	t.Cleanup(func() { gateRand = old })
	c, _ := openRoad(t)
	for n := 1; n <= 3; n++ {
		if err := c.store.RecordStageCleared(n, time.Now()); err != nil {
			t.Fatal(err)
		}
	}
	c.forgetRoad()
	if _, err := c.Road(); err != nil {
		t.Fatal(err)
	}
	w := c.walk
	stages := journey.Stages(w.Counts, w.Cleared)
	w.Stage = &stages[w.Region.Region.To-1]
	w.Stage.Status = journey.StageOpen // as if enough were learned
	lair := w.Region.Lair()
	w.Pos = journey.Point{X: lair.X + 1, Y: lair.Y}
	if w.Region.At(w.Pos) == journey.TileBlock {
		t.Fatalf("no road east of the lair at %v", lair)
	}
	return c
}

// inOrder is the shown tiles' positions that lay the round's sentence.
func inOrder(c *Core) []int {
	var out []int
	for w := range c.gate.Sentence().Words {
		out = append(out, slices.Index(c.gate.Tiles, w))
	}
	return out
}

func TestTheLairsGrammarGateIsLaidInTiles(t *testing.T) {
	c := beforeTheLair(t)
	o := step(t, c, -1, 0)
	g := o.Gate
	if o.Kind != "tribulation" || g == nil || g.Pattern == "" || g.Explain == "" || len(g.Examples) == 0 || g.Rounds != 5 || g.Typed || len(g.Tiles) < 2 || g.English == "" {
		t.Fatalf("walking into the lair: %+v %+v", o, g)
	}
	if _, err := c.Step(-1, 0); err == nil {
		t.Error("a step with the gate open")
	}
	if g = must[*roadGate](t)(c.LayTile(0)); g.Laid != "more" || !g.Tiles[0].Laid || g.Built != g.Tiles[0].Word {
		t.Errorf("one tile laid: %+v", g)
	}
	if g = must[*roadGate](t)(c.UndoTile()); g.Tiles[0].Laid || g.Built != "" {
		t.Errorf("undone: %+v", g)
	}
	wrong := inOrder(c)
	wrong[0], wrong[1] = wrong[1], wrong[0]
	for _, p := range wrong {
		g = must[*roadGate](t)(c.LayTile(p))
	}
	if g.Laid != "wrong" || g.Round != 0 || !strings.HasPrefix(g.Note, "不对") {
		t.Errorf("out of order: %+v", g)
	}
	for round := 0; round < 5; round++ {
		for _, p := range inOrder(c) {
			g = must[*roadGate](t)(c.LayTile(p))
		}
		if g.Laid != "right" || g.Round != round+1 || !strings.HasPrefix(g.Note, "对！") {
			t.Fatalf("round %d laid right: %+v", round, g)
		}
	}
	if !g.Passed || len(g.Tiles) != 0 {
		t.Fatalf("five sentences laid: %+v", g)
	}
	if on := must[stepOutcome](t)(c.WalkOn()); !strings.HasPrefix(on.Text, "The gate stands open") {
		t.Errorf("walking on from the gate: %+v", on)
	}
	if rm := must[roadMap](t)(c.Road()); rm.Gate == nil || !rm.Gate.Passed {
		t.Errorf("the gate on the road: %+v", rm.Gate)
	}
}

func TestAGateCanBeLeftAndCheckedForThePattern(t *testing.T) {
	c := beforeTheLair(t)
	if _, err := c.Gate(); err == nil {
		t.Error("a gate with none open")
	}
	if _, err := c.CheckSentence("我把书放在桌子上。"); err == nil {
		t.Error("a sentence checked with no gate open")
	}
	step(t, c, -1, 0)
	g := must[*roadGate](t)(c.Gate())
	check := must[sentenceCheck](t)(c.CheckSentence("我很好。"))
	if check.Uses || !strings.Contains(check.Note, g.Pattern) || !strings.Contains(check.Task, g.Pattern) || check.Criteria == "" {
		t.Errorf("a sentence without %s: %+v", g.Pattern, check)
	}
	if on := must[stepOutcome](t)(c.WalkOn()); on.Text != "The grammar gate waits before the lair." {
		t.Errorf("leaving the gate: %+v", on)
	}
	if rm := must[roadMap](t)(c.Road()); rm.Gate == nil || rm.Gate.Passed {
		t.Errorf("a gate left unpassed: %+v", rm.Gate)
	}
}
