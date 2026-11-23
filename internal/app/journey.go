package app

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/therealsiege/tuilaoshi/internal/journey"
	"github.com/therealsiege/tuilaoshi/internal/ui"
)

// startJourney begins playing the journey chosen on the title screen: the
// milestones seen so far are those of that journey, so loading one doesn't
// celebrate what it had already reached.
func (m Model) startJourney() (tea.Model, tea.Cmd) {
	if st, err := journey.Calculate(m.store); err == nil {
		m.prevRankName, m.prevLevel, m.prevOpenStage = st.Rank.Name, st.Level.N, openStage(st)
		m.prevBossCount = 0
		for _, b := range st.Bosses {
			if b.Defeated {
				m.prevBossCount++
			}
		}
	}
	m.world = m.world.Refresh()
	return m, ui.SwitchTo(ui.ViewMap)
}

// enterTown opens a town the pilgrim walked into.
func (m Model) enterTown(id string) (tea.Model, tea.Cmd) {
	m.currentView = ui.ViewWorld
	m.world = m.world.Refresh().OpenTown(id)
	return m, nil
}

// isInputComposing checks if the current view's input is in Chinese composition mode.
func (m Model) isInputComposing() bool {
	switch m.currentView {
	case ui.ViewDrill:
		return m.drill.InputComposing()
	case ui.ViewChat:
		return m.chat.InputComposing()
	case ui.ViewReview:
		return m.review.InputComposing()
	case ui.ViewBattle:
		return m.battle.InputComposing()
	}
	return false
}

// resize passes the terminal's size to every screen.
func (m Model) resize(w, h int) Model {
	m.width = w
	m.height = h
	m.dashboard = m.dashboard.SetSize(w, h)
	m.drill = m.drill.SetSize(w, h)
	m.browser = m.browser.SetSize(w, h)
	m.chat = m.chat.SetSize(w, h)
	m.review = m.review.SetSize(w, h)
	m.stats = m.stats.SetSize(w, h)
	m.journey = m.journey.SetSize(w, h)
	m.radicals = m.radicals.SetSize(w, h)
	m.shrine = m.shrine.SetSize(w, h)
	m.cards = m.cards.SetSize(w, h)
	m.party = m.party.SetSize(w, h)
	m.world = m.world.SetSize(w, h)
	m.title = m.title.SetSize(w, h)
	m.roam = m.roam.SetSize(w, h)
	m.theatre = m.theatre.SetSize(w, h)
	m.forge = m.forge.SetSize(w, h)
	m.battle = m.battle.SetSize(w, h)
	m.reader = m.reader.SetSize(w, h)
	m.assess = m.assess.SetSize(w, h)
	m.festival = m.festival.SetSize(w, h)
	return m
}
