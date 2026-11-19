package rpg

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/cj/mandarin-wiki/internal/journey"
	"github.com/cj/mandarin-wiki/internal/model"
	rpgengine "github.com/cj/mandarin-wiki/internal/rpg"
	"github.com/cj/mandarin-wiki/internal/ui"
)

func fightingModel() Model {
	cards := []model.Card{
		{ID: "1", Chinese: "书", Pinyin: "shū", English: "book"},
		{ID: "2", Chinese: "水", Pinyin: "shuǐ", English: "water"},
		{ID: "3", Chinese: "火", Pinyin: "huǒ", English: "fire"},
		{ID: "4", Chinese: "山", Pinyin: "shān", English: "mountain"},
	}
	demon := journey.BestiaryEntry{Name: "白骨精", English: "White Bone Demon", Chapter: "Chapter 27"}
	return Model{
		gkeys:  ui.DefaultGlobalKeys(),
		battle: rpgengine.NewBattle(cards, demon, "", 0),
		cards:  cards,
		state:  stateFighting,
	}
}

func missQuestion(m Model) Model {
	m.battle.Selected = (m.battle.Question.Correct + 1) % len(m.battle.Question.Choices)
	m, _ = m.submitAnswer()
	return m
}

func TestMissShowsCorrectAnswerUntilKeypress(t *testing.T) {
	m := missQuestion(fightingModel())
	q := m.battle.Question

	if m.state != stateEnemyAttack {
		t.Fatalf("state = %v, want stateEnemyAttack", m.state)
	}
	view := m.View()
	for _, want := range []string{q.CorrectAnswer(), q.Chinese, q.Pinyin, q.English} {
		if !strings.Contains(view, want) {
			t.Errorf("view is missing %q", want)
		}
	}

	m, _ = m.Update(animTickMsg{})
	if m.state != stateEnemyAttack {
		t.Errorf("tick advanced past the answer: state = %v", m.state)
	}

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if m.state != stateFighting {
		t.Errorf("state = %v after enter, want stateFighting", m.state)
	}
}

func TestDefeatEnterStartsRematch(t *testing.T) {
	m := fightingModel()
	enemy := m.battle.Enemy.Chinese
	m.battle.Player.HP = 1
	m = missQuestion(m)

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if m.state != stateDefeat {
		t.Fatalf("state = %v, want stateDefeat", m.state)
	}

	m, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd != nil {
		t.Error("rematch should stay in the battle view")
	}
	if m.state != stateFighting {
		t.Errorf("state = %v, want stateFighting", m.state)
	}
	if m.battle.Enemy.Chinese != enemy || m.battle.Player.HP != m.battle.Player.MaxHP {
		t.Errorf("rematch not reset: %+v", m.battle)
	}
	if len(m.log) != 0 {
		t.Errorf("log = %v, want empty", m.log)
	}
}

func TestStaleTickDoesNotSkipIntro(t *testing.T) {
	m := fightingModel()
	m.state = stateIntro

	m, _ = m.Update(animTickMsg{})

	if m.state != stateIntro {
		t.Errorf("state = %v, want stateIntro", m.state)
	}
}

func TestVictoryReportsWholeBattleAccuracy(t *testing.T) {
	m := fightingModel()
	m = missQuestion(m)
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})

	m.battle.Enemy.HP = 1
	m.battle.Selected = m.battle.Question.Correct
	m, _ = m.submitAnswer()
	for m.state == statePlayerAttack {
		m, _ = m.Update(animTickMsg{})
	}

	if m.state != stateVictory {
		t.Fatalf("state = %v, want stateVictory", m.state)
	}
	if view := m.View(); !strings.Contains(view, "50%") {
		t.Errorf("victory view should report 50%% accuracy:\n%s", view)
	}
}
