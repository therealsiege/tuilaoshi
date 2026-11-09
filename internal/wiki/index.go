package wiki

import (
	"os"
	"path/filepath"
	"strings"
	"time"

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
func (s *Store) Stats() ContentStats {
	pages, _ := s.ListPages()
	var stats ContentStats
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
	return stats
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
		LastReviewed:    p.LastReviewed,
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

// DailyStats holds aggregated stats for a single day.
type DailyStats struct {
	Date          string
	CardsReviewed int
	Correct       int
	NewCards      int
}

// TodayStats computes stats from session files for today.
func (s *Store) TodayStats() DailyStats {
	today := time.Now().Format("2006-01-02")
	stats := DailyStats{Date: today}

	sessDir := filepath.Join(s.Dir, "sessions")
	entries, err := os.ReadDir(sessDir)
	if err != nil {
		return stats
	}

	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		if !strings.HasPrefix(e.Name(), today) {
			continue
		}
		data, err := os.ReadFile(filepath.Join(sessDir, e.Name()))
		if err != nil {
			continue
		}
		body := string(data)
		for _, line := range strings.Split(body, "\n") {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "| ") && !strings.HasPrefix(trimmed, "| Card") && !strings.HasPrefix(trimmed, "|---") {
				stats.CardsReviewed++
				if strings.Contains(trimmed, "correct") || strings.Contains(trimmed, "Good") || strings.Contains(trimmed, "Easy") {
					stats.Correct++
				}
			}
		}
	}
	return stats
}

// Streak counts consecutive days (ending today or yesterday) with at least one session.
func (s *Store) Streak() int {
	sessDir := filepath.Join(s.Dir, "sessions")
	entries, err := os.ReadDir(sessDir)
	if err != nil {
		return 0
	}

	// Collect unique session dates
	dates := make(map[string]bool)
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		if len(e.Name()) >= 10 {
			dates[e.Name()[:10]] = true
		}
	}

	if len(dates) == 0 {
		return 0
	}

	streak := 0
	day := time.Now()
	// Allow streak to start from today or yesterday
	if !dates[day.Format("2006-01-02")] {
		day = day.AddDate(0, 0, -1)
		if !dates[day.Format("2006-01-02")] {
			return 0
		}
	}
	for dates[day.Format("2006-01-02")] {
		streak++
		day = day.AddDate(0, 0, -1)
	}
	return streak
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
