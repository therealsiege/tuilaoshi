package ui

import (
	"testing"

	"github.com/therealsiege/tuilaoshi/internal/audio"
)

func TestSpeechOffSpeaksNothing(t *testing.T) {
	defer func(on bool) { Prefs.Speech = on }(Prefs.Speech)
	Prefs.Speech = false
	if SpeakCmd(audio.New(), "你好", "zh") != nil {
		t.Error("speech off still speaks")
	}
}
