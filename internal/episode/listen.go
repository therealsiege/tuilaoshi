package episode

import (
	"fmt"
	"math/rand"
	"strings"
	"unicode"
)

// A listening drill plays one line of the episode's own audio: a short line
// is typed back, a longer one is matched to its meaning.

// typeUpTo is the longest line, in characters, typed back rather than matched.
const typeUpTo = 8

// Drill is one listening question.
type Drill struct {
	Line    int // the line's index in the episode
	Type    bool
	Choices []string // for a matched line: English meanings, shuffled
	Right   int
}

// CardID names a line as a card, so a right answer pays once a day like any
// other word.
func (e Episode) CardID(line int) string { return fmt.Sprintf("episode:%s:line:%d", e.ID, line) }

// NextDrill picks a line to hear: one with sure English and its own clip.
func (e Episode) NextDrill(rng *rand.Rand) (Drill, bool) {
	if e.Audio == "" {
		return Drill{}, false
	}
	var lines []int
	for i, l := range e.Lines {
		if l.End > l.Start && hanCount(l.Chinese) >= 2 && cleanEnglish(l.English) != "" && !strings.Contains(l.English, unsureMark) {
			lines = append(lines, i)
		}
	}
	if len(lines) < choiceCount {
		return Drill{}, false
	}
	i := lines[rng.Intn(len(lines))]
	d := Drill{Line: i, Type: hanCount(e.Lines[i].Chinese) <= typeUpTo}
	if d.Type {
		return d, true
	}
	right := cleanEnglish(e.Lines[i].English)
	d.Choices = []string{right}
	seen := map[string]bool{strings.ToLower(right): true}
	for _, j := range rng.Perm(len(lines)) {
		other := cleanEnglish(e.Lines[lines[j]].English)
		if len(d.Choices) == choiceCount || seen[strings.ToLower(other)] {
			continue
		}
		seen[strings.ToLower(other)] = true
		d.Choices = append(d.Choices, other)
	}
	rng.Shuffle(len(d.Choices), func(a, b int) { d.Choices[a], d.Choices[b] = d.Choices[b], d.Choices[a] })
	for k, c := range d.Choices {
		if c == right {
			d.Right = k
		}
	}
	return d, true
}

// Heard reports whether typed matches the line, counting only Chinese
// characters (punctuation and spaces don't matter).
func Heard(line, typed string) bool { return hanOnly(line) != "" && hanOnly(line) == hanOnly(typed) }

func hanOnly(s string) string {
	var b strings.Builder
	for _, r := range s {
		if unicode.Is(unicode.Han, r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func hanCount(s string) int { return len([]rune(hanOnly(s))) }
