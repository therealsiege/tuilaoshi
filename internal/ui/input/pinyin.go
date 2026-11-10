package input

import (
	"strings"
	"unicode"
)

// Tone mark lookup tables indexed by tone number (1-4). Index 0 is unused.
var toneMarks = map[rune][5]rune{
	'a': {'a', '\u0101', '\u00e1', '\u01ce', '\u00e0'}, // a ā á ǎ à
	'e': {'e', '\u0113', '\u00e9', '\u011b', '\u00e8'}, // e ē é ě è
	'i': {'i', '\u012b', '\u00ed', '\u01d0', '\u00ec'}, // i ī í ǐ ì
	'o': {'o', '\u014d', '\u00f3', '\u01d2', '\u00f2'}, // o ō ó ǒ ò
	'u': {'u', '\u016b', '\u00fa', '\u01d4', '\u00f9'}, // u ū ú ǔ ù
	'v': {'\u00fc', '\u01d6', '\u01d8', '\u01da', '\u01dc'}, // ü ǖ ǘ ǚ ǜ
}

// validInitials lists valid pinyin initial consonants.
var validInitials = map[string]bool{
	"b": true, "p": true, "m": true, "f": true,
	"d": true, "t": true, "n": true, "l": true,
	"g": true, "k": true, "h": true,
	"j": true, "q": true, "x": true,
	"zh": true, "ch": true, "sh": true, "r": true,
	"z": true, "c": true, "s": true,
	"y": true, "w": true,
}

// validFinals lists valid pinyin final vowel groups.
var validFinals = map[string]bool{
	"a": true, "o": true, "e": true, "i": true, "u": true, "v": true,
	"ai": true, "ei": true, "ao": true, "ou": true,
	"an": true, "en": true, "ang": true, "eng": true, "er": true,
	"ia": true, "ie": true, "iao": true, "iu": true, "iou": true,
	"ian": true, "in": true, "iang": true, "ing": true, "iong": true,
	"ua": true, "uo": true, "uai": true, "ui": true, "uei": true,
	"uan": true, "un": true, "uen": true, "uang": true,
	"va": true, "ve": true, "van": true, "vn": true,
	"ong": true,
}

// IsValidSyllable checks whether s looks like a numbered pinyin syllable
// (letters optionally followed by a tone digit 1-5).
func IsValidSyllable(s string) bool {
	if len(s) == 0 {
		return false
	}
	s = strings.ToLower(s)

	// Strip trailing tone digit
	tone := 0
	last := s[len(s)-1]
	if last >= '1' && last <= '5' {
		tone = int(last - '0')
		s = s[:len(s)-1]
	}
	_ = tone

	if len(s) == 0 {
		return false
	}

	// Try to split into initial + final
	initial := ""
	final := s

	// Check two-letter initials first (zh, ch, sh)
	if len(s) >= 3 {
		prefix := s[:2]
		if validInitials[prefix] {
			initial = prefix
			final = s[2:]
		}
	}
	if initial == "" && len(s) >= 2 {
		prefix := s[:1]
		if validInitials[prefix] {
			initial = prefix
			final = s[1:]
		}
	}

	if final == "" {
		return false
	}

	return validFinals[final]
}

// ConvertSyllable converts a numbered pinyin syllable to diacritical form.
// For example: "ni3" → "nǐ", "lv4" → "lǜ".
// If s is not a valid numbered pinyin syllable, it is returned unchanged.
func ConvertSyllable(s string) string {
	if len(s) < 2 {
		return s
	}

	lower := strings.ToLower(s)
	last := lower[len(lower)-1]
	if last < '1' || last > '5' {
		return s
	}
	tone := int(last - '0')
	base := lower[:len(lower)-1]

	if !IsValidSyllable(lower) {
		return s
	}

	// Tone 5 (neutral) — just strip the digit and convert v→ü
	if tone == 5 {
		return replaceV(base)
	}

	// Find which vowel gets the tone mark using standard rules:
	// 1. If 'a' or 'e' is present, it gets the mark.
	// 2. If 'ou' is present, 'o' gets the mark.
	// 3. Otherwise the last vowel gets the mark.
	runes := []rune(base)
	markIdx := -1

	for i, r := range runes {
		if r == 'a' || r == 'e' {
			markIdx = i
			break
		}
	}

	if markIdx == -1 {
		// Check for 'ou'
		ouIdx := strings.Index(base, "ou")
		if ouIdx >= 0 {
			markIdx = ouIdx // mark the 'o'
		}
	}

	if markIdx == -1 {
		// Last vowel
		for i := len(runes) - 1; i >= 0; i-- {
			if isVowel(runes[i]) {
				markIdx = i
				break
			}
		}
	}

	if markIdx == -1 {
		return s // no vowel found — return unchanged
	}

	// Build result
	var result strings.Builder
	for i, r := range runes {
		if i == markIdx {
			vowel := r
			if vowel == 'v' {
				vowel = 'v' // keep as key for lookup
			}
			marks, ok := toneMarks[vowel]
			if ok {
				result.WriteRune(marks[tone])
			} else {
				result.WriteRune(r)
			}
		} else if r == 'v' {
			result.WriteRune('\u00fc') // ü
		} else {
			result.WriteRune(r)
		}
	}
	return result.String()
}

// ConvertText processes a full string, converting valid numbered pinyin
// syllables to their diacritical forms. Non-pinyin words pass through.
func ConvertText(text string) string {
	words := strings.Fields(text)
	for i, w := range words {
		if IsValidSyllable(w) {
			words[i] = ConvertSyllable(w)
		}
	}
	return strings.Join(words, " ")
}

func isVowel(r rune) bool {
	switch r {
	case 'a', 'e', 'i', 'o', 'u', 'v':
		return true
	}
	return false
}

func replaceV(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r == 'v' {
			b.WriteRune('\u00fc')
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// stripToneNumber removes a trailing digit 1-5 from a pinyin syllable.
func stripToneNumber(s string) string {
	if len(s) == 0 {
		return s
	}
	last := s[len(s)-1]
	if last >= '1' && last <= '5' {
		return s[:len(s)-1]
	}
	return s
}

// stripDiacritics converts diacritical pinyin back to plain letters
// and returns the result plus the tone number (0 if neutral/none).
func stripDiacritics(s string) (string, int) {
	var b strings.Builder
	tone := 0
	for _, r := range strings.ToLower(s) {
		found := false
		for base, marks := range toneMarks {
			for t := 0; t < 5; t++ {
				if marks[t] == r {
					if base == 'v' {
						b.WriteRune('v')
					} else {
						b.WriteRune(base)
					}
					if t > 0 && tone == 0 {
						tone = t
					}
					found = true
					break
				}
			}
			if found {
				break
			}
		}
		if !found {
			if unicode.IsLetter(r) {
				b.WriteRune(r)
			}
		}
	}
	return b.String(), tone
}
