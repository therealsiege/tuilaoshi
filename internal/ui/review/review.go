package review

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/therealsiege/tuilaoshi/internal/audio"
	"github.com/therealsiege/tuilaoshi/internal/claude"
	"github.com/therealsiege/tuilaoshi/internal/episode"
	"github.com/therealsiege/tuilaoshi/internal/journey"
	"github.com/therealsiege/tuilaoshi/internal/mnemonic"
	"github.com/therealsiege/tuilaoshi/internal/model"
	"github.com/therealsiege/tuilaoshi/internal/ui"
	"github.com/therealsiege/tuilaoshi/internal/ui/input"
	"github.com/therealsiege/tuilaoshi/internal/wiki"
)

type reviewState int

const (
	stateList       reviewState = iota // showing due items list
	stateDrill                         // drilling a selected item
	stateEval                          // evaluating answer
	stateDifficulty                    // showing feedback + difficulty buttons
)

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
	input    input.Model
	feedback string
	result   model.DrillResult
	correct  int
	total    int

	// Companion dialogue rotation
	dialogueCount int

	// Session tracking
	session model.Session
	saveErr error // first failed save this visit, shown until the queue reloads
	loadErr error // why the review queue could not be read
	ttsErr  error // why the last speech failed

	width  int
	height int
}

// New creates a new review model.
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
		m.items, m.loadErr = wiki.DueItems(m.store, time.Now())
		m.items = m.episodeWordsFirst(m.items)
		m.cursor = 0
		m.state = stateList
		m.saveErr = nil
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
		m.dialogueCount++

		m.state = stateDifficulty
		return m, nil

	case ui.SpeechDoneMsg:
		m.ttsErr = msg.Err
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
					return m, tea.Batch(textinput.Blink, ui.SpeakCmd(m.tts, card.Chinese, "zh"))
				}
			}
			return m, nil

		case stateDrill:
			// Let the input handle Ctrl+Space for mode toggle
			if msg.Type == tea.KeyCtrlAt {
				var cmd tea.Cmd
				m.input, cmd = m.input.Update(msg)
				return m, cmd
			}
			if msg.Type == tea.KeyCtrlR {
				return m, ui.SpeakCmd(m.tts, m.items[m.cursor].Chinese, "zh")
			}

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

			// Move to next due item or back to list
			m.cursor++
			if m.cursor >= len(m.items) {
				// Save session and reload
				m.session.EndedAt = time.Now()
				m.session.CardsTotal = m.total
				m.session.CardsCorrect = m.correct
				m.noteSaveErr(m.store.WriteSession(m.session))

				m.items, m.loadErr = wiki.DueItems(m.store, time.Now())
				m.items = m.episodeWordsFirst(m.items)
				m.cursor = 0
				m.state = stateList
			} else {
				m.state = stateDrill
				m.input.Reset()
				m.input.Focus()
				card := m.items[m.cursor]
				return m, tea.Batch(textinput.Blink, ui.SpeakCmd(m.tts, card.Chinese, "zh"))
			}
			return m, nil
		}
	}

	return m, nil
}

// noteSaveErr keeps the first failed save so the player sees it.
func (m *Model) noteSaveErr(err error) {
	if err != nil && m.saveErr == nil {
		m.saveErr = err
	}
}

func (m Model) View() string {
	var b strings.Builder

	title := ui.TitleStyle.Render("Review Queue 复习")
	b.WriteString(title)
	b.WriteString("\n\n")
	b.WriteString(ui.LoadError(m.loadErr))
	b.WriteString(ui.AudioError(m.ttsErr))
	if m.saveErr != nil {
		b.WriteString(ui.ErrorStyle.Render(fmt.Sprintf("  Could not save progress: %v", m.saveErr)))
		b.WriteString("\n\n")
	}

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

		first, last := 0, len(m.items)
		if room := m.height - 10; m.height > 0 && room < len(m.items) {
			room = max(room, 3)
			first = min(max(m.cursor-room/2, 0), len(m.items)-room)
			last = first + room
		}
		if first > 0 {
			b.WriteString(ui.MutedStyle.Render(fmt.Sprintf("    … %d above", first)) + "\n")
		}
		for i := first; i < last; i++ {
			card := m.items[i]
			line := fmt.Sprintf("  %s  %s",
				card.Chinese,
				ui.MutedStyle.Render(fmt.Sprintf("[%s]", card.Type)),
			)
			if en := card.English; en != "" {
				if m.width > 0 {
					if room := m.width - lipgloss.Width(line) - 6; room > 3 && lipgloss.Width(en) > room {
						en = ui.WrapDisplay(en, room-1)[0] + "…"
					}
				}
				line += "  " + ui.MutedStyle.Render(en)
			}
			if i == m.cursor {
				b.WriteString(ui.SelectedStyle.Render("▸ " + line))
			} else {
				b.WriteString("  " + line)
			}
			b.WriteString("\n")
		}
		if last < len(m.items) {
			b.WriteString(ui.MutedStyle.Render(fmt.Sprintf("    … %d below", len(m.items)-last)) + "\n")
		}
		b.WriteString(ui.HelpStyle.Render("  ↑↓ navigate · enter start drill · esc back"))

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
			b.WriteString(ui.HelpStyle.Render("  enter submit · ctrl+space toggle input · ctrl+r replay · esc back"))
		}

	case stateEval:
		b.WriteString(ui.MutedStyle.Render("  Checking your answer…"))

	case stateDifficulty:
		if m.result.Correct {
			b.WriteString(ui.SuccessStyle.Render("  ✓ Correct!"))
		} else {
			b.WriteString(ui.ErrorStyle.Render("  ✗ Incorrect"))
		}
		b.WriteString("\n\n")
		width := 76
		if m.width > 0 {
			width = min(width, m.width-4)
		}
		b.WriteString("  " + strings.Join(ui.WrapDisplay(m.feedback, width), "\n  "))
		if m.cursor < len(m.items) {
			if line := formatDecomposition(m.items[m.cursor].Chinese); line != "" {
				b.WriteString("\n\n")
				b.WriteString(ui.MutedStyle.Render(line))
			}
		}
		// Companion dialogue
		if m.result.Correct {
			if d := journey.RandomDialogue("streak", m.dialogueCount); d != nil {
				b.WriteString("\n")
				b.WriteString(ui.MutedStyle.Render("  " + ui.ChineseStyle.Render(d.Speaker) + ": " + d.Line))
			}
		} else {
			if d := journey.RandomDialogue("mistake", m.dialogueCount); d != nil {
				b.WriteString("\n")
				b.WriteString(ui.MutedStyle.Render("  " + ui.ChineseStyle.Render(d.Speaker) + ": " + d.Line))
			}
		}
		b.WriteString("\n\n")
		b.WriteString("  Rate difficulty:\n")
		fmt.Fprintf(&b, "  %s  %s  %s  %s\n",
			ui.DiffAgainStyle.Render("[1] Again"),
			ui.DiffHardStyle.Render("[2] Hard"),
			ui.DiffGoodStyle.Render("[3] Good"),
			ui.DiffEasyStyle.Render("[4] Easy"),
		)
		b.WriteString("\n")
		remaining := len(m.items) - m.cursor - 1
		if remaining > 0 {
			b.WriteString(ui.MutedStyle.Render(fmt.Sprintf("  %d items remaining", remaining)))
			b.WriteString("\n")
		}
		b.WriteString(ui.HelpStyle.Render("  1-4 rate · enter default (Good) · esc back"))
	}

	// Progress
	if m.total > 0 {
		b.WriteString("\n\n")
		progress := fmt.Sprintf("  Reviewed: %d  •  Correct: %d/%d", m.total, m.correct, m.total)
		b.WriteString(ui.StatusBarStyle.Render(progress))
	}

	return b.String()
}

// formatDecomposition returns a single mnemonic line for a character, or ""
// if no decomposition is known.
func formatDecomposition(char string) string {
	d, ok := mnemonic.Lookup(char)
	if !ok {
		return ""
	}
	var parts []string
	for _, c := range d.Components {
		parts = append(parts, fmt.Sprintf("%s (%s)", c.Part, c.Meaning))
	}
	return fmt.Sprintf("  部 %s → %s", strings.Join(parts, " + "), d.Story)
}

// InputComposing reports whether the input is in Chinese composition mode.
func (m Model) InputComposing() bool {
	return m.state == stateDrill && m.input.Composing()
}

// SetSize updates terminal dimensions.
func (m Model) SetSize(w, h int) Model {
	m.width = w
	m.height = h
	return m
}

// Refresh reloads due items.
func (m Model) Refresh() Model {
	m.items, m.loadErr = wiki.DueItems(m.store, time.Now())
	m.items = m.episodeWordsFirst(m.items)
	m.cursor = 0
	m.correct = 0
	m.total = 0
	m.state = stateList
	return m
}

// episodeWordsFirst moves the words of the episodes the journey has reached
// to the front of the queue, so they enter review without being marked.
func (m *Model) episodeWordsFirst(due []model.Card) []model.Card {
	due = frequencyNext(due)
	opened, err := episode.OpenedNow(m.store)
	if err != nil || len(opened) == 0 {
		return due
	}
	all, err := m.store.DrillableCards()
	if err != nil {
		return due
	}
	return episode.Ahead(due, opened, all)
}

// frequencyNext moves frequency-list words not yet learned to the front,
// keeping episode words ahead of them.
func frequencyNext(due []model.Card) []model.Card {
	var freq, rest []model.Card
	for _, c := range due {
		if wiki.FrequencyPage(c.ID) && c.ReviewInterval < journey.LearnedInterval() {
			freq = append(freq, c)
		} else {
			rest = append(rest, c)
		}
	}
	return append(freq, rest...)
}

// Leave saves the answers given so far, when the learner leaves before
// the queue is done: their coins, streak, bounties and misses count. The
// next answers start a new session.
func (m Model) Leave() Model {
	if len(m.session.Items) == 0 {
		return m
	}
	m.session.EndedAt = time.Now()
	m.session.CardsTotal, m.session.CardsCorrect = len(m.session.Items), 0
	for _, it := range m.session.Items {
		if it.Correct {
			m.session.CardsCorrect++
		}
	}
	m.noteSaveErr(m.store.WriteSession(m.session))
	m.session = model.Session{StartedAt: time.Now(), Type: "review"}
	return m
}
