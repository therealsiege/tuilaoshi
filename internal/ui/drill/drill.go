package drill

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/therealsiege/tuilaoshi/internal/audio"
	"github.com/therealsiege/tuilaoshi/internal/claude"
	"github.com/therealsiege/tuilaoshi/internal/journey"
	"github.com/therealsiege/tuilaoshi/internal/model"
	"github.com/therealsiege/tuilaoshi/internal/ui"
	"github.com/therealsiege/tuilaoshi/internal/ui/input"
	"github.com/therealsiege/tuilaoshi/internal/wiki"
)

type state int

const (
	stateModePick   state = iota // choosing drill mode
	stateLoading                 // loading cards
	statePrompt                  // showing card, waiting for answer
	stateEval                    // evaluating answer with Claude
	stateFeedback                // showing Claude's feedback
	stateDifficulty              // waiting for difficulty rating
	stateDone                    // no more cards
)

type modeOption struct {
	mode  model.DrillMode
	label string
	desc  string
}

var modeOptions = []modeOption{
	{model.ModeRecognition, "Recognition 认读", "See Chinese → type meaning"},
	{model.ModeProduction, "Production 写作", "See English → type Chinese"},
	{model.ModeListening, "Listening 听力", "Hear Chinese → type what you heard"},
	{model.ModeSentence, "Sentence 造句", "Fill in the missing word"},
	{model.DrillMode("random"), "Random Mix 随机", "Mix of all modes"},
}

// Model is the drill view model.
type Model struct {
	store  *wiki.Store
	claude *claude.Client
	tts    *audio.Client
	gkeys  ui.GlobalKeys

	cards   []model.Card
	current int
	state   state
	mode    model.DrillMode

	// Mode picker
	modeCursor int

	input    input.Model
	feedback string
	result   model.DrillResult

	correct int
	total   int

	// Combo tracking
	combo    int    // current consecutive correct count
	maxCombo int    // best combo this session
	comboMsg string // milestone celebration message (cleared on next card)

	// Companion dialogue rotation
	dialogueCount int

	// Demon encounter
	encounter     *journey.Encounter // active demon encounter
	encounterHits int                // correct answers during encounter (need 3)
	totalAnswered int                // total cards answered (for trigger calculation)

	// Session tracking
	session model.Session

	width   int
	height  int
	err     error
	saveErr error // first failed save this drill, shown until the next drill
	ttsErr  error // why the last speech failed
}

// New creates a new drill model.
func New(store *wiki.Store, claudeClient *claude.Client, tts *audio.Client, dict *input.Dict) Model {
	ti := input.NewWithDict(dict)
	ti.SetPlaceholder("Type your answer...")
	ti.SetCharLimit(200)

	return Model{
		store:  store,
		claude: claudeClient,
		tts:    tts,
		gkeys:  ui.DefaultGlobalKeys(),
		input:  ti,
		mode:   model.ModeRecognition,
	}
}

func (m Model) Init() tea.Cmd {
	return nil // start at mode picker, no async needed
}

type loadCardsMsg struct{}

type evalResultMsg struct {
	result model.DrillResult
	err    error
}

// effectiveMode returns the mode for the current card.
// For "random" mode, picks a random mode per card.
func (m *Model) effectiveMode() model.DrillMode {
	if m.mode == model.DrillMode("random") {
		modes := []model.DrillMode{
			model.ModeRecognition,
			model.ModeProduction,
			model.ModeListening,
			model.ModeSentence,
		}
		return modes[rand.Intn(len(modes))]
	}
	return m.mode
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case loadCardsMsg:
		due, err := wiki.DueItems(m.store, time.Now())
		if err != nil {
			m.err = err
			m.state = stateDone
			return m, nil
		}
		m.saveErr = nil
		if len(due) == 0 {
			m.cards, m.err = m.store.DrillableCards()
		} else {
			m.cards = due
		}

		// Shuffle cards
		rand.Shuffle(len(m.cards), func(i, j int) {
			m.cards[i], m.cards[j] = m.cards[j], m.cards[i]
		})

		if len(m.cards) == 0 {
			m.state = stateDone
			return m, nil
		}
		m.current = 0
		m.state = statePrompt
		m.input.Reset()
		m.input.Focus()
		m.session = model.Session{
			StartedAt: time.Now(),
			Type:      "drill",
		}

		card := m.cards[m.current]
		return m, tea.Batch(textinput.Blink, ui.SpeakCmd(m.tts, card.Chinese, "zh"))

	case evalResultMsg:
		if msg.err != nil {
			m.feedback = fmt.Sprintf("Error: %v", msg.err)
			m.result = model.DrillResult{Feedback: m.feedback}
		} else {
			m.result = msg.result
			m.feedback = msg.result.Feedback
			if msg.result.Correct {
				m.correct++
			}
		}
		m.total++
		m.totalAnswered++
		m.dialogueCount++

		// Track encounter hits on correct answers
		if m.encounter != nil && m.result.Correct {
			m.encounterHits++
		}

		// Check for new encounter trigger (only when not already in one)
		if m.encounter == nil {
			if enc := journey.RandomEncounter(m.totalAnswered); enc != nil {
				m.encounter = enc
				m.encounterHits = 0
			}
		}

		m.state = stateDifficulty
		return m, nil

	case ui.SpeechDoneMsg:
		m.ttsErr = msg.Err
		return m, nil

	case tea.KeyMsg:
		switch m.state {
		case stateModePick:
			switch {
			case key.Matches(msg, m.gkeys.Up):
				if m.modeCursor > 0 {
					m.modeCursor--
				}
			case key.Matches(msg, m.gkeys.Down):
				if m.modeCursor < len(modeOptions)-1 {
					m.modeCursor++
				}
			case key.Matches(msg, m.gkeys.Enter):
				m.mode = modeOptions[m.modeCursor].mode
				m.state = stateLoading
				return m, func() tea.Msg { return loadCardsMsg{} }
			}
			return m, nil

		case statePrompt:
			// Let the input handle Ctrl+Space for mode toggle
			if msg.Type == tea.KeyCtrlAt {
				var cmd tea.Cmd
				m.input, cmd = m.input.Update(msg)
				return m, cmd
			}
			if msg.Type == tea.KeyCtrlR {
				return m, ui.SpeakCmd(m.tts, m.cards[m.current].Chinese, "zh")
			}

			if key.Matches(msg, m.gkeys.Enter) {
				answer := m.input.Value()
				if answer == "" {
					return m, nil
				}
				m.state = stateEval
				card := m.cards[m.current]
				mode := m.effectiveMode()
				if result, ok := exactChinese(card, answer, mode); ok {
					return m, func() tea.Msg { return evalResultMsg{result: result} }
				}
				return m, func() tea.Msg {
					result, err := m.claude.EvaluateDrill(context.Background(), card, answer, mode)
					return evalResultMsg{result: result, err: err}
				}
			}

			var cmd tea.Cmd
			m.input, cmd = m.input.Update(msg)
			return m, cmd

		case stateFeedback:
			if key.Matches(msg, m.gkeys.Enter) || msg.Type == tea.KeySpace {
				return m, m.advanceCard(model.DiffGood)
			}

		case stateDifficulty:
			switch msg.String() {
			case "1":
				return m, m.advanceCard(model.DiffAgain)
			case "2":
				return m, m.advanceCard(model.DiffHard)
			case "3":
				return m, m.advanceCard(model.DiffGood)
			case "4":
				return m, m.advanceCard(model.DiffEasy)
			case "enter", " ":
				// Default to Good
				return m, m.advanceCard(model.DiffGood)
			}
			return m, nil
		}
	}

	return m, nil
}

func (m *Model) advanceCard(diff model.DrillDifficulty) tea.Cmd {
	// Update combo based on difficulty rating
	if diff == model.DiffAgain {
		m.combo = 0
	} else {
		m.combo++
		if m.combo > m.maxCombo {
			m.maxCombo = m.combo
		}
	}

	// Set milestone celebration message
	m.comboMsg = comboMilestone(m.combo)

	// Clear encounter on victory (3 correct hits)
	if m.encounter != nil && m.encounterHits >= 3 {
		m.encounter = nil
		m.encounterHits = 0
	}

	// Record session item
	if m.current < len(m.cards) {
		card := &m.cards[m.current]
		m.noteSaveErr(m.store.RecordReview(card, diff, time.Now()))
		m.session.Items = append(m.session.Items, model.SessionItem{
			CardID:     card.ID,
			UserAnswer: m.input.Value(),
			Correct:    m.result.Correct,
			Score:      m.result.Score,
			Difficulty: diff,
			Feedback:   m.feedback,
		})
	}

	m.current++
	if m.current >= len(m.cards) {
		m.state = stateDone
		// Save session
		m.session.EndedAt = time.Now()
		m.session.CardsTotal = m.total
		m.session.CardsCorrect = m.correct
		m.noteSaveErr(m.store.WriteSession(m.session))
		return nil
	}
	m.state = statePrompt
	m.input.Reset()
	m.input.Focus()

	card := m.cards[m.current]
	return tea.Batch(textinput.Blink, ui.SpeakCmd(m.tts, card.Chinese, "zh"))
}

// exactChinese grades a typed answer locally when it is exactly the word a
// production, listening, or sentence card asks for, so the common case needs
// no Claude call and works offline. Anything else goes to Claude.
func exactChinese(card model.Card, answer string, mode model.DrillMode) (model.DrillResult, bool) {
	switch mode {
	case model.ModeProduction, model.ModeListening, model.ModeSentence:
	default:
		return model.DrillResult{}, false
	}
	if card.Chinese == "" || strings.TrimSpace(answer) != card.Chinese {
		return model.DrillResult{}, false
	}
	return model.DrillResult{Correct: true, Score: 5, Feedback: "完全正确！Exactly right: " + card.Chinese}, true
}

// noteSaveErr keeps the first failed save so the player sees it.
func (m *Model) noteSaveErr(err error) {
	if err != nil && m.saveErr == nil {
		m.saveErr = err
	}
}

// InputComposing reports whether the input is in Chinese composition mode.
func (m Model) InputComposing() bool {
	return m.state == statePrompt && m.input.Composing()
}

// SetSize updates terminal dimensions.
func (m Model) SetSize(w, h int) Model {
	m.width = w
	m.height = h
	return m
}

// Refresh reloads cards.
func (m Model) Refresh() Model {
	m.current = 0
	m.correct = 0
	m.total = 0
	m.totalAnswered = 0
	m.encounter = nil
	m.encounterHits = 0
	m.state = stateModePick
	m.modeCursor = 0
	m.feedback = ""
	m.input.Reset()
	return m
}
