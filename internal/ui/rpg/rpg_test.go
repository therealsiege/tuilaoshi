package rpg

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/therealsiege/tuilaoshi/internal/journey"
	"github.com/therealsiege/tuilaoshi/internal/model"
	rpgengine "github.com/therealsiege/tuilaoshi/internal/rpg"
	"github.com/therealsiege/tuilaoshi/internal/ui"
	"github.com/therealsiege/tuilaoshi/internal/wiki"
)

// storeModel returns a battle backed by a temporary wiki with one
// vocabulary page.
func storeModel(t *testing.T) Model {
	t.Helper()
	dir := t.TempDir()
	page := "---\ntype: vocabulary\nreview_interval: 10\nnext_review: 2026-12-01\n---\n\n# Basics\n\n" +
		"| Word | Pinyin | English |\n|------|--------|---------|\n" +
		"| 书 | shū | book |\n| 水 | shuǐ | water |\n| 火 | huǒ | fire |\n| 山 | shān | mountain |\n"
	if err := os.MkdirAll(filepath.Join(dir, "vocabulary"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "vocabulary", "basics.md"), []byte(page), 0644); err != nil {
		t.Fatal(err)
	}

	m := skipScene(New(wiki.NewStore(dir), nil, nil).Refresh())
	if m.battle == nil || len(m.cards) != 4 {
		t.Fatalf("battle not loaded from wiki: %d cards", len(m.cards))
	}
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	return m
}

func scheduleFor(t *testing.T, m Model, id string) model.Card {
	t.Helper()
	cards, err := m.store.DrillableCards()
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range cards {
		if c.ID == id {
			return c
		}
	}
	t.Fatalf("no card %q", id)
	return model.Card{}
}

func TestMissShortensTheWordsInterval(t *testing.T) {
	m := storeModel(t)
	missed := m.battle.Question.Card.ID
	was := scheduleFor(t, m, missed).ReviewInterval

	m = missQuestion(m)

	if got, want := scheduleFor(t, m, missed).ReviewInterval, max(was/4, 1); got != want {
		t.Errorf("missed word interval %d -> %d, want a quarter (%d)", was, got, want)
	}
}

func TestCorrectAnswerCountsOncePerBattle(t *testing.T) {
	m := storeModel(t)
	q := m.battle.Question
	id := q.Card.ID

	for i := 0; i < 3; i++ {
		m.battle.Question = q
		m.battle.Selected = q.Correct
		m, _ = m.submitAnswer()
		m.state = stateFighting
	}

	if got := scheduleFor(t, m, id); got.ReviewInterval != 12 {
		t.Errorf("interval = %d, want 12 (one Hard review of 10)", got.ReviewInterval)
	}
	if len(m.session.Items) != 3 {
		t.Errorf("session has %d items, want all 3 answers", len(m.session.Items))
	}
}

func TestBattleEndSavesSessionWithXP(t *testing.T) {
	m := storeModel(t)
	m.battle.Enemy.HP = 1
	m.battle.Selected = m.battle.Question.Correct
	m, _ = m.submitAnswer()
	for m.state == statePlayerAttack {
		m, _ = m.Update(animTickMsg{})
	}

	if m.state != stateVictory {
		t.Fatalf("state = %v, want stateVictory", m.state)
	}
	if m.saveErr != nil {
		t.Fatalf("save failed: %v", m.saveErr)
	}
	if got, want := m.store.TotalXP(), m.battle.XPEarned; got != want || got == 0 {
		t.Errorf("TotalXP() = %d, want %d", got, want)
	}
	if !strings.Contains(m.View(), "Total XP") {
		t.Error("victory screen should show lifetime XP")
	}

	m = m.Leave()
	if got := m.store.TotalXP(); got != m.battle.XPEarned {
		t.Errorf("leaving after victory saved again: TotalXP() = %d", got)
	}
}

func TestFleeingSavesAnswersSoFar(t *testing.T) {
	m := storeModel(t)
	m = missQuestion(m)

	m = m.Leave()

	if !m.saved {
		t.Error("fleeing with answers should save the session")
	}
	if m.store.TodayStats().CardsReviewed != 1 {
		t.Errorf("today's stats should include the battle answer")
	}
}

func TestFleeingBeforeAnsweringSavesNothing(t *testing.T) {
	m := storeModel(t)

	m = m.Leave()

	if _, err := os.Stat(filepath.Join(m.store.Dir, "sessions")); !os.IsNotExist(err) {
		t.Error("an empty battle should not write a session")
	}
}

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

// skipScene goes past a stage's story scene to the battle intro.
func skipScene(m Model) Model {
	if m.state == stateScene {
		m, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	}
	return m
}
