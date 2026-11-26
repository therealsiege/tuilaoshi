package assess

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// MockPassPercent of the maximum passes a mock test, as on the HSK
// (180 of 300 at HSK4).
const MockPassPercent = 60

// SectionScore is the points won in one section of a mock test.
type SectionScore struct {
	Points, Max int
	Skipped     bool // no audio for listening, say: left out of the total
}

// Score is the section's mark out of 100, as the HSK reports each section.
func (s SectionScore) Score() int {
	if s.Max == 0 {
		return 0
	}
	return (100*s.Points + s.Max/2) / s.Max
}

// MockTotal adds the section marks of the sections sat, and the most they
// could have been.
func MockTotal(scores [NumSections]SectionScore) (total, most int) {
	for _, s := range scores {
		if s.Skipped {
			continue
		}
		total += s.Score()
		most += 100
	}
	return total, most
}

// MockPassed reports whether total reaches the pass mark of most.
func MockPassed(total, most int) bool {
	return most > 0 && total*100 >= MockPassPercent*most
}

// Estimate is the level check's overall estimate: the middle of the skill
// levels it found (the lower middle for an even count), or 0 with none.
// A skill found below its lowest tested level counts as the level under it.
func Estimate(results []Result) int {
	var levels []int
	for _, r := range results {
		switch {
		case r.Asked == 0 || len(r.Levels) == 0:
		case r.Level == 0:
			levels = append(levels, r.Levels[0]-1)
		default:
			levels = append(levels, r.Level)
		}
	}
	return middle(levels)
}

var hskLabel = regexp.MustCompile(`(below )?HSK(\d)`)

// EstimateFromSection reads the same estimate from the assessment table in
// overview.md, for when the check was run on an earlier day.
func EstimateFromSection(section string) int {
	var levels []int
	for _, line := range strings.Split(section, "\n") {
		cells := strings.Split(strings.Trim(strings.TrimSpace(line), "|"), "|")
		if !strings.HasPrefix(strings.TrimSpace(line), "|") || len(cells) < 2 {
			continue
		}
		m := hskLabel.FindStringSubmatch(cells[1])
		if m == nil {
			continue
		}
		n, err := strconv.Atoi(m[2])
		if err != nil {
			continue
		}
		if m[1] != "" {
			n--
		}
		levels = append(levels, n)
	}
	return middle(levels)
}

func middle(levels []int) int {
	if len(levels) == 0 {
		return 0
	}
	sort.Ints(levels)
	return levels[(len(levels)-1)/2]
}

// Verdict compares a sitting with the level check's estimate.
func Verdict(level int, passed bool, estimate int) string {
	switch {
	case estimate == 0:
		return "No level check to compare with yet: run it from the intro."
	case passed && estimate >= level:
		return fmt.Sprintf("The level check put you at HSK%d; passing HSK%d bears it out.", estimate, level)
	case passed:
		return fmt.Sprintf("The level check put you at HSK%d, yet you passed HSK%d: it looks low.", estimate, level)
	case estimate >= level:
		return fmt.Sprintf("The level check put you at HSK%d, but HSK%d fell short: it looks high.", estimate, level)
	}
	return fmt.Sprintf("The level check put you at HSK%d, below HSK%d, and this sitting agrees.", estimate, level)
}
