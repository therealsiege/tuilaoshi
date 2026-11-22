package browser

import (
	"sort"
	"strings"
	"unicode"

	"github.com/therealsiege/tuilaoshi/internal/model"
)

// entry is one page in the browser list, with what search matches against.
type entry struct {
	page  model.Page
	icon  string // the page type, as one character
	title string
	level string
	key   string // lowercased, tone-free text for search
}

// kinds orders page types in the list and labels each with one character.
var kinds = []struct{ typ, icon string }{
	{"reading", "读"}, {"grammar", "文"}, {"vocabulary", "词"}, {"character", "字"},
	{"reference", "参"}, {"session", "记"},
}

func kindRank(typ string) int {
	for i, k := range kinds {
		if k.typ == typ {
			return i
		}
	}
	return len(kinds)
}

func kindIcon(typ string) string {
	if r := kindRank(typ); r < len(kinds) {
		return kinds[r].icon
	}
	return "·"
}

// buildIndex sorts pages by type, then HSK level, then title, and prepares
// their search text.
func buildIndex(pages []model.Page) []entry {
	entries := make([]entry, 0, len(pages))
	for _, p := range pages {
		title := p.Title
		if title == "" {
			title = p.Path
		}
		text := strings.Join(append([]string{title, p.Path, p.Character, p.Pinyin, p.Meaning, p.Level}, p.Tags...), " ")
		entries = append(entries, entry{page: p, icon: kindIcon(p.Type), title: title, level: p.Level, key: fold(text)})
	}
	sort.SliceStable(entries, func(i, j int) bool {
		a, b := entries[i], entries[j]
		if ra, rb := kindRank(a.page.Type), kindRank(b.page.Type); ra != rb {
			return ra < rb
		}
		if a.level != b.level {
			return levelRank(a.level) < levelRank(b.level)
		}
		return a.title < b.title
	})
	return entries
}

// levelRank puts HSK1..HSK6 first, then anything else.
func levelRank(level string) string {
	if strings.HasPrefix(level, "HSK") {
		return level
	}
	return "~" + level
}

// match returns the entries containing every word of the query, best
// first: those where each word is a whole word of the page (书 for "shu"),
// then where it starts one, then where it appears inside one (水 shuǐ).
// Within a rank, entries keep the list order.
func match(entries []entry, query string) []int {
	words := strings.Fields(fold(query))
	var ranked [3][]int
	for i, e := range entries {
		worst := 0
		for _, w := range words {
			r := wordRank(e.key, w)
			if r < 0 {
				worst = -1
				break
			}
			worst = max(worst, r)
		}
		if worst >= 0 {
			ranked[worst] = append(ranked[worst], i)
		}
	}
	return append(append(ranked[0], ranked[1]...), ranked[2]...)
}

// wordRank is 0 when w is a whole word of key, 1 when it starts one, 2 when
// it is only inside one, and -1 when absent.
func wordRank(key, w string) int {
	if !strings.Contains(key, w) {
		return -1
	}
	best := 2
	for _, f := range strings.FieldsFunc(key, func(r rune) bool { return !unicode.IsLetter(r) && !unicode.IsDigit(r) }) {
		switch {
		case f == w:
			return 0
		case strings.HasPrefix(f, w):
			best = 1
		}
	}
	return best
}

// toneless maps pinyin vowels with tone marks to plain letters, so "shu"
// finds shū.
var toneless = strings.NewReplacer(
	"ā", "a", "á", "a", "ǎ", "a", "à", "a",
	"ē", "e", "é", "e", "ě", "e", "è", "e",
	"ī", "i", "í", "i", "ǐ", "i", "ì", "i",
	"ō", "o", "ó", "o", "ǒ", "o", "ò", "o",
	"ū", "u", "ú", "u", "ǔ", "u", "ù", "u",
	"ǖ", "v", "ǘ", "v", "ǚ", "v", "ǜ", "v", "ü", "v",
)

// fold lowercases text and drops tone marks for searching.
func fold(s string) string { return toneless.Replace(strings.ToLower(s)) }
