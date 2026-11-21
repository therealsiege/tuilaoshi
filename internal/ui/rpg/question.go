package rpg

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/therealsiege/tuilaoshi/internal/ui"
	"github.com/therealsiege/tuilaoshi/internal/ui/input"
)

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

// prepareQuestion readies the current question: it clears the answer box for a
// typed question and speaks a listening one.
func (m Model) prepareQuestion() (Model, tea.Cmd) {
	m.order = nil
	if q := m.battle.Question; q != nil && q.Category == "write" {
		m.input.Reset()
		m.input.SetMode(input.ModeChinese)
		return m, m.input.Focus()
	}
	return m, m.speakQuestion()
}
