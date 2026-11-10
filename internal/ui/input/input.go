package input

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Style constants for mode badges.
var (
	modeNormalStyle = lipgloss.NewStyle().
			Padding(0, 1).
			Bold(true).
			Foreground(lipgloss.Color("#ABB2BF")).
			Background(lipgloss.Color("#3E4451"))

	modePinyinStyle = lipgloss.NewStyle().
			Padding(0, 1).
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#C678DD"))

	modeChineseStyle = lipgloss.NewStyle().
				Padding(0, 1).
				Bold(true).
				Foreground(lipgloss.Color("#FFFFFF")).
				Background(lipgloss.Color("#E06C75"))
)

// Model wraps textinput.Model with pinyin and Chinese input modes.
// It is a drop-in replacement — same public API.
type Model struct {
	inner textinput.Model
	dict  *Dict
	mode  Mode

	// Chinese mode composition state
	composing  bool
	compBuf    string       // raw pinyin being composed
	candidates []Candidate  // current candidate list
	candPage   int          // pagination offset
}

const candidatesPerPage = 9

// New creates a new input model (ModeNormal, no dict).
func New() Model {
	return Model{inner: textinput.New()}
}

// NewWithDict creates a new input model backed by a character dictionary.
func NewWithDict(dict *Dict) Model {
	return Model{inner: textinput.New(), dict: dict}
}

// --- Accessors that mirror textinput.Model ---

func (m Model) Value() string          { return m.inner.Value() }
func (m *Model) SetValue(s string)     { m.inner.SetValue(s) }
func (m *Model) Reset()                { m.cancelComposition(); m.inner.Reset() }
func (m Model) Focused() bool          { return m.inner.Focused() }
func (m *Model) Focus() tea.Cmd        { return m.inner.Focus() }
func (m *Model) Blur()                 { m.inner.Blur() }
func (m Model) Cursor() int            { return m.inner.Position() }
func (m Model) Mode() Mode             { return m.mode }
func (m Model) Composing() bool        { return m.composing }

// Placeholder pass-through.
func (m *Model) SetPlaceholder(s string) { m.inner.Placeholder = s }
func (m *Model) SetCharLimit(n int)      { m.inner.CharLimit = n }
func (m *Model) SetWidth(n int)          { m.inner.Width = n }

// Expose inner fields for direct tweaks in existing code.
func (m *Model) Inner() *textinput.Model { return &m.inner }

// Update handles key events, applying mode-specific behaviour.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	keyMsg, isKey := msg.(tea.KeyMsg)
	if !isKey {
		var cmd tea.Cmd
		m.inner, cmd = m.inner.Update(msg)
		return m, cmd
	}

	// Ctrl+Space toggles mode
	if keyMsg.Type == tea.KeyCtrlAt {
		m.cancelComposition()
		m.mode = m.mode.Next()
		// Skip Chinese mode if no dict available
		if m.mode == ModeChinese && m.dict == nil {
			m.mode = m.mode.Next()
		}
		return m, nil
	}

	switch m.mode {
	case ModePinyin:
		return m.updatePinyin(keyMsg)
	case ModeChinese:
		return m.updateChinese(keyMsg)
	default:
		return m.updateNormal(keyMsg)
	}
}

// View renders the mode badge + input + optional candidate bar.
func (m Model) View() string {
	badge := m.modeBadge()

	var b strings.Builder
	b.WriteString(badge)
	b.WriteString(" ")

	if m.composing && m.mode == ModeChinese {
		// Show the composition buffer inline after existing text
		// Render inner with composition overlay
		b.WriteString(m.inner.View())
		if m.compBuf != "" {
			compStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("#C678DD")).
				Underline(true)
			b.WriteString(compStyle.Render(m.compBuf))
		}
	} else {
		b.WriteString(m.inner.View())
	}

	// Candidate bar for Chinese mode
	if m.composing && len(m.candidates) > 0 {
		b.WriteString("\n")
		b.WriteString(m.candidateBar())
	}

	return b.String()
}

// --- Mode-specific update methods ---

func (m Model) updateNormal(msg tea.KeyMsg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	m.inner, cmd = m.inner.Update(msg)
	return m, cmd
}

func (m Model) updatePinyin(msg tea.KeyMsg) (Model, tea.Cmd) {
	// On space or enter, try to convert the last word
	if msg.Type == tea.KeySpace || msg.Type == tea.KeyEnter {
		val := m.inner.Value()
		converted := m.tryConvertLastWord(val)
		if converted != val {
			m.inner.SetValue(converted)
			m.inner.SetCursor(len([]rune(converted)))
		}
		if msg.Type == tea.KeySpace {
			// Add the space
			m.inner.SetValue(m.inner.Value() + " ")
			m.inner.SetCursor(len([]rune(m.inner.Value())))
			return m, nil
		}
		// Enter — don't consume, let parent handle it
		var cmd tea.Cmd
		m.inner, cmd = m.inner.Update(msg)
		return m, cmd
	}

	var cmd tea.Cmd
	m.inner, cmd = m.inner.Update(msg)
	return m, cmd
}

func (m Model) updateChinese(msg tea.KeyMsg) (Model, tea.Cmd) {
	if m.composing {
		return m.updateComposition(msg)
	}

	// Start composition on letter key
	if msg.Type == tea.KeyRunes && len(msg.Runes) == 1 && unicode.IsLetter(msg.Runes[0]) {
		r := unicode.ToLower(msg.Runes[0])
		if r >= 'a' && r <= 'z' {
			m.composing = true
			m.compBuf = string(r)
			m.updateCandidates()
			return m, nil
		}
	}

	// Non-letter keys pass through normally
	var cmd tea.Cmd
	m.inner, cmd = m.inner.Update(msg)
	return m, cmd
}

func (m Model) updateComposition(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEscape:
		m.cancelComposition()
		return m, nil

	case tea.KeyBackspace:
		if len(m.compBuf) > 0 {
			m.compBuf = m.compBuf[:len(m.compBuf)-1]
			if len(m.compBuf) == 0 {
				m.cancelComposition()
				return m, nil
			}
			m.updateCandidates()
		}
		return m, nil

	case tea.KeySpace:
		// Select first candidate
		if len(m.candidates) > 0 {
			m.selectCandidate(0)
		} else {
			m.cancelComposition()
		}
		return m, nil

	case tea.KeyEnter:
		// Select first candidate and let enter propagate
		if len(m.candidates) > 0 {
			m.selectCandidate(0)
		} else {
			m.cancelComposition()
		}
		return m, nil

	case tea.KeyRight:
		// Next page of candidates
		maxPage := len(m.candidates) / candidatesPerPage
		if m.candPage < maxPage {
			m.candPage++
		}
		return m, nil

	case tea.KeyLeft:
		if m.candPage > 0 {
			m.candPage--
		}
		return m, nil

	case tea.KeyRunes:
		if len(msg.Runes) == 1 {
			r := msg.Runes[0]
			// Number selection 1-9
			if r >= '1' && r <= '9' {
				idx := m.candPage*candidatesPerPage + int(r-'1')
				if idx < len(m.candidates) {
					m.selectCandidate(idx)
					return m, nil
				}
			}
			// More letters → extend composition
			if unicode.IsLetter(r) {
				lr := unicode.ToLower(r)
				if lr >= 'a' && lr <= 'z' {
					m.compBuf += string(lr)
					m.updateCandidates()
					return m, nil
				}
			}
		}
	}

	return m, nil
}

// --- Internal helpers ---

func (m *Model) updateCandidates() {
	m.candPage = 0
	if m.dict == nil {
		m.candidates = nil
		return
	}
	m.candidates = m.dict.Lookup(m.compBuf)
}

func (m *Model) selectCandidate(idx int) {
	if idx >= len(m.candidates) {
		return
	}
	char := m.candidates[idx].Char
	m.inner.SetValue(m.inner.Value() + char)
	m.inner.SetCursor(len([]rune(m.inner.Value())))
	m.cancelComposition()
}

func (m *Model) cancelComposition() {
	m.composing = false
	m.compBuf = ""
	m.candidates = nil
	m.candPage = 0
}

func (m Model) modeBadge() string {
	label := m.mode.Label()
	switch m.mode {
	case ModePinyin:
		return modePinyinStyle.Render(label)
	case ModeChinese:
		return modeChineseStyle.Render(label)
	default:
		return modeNormalStyle.Render(label)
	}
}

func (m Model) candidateBar() string {
	start := m.candPage * candidatesPerPage
	end := start + candidatesPerPage
	if end > len(m.candidates) {
		end = len(m.candidates)
	}

	var parts []string
	for i := start; i < end; i++ {
		c := m.candidates[i]
		num := i - start + 1
		parts = append(parts, fmt.Sprintf("%d.%s", num, c.Char))
	}

	bar := " " + strings.Join(parts, "  ")
	if end < len(m.candidates) {
		bar += "  [→]"
	}
	if m.candPage > 0 {
		bar = " [←]" + bar
	}

	barStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#ABB2BF")).
		Background(lipgloss.Color("#3E4451")).
		Padding(0, 1)
	return barStyle.Render(bar)
}

// tryConvertLastWord finds the last whitespace-delimited token in val and
// converts it if it's a valid numbered pinyin syllable.
func (m Model) tryConvertLastWord(val string) string {
	if val == "" {
		return val
	}

	// Find last word boundary
	lastSpace := strings.LastIndex(val, " ")
	var prefix, lastWord string
	if lastSpace >= 0 {
		prefix = val[:lastSpace+1]
		lastWord = val[lastSpace+1:]
	} else {
		prefix = ""
		lastWord = val
	}

	if lastWord == "" {
		return val
	}

	if IsValidSyllable(lastWord) {
		return prefix + ConvertSyllable(lastWord)
	}
	return val
}
