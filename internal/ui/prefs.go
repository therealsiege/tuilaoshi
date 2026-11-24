package ui

import (
	"testing"

	"github.com/therealsiege/tuilaoshi/internal/audio"
	"github.com/therealsiege/tuilaoshi/internal/ui/input"
	"github.com/therealsiege/tuilaoshi/internal/wiki"
)

// Prefs are the learner's settings in effect: loaded when the app starts
// and replaced the moment the settings screen changes one, so every screen
// reads the same values without passing them around.
var Prefs = wiki.DefaultSettings()

// AnswerMode is the input mode answer boxes start in: Chinese characters
// picked from candidates, or pinyin with tone numbers.
func AnswerMode() input.Mode {
	if Prefs.Input == "pinyin" {
		return input.ModePinyin
	}
	return input.ModeChinese
}

// Cue plays a game event's sound when the learner has sound effects on.
// Tests stay silent.
func Cue(e audio.Effect) {
	if Prefs.Effects && !testing.Testing() {
		audio.Play(e)
	}
}
