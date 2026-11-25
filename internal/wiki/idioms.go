package wiki

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Pages of words the game adds to the reviews: idioms met in idiom-chain
// duels, and the words of festivals the learner has visited.
const (
	idiomPage    = "private/vocabulary/chengyu-duels.md"
	festivalPage = "private/vocabulary/festivals.md"
)

// Word is a word to add to the wiki: Chinese, its pinyin and meaning.
type Word struct {
	Chinese, Pinyin, Meaning string
}

// AddIdioms adds idioms not yet on the duel idiom page, creating it if
// needed, and returns how many were new.
func (s *Store) AddIdioms(words []Word, now time.Time) (int, error) {
	return s.addWords(idiomPage, "成语 from idiom-chain duels", "chengyu, duels", words, now)
}

// AddFestivalWords adds a festival's words to the festival words page.
func (s *Store) AddFestivalWords(words []Word, now time.Time) (int, error) {
	return s.addWords(festivalPage, "节日 Festival words", "festivals, culture", words, now)
}

// addWords adds words not yet on a vocabulary page, creating it with title
// and tags if needed, and returns how many were new.
func (s *Store) addWords(rel, title, tags string, words []Word, now time.Time) (int, error) {
	path := filepath.Join(s.Dir, rel)
	data, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return 0, fmt.Errorf("read %s: %w", rel, err)
	}
	page := string(data)
	if page == "" {
		page = fmt.Sprintf("---\ntype: vocabulary\nlevel: advanced\nstatus: active\ntags: [%s]\nintroduced: %s\n---\n\n"+
			"# %s\n\n| Word | Pinyin | Meaning |\n|------|--------|---------|\n", tags, FormatDate(now), title)
	}
	added := 0
	for _, w := range words {
		w.Chinese = strings.TrimSpace(w.Chinese)
		if w.Chinese == "" || strings.Contains(page, "| "+w.Chinese+" |") {
			continue
		}
		clean := func(s string) string { return strings.ReplaceAll(strings.TrimSpace(s), "|", "/") }
		page = strings.TrimRight(page, "\n") + fmt.Sprintf("\n| %s | %s | %s |\n", w.Chinese, clean(w.Pinyin), clean(w.Meaning))
		added++
	}
	if added == 0 {
		return 0, nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return 0, err
	}
	return added, os.WriteFile(path, []byte(page), 0644)
}
