package app

import (
	"fmt"

	"github.com/therealsiege/tuilaoshi/internal/audio"

	"github.com/therealsiege/tuilaoshi/internal/journey"
	"github.com/therealsiege/tuilaoshi/internal/ui"
	"github.com/therealsiege/tuilaoshi/internal/ui/celebrate"
)

// checkCelebrations compares current journey state against previous snapshots
// and triggers a celebration when a rank-up or boss defeat is detected.
func (m *Model) checkCelebrations() {
	state, err := journey.Calculate(m.store)
	if err != nil {
		return // the screens report read errors; a celebration can wait for the next check
	}

	// Count currently defeated bosses.
	bossCount := 0
	for _, b := range state.Bosses {
		if b.Defeated {
			bossCount++
		}
	}

	// Boss defeat takes priority (more dramatic).
	if bossCount > m.prevBossCount {
		// Find the most recently defeated boss (highest-index defeated one).
		for i := len(state.Bosses) - 1; i >= 0; i-- {
			if state.Bosses[i].Defeated {
				c := celebrate.BossDefeatCelebration(state.Bosses[i].Name, state.Bosses[i].English)
				m.celebration = &c
				ui.Cue(audio.LevelUp)
				m.celebrationTTL = 3
				break
			}
		}
		m.prevBossCount = bossCount
		m.prevRankName = state.Rank.Name
		return
	}

	// Rank change.
	if state.Rank.Name != m.prevRankName {
		c := celebrate.RankUpCelebration(state.Rank.Name, state.Rank.English)
		m.celebration = &c
		ui.Cue(audio.LevelUp)
		m.celebrationTTL = 3
		m.prevRankName = state.Rank.Name
		m.prevBossCount = bossCount
		return
	}

	// Level-up, saying what it brought.
	if state.Level.N > m.prevLevel {
		c := celebrate.LevelUpCelebration(state.Level.N, levelGains(state, m.prevLevel, m.prevOpenStage))
		m.celebration = &c
		ui.Cue(audio.LevelUp)
		m.celebrationTTL = 3
	}

	// Keep snapshots current even when no celebration fires.
	m.prevRankName = state.Rank.Name
	m.prevBossCount = bossCount
	m.prevLevel = state.Level.N
	m.prevOpenStage = openStage(state)
}

// openStage is the number of the stage ready to fight, or 0.
func openStage(st journey.State) int {
	if s := st.Stage; s != nil && s.Status == journey.StageOpen {
		return s.Number
	}
	return 0
}

// levelGains lists what reaching the state's level brought since level from.
func levelGains(st journey.State, from, openBefore int) []string {
	gains := []string{fmt.Sprintf("Max HP +%d (now %d)", journey.HPPerLevel*(st.Level.N-from), journey.LevelHP(st.Level.N))}
	if n := openStage(st); n != 0 && n != openBefore {
		gains = append(gains, fmt.Sprintf("Tribulation %d is open (f)", n))
	}
	return append(gains, fmt.Sprintf("Next level: master %d more", st.Level.Need-st.Level.Into))
}
