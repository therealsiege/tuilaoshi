package wiki

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// FrequencyFile is the learner's private frequency word list, most
// frequent first, imported by cmd/freqlist.
const FrequencyFile = "private/frequency.yaml"

// FreqWord is one word of the frequency list.
type FreqWord struct {
	Word    string `yaml:"w"`
	Count   int    `yaml:"n"`
	Pinyin  string `yaml:"py,omitempty"`
	HSK     int    `yaml:"hsk,omitempty"`
	Name    bool   `yaml:"name,omitempty"` // a name in the source (a character, place or item)
	Meaning string `yaml:"en,omitempty"`
}

// Frequency reads the frequency list; a missing list means none.
func (s *Store) Frequency() ([]FreqWord, error) {
	data, err := os.ReadFile(filepath.Join(s.Dir, FrequencyFile))
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read frequency list: %w", err)
	}
	var words []FreqWord
	if err := yaml.Unmarshal(data, &words); err != nil {
		return nil, fmt.Errorf("parse frequency list: %w", err)
	}
	return words, nil
}

// frequencyPage holds the frequency words introduced so far, as a private
// vocabulary page, so they come up in reviews like any other word.
const frequencyPage = "private/vocabulary/frequency-words.md"

// FrequencyPage reports whether a card comes from the frequency deck.
func FrequencyPage(cardID string) bool { return strings.HasPrefix(cardID, frequencyPage+"#") }

// knownHSK is the HSK level up to which the learner (functional at HSK3,
// per the wiki's profile) already knows the words.
const knownHSK = 3

// IntroduceFrequency adds the next n most frequent words the learner
// doesn't have yet (not a name, not already a card, above knownHSK) to the
// reviews, once a day, and returns how many it added.
func (s *Store) IntroduceFrequency(n int, now time.Time) (int, error) {
	study, err := s.readStudy()
	if err != nil || n <= 0 || study.FreqDay == now.Format(dateFmt) {
		return 0, err
	}
	list, err := s.Frequency()
	if err != nil || len(list) == 0 {
		return 0, err
	}
	cards, err := s.DrillableCards()
	if err != nil {
		return 0, err
	}
	known := map[string]bool{}
	for _, c := range cards {
		known[c.Chinese] = true
	}
	// Basic characters, as the list or the wiki's character pages level
	// them, so untagged words made only of them (他们, 一个) count as known.
	basic := map[rune]bool{}
	for _, w := range list {
		if r := []rune(w.Word); len(r) == 1 && w.HSK >= 1 && w.HSK <= 2 {
			basic[r[0]] = true
		}
	}
	pages, err := s.ListPages()
	if err != nil {
		return 0, err
	}
	for _, p := range pages {
		if r := []rune(p.Character); p.Type == "character" && len(r) == 1 && (p.Level == "HSK1" || p.Level == "HSK2") {
			basic[r[0]] = true
		}
	}
	allBasic := func(word string) bool {
		r := []rune(word)
		for _, c := range r {
			if !basic[c] {
				return false
			}
		}
		return len(r) > 1
	}
	var add []Word
	for _, w := range list {
		if len(add) == n {
			break
		}
		if w.Name || known[w.Word] || w.Pinyin == "" || w.Meaning == "" || (w.HSK >= 1 && w.HSK <= knownHSK) || (w.HSK == 0 && allBasic(w.Word)) {
			continue
		}
		add = append(add, Word{Chinese: w.Word, Pinyin: w.Pinyin, Meaning: shortMeaning(w.Meaning)})
	}
	added, err := s.addWords(frequencyPage, "Words by frequency (private list)", "frequency", add, now)
	if err != nil {
		return 0, err
	}
	study.FreqDay = now.Format(dateFmt)
	return added, s.writeStudy(study)
}

// shortMeaning keeps a word's first two senses.
func shortMeaning(m string) string {
	parts := strings.Split(m, "; ")
	return strings.Join(parts[:min(2, len(parts))], "; ")
}
