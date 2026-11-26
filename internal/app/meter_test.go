package app

import (
	"testing"

	"github.com/therealsiege/tuilaoshi/internal/ui"
	"github.com/therealsiege/tuilaoshi/internal/wiki"
)

func TestTheMeterHoldsTheDailyCap(t *testing.T) {
	m := meter{wiki.NewStore(t.TempDir())}
	saved := ui.Prefs
	defer func() { ui.Prefs = saved }()
	ui.Prefs.ClaudeCap = 2
	for range 2 {
		if !m.Allow() {
			t.Fatal("refused under the cap")
		}
		m.Used("claude-sonnet-5-5", 10, 10)
	}
	if m.Allow() {
		t.Error("allowed past the cap")
	}
	if ui.Prefs.ClaudeCap = 0; !m.Allow() {
		t.Error("no limit refused a call")
	}
}
