package rpg

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/cj/mandarin-wiki/internal/model"
	rpgengine "github.com/cj/mandarin-wiki/internal/rpg"
	"github.com/cj/mandarin-wiki/internal/ui"
	"github.com/cj/mandarin-wiki/internal/wiki"
)

type viewState int

const (
	stateIntro viewState = iota
	stateFighting
	statePlayerAttack
	stateEnemyAttack
	stateVictory
	stateDefeat
)

const (
	animTickInterval = 120 * time.Millisecond
	// playerAttackTicks keeps a correct answer on screen for about a second.
	playerAttackTicks = 8
)

// animTickMsg drives the brief pause after a correct answer.
type animTickMsg struct{}

// Model is the RPG battle view model.
type Model struct {
	store  *wiki.Store
	gkeys  ui.GlobalKeys
	battle *rpgengine.Battle
	cards  []model.Card
	state  viewState
	width  int
	height int

	// Animation timer: how many ticks remain before advancing state.
	animTicks int

	// Combat log lines (most recent last).
	log []logEntry

	// Results of this battle, saved to the wiki when it ends or the player flees.
	session  model.Session
	reviewed map[string]bool // cards whose correct answer has counted this battle
	saved    bool
	saveErr  error
	totalXP  int // lifetime XP, read back after saving
}

type logEntry struct {
	text  string
	style lipgloss.Style
}

// New creates a new RPG battle model.
func New(store *wiki.Store) Model {
	return Model{
		store: store,
		gkeys: ui.DefaultGlobalKeys(),
		state: stateIntro,
	}
}

// Init satisfies the Bubble Tea model interface.
func (m Model) Init() tea.Cmd {
	return nil
}

// SetSize updates terminal dimensions.
func (m Model) SetSize(w, h int) Model {
	m.width = w
	m.height = h
	return m
}

// Refresh reloads battle state with fresh cards from the wiki.
func (m Model) Refresh() Model {
	m.cards = loadCards(m.store)
	mastered := countMastered(m.store)
	m.battle = rpgengine.NewRandomBattle(m.cards, mastered)
	m.state = stateIntro
	m.log = nil
	m.animTicks = 0
	return m.startSession()
}

// Leave saves the battle's results when the player flees mid-fight.
func (m Model) Leave() Model {
	m.saveSession()
	return m
}

// startSession begins recording results for a new battle.
func (m Model) startSession() Model {
	m.session = model.Session{StartedAt: time.Now(), Type: "battle"}
	m.reviewed = map[string]bool{}
	m.saved = false
	m.saveErr = nil
	return m
}

// recordAnswer logs an answer and reschedules its card: a miss makes it due
// again tomorrow, and a hit counts once per battle as a "Hard" review.
func (m *Model) recordAnswer(q *rpgengine.CombatQuestion, choice int, correct bool) {
	diff := model.DiffAgain
	if correct {
		diff = model.DiffHard
	}
	m.session.Items = append(m.session.Items, model.SessionItem{
		CardID:     q.Card.ID,
		UserAnswer: q.Choices[choice],
		Correct:    correct,
		Difficulty: diff,
	})

	if m.store == nil || (correct && m.reviewed[q.Card.ID]) {
		return
	}
	card := q.Card
	if err := m.store.RecordReview(&card, diff, time.Now()); err != nil {
		m.saveErr = err
	}
	m.reviewed[q.Card.ID] = true
}

// saveSession writes the battle to the wiki's session log once, so it counts
// toward stats, streaks, and lifetime XP.
func (m *Model) saveSession() {
	if m.saved || m.store == nil || m.battle == nil || len(m.session.Items) == 0 {
		return
	}
	m.session.EndedAt = time.Now()
	m.session.CardsTotal = m.battle.Answered
	m.session.CardsCorrect = m.battle.CorrectHits
	m.session.XP = m.battle.XPEarned
	if err := m.store.WriteSession(m.session); err != nil {
		m.saveErr = err
		return
	}
	m.saved = true
	m.totalXP = m.store.TotalXP()
}

// loadCards pulls drillable cards from the wiki store.
func loadCards(store *wiki.Store) []model.Card {
	cards, _ := store.DrillableCards()
	return cards
}

// countMastered returns the count of cards with a review interval >= 21 days.
func countMastered(store *wiki.Store) int {
	pages, err := store.ListPages()
	if err != nil {
		return 0
	}
	count := 0
	for _, p := range pages {
		if p.ReviewInterval >= 21 {
			count++
		}
	}
	return count
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case animTickMsg:
		// A tick can outlive its attack if the player flees and starts a new battle.
		if m.state != statePlayerAttack {
			return m, nil
		}
		m.animTicks--
		if m.animTicks > 0 {
			return m, tickAfter(animTickInterval)
		}
		// Animation finished — advance to the next logical state.
		return m.nextTurn()

	case tea.KeyMsg:
		// Global escape: flee battle.
		if key.Matches(msg, m.gkeys.Back) {
			return m, ui.SwitchTo(ui.ViewDashboard)
		}

		switch m.state {
		case stateIntro:
			if key.Matches(msg, m.gkeys.Enter) || msg.Type == tea.KeySpace {
				m.state = stateFighting
				m.battle.State = rpgengine.BattlePlayerTurn
				q := m.battle.GenerateQuestion(m.cards)
				m.battle.Question = q
				return m, nil
			}

		case stateFighting:
			return m.handleFightingKey(msg)

		case statePlayerAttack:
			// Ignore input during animations.
			return m, nil

		case stateEnemyAttack:
			// Hold the correct answer on screen until the player has read it.
			if key.Matches(msg, m.gkeys.Enter) || msg.Type == tea.KeySpace {
				return m.nextTurn()
			}

		case stateVictory:
			if key.Matches(msg, m.gkeys.Enter) || msg.Type == tea.KeySpace {
				return m, ui.SwitchTo(ui.ViewDashboard)
			}

		case stateDefeat:
			if key.Matches(msg, m.gkeys.Enter) || msg.Type == tea.KeySpace {
				return m.rematch(), nil
			}
		}
	}
	return m, nil
}

func (m Model) handleFightingKey(msg tea.KeyMsg) (Model, tea.Cmd) {
	if m.battle == nil || m.battle.Question == nil {
		return m, nil
	}
	q := m.battle.Question
	numChoices := len(q.Choices)

	switch {
	case key.Matches(msg, m.gkeys.Up):
		if m.battle.Selected > 0 {
			m.battle.Selected--
		}
	case key.Matches(msg, m.gkeys.Down):
		if m.battle.Selected < numChoices-1 {
			m.battle.Selected++
		}
	case key.Matches(msg, m.gkeys.Enter):
		return m.submitAnswer()
	default:
		// Number keys 1-5 for direct selection.
		if s := msg.String(); len(s) == 1 && s[0] >= '1' && s[0] <= '5' {
			idx := int(s[0] - '1')
			if idx < numChoices {
				m.battle.Selected = idx
				return m.submitAnswer()
			}
		}
	}
	return m, nil
}

func (m Model) submitAnswer() (Model, tea.Cmd) {
	correct, dmg := m.battle.Answer(m.battle.Selected)
	m.recordAnswer(m.battle.Question, m.battle.Selected, correct)
	if correct {
		m.state = statePlayerAttack
		m.battle.State = rpgengine.BattlePlayerAttack
		entry := logEntry{
			text:  fmt.Sprintf("You strike %s for %d damage!", m.battle.Enemy.Chinese, dmg),
			style: ui.SuccessStyle,
		}
		m.log = appendLog(m.log, entry)
		if m.battle.Combo > 1 {
			comboEntry := logEntry{
				text:  fmt.Sprintf("COMBO x%d! %s", m.battle.Combo, comboLabel(m.battle.Combo)),
				style: ui.WarningStyle,
			}
			m.log = appendLog(m.log, comboEntry)
		}
	} else {
		m.state = stateEnemyAttack
		m.battle.State = rpgengine.BattleEnemyAttack
		entry := logEntry{
			text:  fmt.Sprintf("%s attacks! -%d HP", m.battle.Enemy.Chinese, dmg),
			style: ui.ErrorStyle,
		}
		m.log = appendLog(m.log, entry)
		return m, nil
	}
	m.animTicks = playerAttackTicks
	return m, tickAfter(animTickInterval)
}

// rematch restarts the fight against the same demon.
func (m Model) rematch() Model {
	m.battle = m.battle.Rematch(m.cards)
	m.battle.State = rpgengine.BattlePlayerTurn
	m.state = stateFighting
	m.log = nil
	m.animTicks = 0
	return m.startSession()
}

func (m Model) nextTurn() (Model, tea.Cmd) {
	if m.battle.IsOver() {
		if m.battle.PlayerWon() {
			m.state = stateVictory
			m.battle.State = rpgengine.BattleVictory
		} else {
			m.state = stateDefeat
			m.battle.State = rpgengine.BattleDefeat
		}
		m.saveSession()
		return m, nil
	}
	// Back to player turn with a new question.
	m.state = stateFighting
	m.battle.State = rpgengine.BattlePlayerTurn
	m.battle.Selected = 0
	q := m.battle.GenerateQuestion(m.cards)
	m.battle.Question = q
	return m, nil
}

// tickAfter returns a Cmd that sends animTickMsg after a delay.
func tickAfter(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(_ time.Time) tea.Msg {
		return animTickMsg{}
	})
}
