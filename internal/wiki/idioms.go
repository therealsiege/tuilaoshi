package wiki

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// idiomPage collects the idioms met in idiom-chain duels, as a vocabulary
// page, so they come up in reviews like any other word.
const idiomPage = "vocabulary/chengyu-duels.md"

// Word is a word to add to the wiki: Chinese, its pinyin and meaning.
type Word struct {
	Chinese, Pinyin, Meaning string
}

// AddIdioms adds idioms not yet on the duel idiom page, creating it if
// needed, and returns how many were new.
func (s *Store) AddIdioms(words []Word, now time.Time) (int, error) {
	path := filepath.Join(s.Dir, idiomPage)
	data, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return 0, fmt.Errorf("read %s: %w", idiomPage, err)
	}
	page := string(data)
	if page == "" {
		page = fmt.Sprintf("---\ntype: vocabulary\nlevel: advanced\nstatus: active\ntags: [chengyu, duels]\nintroduced: %s\n---\n\n"+
			"# 成语 from idiom-chain duels\n\n| Word | Pinyin | Meaning |\n|------|--------|---------|\n", FormatDate(now))
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
