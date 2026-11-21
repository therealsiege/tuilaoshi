package rpg

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/therealsiege/tuilaoshi/internal/journey"
	rpgengine "github.com/therealsiege/tuilaoshi/internal/rpg"
	"github.com/therealsiege/tuilaoshi/internal/ui/uitest"
)

// fixedBattle replaces the random demon and question so the screen is stable.
func fixedBattle(t *testing.T) Model {
	t.Helper()
	m := storeModel(t).SetSize(80, 30)
	b := rpgengine.NewBattle(nil, journey.BestiaryEntry{Name: "白骨精", English: "White Bone Demon", Chapter: "Chapter 27"}, "", 0)
	b.Question = &rpgengine.CombatQuestion{
		Prompt:   "What does 书 mean?",
		Choices:  []string{"water", "book", "fire", "mountain"},
		Correct:  1,
		Category: "meaning",
		Damage:   20,
		Chinese:  "书",
		Pinyin:   "shū",
		English:  "book",
		Card:     scheduleFor(t, m, "vocabulary/basics.md#书"),
	}
	m.battle = b
	m.state = stateIntro
	return m
}

func TestBattleScreens(t *testing.T) {
	m := fixedBattle(t)
	uitest.Golden(t, "intro", m.View())

	q := m.battle.Question
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if m.state != stateFighting {
		t.Fatalf("state = %v after enter, want fighting", m.state)
	}
	m.battle.Question = q // entering the fight draws a random question
	uitest.Golden(t, "fighting", m.View())
}
