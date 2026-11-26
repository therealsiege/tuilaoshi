package journey

import (
	"strings"
	"unicode"
)

// Grammar gates (语法关): each region has one grammar pattern standing
// guard before its demon's lair (content/grammar_gates.yaml). The learner
// passes by building sentences with it: arranging word tiles, then, with
// Claude, writing one of their own.

// GrammarGate is the pattern guarding one region's lair.
type GrammarGate struct {
	Region    string           `yaml:"region"`
	Pattern   string           `yaml:"pattern"`
	Pinyin    string           `yaml:"pinyin"`
	Explain   string           `yaml:"explain"`
	Markers   []string         `yaml:"markers"` // the pattern's words in order; "a|b" accepts either
	Tail      int              `yaml:"tail"`    // Han characters needed after the last marker (default 1)
	Ends      bool             `yaml:"ends"`    // the last marker may end the sentence
	Examples  []GrammarExample `yaml:"examples"`
	Sentences []TileSentence   `yaml:"sentences"`
}

// GrammarExample is a model sentence for a pattern.
type GrammarExample struct {
	Chinese string `yaml:"chinese"`
	English string `yaml:"english"`
}

// TileSentence is a sentence to rebuild from its words, in order; Also
// lists other orders of the same words that are just as right.
type TileSentence struct {
	Words   []string `yaml:"words"`
	English string   `yaml:"english"`
	Also    []string `yaml:"also"`
}

// Text is the sentence with its end mark: a question mark for a question.
func (s TileSentence) Text() string {
	end := "。"
	if strings.HasSuffix(s.English, "?") {
		end = "？"
	}
	return strings.Join(s.Words, "") + end
}

// Fits reports whether words joined in this order make the sentence, or
// one of its other right orders.
func (s TileSentence) Fits(words []string) bool {
	got := strings.Join(words, "")
	if got == strings.Join(s.Words, "") {
		return true
	}
	for _, a := range s.Also {
		if got == a {
			return true
		}
	}
	return false
}

// GrammarGates lists every region's gate.
func GrammarGates() []GrammarGate { return content.GrammarGates }

// GrammarGateFor returns the gate guarding a region's lair, if it has one.
func GrammarGateFor(region string) (GrammarGate, bool) {
	for _, g := range content.GrammarGates {
		if g.Region == region {
			return g, true
		}
	}
	return GrammarGate{}, false
}

// han keeps only the Chinese characters of s, so punctuation and spaces
// don't count.
func han(s string) []rune {
	var out []rune
	for _, r := range s {
		if unicode.Is(unicode.Han, r) {
			out = append(out, r)
		}
	}
	return out
}

// Uses reports whether a sentence uses the gate's pattern by the simple
// rules: at least four characters, the markers in order with something
// between each, and enough after the last (two after 把 or 被, for the
// object or doer and the verb). It is a first check; Claude judges the
// rest when it is there.
func (g GrammarGate) Uses(sentence string) bool {
	s := string(han(sentence))
	if len([]rune(s)) < 4 {
		return false
	}
	at := 0 // byte offset in s where the next marker may start
	for i, m := range g.Markers {
		start, end := -1, 0
		for _, alt := range strings.Split(m, "|") {
			if j := strings.Index(s[at:], alt); j >= 0 && (start < 0 || at+j < start) {
				start, end = at+j, at+j+len(alt)
			}
		}
		if start < 0 || (i > 0 && start == at) {
			return false // missing, or straight after the marker before it
		}
		at = end
	}
	tail := g.Tail
	if tail == 0 && !g.Ends {
		tail = 1
	}
	return len([]rune(s[at:])) >= tail
}
