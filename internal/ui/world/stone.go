package world

import (
	"fmt"
	"slices"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/therealsiege/tuilaoshi/internal/ui"
	"github.com/therealsiege/tuilaoshi/internal/wiki"
)

// A town's stone is a secret: its inscription gives a 成语's meaning, a clue
// and the first character, and typing the whole idiom from memory opens it.
// Opened unaided it pays a silver ingot; after asking for the pinyin, copper.

// stoneHintPay is the copper a stone pays once the pinyin hint was used.
const stoneHintPay = 30

// openStone shows the town's stone, unless already opened.
func (m Model) openStone() (Model, tea.Cmd) {
	t := m.towns[m.town]
	if t.Stone == nil {
		return m, nil
	}
	if slices.Contains(m.stones, t.ID) {
		m.note = "The stone already stands open: " + t.Stone.Idiom + " " + t.Stone.Pinyin
		return m, nil
	}
	m.screen, m.hinted, m.answered, m.note = screenStone, false, -1, ""
	m.input.Reset()
	return m, m.input.Focus()
}

// updateStone types the idiom; h shows the pinyin, enter tries the answer.
func (m Model) updateStone(k tea.KeyMsg) (Model, tea.Cmd) {
	t := m.towns[m.town]
	switch {
	case m.answered > 0:
		if k.Type == tea.KeyEnter || k.Type == tea.KeySpace {
			m.screen = screenTown
			return m.Refresh(), nil
		}
		return m, nil
	case k.Type == tea.KeyCtrlH:
		m.hinted = true
		return m, nil
	case k.Type == tea.KeyEnter && !m.input.Composing():
		if !t.Stone.Opens(m.input.Value()) {
			m.note = "The stone stays shut. (ctrl+h shows the pinyin.)"
			return m, nil
		}
		pay := wiki.Wallet{Yuanbao: 1}
		if m.hinted {
			pay = wiki.Wallet{Tongqian: stoneHintPay}
		}
		if err := m.store.OpenStone(t.ID, pay); err != nil {
			m.loadErr = err
		}
		m.answered, m.note = 1, ""
		m.input.Blur()
		return m, nil
	}
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(k)
	return m, cmd
}

func (m Model) stoneView() string {
	t := m.towns[m.town]
	s := t.Stone
	var b strings.Builder
	b.WriteString(ui.TitleStyle.MarginBottom(0).Render("石碑 A stone by the road in "+t.Name) + "\n\n")
	b.WriteString("  " + ui.MutedStyle.Render("Moss covers most of the inscription. You can make out:") + "\n\n")
	b.WriteString("      " + ui.ChineseStyle.Render(s.Inscription()) + "\n\n")
	b.WriteString("  " + s.Meaning + "\n")
	b.WriteString("  " + ui.MutedStyle.Render(s.Clue) + "\n")
	if m.hinted || m.answered > 0 {
		b.WriteString("  " + ui.PinyinStyle.Render(s.Pinyin) + "\n")
	}
	b.WriteString("\n  " + m.input.View() + "\n\n")
	if m.answered > 0 {
		reward := "a silver ingot (1 元宝)"
		if m.hinted {
			reward = fmt.Sprintf("%d 铜钱", stoneHintPay)
		}
		b.WriteString(ui.SuccessStyle.Render(fmt.Sprintf("  %s! The stone slides open; behind it lies %s.", s.Idiom, reward)) + "\n")
		b.WriteString("\n" + ui.HelpStyle.Render("  enter back to town"))
		return b.String()
	}
	help := "  type the whole idiom · enter try · ctrl+h pinyin (pays copper instead of an ingot) · esc leave"
	b.WriteString(ui.HelpStyle.Render(help))
	return b.String()
}
