package rpg

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/therealsiege/tuilaoshi/internal/journey"
	rpgengine "github.com/therealsiege/tuilaoshi/internal/rpg"
	"github.com/therealsiege/tuilaoshi/internal/ui"
	"github.com/therealsiege/tuilaoshi/internal/wiki"
)

// A stage is a gauntlet: its soldiers one after another, then its demon.
// The player's HP carries from fight to fight, with a little rest between,
// and only beating the demon clears the stage. Losing restarts that fight.

// restShare is the share of max HP, in percent, regained between fights.
const restShare = 25

// startGauntlet lines up the stage's soldiers and starts the first fight.
func (m *Model) startGauntlet() {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	m.gauntlet = journey.Gauntlet(m.stage.Number, rng)
	if m.difficulty == rpgengine.Harder {
		m.gauntlet = journey.Reinforce(m.stage.Number, m.gauntlet, rng)
	}
	m.foe = 0
	m.battle = m.stageBattle()
}

// stageBattle is the fight against the current foe of the stage.
func (m Model) stageBattle() *rpgengine.Battle {
	switch {
	case m.foe < len(m.gauntlet):
		return rpgengine.NewSoldierBattle(m.cards, m.gauntlet[m.foe], m.mastered)
	case m.stage.Boss != nil:
		return rpgengine.NewBossBattle(m.cards, m.stage.Opponent, m.mastered)
	}
	return rpgengine.NewBattle(m.cards, m.stage.Opponent, "", m.mastered)
}

// lastFoe reports whether this fight is the one that ends the battle: the
// stage's demon, or any practice battle.
func (m Model) lastFoe() bool { return m.stage == nil || m.foe >= len(m.gauntlet) }

// nextFoe moves on to the stage's next fight, carrying the player's HP.
func (m Model) nextFoe() Model {
	hp := m.battle.Player.HP
	m.foe++
	m.battle = m.stageBattle()
	m.applyAbilities(m.cleared)
	m.battle.Player.HP = min(hp+m.battle.Player.MaxHP*restShare/100, m.battle.Player.MaxHP)
	m.setQuestionKinds()
	m.battle.Radicals = m.radicals
	m.state = stateIntro
	m.log = nil
	return m.startSession()
}

// payWin pays the copper a win earns, and a silver ingot for a boss's
// stage, then works out what this fight paid in all: the right answers
// (paid when the session was saved), any daily bonus, and the win.
//
// A practice win pays only when the fight recalled a word not yet paid for
// today, so replaying easy battles earns nothing.
func (m *Model) payWin(before wiki.Wallet) {
	saved, err := m.store.Wallet()
	if err != nil {
		m.saveErr = err
		return
	}
	pay := wiki.Wallet{Tongqian: m.battle.WinCoins()}
	if m.stage == nil && saved.Tongqian == before.Tongqian {
		pay.Tongqian = 0
	}
	if m.stage != nil && m.stage.Boss != nil && m.lastFoe() && m.battle.PlayerWon() {
		pay.Yuanbao = 1
	}
	if err := m.store.Earn(pay); err != nil {
		m.saveErr = err
		return
	}
	if pay.Tongqian > 0 {
		if err := m.store.Tally("battle_won", 1); err != nil {
			m.saveErr = err
			return
		}
		m.bonded = m.creditBonds()
	}
	after, err := m.store.Wallet()
	if err != nil {
		m.saveErr = err
		return
	}
	m.earned = wiki.Wallet{Tongqian: after.Tongqian - before.Tongqian, Yuanbao: after.Yuanbao - before.Yuanbao}
	m.wallet = after
}

// creditBonds gives a bond point to each companion whose assist helped
// win this fight, and returns their names.
func (m *Model) creditBonds() []string {
	var names []string
	for _, c := range journey.Party(m.cleared) {
		if !c.Unlocked || c.Effect == "" || !m.battle.Assisted[c.Effect] || journey.BondFor(c.Name) == nil {
			continue
		}
		if err := m.store.AddBond(c.Name, 1); err != nil {
			m.saveErr = err
			return names
		}
		names = append(names, c.Name)
	}
	return names
}

// onVictoryKey continues the gauntlet, or leaves once the stage is done.
func (m Model) onVictoryKey() (Model, tea.Cmd) {
	if m.lastFoe() {
		return m, ui.SwitchTo(ui.ViewDashboard)
	}
	return m.nextFoe(), nil
}

// renderGauntlet shows where this fight stands in the stage: the foes in
// order, the current one marked.
func (m Model) renderGauntlet() string {
	if m.stage == nil || len(m.gauntlet) == 0 {
		return ""
	}
	names := make([]string, 0, len(m.gauntlet)+1)
	for _, s := range m.gauntlet {
		names = append(names, s.Name)
	}
	names = append(names, m.stage.Opponent.Name)
	for i := range names {
		switch {
		case i < m.foe:
			names[i] = ui.MutedStyle.Render("✓ " + names[i])
		case i == m.foe:
			names[i] = ui.WarningStyle.Render("▶ " + names[i])
		default:
			names[i] = ui.MutedStyle.Render(names[i])
		}
	}
	return fmt.Sprintf("  %s %s\n", ui.MutedStyle.Render(fmt.Sprintf("Fight %d of %d:", m.foe+1, len(names))),
		strings.Join(names, ui.MutedStyle.Render(" → ")))
}

// renderEarned shows the coins this fight paid and the wallet after it.
func (m Model) renderEarned() string {
	if m.earned == (wiki.Wallet{}) {
		return ""
	}
	got := fmt.Sprintf("+%d 铜钱", m.earned.Tongqian)
	if m.earned.Yuanbao > 0 {
		got += fmt.Sprintf("  +%d 元宝", m.earned.Yuanbao)
	}
	out := fmt.Sprintf("  %s   %s\n", ui.WarningStyle.Render(got),
		ui.MutedStyle.Render(fmt.Sprintf("purse: %d 铜钱 · %d 元宝", m.wallet.Tongqian, m.wallet.Yuanbao)))
	if len(m.bonded) > 0 {
		out += ui.ChineseStyle.Render("  Bond +1: "+strings.Join(m.bonded, " · ")) + "\n"
	}
	return out
}

// nextFoeName is who the stage's next fight is against.
func (m Model) nextFoeName() string {
	if m.foe+1 < len(m.gauntlet) {
		return m.gauntlet[m.foe+1].Name
	}
	return m.stage.Opponent.Name
}

// foeKind says what kind of enemy this is, for the practice intro.
func (m Model) foeKind() string {
	switch m.battle.Rank {
	case journey.RankMinion:
		return "A demon's underling"
	case journey.RankElite:
		return "A demon's lieutenant"
	case journey.RankCaptain:
		return "A demon captain"
	}
	return "A demon"
}
