package model

// Episode is a dialogue to read word by word: a journey episode, or one the
// learner imported from their own subtitle files.
type Episode struct {
	Title  string
	Source string // where it comes from, e.g. "novel, chapter 27" or a file name
	Lines  []DialogueLine
}

// DialogueLine is one spoken line. Speaker and English may be empty.
type DialogueLine struct {
	Speaker string `yaml:"speaker,omitempty"`
	Chinese string `yaml:"zh"`
	English string `yaml:"en,omitempty"`
}
