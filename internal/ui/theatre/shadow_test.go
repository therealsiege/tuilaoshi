package theatre

import (
	"errors"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/therealsiege/tuilaoshi/internal/ui/uitest"
)

func TestShadowingALine(t *testing.T) {
	m := theatreWith(t, 0)
	defer func(r func() string) { shadowReady = r }(shadowReady)
	defer func(l func(time.Duration) (string, error)) { shadowListen = l }(shadowListen)
	playClip = func(string, time.Duration, time.Duration) error { return nil }

	shadowReady = func() string { return "Shadowing needs whisper.cpp to listen." }
	m, _ = m.Update(enter)
	if m, _ = m.Update(press("s")); m.screen == screenShadow || !strings.Contains(m.note, "whisper") {
		t.Fatalf("shadowing started without whisper: %q", m.note)
	}

	shadowReady = func() string { return "" }
	var asked time.Duration
	shadowListen = func(d time.Duration) (string, error) {
		asked = d
		return "第一集第三句台词", nil // a few characters missed
	}
	m, _ = m.Update(press("s"))
	if m.screen != screenShadow {
		t.Fatalf("screen %d, note %q", m.screen, m.note)
	}
	m.shadow.line = 3 // 第1集第3句台词很长
	m, cmd := m.Update(tea.KeyMsg{Type: tea.KeySpace})
	if m.shadow.phase != shadowRecording || !strings.Contains(m.View(), "Listening") {
		t.Fatalf("not recording:\n%s", m.View())
	}
	m, _ = m.Update(cmd())
	if asked < time.Second || m.shadow.phase != shadowHeard || m.shadow.result.Percent >= 100 || m.shadow.result.Percent == 0 {
		t.Fatalf("recorded %s, phase %d, %d%%", asked, m.shadow.phase, m.shadow.result.Percent)
	}
	uitest.Golden(t, "shadow", m.View())

	shadowListen = func(time.Duration) (string, error) { return "", errors.New("recording: no microphone") }
	m, cmd = m.Update(tea.KeyMsg{Type: tea.KeySpace})
	if m, _ = m.Update(cmd()); m.shadow.phase != shadowReadyToSay || !strings.Contains(m.View(), "no microphone") {
		t.Errorf("a failed recording should say why:\n%s", m.View())
	}
	if m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc}); m.screen != screenEpisode || !strings.Contains(m.note, "best") {
		t.Errorf("esc back: screen %d note %q", m.screen, m.note)
	}
}
