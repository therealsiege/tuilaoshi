package rpg

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/therealsiege/tuilaoshi/internal/journey"
	rpgengine "github.com/therealsiege/tuilaoshi/internal/rpg"
	"github.com/therealsiege/tuilaoshi/internal/ui"
)

// timerMsg ticks a burning enemy's answer timer for question seq.
type timerMsg struct{ seq int }

func timerTick(seq int) tea.Cmd {
	return tea.Tick(time.Second, func(time.Time) tea.Msg { return timerMsg{seq} })
}

// applyAbilities puts the enemy's and the joined party's abilities in play.
func (m *Model) applyAbilities(cleared map[int]string) {
	if m.battle == nil {
		return
	}
	enemy := journey.BestiaryEntry{Name: m.battle.Enemy.Chinese}
	if d := journey.EnemyLookup(m.battle.Enemy.Chinese); d != nil {
		enemy = *d
	}
	var e rpgengine.Effects
	e, m.abilities = rpgengine.EffectsFor(enemy, journey.Party(cleared))
	if m.store != nil {
		kit, err := m.store.Kit()
		if err != nil {
			m.loadErr = err
		}
		m.abilities = append(m.abilities, rpgengine.AddGear(&e, journey.Worn(kit.Worn))...)
	}
	if says := m.difficulty.Says(m.accuracy); says != "" {
		if m.difficulty == rpgengine.Easier {
			e.Reveals++
		}
		m.abilities = append(m.abilities, says)
	}
	m.battle.SetEffects(e)
	m.battle.Player.MaxHP += e.BonusHP
	m.battle.Player.HP = m.battle.Player.MaxHP
}

func (m Model) clock() time.Time {
	if m.now != nil {
		return m.now()
	}
	return time.Now()
}

// askQuestion readies the current question and, against a burning enemy,
// starts its answer timer.
func (m Model) askQuestion() (Model, tea.Cmd) {
	m, cmd := m.prepareQuestion()
	limit := m.battle.Effects.TimeLimit
	if limit == 0 || m.battle.Question == nil {
		return m, cmd
	}
	m.qseq++
	m.deadline = m.clock().Add(limit)
	return m, tea.Batch(cmd, timerTick(m.qseq))
}

// onTimer counts down; when time runs out the question is answered wrong.
func (m Model) onTimer(msg timerMsg) (Model, tea.Cmd) {
	if msg.seq != m.qseq || m.state != stateFighting {
		return m, nil
	}
	if m.clock().Before(m.deadline) {
		return m, timerTick(msg.seq)
	}
	q := m.battle.Question
	var dmg int
	if len(q.Choices) > 1 {
		_, dmg = m.battle.Answer((q.Correct + 1) % len(q.Choices))
	} else {
		_, dmg = m.battle.AnswerText("")
	}
	return m.afterAnswer(q, "(time ran out)", false, dmg)
}

// secondsLeft is the time left on a timed question, for display.
func (m Model) secondsLeft() int {
	return max(0, int(m.deadline.Sub(m.clock()).Round(time.Second)/time.Second))
}

// reveal uses Wukong's fiery eyes to strike out wrong answers.
func (m Model) reveal() Model {
	if m.battle.Reveal() {
		m.log = appendLog(m.log, logEntry{text: "孙悟空's fiery eyes see through two lies!", style: ui.SuccessStyle})
	}
	return m
}

// renderAbilities lists the abilities in play for the intro.
func (m Model) renderAbilities() string {
	out := ""
	for _, a := range m.abilities {
		out += ui.MutedStyle.Render("  · "+a) + "\n"
	}
	if out != "" {
		out += "\n"
	}
	return out
}

// renderTimer shows the countdown against a burning enemy.
func (m Model) renderTimer() string {
	if m.battle.Effects.TimeLimit == 0 || m.state != stateFighting {
		return ""
	}
	return ui.WarningStyle.Render(fmt.Sprintf("  ⏳ %ds", m.secondsLeft())) + "\n"
}
