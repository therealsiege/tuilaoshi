package road

import (
	"math/rand"
	"slices"
	"strings"
	"testing"

	"github.com/therealsiege/tuilaoshi/internal/journey"
)

func TestAnsweringOverheardDemons(t *testing.T) {
	w := open(t, 0)
	p := journey.Plots()[0]
	order := Shuffled(p, rand.New(rand.NewSource(2)))
	right := slices.Index(order, 0)
	if Right(order, -1) || Right(order, len(order)) || !Right(order, right) {
		t.Fatalf("Right on %v", order)
	}
	wrong := (right + 1) % len(order)
	if ok, err := w.Answer(order, wrong); ok || err != nil || w.Trail.Forewarned {
		t.Errorf("a wrong answer: %v %v, forewarned %v", ok, err, w.Trail.Forewarned)
	}
	if ok, err := w.Answer(order, right); !ok || err != nil || !w.Trail.Forewarned {
		t.Fatalf("the right answer: %v %v", ok, err)
	}
	if !reloaded(t, w).Trail.Forewarned {
		t.Error("the forewarning wasn't saved")
	}
}

func TestThePlotsTextIsItsLines(t *testing.T) {
	p := journey.Plots()[0]
	text := PlotText(p)
	if lines := strings.Split(text, "\n"); len(lines) != len(p.Lines) || lines[0] != p.Lines[0].Text {
		t.Errorf("plot text %q", text)
	}
}
