// Package episode turns the learner's own episodes (subtitles and audio in
// the git-ignored wiki/private/episodes) into play: each opens at a point
// on the road, its new words become a theme that reviews and battles
// favour, its scenes carry comprehension questions, its lines make
// listening drills, and companions can talk about it. This package holds
// only the rules; the episodes' text never leaves the learner's machine.
package episode

import (
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/therealsiege/tuilaoshi/internal/journey"
	"github.com/therealsiege/tuilaoshi/internal/model"
	"github.com/therealsiege/tuilaoshi/internal/wiki"
)

// Episode is one of the learner's episodes, placed on the road.
type Episode struct {
	model.Episode
	ID     string // its file or folder name, stable across runs
	Number int    // from the name ("西游记 98版 02 龙宫取宝" is 2), else its order
	Stage  int    // the tribulation at which it opens
}

// numbered finds the episode number: the last number before the title.
var numbered = regexp.MustCompile(`(\d{1,3})\D*$`)

// Load reads the learner's episodes and places each on the road.
func Load(store *wiki.Store) ([]Episode, error) {
	eps, err := store.PrivateEpisodes()
	if err != nil {
		return nil, err
	}
	out := make([]Episode, 0, len(eps))
	for i, e := range eps {
		n := i + 1
		if m := numbered.FindStringSubmatch(e.Source); m != nil {
			if v, err := strconv.Atoi(m[1]); err == nil && v > 0 {
				n = v
			}
		}
		out = append(out, Episode{Episode: e, ID: e.Source, Number: n, Stage: journey.EpisodeStage(n)})
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Number < out[j].Number })
	return out, nil
}

// Open reports whether the road has come as far as the episode: its stage
// is being worked on or behind (current 0: the pilgrimage is complete).
func (e Episode) Open(current int) bool { return current == 0 || e.Stage <= current }

// Opened returns the episodes open at the current stage.
func Opened(eps []Episode, current int) []Episode {
	var out []Episode
	for _, e := range eps {
		if e.Open(current) {
			out = append(out, e)
		}
	}
	return out
}

// themeSize caps an episode's theme, its most frequent words first.
const themeSize = 60

// Words is the episode's theme: words of the wiki (two characters or more)
// that its dialogue uses and the learner hasn't learned yet (review
// interval under journey.LearnedInterval), most used first.
func (e Episode) Words(cards []model.Card) []model.Card {
	var text strings.Builder
	for _, l := range e.Lines {
		text.WriteString(l.Chinese)
		text.WriteString("\n")
	}
	all := text.String()
	type hit struct {
		card  model.Card
		count int
	}
	var hits []hit
	seen := map[string]bool{}
	for _, c := range cards {
		if utf8.RuneCountInString(c.Chinese) < 2 || seen[c.Chinese] || c.ReviewInterval >= journey.LearnedInterval() {
			continue
		}
		if n := strings.Count(all, c.Chinese); n > 0 {
			seen[c.Chinese] = true
			hits = append(hits, hit{c, n})
		}
	}
	sort.SliceStable(hits, func(i, j int) bool { return hits[i].count > hits[j].count })
	out := make([]model.Card, 0, min(len(hits), themeSize))
	for _, h := range hits[:min(len(hits), themeSize)] {
		out = append(out, h.card)
	}
	return out
}

// Ahead puts the opened episodes' theme words first in a review queue,
// keeping the rest in order, so new episode words enter review without
// being marked one by one.
func Ahead(due []model.Card, opened []Episode, all []model.Card) []model.Card {
	theme := map[string]bool{}
	for _, e := range opened {
		for _, w := range e.Words(all) {
			theme[w.ID] = true
		}
	}
	if len(theme) == 0 {
		return due
	}
	out := make([]model.Card, 0, len(due))
	for _, c := range due {
		if theme[c.ID] {
			out = append(out, c)
		}
	}
	for _, c := range due {
		if !theme[c.ID] {
			out = append(out, c)
		}
	}
	return out
}

// OpenedNow loads the episodes and returns those the journey being played
// has reached.
func OpenedNow(store *wiki.Store) ([]Episode, error) {
	eps, err := Load(store)
	if err != nil || len(eps) == 0 {
		return nil, err
	}
	pages, err := store.ListPages()
	if err != nil {
		return nil, err
	}
	cleared, err := store.ClearedStages()
	if err != nil {
		return nil, err
	}
	current := 0
	if s := journey.CurrentStage(journey.Stages(journey.CountProgress(pages), cleared)); s != nil {
		current = s.Number
	}
	return Opened(eps, current), nil
}

// WithLatest adds the latest opened episode's theme words to a battle's
// cards, so the road quizzes what the learner has just watched.
func WithLatest(cards, all []model.Card, opened []Episode) []model.Card {
	if len(opened) == 0 {
		return cards
	}
	have := map[string]bool{}
	for _, c := range cards {
		have[c.ID] = true
	}
	out := append([]model.Card(nil), cards...)
	for _, w := range opened[len(opened)-1].Words(all) {
		if !have[w.ID] {
			out = append(out, w)
		}
	}
	return out
}
