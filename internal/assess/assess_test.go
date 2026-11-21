package assess

import (
	"reflect"
	"strings"
	"testing"
)

// run answers each level's questions as know(level) says, and returns the result.
func run(levels []int, know func(int) bool) (*Staircase, []int) {
	s := NewStaircase(levels)
	var visited []int
	for i := 0; !s.Done() && i < 100; i++ {
		if len(visited) == 0 || visited[len(visited)-1] != s.Level() {
			visited = append(visited, s.Level())
		}
		s.Record(know(s.Level()))
	}
	return s, visited
}

func TestStaircaseFindsBoundary(t *testing.T) {
	all := []int{1, 2, 3, 4, 5, 6}
	tests := []struct {
		name    string
		levels  []int
		knowsTo int // highest level answered right
		want    int
		visits  []int
	}{
		{"climbs from HSK3 to the boundary", all, 4, 4, []int{3, 4, 5}},
		{"descends when HSK3 fails", all, 1, 1, []int{3, 2, 1}},
		{"tops out", all, 6, 6, []int{3, 4, 5, 6}},
		{"knows nothing", all, 0, 0, []int{3, 2, 1}},
		{"starts at the nearest testable level", []int{1, 2}, 2, 2, []int{2}},
	}
	for _, tt := range tests {
		s, visited := run(tt.levels, func(l int) bool { return l <= tt.knowsTo })
		if s.Result() != tt.want || !reflect.DeepEqual(visited, tt.visits) {
			t.Errorf("%s: result %d visits %v, want %d visits %v", tt.name, s.Result(), visited, tt.want, tt.visits)
		}
	}
}

func TestStaircasePassNeedsTwoOfThree(t *testing.T) {
	s := NewStaircase([]int{3, 4})
	s.Record(true)
	s.Record(false)
	s.Record(true) // 2/3 passes HSK3
	if s.Done() || s.Level() != 4 {
		t.Fatalf("after 2/3 at HSK3: done %v level %d, want to move to HSK4", s.Done(), s.Level())
	}
	s.Record(true)
	s.Record(false)
	s.Record(false) // 1/3 fails HSK4
	if !s.Done() || s.Result() != 3 {
		t.Errorf("done %v result %d, want HSK3", s.Done(), s.Result())
	}
}

func TestUpdateSectionKeepsUnmeasuredRows(t *testing.T) {
	old := "_Pending_\n\n| Skill | Estimated Level | Confidence | Last Assessed |\n|---|---|---|---|\n" +
		"| Character recognition | ? | — | — |\n| Speaking | HSK3 | self | 2026-01-01 |\n"
	results := []Result{
		{Skill: "Character recognition", Level: 4, Levels: []int{1, 2, 3, 4, 5, 6}, Right: 3, Asked: 3},
		{Skill: "Grammar", Level: 2, Levels: []int{1, 2, 3}, Right: 2, Asked: 3},
	}
	got := UpdateSection(old, results, "2026-11-21")
	for _, want := range []string{
		"| Character recognition | HSK4 | 3/3 right at HSK4 | 2026-11-21 |",
		"| Speaking | HSK3 | self | 2026-01-01 |",
		"| Grammar | HSK2 | 2/3 right at HSK2 | 2026-11-21 |",
		"Priority areas, lowest first: Grammar (HSK2), Character recognition (HSK4).",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in:\n%s", want, got)
		}
	}
	if strings.Contains(got, "_Pending_") {
		t.Errorf("old note kept:\n%s", got)
	}
}

func TestLabel(t *testing.T) {
	for r, want := range map[*Result]string{
		{Level: 0, Levels: []int{2, 3}}:    "below HSK2",
		{Level: 3, Levels: []int{1, 2, 3}}: "HSK3+ (highest in wiki)",
		{Level: 4, Levels: []int{3, 4, 5}}: "HSK4",
		{Levels: nil}:                      "?",
	} {
		if got := r.Label(); got != want {
			t.Errorf("Label(%+v) = %q, want %q", *r, got, want)
		}
	}
}
