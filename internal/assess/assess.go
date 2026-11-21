// Package assess runs a short adaptive diagnostic: for each skill it finds
// the highest HSK level at which the learner still answers most questions,
// starting from the level the learner is expected to be at.
package assess

import (
	"fmt"
	"sort"
	"strings"

	"github.com/therealsiege/tuilaoshi/internal/model"
	"github.com/therealsiege/tuilaoshi/internal/wiki"
)

const (
	// PerLevel questions are asked at each level tried.
	PerLevel = 3
	// PassMark of PerLevel right answers passes a level.
	PassMark = 2
	// startLevel is where every skill starts: CLAUDE.md calibrates the
	// learner as functional at HSK3.
	startLevel = 3
	// minItems a level needs to be tested: enough for distinct questions.
	minItems = 4
)

// Skill is one area the assessment measures, with its question cards by
// HSK level (1-6).
type Skill struct {
	Name    string // matches the skill column in overview.md
	ByLevel map[int][]model.Card
}

// Levels returns the levels with enough cards to test, lowest first.
func (s Skill) Levels() []int {
	var levels []int
	for l, cards := range s.ByLevel {
		if len(cards) >= minItems {
			levels = append(levels, l)
		}
	}
	sort.Ints(levels)
	return levels
}

// Pool returns every card of the skill, for wrong answers.
func (s Skill) Pool() []model.Card {
	var pool []model.Card
	for _, l := range s.Levels() {
		pool = append(pool, s.ByLevel[l]...)
	}
	return pool
}

// Skills builds the four assessed skills from the wiki: character meanings,
// vocabulary meanings, grammar example sentences, and the vocabulary of the
// readings.
func Skills(pages []model.Page, cards []model.Card, readingVocab func(model.Page) []model.Card) []Skill {
	level := map[string]int{}
	for _, p := range pages {
		level[p.Path] = hskLevel(p.Level)
	}
	chars := Skill{Name: "Character recognition", ByLevel: map[int][]model.Card{}}
	vocab := Skill{Name: "Vocabulary (passive)", ByLevel: map[int][]model.Card{}}
	grammar := Skill{Name: "Grammar", ByLevel: map[int][]model.Card{}}
	reading := Skill{Name: "Reading comprehension", ByLevel: map[int][]model.Card{}}
	for _, c := range cards {
		l := level[strings.SplitN(c.ID, "#", 2)[0]]
		if l == 0 {
			continue
		}
		switch c.Type {
		case model.CardCharacter:
			chars.ByLevel[l] = append(chars.ByLevel[l], c)
		case model.CardVocabulary:
			vocab.ByLevel[l] = append(vocab.ByLevel[l], c)
		case model.CardGrammar:
			for i, ex := range c.Examples {
				if zh, en, ok := wiki.ParseExample(ex); ok {
					grammar.ByLevel[l] = append(grammar.ByLevel[l], model.Card{ID: fmt.Sprintf("%s#ex%d", c.ID, i), Chinese: zh, English: en})
				}
			}
		}
	}
	for _, p := range pages {
		if p.Type == "reading" && level[p.Path] > 0 {
			reading.ByLevel[level[p.Path]] = append(reading.ByLevel[level[p.Path]], readingVocab(p)...)
		}
	}
	return []Skill{chars, vocab, grammar, reading}
}

func hskLevel(level string) int {
	var n int
	if _, err := fmt.Sscanf(level, "HSK%d", &n); err != nil || n < 1 || n > 6 {
		return 0
	}
	return n
}

// Staircase walks one skill's levels: pass a level and it tries the next one
// up, fail and it tries the next one down, until it finds the boundary.
type Staircase struct {
	levels []int
	idx    int // current position in levels
	passed map[int]bool
	failed map[int]bool
	right  int // right answers at the current level
	asked  int // questions asked at the current level
	Tally  map[int][2]int
	done   bool
}

// NewStaircase starts at the testable level closest to startLevel.
func NewStaircase(levels []int) *Staircase {
	s := &Staircase{levels: levels, passed: map[int]bool{}, failed: map[int]bool{}, Tally: map[int][2]int{}}
	if len(levels) == 0 {
		s.done = true
		return s
	}
	for i, l := range levels {
		if abs(l-startLevel) < abs(levels[s.idx]-startLevel) {
			s.idx = i
		}
	}
	return s
}

// Level is the level being tested.
func (s *Staircase) Level() int { return s.levels[s.idx] }

// Done reports whether the boundary has been found.
func (s *Staircase) Done() bool { return s.done }

// Record scores one answer at the current level and moves on once the
// level has had its questions.
func (s *Staircase) Record(correct bool) {
	if s.done {
		return
	}
	l := s.Level()
	t := s.Tally[l]
	t[1]++
	s.asked++
	if correct {
		s.right++
		t[0]++
	}
	s.Tally[l] = t
	if s.asked < PerLevel {
		return
	}
	pass := s.right >= PassMark
	s.right, s.asked = 0, 0
	if pass {
		s.passed[l] = true
		if s.idx+1 < len(s.levels) && !s.failed[s.levels[s.idx+1]] {
			s.idx++
			return
		}
	} else {
		s.failed[l] = true
		if s.idx > 0 && !s.passed[s.levels[s.idx-1]] {
			s.idx--
			return
		}
	}
	s.done = true
}

// Result is the highest level passed, or 0 when none was.
func (s *Staircase) Result() int {
	best := 0
	for l := range s.passed {
		if l > best {
			best = l
		}
	}
	return best
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}
