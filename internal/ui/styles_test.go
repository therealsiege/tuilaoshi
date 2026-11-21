package ui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestHeaderIsAClosedBox(t *testing.T) {
	for _, title := range []string{"BATTLE  战斗", "推老师 TuīLǎoShī", "西游记 Journey to the West", "A much longer English title that exceeds thirty cells"} {
		lines := strings.Split(Header(title), "\n")
		if len(lines) != 3 {
			t.Errorf("Header(%q) has %d lines, want 3:\n%s", title, len(lines), Header(title))
			continue
		}
		want := lipgloss.Width(lines[0])
		for i, l := range lines[1:] {
			if w := lipgloss.Width(l); w != want {
				t.Errorf("Header(%q) line %d is %d cells wide, want %d:\n%s", title, i+2, w, want, Header(title))
			}
		}
	}
}
