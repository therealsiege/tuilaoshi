package dashboard

import (
	"strings"
	"testing"
	"time"

	"github.com/therealsiege/tuilaoshi/internal/ui/uitest"
)

func dashboardFor(t *testing.T, width int) Model {
	t.Helper()
	store := uitest.Wiki(t, map[string]string{"characters/shu1.md": uitest.OverdueCharacter})
	m := New(store, nil)
	m.now = func() time.Time { return time.Date(2026, 11, 22, 9, 0, 0, 0, time.Local) } // fixes the day's bounties
	return m.SetSize(width, 40).Refresh()
}

func TestDashboardScreens(t *testing.T) {
	uitest.Golden(t, "wide", dashboardFor(t, 110).View())
	uitest.Golden(t, "narrow", dashboardFor(t, 72).View())
}

func TestStartHereSuggestsOneNextStep(t *testing.T) {
	m := dashboardFor(t, 110)
	if v := uitest.Plain(m.renderStartHere()); !strings.Contains(v, "r  Review your 1 due word") {
		t.Errorf("with a word due: %q", v)
	}
	m.due = 10789
	if v := uitest.Plain(m.renderStartHere()); !strings.Contains(v, "r  Review 20 of your 10,789 due words") {
		t.Errorf("with a backlog: %q", v)
	}
	m.due = 0
	if v := uitest.Plain(m.renderStartHere()); !strings.Contains(v, "esc  Tribulation 1 is open") {
		t.Errorf("nothing due, stage 1 open: %q", v)
	}
	m.journeyState.Stage = nil
	if v := uitest.Plain(m.renderStartHere()); !strings.Contains(v, "t  Read & listen") {
		t.Errorf("nothing due, journey done: %q", v)
	}
}

func TestThousands(t *testing.T) {
	for n, want := range map[int]string{0: "0", 999: "999", 1000: "1,000", 10789: "10,789", 1234567: "1,234,567"} {
		if got := thousands(n); got != want {
			t.Errorf("thousands(%d) = %q, want %q", n, got, want)
		}
	}
}
