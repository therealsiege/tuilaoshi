package rpg

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/therealsiege/tuilaoshi/internal/journey"
	rpgengine "github.com/therealsiege/tuilaoshi/internal/rpg"
	"github.com/therealsiege/tuilaoshi/internal/ui"
	"github.com/therealsiege/tuilaoshi/internal/wiki"
)

// Battles on the road (a tribulation, or a clan met on wild ground) carry
// the pilgrim's wounds in and out, and losing one means waking at the last
// town passed with half HP. Sparring at camp is practice: full HP, a
// rematch after a loss, and no tribulation cleared.

// RoadFight readies the battle for the open tribulation the pilgrim faces.
func (m Model) RoadFight() Model {
	m.road = true
	m = m.Refresh()
	m.applyWounds()
	return m
}

// Encounter readies a fight with a soldier met on wild ground, quizzing the
// coming stage's words.
func (m Model) Encounter(s journey.Soldier) Model {
	m.road = true
	m = m.Refresh()
	m.stage, m.gauntlet, m.foe, m.spirit, m.nightmare = nil, nil, 0, nil, nil
	m.battle = rpgengine.NewSoldierBattle(m.cards, s, m.mastered)
	m.setQuestionKinds()
	m.applyAbilities(m.cleared)
	m.battle.Radicals = m.radicals
	m.applyWounds()
	m.state = stateIntro // no story scene for a wild encounter
	return m.startSession()
}

// Spar readies a practice battle at camp.
func (m Model) Spar() Model {
	m.road = false
	return m.Refresh()
}

// applyWounds starts a road battle at the HP the pilgrim has left.
func (m *Model) applyWounds() {
	if !m.road || m.battle == nil || m.store == nil {
		return
	}
	place, err := m.store.Place()
	if err != nil {
		m.loadErr = err
		return
	}
	m.battle.Player.HP = max(m.battle.Player.MaxHP-place.Wounds, 1)
}

// keepWounds saves the HP lost when a road battle ends or is fled.
func (m *Model) keepWounds() {
	if !m.road || m.battle == nil || m.store == nil {
		return
	}
	if err := m.store.SetWounds(m.battle.Player.MaxHP - m.battle.Player.HP); err != nil {
		m.saveErr = err
	}
}

// fall ends a lost road battle: the companions carry the pilgrim back to
// the last town passed in the region (or its start) to wake with half HP.
func (m Model) fall() (Model, tea.Cmd) {
	place, err := m.store.Place()
	if err == nil {
		if i := journey.RegionByID(place.Region); i >= 0 {
			rm := journey.BuildMap(i)
			wake := rm.Start
			for p := range rm.Towns {
				if p.X >= place.X && (wake == rm.Start || p.X < wake.X) {
					wake = p // the nearest town back east
				}
			}
			err = m.store.SetPlace(wiki.Place{Region: place.Region, X: wake.X, Y: wake.Y})
		}
	}
	if err == nil {
		err = m.store.SetWounds(m.battle.Player.MaxHP / 2)
	}
	if err != nil {
		m.saveErr = err
		return m, nil
	}
	return m, ui.SwitchTo(ui.ViewMap)
}

// PilgrimHP is the pilgrim's HP on the road and its most: level, gear and
// companions' bond upgrades raise the most, wounds lower the HP.
func PilgrimHP(store *wiki.Store) (hp, most int, err error) {
	pages, err := store.ListPages()
	if err != nil {
		return 0, 0, err
	}
	cleared, err := store.ClearedStages()
	if err != nil {
		return 0, 0, err
	}
	kit, err := store.Kit()
	if err != nil {
		return 0, 0, err
	}
	bonds, err := store.BondPoints()
	if err != nil {
		return 0, 0, err
	}
	place, err := store.Place()
	if err != nil {
		return 0, 0, err
	}
	var e rpgengine.Effects
	rpgengine.AddGear(&e, journey.Worn(kit.Worn))
	rpgengine.AddUpgrades(&e, journey.Party(cleared), bonds)
	most = rpgengine.PlayerHP(journey.CountProgress(pages).Mastered) + e.BonusHP
	return max(most-place.Wounds, 1), most, nil
}
