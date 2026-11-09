package review

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/cj/mandarin-wiki/internal/audio"
	"github.com/cj/mandarin-wiki/internal/claude"
	"github.com/cj/mandarin-wiki/internal/model"
	"github.com/cj/mandarin-wiki/internal/ui"
	"github.com/cj/mandarin-wiki/internal/wiki"
)

type reviewState int

const (
	stateList       reviewState = iota // showing due items list
	stateDrill                         // drilling a selected item
	stateEval                          // evaluating answer
	stateDifficulty                    // showing feedback + difficulty buttons
)

type ttsFinishedMsg struct{}

// Model is the review queue view model.
type Model struct {
	store  *wiki.Store
	claude *claude.Client
	tts    *audio.Client
	gkeys  ui.GlobalKeys

	items  []model.Card
	cursor int
	state  reviewState

	// Drill state
	input    textinput.Model
	feedback string
	result   model.DrillResult
	correct  int
	total    int

	// Session tracking
	session model.Session

	width  int
	height int
}

// New creates a new review model.
func New(store *wiki.Store, claudeClient *claude.Client, tts *audio.Client) Model {
	ti := textinput.New()
	ti.Placeholder = "Type your answer..."
	ti.CharLimit = 200

	return Model{
		store:  store,
		claude: claudeClient,
		tts:    tts,
		gkeys:  ui.DefaultGlobalKeys(),
		input:  ti,
	}
}

func (m Model) Init() tea.Cmd {
	return func() tea.Msg { return loadDueMsg{} }
}

type loadDueMsg struct{}

type reviewEvalMsg struct {
	result model.DrillResult
	err    error
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case loadDueMsg:
		due, _ := wiki.DueItems(m.store, time.Now())
		m.items = due
		m.cursor = 0
		m.state = stateList
		m.session = model.Session{
			StartedAt: time.Now(),
			Type:      "review",
		}
		return m, nil

	case reviewEvalMsg:
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
		case stateList:
			switch {
			case key.Matches(msg, m.gkeys.Up):
				if m.cursor > 0 {
					m.cursor--
				}
			case key.Matches(msg, m.gkeys.Down):
				if m.cursor < len(m.items)-1 {
					m.cursor++
				}
			case key.Matches(msg, m.gkeys.Enter):
				if len(m.items) > 0 {
					m.state = stateDrill
					m.input.Reset()
					m.input.Focus()
					card := m.items[m.cursor]
					var cmds []tea.Cmd
					cmds = append(cmds, textinput.Blink)
					if m.tts != nil && m.tts.Available() && card.Chinese != "" {
						cmds = append(cmds, func() tea.Msg {
							_ = m.tts.Speak(card.Chinese, "zh")
							return ttsFinishedMsg{}
						})
					}
					return m, tea.Batch(cmds...)
				}
			}
			return m, nil

		case stateDrill:
			if key.Matches(msg, m.gkeys.Enter) {
				answer := m.input.Value()
				if answer == "" {
					return m, nil
				}
				m.state = stateEval
				card := m.items[m.cursor]
				return m, func() tea.Msg {
					result, err := m.claude.EvaluateDrill(context.Background(), card, answer, model.ModeRecognition)
					return reviewEvalMsg{result: result, err: err}
				}
			}
			var cmd tea.Cmd
			m.input, cmd = m.input.Update(msg)
			return m, cmd

		case stateDifficulty:
			var diff model.DrillDifficulty
			switch msg.String() {
			case "1":
				diff = model.DiffAgain
			case "2":
				diff = model.DiffHard
			case "3":
				diff = model.DiffGood
			case "4":
				diff = model.DiffEasy
			case "enter", " ":
				diff = model.DiffGood
			default:
				return m, nil
			}

			// Update SRS with difficulty
			if m.cursor < len(m.items) {
				card := &m.items[m.cursor]
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

			// Move to next due item or back to list
			m.cursor++
			if m.cursor >= len(m.items) {
				// Save session and reload
				m.session.EndedAt = time.Now()
				m.session.CardsTotal = m.total
				m.session.CardsCorrect = m.correct
				_ = m.store.WriteSession(m.session)

				due, _ := wiki.DueItems(m.store, time.Now())
				m.items = due
				m.cursor = 0
				m.state = stateList
			} else {
				m.state = stateDrill
				m.input.Reset()
				m.input.Focus()
				card := m.items[m.cursor]
				var cmds []tea.Cmd
				cmds = append(cmds, textinput.Blink)
				if m.tts != nil && m.tts.Available() && card.Chinese != "" {
					cmds = append(cmds, func() tea.Msg {
						_ = m.tts.Speak(card.Chinese, "zh")
						return ttsFinishedMsg{}
					})
				}
				return m, tea.Batch(cmds...)
			}
			return m, nil
		}
	}

	return m, nil
}

func (m Model) View() string {
	var b strings.Builder

	title := ui.TitleStyle.Render("Review Queue 复习")
	b.WriteString(title)
	b.WriteString("\n\n")

	switch m.state {
	case stateList:
		if len(m.items) == 0 {
			b.WriteString(ui.SuccessStyle.Render("  All caught up! No items due for review."))
			b.WriteString("\n\n")
			b.WriteString(ui.HelpStyle.Render("  esc back to dashboard"))
			return b.String()
		}

		b.WriteString(ui.SubtitleStyle.Render(fmt.Sprintf("  %d items due for review", len(m.items))))
		b.WriteString("\n\n")

		for i, card := range m.items {
			line := fmt.Sprintf("  %s  %s",
				card.Chinese,
				ui.MutedStyle.Render(fmt.Sprintf("[%s]", card.Type)),
			)
			if card.English != "" {
				line += "  " + ui.MutedStyle.Render(card.English)
			}
			if i == m.cursor {
				b.WriteString(ui.SelectedStyle.Render("▸ " + line))
			} else {
				b.WriteString("  " + line)
			}
			b.WriteString("\n")
		}
		b.WriteString("\n")
		b.WriteString(ui.HelpStyle.Render("  ↑↓ navigate • enter start drill • esc back"))

	case stateDrill:
		if m.cursor < len(m.items) {
			card := m.items[m.cursor]
			cardView := fmt.Sprintf("  What does this mean?\n\n  %s",
				ui.ChineseStyle.Render(card.Chinese))
			if card.Pinyin != "" {
				cardView += "\n  " + ui.PinyinStyle.Render(card.Pinyin)
			}
			b.WriteString(ui.CardStyle.Render(cardView))
			b.WriteString("\n\n")
			b.WriteString("  " + m.input.View())
			b.WriteString("\n\n")
			b.WriteString(ui.HelpStyle.Render("  enter submit • esc back"))
		}

	case stateEval:
		b.WriteString(ui.MutedStyle.Render("  Evaluating with Claude..."))

	case stateDifficulty:
		if m.result.Correct {
			b.WriteString(ui.SuccessStyle.Render("  ✓ Correct!"))
		} else {
			b.WriteString(ui.ErrorStyle.Render("  ✗ Incorrect"))
		}
		b.WriteString("\n\n")
		b.WriteString("  " + m.feedback)
		b.WriteString("\n\n")
		b.WriteString("  Rate difficulty:\n")
		b.WriteString(fmt.Sprintf("  %s  %s  %s  %s\n",
			ui.DiffAgainStyle.Render("[1] Again"),
			ui.DiffHardStyle.Render("[2] Hard"),
			ui.DiffGoodStyle.Render("[3] Good"),
			ui.DiffEasyStyle.Render("[4] Easy"),
		))
		b.WriteString("\n")
		remaining := len(m.items) - m.cursor - 1
		if remaining > 0 {
			b.WriteString(ui.MutedStyle.Render(fmt.Sprintf("  %d items remaining", remaining)))
			b.WriteString("\n")
		}
		b.WriteString(ui.HelpStyle.Render("  1-4 rate • enter default (Good) • esc back"))
	}

	// Progress
	if m.total > 0 {
		b.WriteString("\n\n")
		progress := fmt.Sprintf("  Reviewed: %d  •  Correct: %d/%d", m.total, m.correct, m.total)
		b.WriteString(ui.StatusBarStyle.Render(progress))
	}

	return b.String()
}

// SetSize updates terminal dimensions.
func (m Model) SetSize(w, h int) Model {
	m.width = w
	m.height = h
	return m
}

// Refresh reloads due items.
func (m Model) Refresh() Model {
	due, _ := wiki.DueItems(m.store, time.Now())
	m.items = due
	m.cursor = 0
	m.correct = 0
	m.total = 0
	m.state = stateList
	return m
}
