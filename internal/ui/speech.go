package ui

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/cj/mandarin-wiki/internal/audio"
)

// SpeechDoneMsg reports that text-to-speech finished, with the error if it failed.
type SpeechDoneMsg struct{ Err error }

// SpeakCmd speaks text in the background. It is nil when audio is off or
// there is nothing to say, which tea.Batch skips.
func SpeakCmd(tts *audio.Client, text, lang string) tea.Cmd {
	if tts == nil || !tts.Available() || text == "" {
		return nil
	}
	return func() tea.Msg {
		return SpeechDoneMsg{Err: tts.Speak(text, lang)}
	}
}

// AudioError renders why the last speech failed, or "" when it worked.
func AudioError(err error) string {
	if err == nil {
		return ""
	}
	return ErrorStyle.Render("  Audio failed: "+err.Error()) + "\n\n"
}
