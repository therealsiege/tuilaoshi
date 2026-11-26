package app

import (
	"fmt"
	"strings"

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
		if cmd := m.roam.Replay(); cmd != nil {
			return cmd // overheard demons, in their two voices
		}
		text = m.roam.Hear()
	case ui.ViewWorld:
		text = m.world.Hear()
	case ui.ViewFestival:
		text = m.festival.Hear()
	case ui.ViewDiary:
		text = m.diary.Hear()
	case ui.ViewDashboard:
		text = m.dashboard.Hear()
	case ui.ViewReading:
		text = m.reader.Hear()
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

// fitToWindow keeps a screen taller than the window readable: the title
// (its first lines) and the keys (its last) stay, and the middle shows the
// part around the selected line (▸), with a note that more is above or
// below. Screens that fit, or a window of unknown height, are untouched.
func fitToWindow(screen string, height int) string {
	lines := strings.Split(screen, "\n")
	const head, tail = 2, 2
	if height <= 0 || len(lines) <= height || height < head+tail+3 {
		return screen
	}
	middle := lines[head : len(lines)-tail]
	room := height - head - tail - 1 // one line says what is hidden
	sel := 0
	for i, l := range middle {
		if strings.Contains(l, "▸") {
			sel = i
			break
		}
	}
	first := min(max(sel-room/2, 0), len(middle)-room)
	out := append([]string{}, lines[:head]...)
	out = append(out, middle[first:first+room]...)
	note := fmt.Sprintf("  ⋯ %d more lines above, %d below: a taller window shows them", first, len(middle)-first-room)
	out = append(out, ui.MutedStyle.Render(note))
	return strings.Join(append(out, lines[len(lines)-tail:]...), "\n")
}
