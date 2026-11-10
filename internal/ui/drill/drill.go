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
	"github.com/cj/mandarin-wiki/internal/audio"
	"github.com/cj/mandarin-wiki/internal/claude"
	"github.com/cj/mandarin-wiki/internal/model"
	"github.com/cj/mandarin-wiki/internal/ui"
	"github.com/cj/mandarin-wiki/internal/ui/input"
	"github.com/cj/mandarin-wiki/internal/wiki"
)

type state int

const (
	stateModePick state = iota // choosing drill mode
	stateLoading               // loading cards
	statePrompt                // showing card, waiting for answer
	stateEval                  // evaluating answer with Claude
	stateFeedback              // showing Claude's feedback
	stateDifficulty            // waiting for difficulty rating
	stateDone                  // no more cards
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

	// Session tracking
	session model.Session

	width  int
	height int
	err    error
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

type ttsFinishedMsg struct{}

// speakCmd returns a Cmd that speaks text in the background.
func speakCmd(tts *audio.Client, text, lang string) tea.Cmd {
	if tts == nil || !tts.Available() || text == "" {
		return nil
	}
	return func() tea.Msg {
		_ = tts.Speak(text, lang)
		return ttsFinishedMsg{}
	}
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
		if len(due) == 0 {
			pages, _ := m.store.ListPages()
			var drillable []model.Page
			for _, p := range pages {
				switch p.Type {
				case "character", "vocabulary", "grammar":
					drillable = append(drillable, p)
				}
			}
			m.cards = wiki.CardsFromPages(drillable)
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
		return m, tea.Batch(textinput.Blink, speakCmd(m.tts, card.Chinese, "zh"))

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

		m.state = stateDifficulty
		return m, nil

	case ttsFinishedMsg:
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

			if key.Matches(msg, m.gkeys.Enter) {
				answer := m.input.Value()
				if answer == "" {
					return m, nil
				}
				m.state = stateEval
				card := m.cards[m.current]
				mode := m.effectiveMode()
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
	// Record session item
	if m.current < len(m.cards) {
		card := &m.cards[m.current]
		wiki.UpdateSRS(card, diff, time.Now())
		page, err := m.store.ReadPage(card.ID)
		if err == nil {
			wiki.ApplySRSToPage(&page, *card)
			_ = m.store.WriteFrontmatter(page)
		}
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
		_ = m.store.WriteSession(m.session)
		return nil
	}
	m.state = statePrompt
	m.input.Reset()
	m.input.Focus()

	card := m.cards[m.current]
	return tea.Batch(textinput.Blink, speakCmd(m.tts, card.Chinese, "zh"))
}

func (m Model) View() string {
	var b strings.Builder

	title := ui.TitleStyle.Render("Drill 练习")
	b.WriteString(title)
	b.WriteString("\n\n")

	switch m.state {
	case stateModePick:
		b.WriteString(ui.SubtitleStyle.Render("  Select Drill Mode"))
		b.WriteString("\n\n")
		for i, opt := range modeOptions {
			label := fmt.Sprintf("  %s", opt.label)
			desc := ui.MutedStyle.Render("  " + opt.desc)
			if i == m.modeCursor {
				b.WriteString(ui.SelectedStyle.Render("▸ "+label) + "\n")
				b.WriteString("    " + desc + "\n")
			} else {
				b.WriteString(ui.UnselectedStyle.Render("  "+label) + "\n")
			}
		}
		b.WriteString("\n")
		b.WriteString(ui.HelpStyle.Render("  ↑↓ select • enter confirm • esc back"))

	case stateLoading:
		b.WriteString(ui.MutedStyle.Render("Loading cards..."))

	case statePrompt:
		if m.current < len(m.cards) {
			card := m.cards[m.current]
			b.WriteString(m.renderCard(card))
			b.WriteString("\n\n")
			b.WriteString("  " + m.input.View())
			b.WriteString("\n\n")
			b.WriteString(ui.HelpStyle.Render("  enter submit • ctrl+space toggle input • esc back"))
		}

	case stateEval:
		b.WriteString(ui.MutedStyle.Render("  Evaluating with Claude..."))

	case stateDifficulty:
		if m.current < len(m.cards) {
			card := m.cards[m.current]
			b.WriteString(m.renderCard(card))
			b.WriteString("\n\n")
		}
		if m.result.Correct {
			b.WriteString(ui.SuccessStyle.Render("  ✓ Correct!"))
		} else {
			b.WriteString(ui.ErrorStyle.Render("  ✗ Incorrect"))
		}
		b.WriteString("\n\n  ")
		b.WriteString(m.feedback)
		b.WriteString("\n\n")
		b.WriteString("  Rate difficulty:\n")
		b.WriteString(fmt.Sprintf("  %s  %s  %s  %s\n",
			ui.DiffAgainStyle.Render("[1] Again"),
			ui.DiffHardStyle.Render("[2] Hard"),
			ui.DiffGoodStyle.Render("[3] Good"),
			ui.DiffEasyStyle.Render("[4] Easy"),
		))
		b.WriteString("\n")
		b.WriteString(ui.HelpStyle.Render("  1-4 rate • enter default (Good) • esc back"))

	case stateFeedback:
		// Transitional state, same as difficulty
		b.WriteString(ui.MutedStyle.Render("  Processing..."))

	case stateDone:
		if m.err != nil {
			b.WriteString(ui.ErrorStyle.Render(fmt.Sprintf("Error: %v", m.err)))
		} else if m.total == 0 {
			b.WriteString(ui.MutedStyle.Render("  No cards available for drilling."))
			b.WriteString("\n\n")
			b.WriteString(ui.MutedStyle.Render("  Add character, vocabulary, or grammar pages to the wiki to start drilling."))
		} else {
			b.WriteString(ui.SubtitleStyle.Render("  Session Complete 完成"))
			b.WriteString("\n\n")
			b.WriteString(fmt.Sprintf("  Cards:   %d\n", m.total))
			b.WriteString(fmt.Sprintf("  Correct: %d\n", m.correct))
			pct := float64(m.correct) / float64(m.total) * 100
			scoreText := fmt.Sprintf("  Score:   %.0f%%", pct)
			if pct >= 80 {
				b.WriteString(ui.SuccessStyle.Render(scoreText))
			} else if pct >= 50 {
				b.WriteString(ui.WarningStyle.Render(scoreText))
			} else {
				b.WriteString(ui.ErrorStyle.Render(scoreText))
			}
			b.WriteString("\n\n")
			b.WriteString("  " + ui.ProgressBar(m.correct, m.total, 30))
		}
		b.WriteString("\n\n")
		b.WriteString(ui.HelpStyle.Render("  esc back to dashboard"))
	}

	// Progress bar
	if m.state != stateDone && m.state != stateModePick && len(m.cards) > 0 {
		b.WriteString("\n\n")
		progress := fmt.Sprintf("  Card %d/%d", m.current+1, len(m.cards))
		if m.total > 0 {
			progress += fmt.Sprintf("  •  %d/%d correct", m.correct, m.total)
		}
		b.WriteString(ui.StatusBarStyle.Render(progress))
	}

	return b.String()
}

func (m Model) renderCard(card model.Card) string {
	mode := m.effectiveMode()

	var content string
	switch mode {
	case model.ModeRecognition:
		chinese := ui.ChineseStyle.Render(card.Chinese)
		content = fmt.Sprintf("  What does this mean?\n\n  %s", chinese)
		if card.Pinyin != "" {
			content += "\n  " + ui.PinyinStyle.Render(card.Pinyin)
		}
	case model.ModeProduction:
		content = fmt.Sprintf("  How do you say this in Chinese?\n\n  %s", card.English)
		if card.Pinyin != "" {
			content += "\n  " + ui.PinyinStyle.Render("("+card.Pinyin+")")
		}
	case model.ModeListening:
		content = "  What did you hear? Type it in Chinese.\n\n"
		content += ui.ChineseStyle.Render("  🔊 [Listen]")
	case model.ModeSentence:
		if len(card.Examples) > 0 {
			// Pick a random example and blank out part of it
			ex := card.Examples[rand.Intn(len(card.Examples))]
			// Remove the "- " prefix if present
			ex = strings.TrimPrefix(ex, "- ")
			content = fmt.Sprintf("  Fill in the blank:\n\n  %s\n\n  Hint: %s",
				blankOut(ex, card.Chinese),
				ui.PinyinStyle.Render(card.English))
		} else {
			// Fallback to recognition
			chinese := ui.ChineseStyle.Render(card.Chinese)
			content = fmt.Sprintf("  What does this mean?\n\n  %s", chinese)
		}
	}

	typeLabel := ui.MutedStyle.Render(fmt.Sprintf("[%s]", card.Type))
	modeLabel := ui.MutedStyle.Render(fmt.Sprintf(" %s", mode))
	return ui.CardStyle.Render(typeLabel + modeLabel + "\n" + content)
}

// blankOut replaces the target word in a sentence with _____.
func blankOut(sentence, word string) string {
	if word == "" {
		return sentence
	}
	if idx := strings.Index(sentence, word); idx >= 0 {
		return sentence[:idx] + ui.ChineseStyle.Render("_____") + sentence[idx+len(word):]
	}
	return sentence + " " + ui.ChineseStyle.Render("_____")
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
	m.state = stateModePick
	m.modeCursor = 0
	m.feedback = ""
	m.input.Reset()
	return m
}
