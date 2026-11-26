package episode

import "unicode"

// Shadowing compares a line said back (as whisper heard it) with the line
// itself, character by character, ignoring punctuation and spaces. A 儿
// ending (一点儿) counts as said even when unheard: speech-to-text often
// drops it.

// ShadowChar is one character of the line and whether it was heard.
type ShadowChar struct {
	Char  rune
	Heard bool
}

// Shadowing is how close a line said back was.
type Shadowing struct {
	Chars   []ShadowChar
	Percent int // characters of the line heard, of all its characters
}

// numerals are digits as speech-to-text writes them in Chinese.
var numerals = []rune("〇一二三四五六七八九")

// spoken keeps the characters a speaker says: no punctuation or spaces,
// and digits as Chinese numerals (subtitles write 第1集, whisper 第一集).
func spoken(s string) []rune {
	var out []rune
	for _, r := range s {
		switch {
		case r >= '0' && r <= '9':
			out = append(out, numerals[r-'0'])
		case r >= '０' && r <= '９':
			out = append(out, numerals[r-'０'])
		case unicode.Is(unicode.Han, r) || unicode.IsLetter(r) || unicode.IsDigit(r):
			out = append(out, r)
		}
	}
	return out
}

// CompareShadow lines up the line and what was heard by their longest
// common sequence of characters.
func CompareShadow(line, heard string) Shadowing {
	a, b := spoken(line), spoken(heard)
	// lcs[i][j]: the longest common sequence of a[i:] and b[j:].
	lcs := make([][]int, len(a)+1)
	for i := range lcs {
		lcs[i] = make([]int, len(b)+1)
	}
	for i := len(a) - 1; i >= 0; i-- {
		for j := len(b) - 1; j >= 0; j-- {
			if a[i] == b[j] {
				lcs[i][j] = lcs[i+1][j+1] + 1
			} else {
				lcs[i][j] = max(lcs[i+1][j], lcs[i][j+1])
			}
		}
	}
	s := Shadowing{Chars: make([]ShadowChar, len(a))}
	for i, j := 0, 0; i < len(a); {
		switch {
		case j < len(b) && a[i] == b[j]:
			s.Chars[i] = ShadowChar{a[i], true}
			i, j = i+1, j+1
		case j < len(b) && lcs[i][j+1] >= lcs[i+1][j]:
			j++
		default:
			s.Chars[i] = ShadowChar{a[i], a[i] == '儿' && i > 0}
			i++
		}
	}
	heardN := 0
	for _, c := range s.Chars {
		if c.Heard {
			heardN++
		}
	}
	if len(a) > 0 {
		s.Percent = heardN * 100 / len(a)
	}
	return s
}

// shadowMax is the longest line, in characters, offered for shadowing.
const shadowMax = 24

// ShadowLines are the lines to shadow: two to shadowMax characters, and,
// when the episode has audio, a clip of their own to hear first.
func (e Episode) ShadowLines() []int {
	var lines []int
	for i, l := range e.Lines {
		n := hanCount(l.Chinese)
		if n >= 2 && n <= shadowMax && (e.Audio == "" || l.End > l.Start) {
			lines = append(lines, i)
		}
	}
	return lines
}
