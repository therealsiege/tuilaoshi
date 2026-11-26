package wiki

import (
	"encoding/csv"
	"fmt"
	"io"
	"strings"

	"github.com/therealsiege/tuilaoshi/internal/model"
)

// deckHeader names the deck's columns, for Anki's and Pleco's importers.
var deckHeader = []string{"Chinese", "Pinyin", "Meaning", "Tags"}

// WriteDeck writes the learner's words as a CSV deck for Anki or Pleco:
// every character and vocabulary word reviewed in the game at least once,
// tagged with its page's HSK level and "hard" when it has been missed
// HardWordLapses times or more. It returns how many words it wrote.
func (s *Store) WriteDeck(w io.Writer) (int, error) {
	cards, err := s.DrillableCards()
	if err != nil {
		return 0, err
	}
	pages, err := s.ListPages()
	if err != nil {
		return 0, err
	}
	levels := make(map[string]string, len(pages))
	for _, p := range pages {
		levels[p.Path] = p.Level
	}
	hard, err := s.HardWords()
	if err != nil {
		return 0, err
	}
	isHard := make(map[string]bool, len(hard))
	for _, h := range hard {
		isHard[h.CardID] = true
	}

	out := csv.NewWriter(w)
	if err := out.Write(deckHeader); err != nil {
		return 0, err
	}
	n := 0
	for _, c := range cards {
		if c.Type == model.CardGrammar || (c.LastReviewed.IsZero() && !isHard[c.ID]) {
			continue
		}
		var tags []string
		if isHard[c.ID] {
			tags = append(tags, "hard")
		}
		page, _, _ := strings.Cut(c.ID, "#")
		if level := levels[page]; level != "" {
			tags = append(tags, level)
		}
		if err := out.Write([]string{c.Chinese, c.Pinyin, c.English, strings.Join(tags, " ")}); err != nil {
			return n, err
		}
		n++
	}
	out.Flush()
	if err := out.Error(); err != nil {
		return n, fmt.Errorf("write deck: %w", err)
	}
	return n, nil
}
