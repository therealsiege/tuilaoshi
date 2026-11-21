package wiki

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/cj/mandarin-wiki/internal/model"
)

// IndexEntry represents a single item in index.md.
type IndexEntry struct {
	Link        string // e.g. "grammar/ba-construction"
	Description string
}

// ReadIndex returns the raw content of index.md.
func (s *Store) ReadIndex() (string, error) {
	data, err := os.ReadFile(filepath.Join(s.Dir, "index.md"))
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// ContentStats returns summary counts from the wiki.
type ContentStats struct {
	TotalPages     int
	GrammarPages   int
	VocabPages     int
	CharacterPages int
	ReadingPages   int
	SessionPages   int
	DueForReview   int
}

// Stats computes summary statistics about the wiki content.
func (s *Store) Stats() (ContentStats, error) {
	var stats ContentStats
	pages, err := s.ListPages()
	if err != nil {
		return stats, err
	}
	stats.TotalPages = len(pages)
	for _, p := range pages {
		switch p.Type {
		case "grammar":
			stats.GrammarPages++
		case "vocabulary":
			stats.VocabPages++
		case "character":
			stats.CharacterPages++
		case "reading":
			stats.ReadingPages++
		case "session":
			stats.SessionPages++
		}
	}
	return stats, nil
}

// CardsFromPages converts wiki pages into flashcards for drilling.
// Vocabulary pages with tables produce one card per table row.
func CardsFromPages(pages []model.Page) []model.Card {
	var cards []model.Card
	for _, p := range pages {
		cc := pageToCards(p)
		cards = append(cards, cc...)
	}
	return cards
}

func pageToCards(p model.Page) []model.Card {
	base := model.Card{
		ID:             p.Path,
		Introduced:     p.Introduced,
		LastReviewed:   p.LastReviewed,
		ReviewInterval: p.ReviewInterval,
		NextReview:     p.NextReview,
	}

	switch p.Type {
	case "character":
		c := base
		c.Type = model.CardCharacter
		c.Chinese = p.Character
		c.Pinyin = p.Pinyin
		c.English = p.Meaning
		c.Examples = extractExampleSentences(p.Body)
		if c.Chinese != "" || c.English != "" {
			return []model.Card{c}
		}
	case "vocabulary":
		cards := extractAllVocabCards(p.Body, base)
		if len(cards) > 0 {
			return cards
		}
		// Fallback to single card extraction
		c := base
		c.Type = model.CardVocabulary
		c.Chinese = extractFirstVocabChinese(p.Body)
		c.Pinyin = extractFirstVocabPinyin(p.Body)
		c.English = extractFirstVocabEnglish(p.Body)
		c.Examples = extractExampleSentences(p.Body)
		if c.Chinese != "" || c.English != "" {
			return []model.Card{c}
		}
	case "grammar":
		c := base
		c.Type = model.CardGrammar
		c.Chinese = extractGrammarPattern(p.Body)
		c.English = p.Title
		c.Examples = extractExampleSentences(p.Body)
		if c.Chinese != "" || c.English != "" {
			return []model.Card{c}
		}
	}
	return nil
}

// extractAllVocabCards extracts one card per table row from a vocabulary page.
func extractAllVocabCards(body string, base model.Card) []model.Card {
	var cards []model.Card
	for _, line := range strings.Split(body, "\n") {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "| ") {
			continue
		}
		// Skip header and separator rows
		if strings.HasPrefix(trimmed, "| Word") || strings.HasPrefix(trimmed, "|---") ||
			strings.HasPrefix(trimmed, "| --") {
			continue
		}
		parts := strings.Split(trimmed, "|")
		if len(parts) < 4 {
			continue
		}
		chinese := strings.TrimSpace(parts[1])
		pinyin := strings.TrimSpace(parts[2])
		english := strings.TrimSpace(parts[3])
		if chinese == "" {
			continue
		}
		c := base
		c.Type = model.CardVocabulary
		c.ID = base.ID + "#" + chinese
		c.Chinese = chinese
		c.Pinyin = pinyin
		c.English = english
		if len(parts) >= 5 {
			example := strings.TrimSpace(parts[4])
			if example != "" {
				c.Examples = []string{example}
			}
		}
		cards = append(cards, c)
	}
	return cards
}

// extractFirstVocabChinese pulls the first Chinese word from a vocabulary page body.
func extractFirstVocabChinese(body string) string {
	for _, line := range strings.Split(body, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "| ") && !strings.HasPrefix(trimmed, "| Word") && !strings.HasPrefix(trimmed, "|---") {
			parts := strings.Split(trimmed, "|")
			if len(parts) >= 3 {
				return strings.TrimSpace(parts[1])
			}
		}
		if strings.HasPrefix(strings.ToLower(trimmed), "- **") {
			val := extractFieldValue(trimmed)
			if val != "" && containsCJK(val) {
				return val
			}
		}
	}
	return ""
}

func extractFirstVocabPinyin(body string) string {
	for _, line := range strings.Split(body, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "| ") && !strings.HasPrefix(trimmed, "| Word") && !strings.HasPrefix(trimmed, "|---") {
			parts := strings.Split(trimmed, "|")
			if len(parts) >= 4 {
				return strings.TrimSpace(parts[2])
			}
		}
	}
	return ""
}

func extractFirstVocabEnglish(body string) string {
	for _, line := range strings.Split(body, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "| ") && !strings.HasPrefix(trimmed, "| Word") && !strings.HasPrefix(trimmed, "|---") {
			parts := strings.Split(trimmed, "|")
			if len(parts) >= 5 {
				return strings.TrimSpace(parts[3])
			}
		}
	}
	return ""
}

func extractGrammarPattern(body string) string {
	for _, line := range strings.Split(body, "\n") {
		trimmed := strings.TrimSpace(line)
		lower := strings.ToLower(trimmed)
		if strings.HasPrefix(lower, "- **pattern") || strings.HasPrefix(lower, "- **structure") {
			return extractFieldValue(trimmed)
		}
	}
	return ""
}

func extractExampleSentences(body string) []string {
	var examples []string
	inExamples := false
	for _, line := range strings.Split(body, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "## ") && strings.Contains(strings.ToLower(trimmed), "example") {
			inExamples = true
			continue
		}
		if strings.HasPrefix(trimmed, "## ") {
			inExamples = false
			continue
		}
		if inExamples && trimmed != "" && !strings.HasPrefix(trimmed, "<!--") {
			examples = append(examples, trimmed)
		}
	}
	return examples
}

func containsCJK(s string) bool {
	for _, r := range s {
		if r >= 0x4E00 && r <= 0x9FFF {
			return true
		}
	}
	return false
}
