package journey

import "testing"

func TestLevelCurve(t *testing.T) {
	for _, c := range []struct{ mastered, level, into, need int }{
		{0, 1, 0, 5}, {4, 1, 4, 5}, {5, 2, 0, 10}, {225, 10, 0, 50}, {949, 19, 94, 95}, {950, 20, 0, 100},
	} {
		if l := LevelFor(c.mastered); l != (Level{c.level, c.into, c.need}) {
			t.Errorf("%d mastered: %+v, want level %d (%d of %d)", c.mastered, l, c.level, c.into, c.need)
		}
	}
	for m := 1; m < 5000; m++ {
		if LevelFor(m).N < LevelFor(m-1).N {
			t.Fatalf("level falls at %d mastered", m)
		}
	}
}
