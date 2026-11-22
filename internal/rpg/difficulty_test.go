package rpg

import (
	"slices"
	"testing"
)

func TestDifficultyFollowsRecentAccuracy(t *testing.T) {
	for _, c := range []struct {
		pct, answered int
		want          Difficulty
	}{{50, 10, Normal}, {50, 60, Easier}, {80, 60, Normal}, {95, 60, Harder}} {
		if got := DifficultyFor(c.pct, c.answered); got != c.want {
			t.Errorf("%d%% of %d: %d, want %d", c.pct, c.answered, got, c.want)
		}
	}
}

func TestTuneShiftsTheQuestionMix(t *testing.T) {
	all := []string{"meaning", "pinyin", "write", "listening"}
	if got := Easier.tune(all); slices.Contains(got, "write") || slices.Contains(got, "listening") || len(got) != 2 {
		t.Errorf("easier: %v", got)
	}
	if got := Easier.tune([]string{"order"}); !slices.Equal(got, []string{"order"}) {
		t.Errorf("easier with only production kinds should keep them: %v", got)
	}
	if got := Harder.tune(all); len(got) != 6 {
		t.Errorf("harder should offer production kinds twice: %v", got)
	}
	if got := Normal.tune(all); !slices.Equal(got, all) {
		t.Errorf("normal changed the mix: %v", got)
	}
}
