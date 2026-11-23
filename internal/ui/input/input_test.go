package input

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestBatchedLettersCompose(t *testing.T) {
	dict := NewDict()
	dict.add("xiang", Candidate{Char: "想", Pinyin: "xiǎng"})
	m := NewWithDict(dict)
	m.SetMode(ModeChinese)
	m.Focus()
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("xiang")})
	if !m.Composing() || len(m.candidates) != 1 {
		t.Fatalf("composing %v with %d candidates; want xiang composed", m.Composing(), len(m.candidates))
	}
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeySpace})
	if m.Value() != "想" {
		t.Errorf("value %q, want 想", m.Value())
	}
}
