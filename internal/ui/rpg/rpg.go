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
	"github.com/therealsiege/tuilaoshi/internal/episode"
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

	// Abilities in play, as the intro describes them, and the answer timer
	// of a burning enemy: qseq numbers questions so stale ticks are ignored.
	abilities []string
	words     journey.WordProgress // learned words of the stage's theme
	deadline  time.Time
	qseq      int
	now       func() time.Time
	mastered  int            // level points (learned + mastered characters), which set player HP
	counts    journey.Counts // characters learned and mastered, which open stages

	// A stage's soldiers, fought before its demon (see gauntlet.go); foe is
	// the current fight, len(gauntlet) meaning the demon.
	gauntlet []journey.Soldier
	foe      int
	cleared  map[int]string    // cleared stages, whose companions have joined
	radicals map[string]string // character -> radical, for answer hints

	// Coins paid by the last fight, and the wallet after it.
	earned, wallet wiki.Wallet
	bonded         []string // companions whose bond the last fight raised

	// Battles adapt to recent accuracy (accuracy percent right).
	difficulty rpgengine.Difficulty
	accuracy   int

	// spirit is the hidden spirit fought in this practice battle, if any.
	spirit *journey.Spirit
	// nightmare is the beaten boss brought back in a nightmare rematch, if any.
	nightmare *journey.Boss
	// road is set for battles fought on the road (see road.go).
	road bool
	// loot is the materials the last fight earned (see road.go).
	loot []string
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
	m.mastered = m.counts.LevelPoints()
	stages := journey.Stages(m.counts, cleared)
	m.stage, m.next = nil, journey.CurrentStage(stages)
	if m.road && m.next != nil && m.next.Status == journey.StageOpen {
		m.stage = m.next
	}
	// Battles use the words of the stage being fought, or practise the
	// coming stage's words.
	if s := m.next; s != nil {
		m.cards = journey.StageCards(*s, pages, cards)
		if opened, err := episode.OpenedNow(m.store); err == nil {
			m.cards = episode.WithLatest(m.cards, cards, opened)
		}
		m.words = journey.StageProgress(*s, pages, cards)
	}
	m.gauntlet, m.foe, m.cleared, m.nightmare = nil, 0, cleared, nil
	if m.store != nil {
		pct, answered := m.store.RecentAccuracy(rpgengine.RecentAnswers)
		m.accuracy, m.difficulty = pct, rpgengine.DifficultyFor(pct, answered)
		switch ui.Prefs.Difficulty { // the learner's setting overrides adapting
		case "gentler":
			m.difficulty = rpgengine.Easier
		case "sterner":
			m.difficulty = rpgengine.Harder
		}
	}
	switch {
	case m.stage != nil:
		if m.stage.Boss != nil {
			// Bosses, and their soldiers, quiz the learner's weakest of those words.
			m.cards = rpgengine.WeakCards(m.cards)
		}
		m.startGauntlet()
	case m.meetSpirit(pages, cards):
	default:
		reached := len(stages)
		if m.next != nil {
			reached = m.next.Number
		}
		m.battle = rpgengine.NewPracticeBattle(m.cards, m.mastered,
			rpgengine.ReachedDemons(stages, reached), rpgengine.ReachedSoldiers(reached))
	}
	m.setQuestionKinds()
	m.applyAbilities(cleared)
	m.radicals = journey.RadicalIndex(pages)
	if m.battle != nil {
		m.battle.Radicals = m.radicals
	}
	m.state = stateIntro
	m.log = nil
	m.animTicks = 0
	m = m.startSession()
	return m.openScene()
}

// Leave saves the battle's results when the player flees mid-fight.
func (m Model) Leave() Model {
	m.keepWounds()
	m.saveSession()
	return m
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case ui.SpeechDoneMsg:
		m.ttsErr = msg.Err
		return m, nil

	case timerMsg:
		return m.onTimer(msg)

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
			return m, ui.SwitchTo(ui.ViewMap)
		}

		switch m.state {
		case stateIntro:
			if msg.String() == "n" && m.nightmare == nil && m.nightmareBoss() != nil {
				return m.startNightmare(), nil
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
				return m.onVictoryKey()
			}

		case stateDefeat:
			if key.Matches(msg, m.gkeys.Enter) || msg.Type == tea.KeySpace {
				if m.road {
					return m.fall()
				}
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
	case msg.String() == "w":
		return m.reveal(), nil
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
	m.qseq++ // stops the answer timer
	m.recordAnswer(q, answer, correct)
	if e := m.battle.Event; e != "" {
		m.log = appendLog(m.log, logEntry{text: e, style: ui.SuccessStyle})
	}
	if correct {
		ui.Cue(audio.Right)
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
		ui.Cue(audio.Wrong)
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
			ui.Cue(audio.Victory)
			m.state = stateVictory
			m.battle.State = rpgengine.BattleVictory
		} else {
			ui.Cue(audio.Defeat)
			m.state = stateDefeat
			m.battle.State = rpgengine.BattleDefeat
		}
		if m.battle.PlayerWon() {
			m.keepWounds() // before saving: the session's fresh answers then heal
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
