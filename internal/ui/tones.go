package ui

import (
	"strings"
	"unicode"

	"github.com/charmbracelet/lipgloss"
)

// Tone colours: pinyin is drawn a syllable at a time in its tone's colour
// (first red, second orange, third green, fourth blue, neutral grey), as
// Pleco does, so tones are seen every time a word is. The tone marks stay,
// so nothing is lost without colour. Prefs.ToneColours turns it off.

var toneColours = [5]lipgloss.Color{"#9AA0A6", "#E06C5A", "#E5A84B", "#6FBF73", "#6AA0E0"}

// toneOf maps each marked vowel to its tone.
var toneOf = map[rune]int{}

func init() {
	for tone, vowels := range []string{"āēīōūǖĀĒĪŌŪǕ", "áéíóúǘÁÉÍÓÚǗ", "ǎěǐǒǔǚǍĚǏǑǓǙ", "àèìòùǜÀÈÌÒÙǛ"} {
		for _, v := range vowels {
			toneOf[v] = tone + 1
		}
	}
	for _, i := range []string{"", "b", "p", "m", "f", "d", "t", "n", "l", "g", "k", "h", "j", "q", "x", "zh", "ch", "sh", "r", "z", "c", "s", "y", "w"} {
		for _, f := range strings.Fields("a o e i u v ai ei ao ou an en ang eng ong er ia ie iao iu ian in iang ing iong ua uo uai ui uan un uang ueng ve van vn ue uan") {
			syllables[i+f] = true
		}
	}
}

// syllables are the toneless pinyin syllables, written with v for ü.
var syllables = map[string]bool{}

// plain is a pinyin letter without its tone, lower case, ü as v.
func plain(r rune) rune {
	switch {
	case strings.ContainsRune("āáǎàa", unicode.ToLower(r)):
		return 'a'
	case strings.ContainsRune("ēéěèe", unicode.ToLower(r)):
		return 'e'
	case strings.ContainsRune("īíǐìi", unicode.ToLower(r)):
		return 'i'
	case strings.ContainsRune("ōóǒòo", unicode.ToLower(r)):
		return 'o'
	case strings.ContainsRune("ūúǔùu", unicode.ToLower(r)):
		return 'u'
	case strings.ContainsRune("ǖǘǚǜü", unicode.ToLower(r)):
		return 'v'
	}
	return unicode.ToLower(r)
}

// SplitPinyin splits a pinyin word into syllables by longest match, so
// "xuéxí" gives "xué" and "xí". Anything that isn't pinyin is one piece.
func SplitPinyin(word string) []string {
	runes := []rune(word)
	var out []string
	for i := 0; i < len(runes); {
		n := 0
		for j := min(len(runes), i+6); j > i; j-- {
			key := make([]rune, 0, j-i)
			for _, r := range runes[i:j] {
				key = append(key, plain(r))
			}
			if syllables[string(key)] {
				n = j - i
				break
			}
		}
		if n == 0 {
			n = 1
		}
		out = append(out, string(runes[i:i+n]))
		i += n
	}
	return out
}

// Pinyin draws pinyin in tone colours (or in the plain pinyin style when
// tone colours are off).
func Pinyin(s string) string {
	if !Prefs.ToneColours {
		return PinyinStyle.Render(s)
	}
	var b strings.Builder
	for i, word := range strings.Split(s, " ") {
		if i > 0 {
			b.WriteString(" ")
		}
		for _, syl := range SplitPinyin(word) {
			tone := 0
			for _, r := range syl {
				if t, ok := toneOf[r]; ok {
					tone = t
				}
			}
			b.WriteString(lipgloss.NewStyle().Foreground(toneColours[tone]).Italic(true).Render(syl))
		}
	}
	return b.String()
}
