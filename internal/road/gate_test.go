package road

import (
	"math/rand"
	"slices"
	"strings"
	"testing"

	"github.com/therealsiege/tuilaoshi/internal/journey"
)

// inOrder is the shown tiles' positions that lay the round's sentence.
func inOrder(g *Gate) []int {
	var out []int
	for w := range g.Sentence().Words {
		out = append(out, slices.Index(g.Tiles, w))
	}
	return out
}

func lay(g *Gate, positions []int) Laid {
	var last Laid
	for _, p := range positions {
		last = g.Lay(p)
	}
	return last
}

func gateOf(t *testing.T, typed bool) Gate {
	t.Helper()
	g, ok := OpenGate(journey.BuildMap(0).Region.ID, typed, rand.New(rand.NewSource(1)))
	if !ok {
		t.Fatal("the first region has no grammar gate")
	}
	return g
}

func TestAGateIsFiveTileRoundsWithoutAGrader(t *testing.T) {
	g := gateOf(t, false)
	if len(g.Rounds) != GateRounds || g.Typed() || g.Passed() {
		t.Fatalf("%d tile rounds, typed %v", len(g.Rounds), g.Typed())
	}
	if typed := gateOf(t, true); len(typed.Rounds) != GateRounds-1 {
		t.Errorf("with a grader %d tile rounds, want %d", len(typed.Rounds), GateRounds-1)
	}
	if _, ok := OpenGate("nowhere", false, rand.New(rand.NewSource(1))); ok {
		t.Error("a region that doesn't exist has a gate")
	}
}

func TestTilesAreLaidInOrder(t *testing.T) {
	g := gateOf(t, false)
	if g.Sentence().Fits(g.Chosen([]int{0, 1, 2, 3, 4, 5, 6, 7, 8}[:len(g.Tiles)])) {
		t.Fatal("the tiles were laid out in the right order")
	}
	if g.Lay(0) != LaidMore || g.Lay(0) != LaidNot || g.Lay(len(g.Tiles)) != LaidNot {
		t.Fatal("laying a tile twice, or one that isn't there")
	}
	g.Undo()
	if len(g.Picked) != 0 {
		t.Fatalf("undo left %v", g.Picked)
	}
	for miss := 1; miss <= 3; miss++ {
		wrong := inOrder(&g)
		wrong[0], wrong[1] = wrong[1], wrong[0]
		if lay(&g, wrong) != LaidWrong || g.Round != 0 || len(g.Picked) != 0 || g.Misses != miss {
			t.Fatalf("miss %d: round %d, picked %v, note %q", miss, g.Round, g.Picked, g.Note)
		}
	}
	if !strings.Contains(g.Note, "It reads: "+g.Sentence().Text()) {
		t.Errorf("after three misses the note should give the sentence: %q", g.Note)
	}
	for round := 0; round < GateRounds; round++ {
		if lay(&g, inOrder(&g)) != LaidRight || g.Round != round+1 || g.Misses != 0 || !strings.HasPrefix(g.Note, "对！") {
			t.Fatalf("round %d laid right: round %d, note %q", round, g.Round, g.Note)
		}
	}
	if !g.Passed() || g.Lay(0) != LaidNot {
		t.Error("five sentences laid and the gate isn't passed")
	}
}

func TestTheTypedRoundKeepsThePatternsRules(t *testing.T) {
	g := gateOf(t, true)
	for range GateRounds - 1 {
		lay(&g, inOrder(&g))
	}
	if !g.Typed() || g.Lay(0) != LaidNot {
		t.Fatal("the last round should be typed")
	}
	if note := g.Unused("我很好。"); !strings.Contains(note, "doesn't use "+g.Pattern) {
		t.Errorf("a sentence without the pattern: %q", note)
	}
	if !strings.Contains(g.Task(), g.Pattern) || !strings.Contains(g.Criteria(), g.Pattern) {
		t.Errorf("the grader's task doesn't name the pattern: %q / %q", g.Task(), g.Criteria())
	}
}

func TestAPassedGateIsKept(t *testing.T) {
	w := open(t, 0)
	g := gateOf(t, false)
	if w.GatePassed() {
		t.Fatal("a new journey's gate is passed")
	}
	if err := w.PassGate(&g); err != nil || !g.Passed() || !w.GatePassed() {
		t.Fatalf("passing: %v, passed %v", err, w.GatePassed())
	}
	if !reloaded(t, w).GatePassed() {
		t.Error("the passed gate wasn't saved")
	}
}
