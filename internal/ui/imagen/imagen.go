package imagen

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	imagenClient "github.com/cj/mandarin-wiki/internal/imagen"
	"github.com/cj/mandarin-wiki/internal/ui"
	"github.com/cj/mandarin-wiki/internal/ui/input"
)

type imagenState int

const (
	stateReady      imagenState = iota
	stateGenerating
)

// HistoryItem records a single generation request and its result.
type HistoryItem struct {
	Prompt    string
	ImagePath string
	Text      string // model commentary
	Error     string
	Time      time.Time
}

type imagenResultMsg struct {
	result *imagenClient.Result
	err    error
}

type initMsg struct{}
type openDoneMsg struct{}

// Model is the imagen view model.
type Model struct {
	client   *imagenClient.Client
	history  []HistoryItem
	input    input.Model
	viewport viewport.Model
	state    imagenState
	width    int
	height   int
	gkeys    ui.GlobalKeys
}

// New creates a new imagen view model.
func New(client *imagenClient.Client, dict *input.Dict) Model {
	ti := input.NewWithDict(dict)
	ti.SetPlaceholder("Describe an image to generate...")
	ti.SetCharLimit(500)
	ti.SetWidth(60)

	vp := viewport.New(80, 20)

	return Model{
		client:   client,
		input:    ti,
		viewport: vp,
		gkeys:    ui.DefaultGlobalKeys(),
	}
}

func (m Model) Init() tea.Cmd {
	return func() tea.Msg { return initMsg{} }
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case initMsg:
		cmd := m.input.Focus()
		return m, tea.Batch(cmd, textinput.Blink)

	case imagenResultMsg:
		m.state = stateReady
		if len(m.history) > 0 {
			last := &m.history[len(m.history)-1]
			if msg.err != nil {
				last.Error = msg.err.Error()
			} else if msg.result != nil {
				last.ImagePath = msg.result.ImagePath
				last.Text = msg.result.Text
			}
		}
		m.refreshViewport()
		focusCmd := m.input.Focus()
		return m, tea.Batch(focusCmd, textinput.Blink)

	case openDoneMsg:
		return m, nil

	case tea.KeyMsg:
		// Let input handle Ctrl+Space for mode toggle
		if msg.Type == tea.KeyCtrlAt {
			var cmd tea.Cmd
			m.input, cmd = m.input.Update(msg)
			return m, cmd
		}

		switch {
		case msg.Type == tea.KeyEnter:
			if m.state == stateGenerating {
				return m, nil
			}
			text := strings.TrimSpace(m.input.Value())
			if text == "" {
				return m, nil
			}
			m.history = append(m.history, HistoryItem{
				Prompt: text,
				Time:   time.Now(),
			})
			m.input.Reset()
			m.state = stateGenerating
			m.refreshViewport()

			prompt := text
			return m, func() tea.Msg {
				result, err := m.client.Generate(context.Background(), prompt)
				return imagenResultMsg{result: result, err: err}
			}

		case key.Matches(msg, key.NewBinding(key.WithKeys("o"))):
			if m.state == stateReady {
				// Open the last generated image.
				for i := len(m.history) - 1; i >= 0; i-- {
					if m.history[i].ImagePath != "" {
						path := m.history[i].ImagePath
						return m, func() tea.Msg {
							_ = imagenClient.OpenImage(path)
							return openDoneMsg{}
						}
					}
				}
			}

		case msg.Type == tea.KeyPgUp:
			m.viewport.HalfViewUp()
			return m, nil
		case msg.Type == tea.KeyPgDown:
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

	title := ui.TitleStyle.Render("Image Studio 图片生成")
	b.WriteString(title)
	b.WriteString("\n\n")

	// Viewport for history.
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

	// Input area.
	if m.state == stateGenerating {
		b.WriteString(ui.MutedStyle.Render("  Generating image..."))
	} else if m.client == nil || !m.client.Available() {
		b.WriteString(ui.ErrorStyle.Render("  GOOGLE_API_KEY not set"))
	} else {
		m.input.SetWidth(contentWidth - 8)
		b.WriteString("  " + m.input.View())
	}
	b.WriteString("\n")
	b.WriteString(ui.HelpStyle.Render("  enter generate • ctrl+space toggle input • o open image • pgup/pgdn scroll • esc back"))

	return b.String()
}

func (m *Model) refreshViewport() {
	contentWidth := m.width - 12
	if contentWidth < 40 {
		contentWidth = 56
	}

	var lines []string
	for _, item := range m.history {
		timestamp := ui.ChatTimestampStyle.Render(item.Time.Format("15:04"))

		// Prompt line.
		promptText := ui.ChineseStyle.Render(fmt.Sprintf("> \"%s\"", item.Prompt))
		lines = append(lines, fmt.Sprintf("  %s  %s", timestamp, promptText))

		// Result line.
		if item.Error != "" {
			lines = append(lines, "    "+ui.ErrorStyle.Render("Error: "+item.Error))
		} else if item.ImagePath != "" {
			shortPath := item.ImagePath
			if base := filepath.Base(item.ImagePath); len(base) > 0 {
				shortPath = base
			}
			lines = append(lines, "    "+ui.SuccessStyle.Render("Saved: "+shortPath))
			if item.Text != "" {
				wrapped := ui.WrapText(item.Text, contentWidth-8)
				lines = append(lines, "    "+ui.MutedStyle.Render(wrapped))
			}
		} else {
			lines = append(lines, "    "+ui.WarningStyle.Render("Generating..."))
		}
		lines = append(lines, "")
	}

	if len(m.history) == 0 {
		lines = append(lines, ui.MutedStyle.Render("  Type a prompt and press enter to generate an image."))
		lines = append(lines, "")
		lines = append(lines, ui.MutedStyle.Render("  Examples:"))
		lines = append(lines, ui.MutedStyle.Render("    \"A Chinese ink wash painting of mountains and bamboo\""))
		lines = append(lines, ui.MutedStyle.Render("    \"Watercolor of a traditional tea ceremony\""))
		lines = append(lines, ui.MutedStyle.Render("    \"Neon Chinese characters on a cyberpunk street\""))
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

// Refresh is called when switching to this view.
func (m Model) Refresh() Model {
	return m
}
