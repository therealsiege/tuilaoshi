package rpg

import "fmt"

// Difficulty tunes battles to recent accuracy, keeping recall effortful
// but winnable: shaky players get recognition questions and an extra
// reveal, strong ones more writing, listening and sentence building.
type Difficulty int

const (
	Easier Difficulty = iota - 1
	Normal
	Harder
)

const (
	// RecentAnswers is how many recent answers set the difficulty.
	RecentAnswers = 60
	// minAnswers is the fewest recent answers worth adapting to.
	minAnswers  = 20
	easierBelow = 70 // percent right
	harderAbove = 90
)

// DifficultyFor picks the difficulty for a recent accuracy.
func DifficultyFor(pct, answered int) Difficulty {
	switch {
	case answered < minAnswers:
		return Normal
	case pct < easierBelow:
		return Easier
	case pct > harderAbove:
		return Harder
	}
	return Normal
}

// Says describes the difficulty for the battle intro, or "" when normal.
func (d Difficulty) Says(pct int) string {
	switch d {
	case Easier:
		return fmt.Sprintf("Recent answers %d%% right: recognition questions and an extra reveal (w)", pct)
	case Harder:
		return fmt.Sprintf("Recent answers %d%% right: more writing, listening and sentences, and a stronger gauntlet", pct)
	}
	return ""
}

// production question kinds make the player produce or decode rather than recognise.
var production = map[string]bool{"write": true, "listening": true, "order": true}

// tune adjusts the question kinds on offer for the difficulty: easier drops
// the production kinds when others remain; harder offers them twice as often.
func (d Difficulty) tune(categories []string) []string {
	var out []string
	for _, c := range categories {
		switch {
		case d == Easier && production[c]:
			continue
		case d == Harder && production[c]:
			out = append(out, c, c)
		default:
			out = append(out, c)
		}
	}
	if len(out) == 0 {
		return categories
	}
	return out
}
