package ui

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/therealsiege/tuilaoshi/internal/audio"
)

// SpeechDoneMsg reports that text-to-speech finished, with the error if it failed.
type SpeechDoneMsg struct{ Err error }

// SpeakCmd speaks text in the background. It is nil when audio is off or
// there is nothing to say, which tea.Batch skips.
func SpeakCmd(tts *audio.Client, text, lang string) tea.Cmd {
	if tts == nil || !tts.Available() || text == "" || !Prefs.Speech {
		return nil
	}
	return func() tea.Msg {
		return SpeechDoneMsg{Err: tts.Speak(text, lang)}
	}
}

// SpeakLinesCmd speaks a conversation in the background, each line in its
// own voice. It is nil when audio is off or there is nothing to say.
func SpeakLinesCmd(tts *audio.Client, lines []audio.Line) tea.Cmd {
	if !Audible(tts) || len(lines) == 0 {
		return nil
	}
	return func() tea.Msg {
		return SpeechDoneMsg{Err: tts.SpeakLines(lines)}
	}
}

// Audible reports whether speech will be heard: a voice, and speech on.
func Audible(tts *audio.Client) bool { return tts != nil && tts.Available() && Prefs.Speech }

// AudioError renders why the last speech failed, or "" when it worked.
func AudioError(err error) string {
	if err == nil {
		return ""
	}
	return ErrorStyle.Render("  Audio failed: "+err.Error()) + "\n\n"
}
