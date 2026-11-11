package chat

import (
	"context"
	"fmt"
	"strings"
	"time"
	"unicode"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/cj/mandarin-wiki/internal/audio"
	"github.com/cj/mandarin-wiki/internal/claude"
	"github.com/cj/mandarin-wiki/internal/ui"
	"github.com/cj/mandarin-wiki/internal/ui/input"
)

type chatState int

const (
	stateReady   chatState = iota
	stateWaiting           // waiting for Claude response
)

// Message is a single chat message.
type Message struct {
	Role    string // "user" or "assistant"
	Content string
	Time    time.Time
}

type ttsFinishedMsg struct{}

// Model is the chat view model.
type Model struct {
	claude   *claude.Client
	tts      *audio.Client
	messages []Message
	input    input.Model
	viewport viewport.Model
	state    chatState
	width    int
	height   int
	gkeys    ui.GlobalKeys
}

// New creates a new chat model.
func New(claudeClient *claude.Client, tts *audio.Client, dict *input.Dict) Model {
	ti := input.NewWithDict(dict)
	ti.SetPlaceholder("Type a message in Chinese or English...")
	ti.SetCharLimit(500)
	ti.SetWidth(60)

	vp := viewport.New(80, 20)

	return Model{
		claude:   claudeClient,
		tts:      tts,
		input:    ti,
		viewport: vp,
		gkeys:    ui.DefaultGlobalKeys(),
	}
}

type initChatMsg struct{}

func (m Model) Init() tea.Cmd {
	return func() tea.Msg { return initChatMsg{} }
}

type chatResponseMsg struct {
	content string
	err     error
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case initChatMsg:
		cmd := m.input.Focus()
		return m, tea.Batch(cmd, textinput.Blink)

	case chatResponseMsg:
		m.state = stateReady
		var spokenText string
		if msg.err != nil {
			m.messages = append(m.messages, Message{
				Role:    "assistant",
				Content: fmt.Sprintf("Error: %v", msg.err),
				Time:    time.Now(),
			})
		} else {
			m.messages = append(m.messages, Message{
				Role:    "assistant",
				Content: msg.content,
				Time:    time.Now(),
			})
			spokenText = msg.content
		}
		m.refreshViewport()
		focusCmd := m.input.Focus()

		var cmds []tea.Cmd
		cmds = append(cmds, focusCmd, textinput.Blink)
		if m.tts != nil && m.tts.Available() && spokenText != "" {
			chinese := extractChinese(spokenText)
			if chinese != "" {
				cmds = append(cmds, func() tea.Msg {
					_ = m.tts.Speak(chinese, "zh")
					return ttsFinishedMsg{}
				})
			}
		}
		return m, tea.Batch(cmds...)

	case ttsFinishedMsg:
		return m, nil

	case tea.KeyMsg:
		// Let input handle Ctrl+Space for mode toggle
		if msg.Type == tea.KeyCtrlAt {
			var cmd tea.Cmd
			m.input, cmd = m.input.Update(msg)
			return m, cmd
		}

		switch msg.Type {
		case tea.KeyEnter:
			if m.state == stateWaiting {
				return m, nil
			}
			text := strings.TrimSpace(m.input.Value())
			if text == "" {
				return m, nil
			}
			m.messages = append(m.messages, Message{
				Role:    "user",
				Content: text,
				Time:    time.Now(),
			})
			m.input.Reset()
			m.state = stateWaiting
			m.refreshViewport()

			// Build history for Claude
			history := make([]claude.ChatMessage, len(m.messages)-1)
			for i, msg := range m.messages[:len(m.messages)-1] {
				history[i] = claude.ChatMessage{
					Role:    msg.Role,
					Content: msg.Content,
				}
			}
			userMsg := m.messages[len(m.messages)-1].Content

			return m, func() tea.Msg {
				resp, err := m.claude.Chat(context.Background(), history, userMsg)
				return chatResponseMsg{content: resp, err: err}
			}

		case tea.KeyPgUp:
			m.viewport.HalfViewUp()
			return m, nil
		case tea.KeyPgDown:
			m.viewport.HalfViewDown()
			return m, nil
		}

		if m.state == stateReady {
			var cmd tea.Cmd
			m.input, cmd = m.input.Update(msg)
			return m, cmd
		}
	}

	return m, nil
}

func (m Model) View() string {
	var b strings.Builder

	title := ui.TitleStyle.Render("Chat 对话")
	b.WriteString(title)
	b.WriteString("\n\n")

	// Messages viewport
	contentHeight := m.height - 8
	if contentHeight < 10 {
		contentHeight = 10
	}
	contentWidth := m.width - 4
	if contentWidth < 40 {
		contentWidth = 60
	}
	m.viewport.Width = contentWidth
	m.viewport.Height = contentHeight

	b.WriteString(ui.BorderStyle.Width(contentWidth).Render(m.viewport.View()))
	b.WriteString("\n")

	// Input area
	if m.state == stateWaiting {
		b.WriteString(ui.MutedStyle.Render("  Claude is thinking..."))
	} else {
		m.input.SetWidth(contentWidth - 8)
		b.WriteString("  " + m.input.View())
	}
	b.WriteString("\n")
	b.WriteString(ui.HelpStyle.Render("  enter send · ctrl+space toggle input · pgup/pgdn scroll · esc back"))

	return b.String()
}

func (m *Model) refreshViewport() {
	contentWidth := m.width - 12
	if contentWidth < 40 {
		contentWidth = 56
	}
	bubbleWidth := contentWidth - 8
	if bubbleWidth < 30 {
		bubbleWidth = 30
	}

	var lines []string
	for _, msg := range m.messages {
		timestamp := ui.ChatTimestampStyle.Render(msg.Time.Format("15:04"))

		if msg.Role == "user" {
			// Right-aligned user bubble
			wrapped := ui.WrapText(msg.Content, bubbleWidth)
			bubble := ui.ChatBubbleUserStyle.Render(wrapped)
			// Right-align by padding
			padLen := contentWidth - len(msg.Content)
			if padLen < 2 {
				padLen = 2
			}
			if padLen > contentWidth-10 {
				padLen = 4
			}
			lines = append(lines, fmt.Sprintf("%s  %s", timestamp, ui.SubtitleStyle.Render("You")))
			lines = append(lines, "    "+bubble)
		} else {
			// Left-aligned assistant bubble
			wrapped := ui.WrapText(msg.Content, bubbleWidth)
			bubble := ui.ChatBubbleAssistStyle.Render(wrapped)
			lines = append(lines, fmt.Sprintf("%s  %s", timestamp, ui.ChineseStyle.Render("Tutor 老师")))
			lines = append(lines, "    "+bubble)
		}
		lines = append(lines, "")
	}
	if m.state == stateWaiting {
		lines = append(lines, ui.MutedStyle.Render("  ..."))
	}
	m.viewport.SetContent(strings.Join(lines, "\n"))
	m.viewport.GotoBottom()
}

// InputComposing reports whether the input is in Chinese composition mode.
func (m Model) InputComposing() bool {
	return m.state == stateReady && m.input.Composing()
}

// SetSize updates terminal dimensions.
func (m Model) SetSize(w, h int) Model {
	m.width = w
	m.height = h
	return m
}

// extractChinese pulls out Chinese character runs from mixed text for TTS.
// Returns only the CJK portions joined by pauses, so the Chinese voice
// doesn't attempt to read English words.
func extractChinese(text string) string {
	var segments []string
	var current strings.Builder
	inChinese := false

	for _, r := range text {
		isCJK := unicode.Is(unicode.Han, r) ||
			unicode.Is(unicode.P, r) && inChinese // keep Chinese punctuation
		if isCJK {
			current.WriteRune(r)
			inChinese = true
		} else {
			if inChinese && current.Len() > 0 {
				seg := strings.TrimSpace(current.String())
				if len([]rune(seg)) >= 2 { // skip single-char fragments
					segments = append(segments, seg)
				}
				current.Reset()
			}
			inChinese = false
		}
	}
	if current.Len() > 0 {
		seg := strings.TrimSpace(current.String())
		if len([]rune(seg)) >= 2 {
			segments = append(segments, seg)
		}
	}
	return strings.Join(segments, "。")
}
