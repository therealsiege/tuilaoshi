package camp

import (
	"sort"
	"time"
	"unicode"

	"github.com/therealsiege/tuilaoshi/internal/model"
	"github.com/therealsiege/tuilaoshi/internal/wiki"
)

// The spider's web (盘丝洞): the learner's studied words, each tied by a
// strand to every other studied word sharing one of its characters
// (学生 ── 学习 ── 习惯). Shared by the Mac's stats screen (internal/ui/stats/webview.go draws
// it) and the phone.

// WeakInterval is the review interval (days) under which a word is still
// slipping.
const WeakInterval = 4

// WalkMax caps how many words one walk of the web revives.
const WalkMax = 20

// WebWord is one studied word on the web and whether it is slipping.
type WebWord struct {
	Card    model.Card
	Lapses  int
	Overdue bool
	Weak    bool // overdue, under WeakInterval days, or a hard word
}

// Web is the graph of studied words. links[i] lists the words sharing a
// character with word i, slipping words first.
type Web struct {
	Words       []WebWord
	Links       [][]int
	Index       map[string]int // word text -> its place in Words
	Strands     int            // pairs of words sharing a character
	WeakStrands int            // strands with a slipping word at either end
}

// BuildWeb ties the studied words together by shared characters. A word
// studied twice (a character page and a vocabulary row, say) appears once,
// as its weaker card.
func BuildWeb(studied []wiki.StudiedWord, now time.Time) Web {
	w := Web{Index: map[string]int{}}
	for _, s := range studied {
		ww := WebWord{Card: s.Card, Lapses: s.Lapses, Overdue: s.Card.IsDue(now)}
		ww.Weak = ww.Overdue || s.Card.ReviewInterval < WeakInterval || s.Lapses >= wiki.HardWordLapses
		if i, ok := w.Index[s.Card.Chinese]; ok {
			if weaker(ww, w.Words[i]) {
				w.Words[i] = ww
			}
			continue
		}
		w.Index[s.Card.Chinese] = len(w.Words)
		w.Words = append(w.Words, ww)
	}

	byChar := map[rune][]int{}
	for i, ww := range w.Words {
		for _, r := range hanSet(ww.Card.Chinese) {
			byChar[r] = append(byChar[r], i)
		}
	}
	w.Links = make([][]int, len(w.Words))
	seen := make([]int, len(w.Words)) // seen[j] == i+1 once j is linked to i
	for i, ww := range w.Words {
		for _, r := range hanSet(ww.Card.Chinese) {
			for _, j := range byChar[r] {
				if j == i || seen[j] == i+1 {
					continue
				}
				seen[j] = i + 1
				w.Links[i] = append(w.Links[i], j)
				if j > i {
					w.Strands++
					if ww.Weak || w.Words[j].Weak {
						w.WeakStrands++
					}
				}
			}
		}
		sort.Slice(w.Links[i], func(a, b int) bool {
			x, y := w.Words[w.Links[i][a]], w.Words[w.Links[i][b]]
			if x.Weak != y.Weak {
				return x.Weak
			}
			return x.Card.Chinese < y.Card.Chinese
		})
	}
	return w
}

// hanSet lists the distinct Chinese characters of a word, in order.
func hanSet(s string) []rune {
	var out []rune
	for _, r := range s {
		if unicode.Is(unicode.Han, r) && !containsRune(out, r) {
			out = append(out, r)
		}
	}
	return out
}

// containsRune reports whether r is in rs.
func containsRune(rs []rune, r rune) bool {
	for _, x := range rs {
		if x == r {
			return true
		}
	}
	return false
}

// weaker reports whether a is slipping worse than b: hard words first,
// then overdue ones (longest overdue first), then the shortest interval.
func weaker(a, b WebWord) bool {
	ah, bh := a.Lapses >= wiki.HardWordLapses, b.Lapses >= wiki.HardWordLapses
	switch {
	case ah != bh:
		return ah
	case a.Overdue != b.Overdue:
		return a.Overdue
	case a.Overdue && !a.Card.NextReview.Equal(b.Card.NextReview):
		return a.Card.NextReview.Before(b.Card.NextReview)
	case a.Card.ReviewInterval != b.Card.ReviewInterval:
		return a.Card.ReviewInterval < b.Card.ReviewInterval
	case a.Lapses != b.Lapses:
		return a.Lapses > b.Lapses
	}
	return a.Card.Chinese < b.Card.Chinese
}

// Weakest is where the web opens: the weakest word with a strand, or the
// weakest word when none has one; -1 for an empty web.
func (w Web) Weakest() int {
	best, bestLinked := -1, -1
	for i, ww := range w.Words {
		if best < 0 || weaker(ww, w.Words[best]) {
			best = i
		}
		if len(w.Links[i]) > 0 && (bestLinked < 0 || weaker(ww, w.Words[bestLinked])) {
			bestLinked = i
		}
	}
	if bestLinked >= 0 {
		return bestLinked
	}
	return best
}

// Cluster is what walking the web from a centre revives: the centre if it
// is slipping, then its slipping neighbours, at most WalkMax words.
func (w Web) Cluster(centre int) []model.Card {
	if centre < 0 || centre >= len(w.Words) {
		return nil
	}
	var out []model.Card
	if w.Words[centre].Weak {
		out = append(out, w.Words[centre].Card)
	}
	for _, j := range w.Links[centre] {
		if len(out) >= WalkMax {
			break
		}
		if w.Words[j].Weak {
			out = append(out, w.Words[j].Card)
		}
	}
	return out
}

// Shared is the characters two words have in common.
func Shared(a, b string) string {
	var out []rune
	bs := hanSet(b)
	for _, r := range hanSet(a) {
		if containsRune(bs, r) {
			out = append(out, r)
		}
	}
	return string(out)
}
