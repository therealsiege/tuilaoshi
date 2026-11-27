package mobile

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/therealsiege/tuilaoshi/internal/journey"
	"github.com/therealsiege/tuilaoshi/internal/model"
	"github.com/therealsiege/tuilaoshi/internal/segment"
	"github.com/therealsiege/tuilaoshi/internal/wiki"
)

// Reading on the phone: the wiki's readings and the journey episodes the
// pilgrimage has reached, split into words and glossed, as the Mac's
// reader does. Private episodes (wiki/private) are never in the app.

type readingItem struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Tag   string `json:"tag"`
}

type token struct {
	Text    string `json:"text"`
	Word    bool   `json:"word"`
	Pinyin  string `json:"pinyin,omitempty"`
	English string `json:"english,omitempty"`
	InDeck  bool   `json:"inDeck,omitempty"`
}

type readingLine struct {
	Speaker string  `json:"speaker,omitempty"`
	Tokens  []token `json:"tokens"`
	English string  `json:"english,omitempty"`
}

// Readings lists what can be read, as JSON: the wiki's readings, then the
// journey episodes reached so far.
func (c *Core) Readings() (string, error) {
	readings, err := c.store.ListPagesByType("reading")
	if err != nil {
		return "", err
	}
	out := []readingItem{}
	for _, p := range readings {
		out = append(out, readingItem{ID: "page:" + p.Path, Title: p.Title, Tag: p.Level})
	}
	eps, err := c.reachedEpisodes()
	if err != nil {
		return "", err
	}
	for _, e := range eps {
		out = append(out, readingItem{ID: "episode:" + e.Title, Title: e.Title, Tag: "journey episode"})
	}
	data, err := json.Marshal(out)
	return string(data), err
}

func (c *Core) reachedEpisodes() ([]model.Episode, error) {
	pages, err := c.store.ListPages()
	if err != nil {
		return nil, err
	}
	cleared, err := c.store.ClearedStages()
	if err != nil {
		return nil, err
	}
	return journey.ReachedEpisodes(journey.Pilgrimage(pages, cleared).Tribulation), nil
}

// Reading opens one item from Readings as JSON: its title and its lines
// split into words, each word glossed by the text's own vocabulary first
// and then the learner's decks.
func (c *Core) Reading(id string) (string, error) {
	kind, key, _ := strings.Cut(id, ":")
	var title string
	var lines []readingLine
	var vocab []model.Card
	switch kind {
	case "page":
		p, err := c.store.ReadPage(key)
		if err != nil {
			return "", err
		}
		title, vocab = p.Title, wiki.ReadingVocab(p)
		for _, l := range wiki.ReadingText(p.Body) {
			lines = append(lines, readingLine{Tokens: []token{{Text: l}}})
		}
	case "episode":
		eps, err := c.reachedEpisodes()
		if err != nil {
			return "", err
		}
		for _, e := range eps {
			if e.Title != key {
				continue
			}
			title, vocab = e.Title, journey.StoryGlossary()
			for _, l := range e.Lines {
				en := l.English
				if l.Original {
					en = "原文, from the novel: " + en
				}
				lines = append(lines, readingLine{Speaker: l.Speaker, Tokens: []token{{Text: l.Chinese}}, English: en})
			}
		}
	}
	if title == "" {
		return "", fmt.Errorf("nothing to read called %q", id)
	}
	decks, err := c.decks()
	if err != nil {
		return "", err
	}
	glossary := map[string]model.Card{}
	known := map[string]bool{}
	for w, cs := range decks {
		glossary[w], known[w] = cs[0], true
	}
	for _, v := range vocab {
		glossary[v.Chinese], known[v.Chinese] = v, true // the text's own meaning first
	}
	for i := range lines {
		var toks []token
		for _, t := range segment.Split(lines[i].Tokens[0].Text, known) {
			tok := token{Text: t.Text, Word: t.Word}
			if g, ok := glossary[t.Text]; ok && t.Word {
				tok.Pinyin, tok.English, tok.InDeck = g.Pinyin, g.English, len(decks[t.Text]) > 0
			}
			toks = append(toks, tok)
		}
		lines[i].Tokens = toks
	}
	data, err := json.Marshal(map[string]any{"title": title, "lines": lines})
	return string(data), err
}

func (c *Core) decks() (map[string][]model.Card, error) {
	cards, err := c.store.DrillableCards()
	if err != nil {
		return nil, err
	}
	decks := map[string][]model.Card{}
	for _, cd := range cards {
		decks[cd.Chinese] = append(decks[cd.Chinese], cd)
	}
	return decks, nil
}

// MarkUnknown records that the learner didn't know a word met in reading:
// each of its deck cards is reviewed as "Again", so it comes back tomorrow.
// It reports false when the word isn't in the decks.
func (c *Core) MarkUnknown(word string) (bool, error) {
	decks, err := c.decks()
	if err != nil {
		return false, err
	}
	for _, cd := range decks[word] {
		card := cd
		if err := c.store.RecordForgotten(&card, time.Now()); err != nil {
			return false, err
		}
	}
	return len(decks[word]) > 0, nil
}

// AddSentenceCard keeps a word with the line it was met in, as a card whose
// review blanks the word. It reports false when the word already has one.
func (c *Core) AddSentenceCard(word, pinyin, english, line string) (bool, error) {
	return c.store.AddSentenceCard(wiki.Word{Chinese: word, Pinyin: pinyin, Meaning: english}, line, time.Now())
}
