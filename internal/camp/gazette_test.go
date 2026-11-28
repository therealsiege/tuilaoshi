package camp

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/therealsiege/tuilaoshi/internal/wiki"
)

func TestPrintGazetteByClaudeOrTheClerk(t *testing.T) {
	now := time.Date(2026, 11, 24, 20, 0, 0, 0, time.Local)
	fail := func(context.Context, string) (string, string, error) { return "", "", errors.New("no voice") }
	store := wiki.NewStore(t.TempDir())
	g, missed, err := PrintGazette(t.Context(), store, fail, now)
	if err != nil || missed == nil || !g.ByHand || g.Text == "" {
		t.Fatalf("when Claude fails the clerk sets it: %+v %v %v", g, missed, err)
	}

	store = wiki.NewStore(t.TempDir())
	calls := 0
	write := func(_ context.Context, facts string) (string, string, error) {
		calls++
		if facts == "" {
			t.Error("no facts")
		}
		return "本报讯。", "Our report.", nil
	}
	g, missed, err = PrintGazette(t.Context(), store, write, now)
	if err != nil || missed != nil || g.ByHand || g.Text != "本报讯。" {
		t.Fatalf("Claude's issue: %+v %v %v", g, missed, err)
	}
	if g, _, _ = PrintGazette(t.Context(), store, write, now); calls != 1 || g.Text != "本报讯。" {
		t.Errorf("the week's issue should be kept, not written again: %d calls", calls)
	}
}
