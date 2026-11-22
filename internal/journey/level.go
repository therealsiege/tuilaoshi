package journey

// The learner's level comes only from mastered characters, so the one way
// up is spaced recall. Early levels come quickly; each later one needs a
// few more characters than the last (level n needs 5·n·(n-1)/2 in all:
// 5 for level 2, 225 for level 10, 950 for level 20).

// Level is where the learner stands on the level curve.
type Level struct {
	N    int // 1 and up
	Into int // characters mastered since reaching this level
	Need int // characters this level spans before the next
}

// levelStart is how many mastered characters reach level n.
func levelStart(n int) int { return 5 * n * (n - 1) / 2 }

// LevelFor places a mastered count on the level curve.
func LevelFor(mastered int) Level {
	n := 1
	for levelStart(n+1) <= mastered {
		n++
	}
	return Level{N: n, Into: mastered - levelStart(n), Need: levelStart(n+1) - levelStart(n)}
}

// HPPerLevel is the max HP each level adds to the player in battle.
const HPPerLevel = 4

// LevelHP is the player's max HP at level n, before gear.
func LevelHP(n int) int { return 100 + HPPerLevel*(n-1) }
