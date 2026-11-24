package rpg

import (
	"time"

	"github.com/therealsiege/tuilaoshi/internal/journey"
	"github.com/therealsiege/tuilaoshi/internal/model"
	rpgengine "github.com/therealsiege/tuilaoshi/internal/rpg"
	"github.com/therealsiege/tuilaoshi/internal/wiki"
)

// Fighting the Inner Demon: a boss battle at camp (full HP, a rematch after
// a loss) quizzing only the week's missed words, stronger the more there
// are. The first win each week pays copper and an ingot.

const (
	innerDemonHPEach = 4  // boss HP per missed word
	innerDemonPay    = 40 // copper for the week's first win
)

// MissedThisWeek returns the cards answered wrong since Monday.
func MissedThisWeek(store *wiki.Store, now time.Time) ([]model.Card, error) {
	cards, err := store.DrillableCards()
	if err != nil {
		return nil, err
	}
	byID := map[string]model.Card{}
	for _, c := range cards {
		byID[c.ID] = c
	}
	var missed []model.Card
	for _, id := range store.MissedSince(journey.WeekStart(now)) {
		if c, ok := byID[id]; ok && c.Chinese != "" {
			missed = append(missed, c)
		}
	}
	return missed, nil
}

// InnerDemon readies the fight with the week's Inner Demon.
func (m Model) InnerDemon() Model {
	m.road = false
	m = m.Refresh()
	missed, err := MissedThisWeek(m.store, time.Now())
	if err != nil {
		m.loadErr = err
		return m
	}
	m.stage, m.gauntlet, m.foe, m.spirit, m.nightmare = nil, nil, 0, nil, nil
	m.inner, m.cards = true, missed
	m.battle = rpgengine.NewBossBattle(missed, journey.InnerDemon, m.mastered)
	m.battle.Enemy.HP += innerDemonHPEach * len(missed)
	m.battle.Enemy.MaxHP = m.battle.Enemy.HP
	m.setQuestionKinds()
	m.applyAbilities(m.cleared)
	m.battle.Radicals = m.radicals
	m.state = stateIntro
	return m.startSession()
}
