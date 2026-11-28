package road

import (
	"fmt"
	"math/rand"
	"slices"
	"strings"

	"github.com/therealsiege/tuilaoshi/internal/journey"
)

// Grammar gates (语法关): before a region's demon's lair can be fought,
// the learner passes the grammar pattern guarding it by building
// GateRounds sentences with it. The first rounds lay word tiles in order;
// with a grader the last is a sentence of their own, typed, which passes
// when it keeps the pattern's simple rules (Unused) and the grader judges
// it right (Task and Criteria). Without one the rounds are all tiles.

// GateRounds is how many sentences pass a gate.
const GateRounds = 5

// Gate is a grammar gate being taken.
type Gate struct {
	journey.GrammarGate
	Rng    *rand.Rand // chooses the sentences and shuffles their tiles
	Rounds []int      // the tile rounds' sentences, by index into Sentences
	Round  int        // the round under way, from 0; GateRounds once passed
	Tiles  []int      // the sentence's words in the order shown
	Picked []int      // the shown tiles laid so far, in order
	Misses int        // wrong orders of this round's sentence
	Note   string     // what came of the last try
}

// OpenGate starts the grammar gate of region from its first round: all
// tiles, or with typed the last round a sentence of the learner's own.
// ok is false when the region has no gate.
func OpenGate(region string, typed bool, rng *rand.Rand) (Gate, bool) {
	g, ok := journey.GrammarGateFor(region)
	if !ok {
		return Gate{}, false
	}
	tiles := GateRounds
	if typed {
		tiles-- // the last round is typed
	}
	run := Gate{GrammarGate: g, Rng: rng, Rounds: rng.Perm(len(g.Sentences))[:min(tiles, len(g.Sentences))]}
	run.Shuffle()
	return run, true
}

// Passed reports whether every round is done.
func (g *Gate) Passed() bool { return g.Round >= GateRounds }

// Typed reports whether this round is the sentence of the learner's own.
func (g *Gate) Typed() bool { return !g.Passed() && g.Round >= len(g.Rounds) }

// Sentence is the tile round's sentence.
func (g *Gate) Sentence() journey.TileSentence { return g.Sentences[g.Rounds[g.Round]] }

// Shuffle lays the round's tiles out afresh, never in the right order.
func (g *Gate) Shuffle() {
	s := g.Sentence()
	g.Picked = nil
	for {
		g.Tiles = g.Rng.Perm(len(s.Words))
		if len(s.Words) < 2 || !s.Fits(g.Chosen(g.Tiles)) {
			return
		}
	}
}

// Chosen is the words of the shown tiles at positions, in that order.
func (g *Gate) Chosen(positions []int) []string {
	s := g.Sentence()
	words := make([]string, len(positions))
	for i, p := range positions {
		words[i] = s.Words[g.Tiles[p]]
	}
	return words
}

// Laid is what laying a tile came to.
type Laid int

const (
	LaidNot   Laid = iota // no such tile, or laid already
	LaidMore              // laid, with more to lay
	LaidRight             // the sentence is laid in order: on to the next round
	LaidWrong             // all laid but out of order: shuffled again
)

// Lay lays the shown tile p next. When every tile is laid the sentence is
// checked: in order, the gate moves on a round (see Advance); out of
// order, the tiles are shuffled again, and after three misses the note
// gives the sentence.
func (g *Gate) Lay(p int) Laid {
	if g.Passed() || g.Typed() || p < 0 || p >= len(g.Tiles) || slices.Contains(g.Picked, p) {
		return LaidNot
	}
	g.Picked = append(g.Picked, p)
	if len(g.Picked) < len(g.Tiles) {
		return LaidMore
	}
	s := g.Sentence()
	if s.Fits(g.Chosen(g.Picked)) {
		g.Note = "对！" + strings.Join(g.Chosen(g.Picked), "") + " " + s.English
		g.Advance()
		return LaidRight
	}
	g.Misses++
	g.Note = "不对. The tiles are shuffled again: try once more."
	if g.Misses >= 3 {
		g.Note += " It reads: " + s.Text()
	}
	g.Shuffle()
	return LaidWrong
}

// Undo takes back the last tile laid.
func (g *Gate) Undo() {
	if len(g.Picked) > 0 {
		g.Picked = g.Picked[:len(g.Picked)-1]
	}
}

// Advance moves on a round: the next tiles laid out, or the typed round,
// or through the gate (Passed; the caller records it with PassGate).
func (g *Gate) Advance() {
	g.Round++
	g.Misses = 0
	if !g.Passed() && !g.Typed() {
		g.Shuffle()
	}
}

// Unused is what to say when a typed sentence doesn't use the gate's
// pattern by its simple rules; "" when it does, and may go to the grader.
func (g *Gate) Unused(sentence string) string {
	if g.Uses(sentence) {
		return ""
	}
	return fmt.Sprintf("That doesn't use %s yet. %s", g.Pattern, g.Explain)
}

// Task is the typed round's task, for the grader.
func (g *Gate) Task() string {
	return fmt.Sprintf("Write one Chinese sentence of your own using the grammar pattern %s (%s): %s", g.Pattern, g.Pinyin, g.Explain)
}

// Criteria is what the grader holds the typed sentence to.
func (g *Gate) Criteria() string {
	return fmt.Sprintf("A complete, grammatical sentence that uses %s correctly and naturally, at about HSK3-4; any topic. Score 0 if %s is missing or misused.", g.Pattern, g.Pattern)
}

// PassGate records this region's gate passed for the journey, and the gate
// run g (if any) as through.
func (w *Walk) PassGate(g *Gate) error {
	id := w.Region.Region.ID
	if !slices.Contains(w.Gates, id) {
		w.Gates = append(w.Gates, id)
	}
	if g != nil {
		g.Round = GateRounds
	}
	return w.Store.PassGrammarGate(id)
}
