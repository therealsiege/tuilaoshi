package rpg

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	rpgengine "github.com/therealsiege/tuilaoshi/internal/rpg"
	"github.com/therealsiege/tuilaoshi/internal/ui/uitest"
)

// orderBattle is a fight on an "order" question for 我在看书, shuffled as 看 书 我 在.
func orderBattle(t *testing.T) Model {
	t.Helper()
	m := fixedBattle(t)
	q := m.battle.Question
	m, _ = enter(m)
	m.battle.Question = &rpgengine.CombatQuestion{
		Prompt: "Put the words in order: I'm reading.", Category: "order", Damage: 20,
		Chinese: "我在看书", English: "I'm reading.", Card: q.Card,
		Accepted: []string{"我在看书"}, Pieces: []string{"看", "书", "我", "在"},
	}
	m, _ = m.askQuestion()
	return m
}

func keys(m Model, ks ...string) Model {
	for _, k := range ks {
		if k == "backspace" {
			m, _ = m.Update(tea.KeyMsg{Type: tea.KeyBackspace})
			continue
		}
		m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(k)})
	}
	return m
}

func TestOrderQuestionScreen(t *testing.T) {
	uitest.Golden(t, "order", keys(orderBattle(t), "3", "4").View())
}

func TestOrderingWords(t *testing.T) {
	m := keys(orderBattle(t), "3", "3", "9", "4", "2", "backspace", "1")
	if m.orderText() != "我在看" || m.state != stateFighting {
		t.Fatalf("built %q in state %s; repeats, out-of-range keys and backspace should leave 我在看", m.orderText(), stateNames[m.state])
	}
	if m = keys(m, "2"); m.state != statePlayerAttack {
		t.Errorf("我在看书: state %s, want PlayerAttack", stateNames[m.state])
	}
	if m = keys(orderBattle(t), "2", "1", "4", "3"); m.state != stateEnemyAttack || m.lastAnswer != "书看在我" {
		t.Errorf("wrong order: state %s, last answer %q", stateNames[m.state], m.lastAnswer)
	}
}
