package app

import (
	"fmt"
	"strings"
	"testing"
)

func TestFitToWindowKeepsTitleKeysAndSelection(t *testing.T) {
	lines := []string{"Title", ""}
	for i := range 40 {
		marker := "  "
		if i == 30 {
			marker = "▸ "
		}
		lines = append(lines, fmt.Sprintf("%sitem %d", marker, i))
	}
	lines = append(lines, "", "↑↓ keys")
	out := fitToWindow(strings.Join(lines, "\n"), 24)
	got := strings.Split(out, "\n")
	if len(got) != 24 || got[0] != "Title" || got[23] != "↑↓ keys" || !strings.Contains(out, "▸ item 30") || !strings.Contains(out, "taller window") {
		t.Errorf("%d lines:\n%s", len(got), out)
	}
	if short := "a\nb\nc"; fitToWindow(short, 24) != short || fitToWindow(strings.Join(lines, "\n"), 0) != strings.Join(lines, "\n") {
		t.Error("a screen that fits, or an unknown height, should be left alone")
	}
}
