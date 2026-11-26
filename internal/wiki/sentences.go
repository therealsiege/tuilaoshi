package wiki

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Sentence cards: a word met in a reading, kept with the sentence it came
// in, on a private vocabulary page whose Example column the reviews show
// (the word blanked) as the card's context.
const sentencePage = "private/vocabulary/sentence-cards.md"

// AddSentenceCard keeps a word with the sentence it was met in; it
// reports false when the word already has a sentence card.
func (s *Store) AddSentenceCard(w Word, sentence string, now time.Time) (bool, error) {
	path := filepath.Join(s.Dir, sentencePage)
	data, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return false, fmt.Errorf("read %s: %w", sentencePage, err)
	}
	page := string(data)
	if page == "" {
		page = fmt.Sprintf("---\ntype: vocabulary\nlevel: advanced\nstatus: active\ntags: [sentences]\nintroduced: %s\n---\n\n"+
			"# Sentence cards\n\nWords kept with the sentence they were met in.\n\n| Word | Pinyin | Meaning | Example |\n|------|--------|---------|---------|\n", FormatDate(now))
	}
	word := strings.TrimSpace(w.Chinese)
	if word == "" || strings.Contains(page, "| "+word+" |") {
		return false, nil
	}
	clean := func(s string) string {
		return strings.ReplaceAll(strings.ReplaceAll(strings.TrimSpace(s), "|", "/"), "\n", " ")
	}
	page = strings.TrimRight(page, "\n") + fmt.Sprintf("\n| %s | %s | %s | %s |\n", word, clean(w.Pinyin), clean(w.Meaning), clean(sentence))
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return false, err
	}
	return true, os.WriteFile(path, []byte(page), 0644)
}

// Cloze blanks a word in a sentence, or returns "" when the sentence
// doesn't contain it.
func Cloze(sentence, word string) string {
	if word == "" || !strings.Contains(sentence, word) {
		return ""
	}
	return strings.Replace(sentence, word, strings.Repeat("＿", len([]rune(word))), 1)
}
