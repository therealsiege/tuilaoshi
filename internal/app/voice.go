package app

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/therealsiege/tuilaoshi/internal/audio"
	"github.com/therealsiege/tuilaoshi/internal/ui"
)

func (m *Model) updateVoicePicker(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	zhVoices := audio.ChineseVoices()
	enVoices := audio.EnglishVoices()

	var maxItems int
	if m.voiceCategory == 0 {
		maxItems = len(zhVoices)
	} else {
		maxItems = len(enVoices)
	}

	switch {
	case key.Matches(msg, m.keys.Up):
		if m.voiceCursor > 0 {
			m.voiceCursor--
		}
	case key.Matches(msg, m.keys.Down):
		if m.voiceCursor < maxItems-1 {
			m.voiceCursor++
		}
	case key.Matches(msg, m.keys.Tab):
		if m.voiceCategory == 0 {
			m.voiceCategory = 1
		} else {
			m.voiceCategory = 0
		}
		m.voiceCursor = 0
	case key.Matches(msg, m.keys.Enter):
		if m.voiceCategory == 0 && m.voiceCursor < len(zhVoices) {
			m.tts.SetChineseVoice(zhVoices[m.voiceCursor])
		} else if m.voiceCategory == 1 && m.voiceCursor < len(enVoices) {
			m.tts.SetEnglishVoice(enVoices[m.voiceCursor])
		}
		m.currentView = ui.ViewDashboard
		m.dashboard = m.dashboard.Refresh()
		return m, m.dashboard.Init()
	case key.Matches(msg, m.keys.Back):
		m.currentView = ui.ViewDashboard
		m.dashboard = m.dashboard.Refresh()
		return m, m.dashboard.Init()
	}
	return m, nil
}

func (m Model) voicePickerView() string {
	var b strings.Builder

	title := ui.TitleStyle.Render("Voice Picker 语音选择")
	b.WriteString(title)
	b.WriteString("\n\n")

	zhVoices := audio.ChineseVoices()
	enVoices := audio.EnglishVoices()

	// Category tabs
	if m.voiceCategory == 0 {
		b.WriteString(ui.ActiveTabStyle.Render("Chinese 中文"))
	} else {
		b.WriteString(ui.TabStyle.Render("Chinese 中文"))
	}
	b.WriteString("  ")
	if m.voiceCategory == 1 {
		b.WriteString(ui.ActiveTabStyle.Render("English"))
	} else {
		b.WriteString(ui.TabStyle.Render("English"))
	}
	b.WriteString("\n\n")

	if m.voiceCategory == 0 {
		current := m.tts.ChineseVoice()
		for i, v := range zhVoices {
			label := fmt.Sprintf("  %s", v.Name)
			if v.ID == current.ID {
				label += " " + ui.SuccessStyle.Render("(current)")
			}
			if i == m.voiceCursor {
				b.WriteString(ui.SelectedStyle.Render("▸ " + label))
			} else {
				b.WriteString(ui.UnselectedStyle.Render("  " + label))
			}
			b.WriteString("\n")
		}
	} else {
		current := m.tts.EnglishVoice()
		for i, v := range enVoices {
			label := fmt.Sprintf("  %s", v.Name)
			if v.ID == current.ID {
				label += " " + ui.SuccessStyle.Render("(current)")
			}
			if i == m.voiceCursor {
				b.WriteString(ui.SelectedStyle.Render("▸ " + label))
			} else {
				b.WriteString(ui.UnselectedStyle.Render("  " + label))
			}
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(ui.HelpStyle.Render("  ↑↓ select · tab switch category · enter confirm · esc back"))

	return b.String()
}
