package rpg

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	rpgengine "github.com/therealsiege/tuilaoshi/internal/rpg"
)

// timeRunsOut puts a burning enemy's timer on the current question and lets
// it expire.
func timeRunsOut(m Model) (Model, tea.Cmd) {
	clock := time.Date(2026, 11, 21, 12, 0, 0, 0, time.UTC)
	m.now = func() time.Time { return clock }
	m.battle.SetEffects(rpgengine.Effects{TimeLimit: 12 * time.Second})
	m, _ = m.askQuestion()
	clock = clock.Add(13 * time.Second)
	return m.Update(timerMsg{m.qseq})
}

func TestBurningEnemyTimesOutTheQuestion(t *testing.T) {
	m := fixedBattle(t)
	m, _ = enter(m)
	clock := time.Date(2026, 11, 21, 12, 0, 0, 0, time.UTC)
	m.now = func() time.Time { return clock }
	m.battle.SetEffects(rpgengine.Effects{TimeLimit: 12 * time.Second})
	m, _ = m.askQuestion()
	if !strings.Contains(m.View(), "⏳ 12s") {
		t.Errorf("no countdown shown:\n%s", m.View())
	}
	clock = clock.Add(5 * time.Second)
	if m2, cmd := m.Update(timerMsg{m.qseq}); m2.state != stateFighting || cmd == nil {
		t.Fatal("timer fired early or stopped ticking")
	}
	if m2, _ := m.Update(timerMsg{m.qseq - 1}); m2.state != stateFighting {
		t.Fatal("a stale tick ended the question")
	}
	m, _ = timeRunsOut(m)
	if m.state != stateEnemyAttack || m.lastAnswer != "(time ran out)" || m.battle.Answered != 1 {
		t.Errorf("after the deadline: state %s, last answer %q", stateNames[m.state], m.lastAnswer)
	}
}

func TestFieryEyesStrikeOutWrongAnswers(t *testing.T) {
	m := fixedBattle(t)
	q := m.battle.Question
	m, _ = enter(m)
	m.battle.Question = q
	m.battle.SetEffects(rpgengine.Effects{Reveals: 1})
	if !strings.Contains(m.View(), "w fiery eyes") {
		t.Error("help doesn't offer the reveal")
	}
	m = keys(m, "w")
	if len(m.battle.Question.Choices) != 2 || m.battle.Question.CorrectAnswer() != "book" {
		t.Errorf("after w: choices %v", m.battle.Question.Choices)
	}
}

func TestIntroListsAbilitiesInPlay(t *testing.T) {
	m := fixedBattle(t) // fights the White Bone Demon, who disguises
	m.applyAbilities(nil)
	if v := m.View(); !strings.Contains(v, "白骨精 disguises its lies") {
		t.Errorf("intro doesn't list the enemy's ability:\n%s", v)
	}
}
