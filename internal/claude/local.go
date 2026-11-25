package claude

import (
	"os"
	"strings"
	"unicode"

	"github.com/therealsiege/tuilaoshi/internal/model"
)

// Without an API key the tutor still works offline: drill and review
// answers are checked here, against the card itself.

// NeedsKey says where the key goes, for screens that can't work without it.
const NeedsKey = "needs ANTHROPIC_API_KEY (in ~/.config/tuilaoshi/env)"

// KeyPresent reports whether an API key is set.
func KeyPresent() bool { return os.Getenv("ANTHROPIC_API_KEY") != "" }

// Ready reports whether Claude can be reached: there is a client and a key.
func (c *Client) Ready() bool { return c != nil && KeyPresent() }

// IfReady returns the client when it can reach Claude, else nil, for
// screens that say what they need when given none.
func (c *Client) IfReady() *Client {
	if c.Ready() {
		return c
	}
	return nil
}

// toneless maps pinyin vowels with tone marks to plain letters.
var toneless = strings.NewReplacer(
	"ā", "a", "á", "a", "ǎ", "a", "à", "a",
	"ē", "e", "é", "e", "ě", "e", "è", "e",
	"ī", "i", "í", "i", "ǐ", "i", "ì", "i",
	"ō", "o", "ó", "o", "ǒ", "o", "ò", "o",
	"ū", "u", "ú", "u", "ǔ", "u", "ù", "u",
	"ǖ", "v", "ǘ", "v", "ǚ", "v", "ǜ", "v", "ü", "v",
)

// plain lowercases text and keeps only its letters and characters, so
// "Shū!" and "shu" compare equal.
func plain(s string) string {
	var b strings.Builder
	for _, r := range toneless.Replace(strings.ToLower(s)) {
		if unicode.IsLetter(r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// glosses splits a card's English into the meanings it lists: "book;
// letter (to write)" gives "book", "letter", "to write", "letter (to write)".
func glosses(english string) []string {
	parts := strings.FieldsFunc(english, func(r rune) bool { return r == ',' || r == ';' || r == '/' })
	var out []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		out = append(out, p)
		if i := strings.Index(p, "("); i > 0 {
			out = append(out, strings.TrimSpace(p[:i]))
		}
		out = append(out, strings.TrimPrefix(strings.TrimPrefix(p, "to "), "a "))
	}
	return out
}

// LocalGrade checks an answer against the card without Claude: a meaning
// against the English glosses, Chinese against the characters or their
// toneless pinyin.
func LocalGrade(card model.Card, answer string, mode model.DrillMode) model.DrillResult {
	got := plain(answer)
	ok := false
	if mode == model.ModeRecognition {
		for _, g := range glosses(card.English) {
			if g := plain(g); g != "" && got == g {
				ok = true
			}
		}
	} else {
		ok = got != "" && (got == plain(card.Chinese) || got == plain(card.Pinyin))
	}
	res := model.DrillResult{Correct: ok, Score: 1}
	if ok {
		res.Score = 4
	}
	res.Feedback = "Checked without Claude: " + card.Chinese + " " + card.Pinyin + " · " + card.English
	return res
}
