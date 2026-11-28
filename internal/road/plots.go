package road

import (
	"math/rand"
	"strings"

	"github.com/therealsiege/tuilaoshi/internal/journey"
)

// Overheard demons (妖怪密谋): some nights without an event two demons are
// heard plotting an ambush, each in their own voice. The learner answers
// where or when they will strike, from the choices shuffled; answered
// rightly, the party is forewarned and the next danger on the road does
// half its harm. A party already forewarned overhears nothing more.

// pickPlot may bring demons plotting at the end of a day without an event
// (roll as for Walked), with their choices shuffled.
func pickPlot(forewarned bool, roll func() int) (*journey.Plot, []int) {
	if forewarned || len(journey.Plots()) == 0 || orRandom(roll)() >= journey.PlotChance {
		return nil, nil
	}
	rng := fresh()
	p := journey.PickPlot(rng)
	return &p, Shuffled(p, rng)
}

// Shuffled is the order a plot's choices are shown in, by rng: indices
// into Choices, 0 the right one.
func Shuffled(p journey.Plot, rng *rand.Rand) []int { return rng.Perm(len(p.Choices)) }

// Right reports whether the shown choice chose (-1 before answering) is
// the right one.
func Right(order []int, chose int) bool { return chose >= 0 && chose < len(order) && order[chose] == 0 }

// Answer takes the shown choice chose of a plot shown in order: the right
// one forewarns the party, and the trail is saved.
func (w *Walk) Answer(order []int, chose int) (bool, error) {
	if !Right(order, chose) {
		return false, nil
	}
	w.Trail.Forewarned = true
	return true, w.SaveTrail()
}

// PlotText is the demons' Chinese, line by line, to read aloud.
func PlotText(p journey.Plot) string {
	var lines []string
	for _, l := range p.Lines {
		lines = append(lines, l.Text)
	}
	return strings.Join(lines, "\n")
}
