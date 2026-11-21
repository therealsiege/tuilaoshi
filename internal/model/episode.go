package model

import "time"

// Episode is a dialogue to read word by word: a journey episode, or one the
// learner imported from their own subtitle files.
type Episode struct {
	Title  string
	Source string // where it comes from, e.g. "novel, chapter 27" or a file name
	Lines  []DialogueLine
	Audio  string // the episode's original audio file, when the learner supplied one
}

// DialogueLine is one spoken line. Speaker and English may be empty.
// Original marks a line quoted verbatim from a source text (the novel).
type DialogueLine struct {
	Speaker  string `yaml:"speaker,omitempty"`
	Chinese  string `yaml:"zh"`
	English  string `yaml:"en,omitempty"`
	Original bool   `yaml:"original,omitempty"`

	// Where the line is spoken in the episode's audio (subtitle timings).
	Start, End time.Duration `yaml:"-"`
}
