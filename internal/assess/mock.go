package assess

import (
	"math/rand"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/charmbracelet/lipgloss"

	"github.com/therealsiege/tuilaoshi/internal/model"
	"github.com/therealsiege/tuilaoshi/internal/wiki"
)

// The sections of the 模拟考 mock test, in the order they are sat.
const (
	Listening = iota
	Reading
	Writing
	NumSections
)

// SectionNames name the sections on screen and in saved sittings.
var SectionNames = [NumSections]string{"Listening 听力", "Reading 阅读", "Writing 书写"}

// SectionTimes are the minutes each section allows, about 15 in all.
var SectionTimes = [NumSections]time.Duration{4 * time.Minute, 7 * time.Minute, 4 * time.Minute}

// Item counts per section, scaled down from a real HSK paper.
const (
	ListeningItems = 10
	GapItems       = 10
	PassageItems   = 5
	TileItems      = 3
	maxTiles       = 6
	// maxChoiceWidth keeps an English choice on one line at 80 columns.
	maxChoiceWidth = 70
	// maxPassage characters keep a passage to two lines at 80 columns.
	maxPassage = 72
	// enoughWords and enoughSentences make a full paper; below them the
	// test reaches down a level for more.
	enoughWords     = 12
	enoughSentences = 30
)

// ItemKind is the kind of question an item asks.
type ItemKind int

// The kinds of item.
const (
	HeardWord    ItemKind = iota // which word was in the sentence heard
	HeardMeaning                 // what the sentence heard means
	Gap                          // which word fills the gap
	Passage                      // which sentence the passage says
	Tiles                        // arrange word tiles into the sentence
	Compose                      // write a sentence with a word, graded by Claude
)

// Item is one question of the mock test.
type Item struct {
	Kind    ItemKind
	Audio   string // the sentence spoken, for listening
	Text    string // Chinese shown: a gapped sentence or a passage
	Prompt  string
	Choices []string
	Correct int
	Tiles   []string // shuffled, for Tiles
	Answer  string   // the sentence (listening, tiles) or the word to use (compose)
	English string   // shown once answered
}

// Mock is a generated paper: its items per section, and the optional
// sentence for Claude to grade.
type Mock struct {
	Level    int
	Sections [NumSections][]Item
	Compose  *Item
}

type word struct{ zh, py, en string }

type sentence struct{ zh, en, page, word string }

// pools are the words and sentences a paper is made from.
type pools struct {
	words    []word
	withWord []sentence // vocabulary examples, each with the word it shows
	withEn   []sentence // example lines with their English
	known    map[string]bool
	used     map[string]bool
}

// BuildMock makes a paper at an HSK level from the wiki's own pages at
// that level: vocabulary table examples become gap and listening items,
// example lines with English become meaning and passage items, and short
// sentences become tile items. Wrong choices are words of the same level
// and kind. When the level has too little, lower levels fill in.
func BuildMock(level int, pages []model.Page, cards []model.Card, rng *rand.Rand) Mock {
	lo := level
	p := collect(level, lo, pages, cards)
	for lo > 1 && (len(p.words) < enoughWords || len(p.withWord)+len(p.withEn) < enoughSentences) {
		lo--
		p = collect(level, lo, pages, cards)
	}
	shuffle(rng, p.words)
	shuffle(rng, p.withWord)
	shuffle(rng, p.withEn)
	m := Mock{Level: level}

	// Tile sentences are the rarest, so they are picked first.
	for _, s := range append(append([]sentence{}, p.withWord...), p.withEn...) {
		if len(m.Sections[Writing]) >= TileItems {
			break
		}
		if it, ok := p.tiles(s, rng); ok {
			m.Sections[Writing] = append(m.Sections[Writing], it)
		}
	}

	for _, s := range p.withEn {
		if len(m.Sections[Listening]) >= ListeningItems*3/5 {
			break
		}
		if it, ok := p.meaning(s, rng); ok {
			m.Sections[Listening] = append(m.Sections[Listening], it)
		}
	}
	for _, s := range p.withWord {
		if len(m.Sections[Listening]) >= ListeningItems {
			break
		}
		if it, ok := p.heard(s, rng); ok {
			m.Sections[Listening] = append(m.Sections[Listening], it)
		}
	}
	for _, s := range p.withEn {
		if len(m.Sections[Listening]) >= ListeningItems {
			break
		}
		if it, ok := p.meaning(s, rng); ok {
			m.Sections[Listening] = append(m.Sections[Listening], it)
		}
	}

	m.Sections[Reading] = p.passages(rng)
	for _, s := range p.withWord {
		if len(m.Sections[Reading]) >= GapItems+PassageItems {
			break
		}
		if it, ok := p.gap(s, rng); ok {
			m.Sections[Reading] = append(m.Sections[Reading], it)
		}
	}
	// Gaps first, then the passages, as on the paper.
	n := len(m.Sections[Reading])
	reading := make([]Item, 0, n)
	for _, kind := range []ItemKind{Gap, Passage} {
		for _, it := range m.Sections[Reading] {
			if it.Kind == kind {
				reading = append(reading, it)
			}
		}
	}
	m.Sections[Reading] = reading

	if len(p.words) > 0 {
		w := p.words[0]
		m.Compose = &Item{Kind: Compose, Answer: w.zh, English: w.en,
			Prompt: "Write one short sentence using " + w.zh + " (" + w.py + ", " + w.en + ")."}
	}
	return m
}

// collect gathers the words and sentences of pages from level lo to level.
func collect(level, lo int, pages []model.Page, cards []model.Card) pools {
	in := map[string]bool{}
	for _, pg := range pages {
		if l := hskLevel(pg.Level); l >= lo && l <= level {
			in[pg.Path] = true
		}
	}
	p := pools{known: map[string]bool{}, used: map[string]bool{}}
	seenWord, seenSent := map[string]bool{}, map[string]bool{}
	for _, c := range cards {
		if c.Type != model.CardVocabulary {
			continue
		}
		p.known[c.Chinese] = true
		zh := c.Chinese
		if !in[strings.SplitN(c.ID, "#", 2)[0]] || !allHan(zh) || utf8.RuneCountInString(zh) > 4 || seenWord[zh] {
			continue
		}
		seenWord[zh] = true
		p.words = append(p.words, word{zh: zh, py: c.Pinyin, en: c.English})
		if len(c.Examples) > 0 {
			ex := strings.TrimSpace(c.Examples[0])
			if strings.Contains(ex, zh) && ex != zh && hasHan(ex) && !seenSent[ex] {
				seenSent[ex] = true
				p.withWord = append(p.withWord, sentence{zh: ex, word: zh})
			}
		}
	}
	for _, pg := range pages {
		if !in[pg.Path] || pg.Type == "session" {
			continue
		}
		for _, line := range strings.Split(pg.Body, "\n") {
			if !strings.HasPrefix(strings.TrimSpace(line), "- ") {
				continue
			}
			zh, en, ok := wiki.ParseExample(line)
			if !ok || !isSentence(zh) || seenSent[zh] {
				continue
			}
			seenSent[zh] = true
			p.withEn = append(p.withEn, sentence{zh: zh, en: en, page: pg.Path})
		}
	}
	// Example lines that use a word of the level can gap that word too:
	// pages such as HSK5 word lists have no example column.
	for _, s := range p.withEn {
		best := ""
		for _, w := range p.words {
			if utf8.RuneCountInString(w.zh) > max(1, utf8.RuneCountInString(best)) && strings.Contains(s.zh, w.zh) {
				best = w.zh
			}
		}
		if best != "" {
			s.word = best
			p.withWord = append(p.withWord, s)
		}
	}
	return p
}

// isSentence tells an example sentence from a compound listed with its meaning.
func isSentence(zh string) bool {
	return utf8.RuneCountInString(zh) >= 5 && strings.ContainsAny(zh, "。！？") && hasHan(zh)
}

// heard asks which word was in a sentence heard.
func (p *pools) heard(s sentence, rng *rand.Rand) (Item, bool) {
	if p.used[s.zh] {
		return Item{}, false
	}
	wrong := p.distractors(s.word, s.zh, rng)
	if len(wrong) < 3 {
		return Item{}, false
	}
	p.used[s.zh] = true
	choices, correct := arrange(s.word, wrong, rng)
	return Item{Kind: HeardWord, Audio: s.zh, Answer: s.zh, Prompt: "Which word did you hear?",
		Choices: choices, Correct: correct}, true
}

// meaning asks what a sentence heard means.
func (p *pools) meaning(s sentence, rng *rand.Rand) (Item, bool) {
	if p.used[s.zh] || lipgloss.Width(s.en) > maxChoiceWidth {
		return Item{}, false
	}
	wrong := p.otherEnglish(s.en, "", rng)
	if len(wrong) < 3 {
		return Item{}, false
	}
	p.used[s.zh] = true
	choices, correct := arrange(s.en, wrong, rng)
	return Item{Kind: HeardMeaning, Audio: s.zh, Answer: s.zh, English: s.en,
		Prompt: "What did the speaker say?", Choices: choices, Correct: correct}, true
}

// gap blanks a sentence's word and asks which word fills it.
func (p *pools) gap(s sentence, rng *rand.Rand) (Item, bool) {
	if p.used[s.zh] {
		return Item{}, false
	}
	wrong := p.distractors(s.word, s.zh, rng)
	if len(wrong) < 3 {
		return Item{}, false
	}
	p.used[s.zh] = true
	choices, correct := arrange(s.word, wrong, rng)
	return Item{Kind: Gap, Text: strings.Replace(s.zh, s.word, "____", 1), Answer: s.zh,
		Prompt: "Which word fills the gap?", Choices: choices, Correct: correct}, true
}

// passages makes up to PassageItems short passages of two sentences from
// one page each, asking which English sentence the passage says.
func (p *pools) passages(rng *rand.Rand) []Item {
	byPage := map[string][]sentence{}
	var order []string
	for _, s := range p.withEn {
		if p.used[s.zh] {
			continue
		}
		if byPage[s.page] == nil {
			order = append(order, s.page)
		}
		byPage[s.page] = append(byPage[s.page], s)
	}
	var items []Item
	for _, page := range order {
		ss := byPage[page]
		for i := 0; i+1 < len(ss) && len(items) < PassageItems; i += 2 {
			if utf8.RuneCountInString(ss[i].zh+ss[i+1].zh) > maxPassage {
				continue
			}
			var fit []sentence
			for _, c := range ss[i : i+2] {
				if lipgloss.Width(c.en) <= maxChoiceWidth {
					fit = append(fit, c)
				}
			}
			if len(fit) == 0 {
				continue
			}
			pick := fit[rng.Intn(len(fit))]
			wrong := p.otherEnglish(pick.en, page, rng)
			if len(wrong) < 3 {
				continue
			}
			p.used[ss[i].zh], p.used[ss[i+1].zh] = true, true
			choices, correct := arrange(pick.en, wrong, rng)
			items = append(items, Item{Kind: Passage, Text: ss[i].zh + ss[i+1].zh, English: ss[i].en + " " + ss[i+1].en,
				Prompt: "Which of these does the passage say?", Choices: choices, Correct: correct})
		}
	}
	return items
}

// distractors are up to three wrong words for target: of the same kind
// (verbs for a verb) and length where possible, and not in the sentence.
func (p *pools) distractors(target, avoid string, rng *rand.Rand) []string {
	var tw word
	for _, w := range p.words {
		if w.zh == target {
			tw = w
		}
	}
	var ranked [3][]string
	for _, w := range p.words {
		if strings.Contains(avoid, w.zh) || strings.Contains(w.zh, target) || strings.Contains(target, w.zh) {
			continue
		}
		r := 2
		if isVerb(w.en) == isVerb(tw.en) {
			r = 1
			if utf8.RuneCountInString(w.zh) == utf8.RuneCountInString(target) {
				r = 0
			}
		}
		ranked[r] = append(ranked[r], w.zh)
	}
	var out []string
	for _, group := range ranked {
		shuffle(rng, group)
		out = append(out, group...)
	}
	return out[:min(3, len(out))]
}

// otherEnglish are up to three other short English sentences, from pages
// other than page when one is given.
func (p *pools) otherEnglish(en, page string, rng *rand.Rand) []string {
	var others []string
	for _, s := range p.withEn {
		if s.en != en && (page == "" || s.page != page) && lipgloss.Width(s.en) <= maxChoiceWidth && !contains(others, s.en) {
			others = append(others, s.en)
		}
	}
	shuffle(rng, others)
	return others[:min(3, len(others))]
}

// arrange shuffles the right answer in among the wrong ones.
func arrange(right string, wrong []string, rng *rand.Rand) ([]string, int) {
	choices := append([]string{right}, wrong...)
	shuffle(rng, choices)
	for i, c := range choices {
		if c == right {
			return choices, i
		}
	}
	return choices, 0
}

func shuffle[T any](rng *rand.Rand, s []T) {
	rng.Shuffle(len(s), func(i, j int) { s[i], s[j] = s[j], s[i] })
}

func isVerb(en string) bool { return strings.HasPrefix(strings.ToLower(en), "to ") }

func allHan(s string) bool {
	for _, r := range s {
		if !unicode.Is(unicode.Han, r) {
			return false
		}
	}
	return s != ""
}

func hasHan(s string) bool {
	return strings.IndexFunc(s, func(r rune) bool { return unicode.Is(unicode.Han, r) }) >= 0
}

func contains(list []string, s string) bool {
	for _, l := range list {
		if l == s {
			return true
		}
	}
	return false
}
