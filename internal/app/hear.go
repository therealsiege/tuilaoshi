package app

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/therealsiege/tuilaoshi/internal/ui"
)

// hear speaks the Chinese on screen, for screens without a voice of their
// own (reviews, flashcards, battles and chat replay with ctrl+r themselves).
// It is nil when there is nothing to say or audio is off.
func (m Model) hear() tea.Cmd {
	var text string
	switch m.currentView {
	case ui.ViewMap:
		text = m.roam.Hear()
	case ui.ViewWorld:
		text = m.world.Hear()
	case ui.ViewFestival:
		text = m.festival.Hear()
	case ui.ViewDiary:
		text = m.diary.Hear()
	}
	return ui.SpeakCmd(m.tts, text, "zh")
}

// ownsEsc reports whether the screen uses esc itself (closing what is open
// on it), so esc doesn't go back to camp.
func (m Model) ownsEsc() bool {
	switch m.currentView {
	case ui.ViewWorld, ui.ViewTitle, ui.ViewMap, ui.ViewForge, ui.ViewDiary:
		return true
	case ui.ViewReading:
		return m.reader.Open()
	case ui.ViewBrowser:
		return m.browser.Searching()
	case ui.ViewParty:
		return m.party.Busy()
	case ui.ViewTheatre:
		return m.theatre.Busy()
	case ui.ViewFestival:
		return m.festival.Busy()
	case ui.ViewStats:
		return m.stats.Walking()
	}
	return false
}
