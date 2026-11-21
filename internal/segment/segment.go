// Package segment splits Chinese text into words by longest match against a
// set of known words. It runs offline on the wiki's own vocabulary.
package segment

import (
	"strings"
	"unicode"
)

// MaxWordLen is the longest word, in characters, that segmentation tries.
const MaxWordLen = 4

// Token is a piece of reading text: a word (or lone character) that can be
// selected and glossed, or a run of punctuation and other text.
type Token struct {
	Text string
	Word bool
}

// Split splits a line into tokens, taking the longest known word at each
// point. Unknown Chinese characters become one-character words; anything
// else is grouped into non-word runs.
func Split(line string, known map[string]bool) []Token {
	runes := []rune(line)
	var tokens []Token
	var other strings.Builder
	flush := func() {
		if other.Len() > 0 {
			tokens = append(tokens, Token{Text: other.String()})
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
		for l := MaxWordLen; l > 1; l-- {
			if i+l <= len(runes) && allHan(runes[i:i+l]) && known[string(runes[i:i+l])] {
				n = l
				break
			}
		}
		tokens = append(tokens, Token{Text: string(runes[i : i+n]), Word: true})
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
