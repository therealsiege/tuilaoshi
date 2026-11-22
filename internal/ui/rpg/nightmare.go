package rpg

import (
	"slices"

	"github.com/therealsiege/tuilaoshi/internal/journey"
	rpgengine "github.com/therealsiege/tuilaoshi/internal/rpg"
)

// A nightmare rematch brings back a beaten boss, half as strong again,
// asking only the learner's weakest words. Pressing n on a practice
// battle's intro starts one; the first nightmare win against each boss
// pays a silver ingot.

// nightmareBoss is the boss a nightmare would bring back: the first
// beaten boss not yet beaten in a nightmare, else the latest beaten.
func (m Model) nightmareBoss() *journey.Boss {
	if m.store == nil || m.stage != nil || m.spirit != nil {
		return nil
	}
	col, err := m.store.Nightmares()
	if err != nil {
		return nil
	}
	var last *journey.Boss
	for _, b := range journey.Bosses(m.counts, m.cleared) {
		if !b.Defeated {
			continue
		}
		b := b
		if !slices.Contains(col, b.Name) {
			return &b
		}
		last = &b
	}
	return last
}

// startNightmare replaces the practice battle with a nightmare rematch.
func (m Model) startNightmare() Model {
	boss := m.nightmareBoss()
	if boss == nil {
		return m
	}
	cards, err := m.store.DrillableCards()
	if err != nil {
		m.loadErr = err
		return m
	}
	m.nightmare, m.cards = boss, rpgengine.WeakCards(cards)
	m.battle = rpgengine.NewNightmareBattle(m.cards, journey.BossEntry(*boss), m.mastered)
	m.setQuestionKinds()
	m.applyAbilities(m.cleared)
	m.battle.Radicals = m.radicals
	return m.startSession()
}
