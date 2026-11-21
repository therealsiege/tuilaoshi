package reader

import (
	"strings"
	"unicode"
)

// maxWordLen is the longest word, in characters, that segmentation tries.
const maxWordLen = 4

// token is a piece of reading text: a word (or lone character) that can be
// selected and glossed, or a run of punctuation and other text.
type token struct {
	text string
	word bool
}

// segment splits a line into tokens, taking the longest known word at each
// point. Unknown Chinese characters become one-character words; anything
// else is grouped into non-word runs.
func segment(line string, known map[string]bool) []token {
	runes := []rune(line)
	var tokens []token
	var other strings.Builder
	flush := func() {
		if other.Len() > 0 {
			tokens = append(tokens, token{text: other.String()})
			other.Reset()
		}
	}
	for i := 0; i < len(runes); {
		if !unicode.Is(unicode.Han, runes[i]) {
			other.WriteRune(runes[i])
			i++
			continue
		}
		flush()
		n := 1
		for l := maxWordLen; l > 1; l-- {
			if i+l <= len(runes) && allHan(runes[i:i+l]) && known[string(runes[i:i+l])] {
				n = l
				break
			}
		}
		tokens = append(tokens, token{text: string(runes[i : i+n]), word: true})
		i += n
	}
	flush()
	return tokens
}

func allHan(rs []rune) bool {
	for _, r := range rs {
		if !unicode.Is(unicode.Han, r) {
			return false
		}
	}
	return true
}

// textSection returns the lines under a page's "## Text" heading, without
// blank lines.
func textSection(body string) []string {
	var lines []string
	in := false
	for _, l := range strings.Split(body, "\n") {
		if strings.HasPrefix(l, "## ") {
			in = strings.TrimSpace(l) == "## Text"
			continue
		}
		if in && strings.TrimSpace(l) != "" {
			lines = append(lines, strings.TrimSpace(l))
		}
	}
	return lines
}
