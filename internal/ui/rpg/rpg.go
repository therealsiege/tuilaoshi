package rpg

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/therealsiege/tuilaoshi/internal/audio"
	"github.com/therealsiege/tuilaoshi/internal/journey"
	"github.com/therealsiege/tuilaoshi/internal/model"
	rpgengine "github.com/therealsiege/tuilaoshi/internal/rpg"
	"github.com/therealsiege/tuilaoshi/internal/ui"
	"github.com/therealsiege/tuilaoshi/internal/ui/input"
	"github.com/therealsiege/tuilaoshi/internal/wiki"
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
	tts    *audio.Client
	ttsErr error // why the last speech failed

	// Typed answers for "write" questions; writing is off without a dictionary.
	input      input.Model
	writing    bool
	lastAnswer string // what the player chose or typed, shown after a miss
	gkeys      ui.GlobalKeys
	battle     *rpgengine.Battle
	cards      []model.Card
	state      viewState
	width      int
	height     int

	// Animation timer: how many ticks remain before advancing state.
	animTicks int

	// Combat log lines (most recent last).
	log []logEntry

	// Results of this battle, saved to the wiki when it ends or the player flees.
	session  model.Session
	reviewed map[string]bool // cards whose correct answer has counted this battle
	saved    bool
	saveErr  error
	loadErr  error // why the wiki could not be read for this battle
	totalXP  int   // lifetime XP, read back after saving

	// The next unbeaten boss, when enough characters are mastered to
	// challenge it from the intro; bossFight is set once challenged.
	boss      *journey.Boss
	bossFight bool
	mastered  int // the journey's mastered count, which scales player stats
}

type logEntry struct {
	text  string
	style lipgloss.Style
}

// New creates a new RPG battle model.
func New(store *wiki.Store, tts *audio.Client, dict *input.Dict) Model {
	ti := input.NewWithDict(dict)
	ti.SetPlaceholder("type pinyin, then pick characters")
	ti.SetCharLimit(20)
	return Model{
		store:   store,
		tts:     tts,
		input:   ti,
		writing: dict != nil,
		gkeys:   ui.DefaultGlobalKeys(),
		state:   stateIntro,
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
	cards, cardsErr := m.store.DrillableCards()
	pages, pagesErr := m.store.ListPages()
	victories, bossErr := m.store.BossVictories()
	m.cards = cards
	m.loadErr = errors.Join(cardsErr, pagesErr, bossErr)
	m.mastered = journey.CountMastered(pages)
	m.boss, m.bossFight = nil, false
	if next := journey.NextBoss(journey.Bosses(m.mastered, victories)); next != nil && next.Unlocked {
		m.boss = next
	}
	// Demons come from the chapters the pilgrimage has reached.
	m.battle = rpgengine.NewRandomBattle(m.cards, m.mastered, journey.Pilgrimage(pages).Percent)
	m.setQuestionKinds()
	m.state = stateIntro
	m.log = nil
	m.animTicks = 0
	return m.startSession()
}

// challengeBoss swaps the road demon for the unlocked boss, which quizzes
// the learner's weakest words.
func (m Model) challengeBoss() Model {
	m.cards = rpgengine.WeakCards(m.cards)
	m.battle = rpgengine.NewBossBattle(m.cards, journey.BossEntry(*m.boss), m.mastered)
	m.setQuestionKinds()
	m.bossFight = true
	return m
}

// setQuestionKinds asks for spoken questions when audio works, and typed
// ones when Chinese input is available.
func (m *Model) setQuestionKinds() {
	if m.battle != nil {
		m.battle.Listening = m.tts != nil && m.tts.Available()
		m.battle.Writing = m.writing
	}
}

// speakQuestion plays a listening question's word; nil for other questions.
func (m Model) speakQuestion() tea.Cmd {
	if q := m.battle.Question; q != nil && q.Category == "listening" {
		return ui.SpeakCmd(m.tts, q.Chinese, "zh")
	}
	return nil
}

// askQuestion readies the current question: it clears the answer box for a
// typed question and speaks a listening one.
func (m Model) askQuestion() (Model, tea.Cmd) {
	if q := m.battle.Question; q != nil && q.Category == "write" {
		m.input.Reset()
		m.input.SetMode(input.ModeChinese)
		return m, m.input.Focus()
	}
	return m, m.speakQuestion()
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
func (m *Model) recordAnswer(q *rpgengine.CombatQuestion, answer string, correct bool) {
	diff := model.DiffAgain
	if correct {
		diff = model.DiffHard
	}
	m.session.Items = append(m.session.Items, model.SessionItem{
		CardID:     q.Card.ID,
		UserAnswer: answer,
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
	if m.bossFight && m.battle.PlayerWon() {
		if err := m.store.RecordBossVictory(m.boss.Name, time.Now()); err != nil {
			m.saveErr = err
			return
		}
	}
	m.saved = true
	m.totalXP = m.store.TotalXP()
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case ui.SpeechDoneMsg:
		m.ttsErr = msg.Err
		return m, nil

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
			if msg.String() == "b" && m.boss != nil && !m.bossFight {
				return m.challengeBoss(), nil
			}
			if key.Matches(msg, m.gkeys.Enter) || msg.Type == tea.KeySpace {
				m.state = stateFighting
				m.battle.State = rpgengine.BattlePlayerTurn
				q := m.battle.GenerateQuestion(m.cards)
				m.battle.Question = q
				return m.askQuestion()
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
				return m.rematch().askQuestion()
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
	if q.Category == "write" {
		if key.Matches(msg, m.gkeys.Enter) && !m.input.Composing() {
			return m.submitText()
		}
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		return m, cmd
	}
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
	case msg.String() == "r":
		return m, m.speakQuestion()
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
	q := m.battle.Question
	correct, dmg := m.battle.Answer(m.battle.Selected)
	return m.afterAnswer(q, q.Choices[m.battle.Selected], correct, dmg)
}

// submitText grades a typed answer to a "write" question.
func (m Model) submitText() (Model, tea.Cmd) {
	text := strings.TrimSpace(m.input.Value())
	if text == "" {
		return m, nil
	}
	m.input.Blur()
	q := m.battle.Question
	correct, dmg := m.battle.AnswerText(text)
	return m.afterAnswer(q, text, correct, dmg)
}

// afterAnswer records an answer and starts the attack it earned.
func (m Model) afterAnswer(q *rpgengine.CombatQuestion, answer string, correct bool, dmg int) (Model, tea.Cmd) {
	m.lastAnswer = answer
	m.recordAnswer(q, answer, correct)
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
	return m.askQuestion()
}

// tickAfter returns a Cmd that sends animTickMsg after a delay.
func tickAfter(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(_ time.Time) tea.Msg {
		return animTickMsg{}
	})
}
