package stats

import (
	"sort"
	"time"
	"unicode"

	"github.com/therealsiege/tuilaoshi/internal/model"
	"github.com/therealsiege/tuilaoshi/internal/wiki"
)

// The spider's web (盘丝洞): the learner's studied words, each tied by a
// strand to every other studied word sharing one of its characters
// (学生 ── 学习 ── 习惯). This file builds the web; webview.go draws it.

// weakInterval is the review interval (days) under which a word is still
// slipping.
const weakInterval = 4

// walkMax caps how many words one walk of the web revives.
const walkMax = 20

// webWord is one studied word on the web and whether it is slipping.
type webWord struct {
	card    model.Card
	lapses  int
	overdue bool
	weak    bool // overdue, under weakInterval days, or a hard word
}

// web is the graph of studied words. links[i] lists the words sharing a
// character with word i, slipping words first.
type web struct {
	words       []webWord
	links       [][]int
	index       map[string]int // word text -> its place in words
	strands     int            // pairs of words sharing a character
	weakStrands int            // strands with a slipping word at either end
}

// buildWeb ties the studied words together by shared characters. A word
// studied twice (a character page and a vocabulary row, say) appears once,
// as its weaker card.
func buildWeb(studied []wiki.StudiedWord, now time.Time) web {
	w := web{index: map[string]int{}}
	for _, s := range studied {
		ww := webWord{card: s.Card, lapses: s.Lapses, overdue: s.Card.IsDue(now)}
		ww.weak = ww.overdue || s.Card.ReviewInterval < weakInterval || s.Lapses >= wiki.HardWordLapses
		if i, ok := w.index[s.Card.Chinese]; ok {
			if weaker(ww, w.words[i]) {
				w.words[i] = ww
			}
			continue
		}
		w.index[s.Card.Chinese] = len(w.words)
		w.words = append(w.words, ww)
	}

	byChar := map[rune][]int{}
	for i, ww := range w.words {
		for _, r := range hanSet(ww.card.Chinese) {
			byChar[r] = append(byChar[r], i)
		}
	}
	w.links = make([][]int, len(w.words))
	seen := make([]int, len(w.words)) // seen[j] == i+1 once j is linked to i
	for i, ww := range w.words {
		for _, r := range hanSet(ww.card.Chinese) {
			for _, j := range byChar[r] {
				if j == i || seen[j] == i+1 {
					continue
				}
				seen[j] = i + 1
				w.links[i] = append(w.links[i], j)
				if j > i {
					w.strands++
					if ww.weak || w.words[j].weak {
						w.weakStrands++
					}
				}
			}
		}
		sort.Slice(w.links[i], func(a, b int) bool {
			x, y := w.words[w.links[i][a]], w.words[w.links[i][b]]
			if x.weak != y.weak {
				return x.weak
			}
			return x.card.Chinese < y.card.Chinese
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
func weaker(a, b webWord) bool {
	ah, bh := a.lapses >= wiki.HardWordLapses, b.lapses >= wiki.HardWordLapses
	switch {
	case ah != bh:
		return ah
	case a.overdue != b.overdue:
		return a.overdue
	case a.overdue && !a.card.NextReview.Equal(b.card.NextReview):
		return a.card.NextReview.Before(b.card.NextReview)
	case a.card.ReviewInterval != b.card.ReviewInterval:
		return a.card.ReviewInterval < b.card.ReviewInterval
	case a.lapses != b.lapses:
		return a.lapses > b.lapses
	}
	return a.card.Chinese < b.card.Chinese
}

// weakest is where the web opens: the weakest word with a strand, or the
// weakest word when none has one; -1 for an empty web.
func (w web) weakest() int {
	best, bestLinked := -1, -1
	for i, ww := range w.words {
		if best < 0 || weaker(ww, w.words[best]) {
			best = i
		}
		if len(w.links[i]) > 0 && (bestLinked < 0 || weaker(ww, w.words[bestLinked])) {
			bestLinked = i
		}
	}
	if bestLinked >= 0 {
		return bestLinked
	}
	return best
}

// cluster is what walking the web from a centre revives: the centre if it
// is slipping, then its slipping neighbours, at most walkMax words.
func (w web) cluster(centre int) []model.Card {
	if centre < 0 || centre >= len(w.words) {
		return nil
	}
	var out []model.Card
	if w.words[centre].weak {
		out = append(out, w.words[centre].card)
	}
	for _, j := range w.links[centre] {
		if len(out) >= walkMax {
			break
		}
		if w.words[j].weak {
			out = append(out, w.words[j].card)
		}
	}
	return out
}

// shared is the characters two words have in common.
func shared(a, b string) string {
	var out []rune
	bs := hanSet(b)
	for _, r := range hanSet(a) {
		if containsRune(bs, r) {
			out = append(out, r)
		}
	}
	return string(out)
}
