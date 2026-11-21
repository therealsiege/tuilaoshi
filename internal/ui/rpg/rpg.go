package rpg

import (
	"errors"
	"fmt"
	"slices"
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
	"github.com/therealsiege/tuilaoshi/internal/ui/reader"
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
	stateScene // the open stage's story scene, before its first fight
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
	order      []int  // pieces picked so far for an "order" question, by index
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

	// The stage this battle can clear, or nil for a practice battle; next
	// is the stage being worked on (nil once all are cleared).
	stage, next *journey.Stage

	// The story scene shown before a stage's first fight.
	scene       reader.Model
	sceneOf     *journey.Scene
	showEnglish bool
	mastered    int            // the journey's mastered count, which scales player stats
	counts      journey.Counts // characters learned and mastered, which open stages
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
	cleared, clearedErr := m.store.ClearedStages()
	m.cards = cards
	m.loadErr = errors.Join(cardsErr, pagesErr, clearedErr)
	m.counts = journey.CountProgress(pages)
	m.mastered = m.counts.Mastered
	stages := journey.Stages(m.counts, cleared)
	m.stage, m.next = nil, journey.CurrentStage(stages)
	if m.next != nil && m.next.Status == journey.StageOpen {
		m.stage = m.next
	}
	// Battles use the words of the stage being fought, or practise the
	// coming stage's words.
	if s := m.next; s != nil {
		m.cards = journey.StageCards(*s, pages, cards)
	}
	switch {
	case m.stage != nil && m.stage.Boss != nil:
		// Bosses quiz the learner's weakest of those words.
		m.cards = rpgengine.WeakCards(m.cards)
		m.battle = rpgengine.NewBossBattle(m.cards, m.stage.Opponent, m.mastered)
	case m.stage != nil:
		m.battle = rpgengine.NewBattle(m.cards, m.stage.Opponent, "", m.mastered)
	default:
		reached := len(stages)
		if m.next != nil {
			reached = m.next.Number
		}
		m.battle = rpgengine.NewRandomBattle(m.cards, m.mastered, rpgengine.ReachedDemons(stages, reached))
	}
	m.setQuestionKinds()
	m.state = stateIntro
	m.log = nil
	m.animTicks = 0
	m = m.startSession()
	return m.openScene()
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
	m.order = nil
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
	if m.stage != nil && m.battle.PlayerWon() {
		if err := m.store.RecordStageCleared(m.stage.Number, time.Now()); err != nil {
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
		if m.state == stateScene && !key.Matches(msg, m.gkeys.Back) {
			return m.updateScene(msg)
		}
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
	if q.Category == "order" {
		return m.handleOrderKey(msg)
	}
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

// handleOrderKey picks pieces with 1-9 and undoes with backspace; the
// sentence is graded once every piece is placed.
func (m Model) handleOrderKey(msg tea.KeyMsg) (Model, tea.Cmd) {
	q := m.battle.Question
	if msg.Type == tea.KeyBackspace {
		if len(m.order) > 0 {
			m.order = m.order[:len(m.order)-1]
		}
		return m, nil
	}
	s := msg.String()
	if len(s) != 1 || s[0] < '1' || s[0] > '9' {
		return m, nil
	}
	idx := int(s[0] - '1')
	if idx >= len(q.Pieces) || slices.Contains(m.order, idx) {
		return m, nil
	}
	m.order = append(m.order, idx)
	if len(m.order) < len(q.Pieces) {
		return m, nil
	}
	text := m.orderText()
	correct, dmg := m.battle.AnswerText(text)
	return m.afterAnswer(q, text, correct, dmg)
}

// orderText is the sentence built so far.
func (m Model) orderText() string {
	var b strings.Builder
	for _, i := range m.order {
		b.WriteString(m.battle.Question.Pieces[i])
	}
	return b.String()
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

// InputComposing reports whether a typed answer is mid-composition, so esc
// cancels it rather than fleeing.
func (m Model) InputComposing() bool { return m.input.Composing() }
