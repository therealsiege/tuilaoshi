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
	return m, ui.SwitchTo(ui.ViewDashboard)
}
