package rpg

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	rpgengine "github.com/therealsiege/tuilaoshi/internal/rpg"
	"github.com/therealsiege/tuilaoshi/internal/ui/uitest"
)

// writeBattle is a fight on a "write" question for 书.
func writeBattle(t *testing.T) Model {
	t.Helper()
	m := fixedBattle(t)
	q := m.battle.Question
	m, _ = enter(m)
	m.battle.Question = &rpgengine.CombatQuestion{
		Prompt: "Write in Chinese: book (shū)", Category: "write", Damage: 20,
		Chinese: "书", Pinyin: "shū", English: "book", Card: q.Card, Accepted: []string{"书"},
	}
	m, _ = m.askQuestion()
	return m
}

func typeText(m Model, s string) Model {
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)})
	return m
}

func TestWriteQuestionScreen(t *testing.T) {
	uitest.Golden(t, "write", writeBattle(t).View())
}

func TestTypedAnswers(t *testing.T) {
	m := typeText(writeBattle(t), "书")
	m, _ = enter(m)
	if m.state != statePlayerAttack {
		t.Fatalf("right typed answer: state %s, want PlayerAttack", stateNames[m.state])
	}

	m = typeText(writeBattle(t), "水")
	m, _ = enter(m)
	if m.state != stateEnemyAttack || m.lastAnswer != "水" {
		t.Fatalf("wrong typed answer: state %s, last answer %q", stateNames[m.state], m.lastAnswer)
	}
	if got := m.session.Items[len(m.session.Items)-1].UserAnswer; got != "水" {
		t.Errorf("session recorded answer %q, want 水", got)
	}
}

func TestBlankTypedAnswerIsIgnored(t *testing.T) {
	m, _ := enter(writeBattle(t))
	if m.state != stateFighting || m.battle.Answered != 0 {
		t.Errorf("enter on an empty answer: state %s, answered %d", stateNames[m.state], m.battle.Answered)
	}
}
