package journey

// The learner's level comes only from characters held in memory, so the
// one way up is spaced recall: a character counts once when learned (a
// review has moved it to a 2-day interval) and again when mastered (7
// days). Early levels come quickly; each later one needs a few more points
// than the last (level n needs 5·n·(n-1)/2 in all: 5 for level 2, 225 for
// level 10, 950 for level 20).

// Level is where the learner stands on the level curve.
type Level struct {
	N    int // 1 and up
	Into int // characters mastered since reaching this level
	Need int // characters this level spans before the next
}

// levelStart is how many mastered characters reach level n.
func levelStart(n int) int { return 5 * n * (n - 1) / 2 }

// LevelPoints are the points the level curve counts: learned characters,
// plus mastered ones again.
func (c Counts) LevelPoints() int { return c.Learned + c.Mastered }

// LevelFor places level points (see LevelPoints) on the level curve.
func LevelFor(points int) Level {
	n := 1
	for levelStart(n+1) <= points {
		n++
	}
	return Level{N: n, Into: points - levelStart(n), Need: levelStart(n+1) - levelStart(n)}
}

// HPPerLevel is the max HP each level adds to the player in battle.
const HPPerLevel = 4

// LevelHP is the player's max HP at level n, before gear.
func LevelHP(n int) int { return 100 + HPPerLevel*(n-1) }
