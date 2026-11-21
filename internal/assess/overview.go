package assess

import (
	"fmt"
	"sort"
	"strings"
)

// Result is one skill's outcome.
type Result struct {
	Skill  string
	Level  int   // highest level passed; 0 for none
	Levels []int // levels the skill could be tested at
	Right  int   // right answers at Level (or at the lowest level tried, when none passed)
	Asked  int
}

// Label is the level as written in overview.md.
func (r Result) Label() string {
	switch {
	case len(r.Levels) == 0:
		return "?"
	case r.Level == 0:
		return fmt.Sprintf("below HSK%d", r.Levels[0])
	case r.Level == r.Levels[len(r.Levels)-1]:
		return fmt.Sprintf("HSK%d+ (highest in wiki)", r.Level)
	}
	return fmt.Sprintf("HSK%d", r.Level)
}

// FromStaircase summarizes a finished staircase.
func FromStaircase(skill Skill, s *Staircase) Result {
	r := Result{Skill: skill.Name, Level: s.Result(), Levels: skill.Levels()}
	at := r.Level
	if at == 0 && len(r.Levels) > 0 {
		at = r.Levels[0]
	}
	t := s.Tally[at]
	r.Right, r.Asked = t[0], t[1]
	return r
}

// UpdateSection rewrites the "Current Level Assessment" section body:
// measured skills replace their table rows, other rows stay as they were,
// and priority areas are listed lowest first. Skills that couldn't be
// tested leave their rows alone.
func UpdateSection(old string, results []Result, date string) string {
	var tested []Result
	byName := map[string]Result{}
	for _, r := range results {
		if len(r.Levels) > 0 {
			tested = append(tested, r)
			byName[r.Skill] = r
		}
	}
	results = tested
	var b strings.Builder
	fmt.Fprintf(&b, "_Last in-app diagnostic: %s (tuilaoshi, a to assess)_\n\n", date)
	b.WriteString("| Skill | Estimated Level | Confidence | Last Assessed |\n")
	b.WriteString("|-------|----------------|------------|---------------|\n")
	written := map[string]bool{}
	for _, line := range strings.Split(old, "\n") {
		cells := strings.Split(strings.Trim(strings.TrimSpace(line), "|"), "|")
		name := strings.TrimSpace(cells[0])
		if !strings.HasPrefix(strings.TrimSpace(line), "|") || name == "Skill" || strings.HasPrefix(name, "---") {
			continue
		}
		if r, ok := byName[name]; ok {
			b.WriteString(row(r, date))
			written[name] = true
		} else {
			b.WriteString(strings.TrimSpace(line) + "\n")
		}
	}
	for _, r := range results {
		if !written[r.Skill] {
			b.WriteString(row(r, date))
		}
	}

	sorted := append([]Result{}, results...)
	sort.SliceStable(sorted, func(i, j int) bool { return sorted[i].Level < sorted[j].Level })
	b.WriteString("\nPriority areas, lowest first: ")
	var parts []string
	for _, r := range sorted {
		if len(r.Levels) > 0 {
			parts = append(parts, fmt.Sprintf("%s (%s)", r.Skill, r.Label()))
		}
	}
	b.WriteString(strings.Join(parts, ", ") + ".\n")
	return b.String()
}

func row(r Result, date string) string {
	confidence := "not enough cards to test"
	if r.Asked > 0 {
		confidence = fmt.Sprintf("%d/%d right at HSK%d", r.Right, r.Asked, max(r.Level, r.Levels[0]))
	}
	return fmt.Sprintf("| %s | %s | %s | %s |\n", r.Skill, r.Label(), confidence, date)
}
