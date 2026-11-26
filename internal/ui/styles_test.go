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

func TestWrapDisplay(t *testing.T) {
	got := WrapDisplay("师父别客气，出家人不容易。", 10)
	if len(got) != 3 || got[0] != "师父别客气，" {
		t.Errorf("Chinese wraps between characters by width: %q", got)
	}
	if got := WrapDisplay("路很好走。", 8); len(got) != 1 {
		t.Errorf("a closing 。 stays on its line: %q", got)
	}
	if got := WrapDisplay("第一难「金蝉」、第二难", 12); lipgloss.Width(got[0]) > 14 {
		t.Errorf("only one closing mark may hang past the width: %q", got)
	}
	if got := WrapDisplay("a bit more for you", 8); len(got) != 3 || got[0] != "a bit" {
		t.Errorf("English wraps at spaces: %q", got)
	}
}
