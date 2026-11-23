package theatre

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/therealsiege/tuilaoshi/internal/episode"
	"github.com/therealsiege/tuilaoshi/internal/ui/uitest"
	"github.com/therealsiege/tuilaoshi/internal/wiki"
)

var day = time.Date(2026, 11, 22, 9, 0, 0, 0, time.Local)

// theatreWith is a wiki with two made-up episodes (with an audio file) and
// stages 1..cleared won.
func theatreWith(t *testing.T, cleared int) Model {
	t.Helper()
	store := uitest.Wiki(t, nil)
	for n := 1; n <= 2; n++ {
		dir := filepath.Join(store.Dir, wiki.PrivateEpisodesDir, fmt.Sprintf("Show %02d Title%d", n, n))
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
		var zh, en strings.Builder
		for i := 0; i < 30; i++ {
			fmt.Fprintf(&zh, "%d\n00:00:%02d,000 --> 00:00:%02d,800\n第%d集第%d句台词很长\n\n", i+1, i, i, n, i)
			fmt.Fprintf(&en, "%d\n00:00:%02d,000 --> 00:00:%02d,800\nEpisode %d, line %d.\n\n", i+1, i, i, n, i)
		}
		for name, body := range map[string]string{"zh.srt": zh.String(), "en.srt": en.String(), "audio.mp3": "fake"} {
			if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0644); err != nil {
				t.Fatal(err)
			}
		}
	}
	if _, err := store.NewSave("test", day); err != nil {
		t.Fatal(err)
	}
	for n := 1; n <= cleared; n++ {
		if err := store.RecordStageCleared(n, day); err != nil {
			t.Fatal(err)
		}
	}
	m := New(store, nil)
	m.now = func() time.Time { return day }
	return m.SetSize(100, 40).Refresh()
}

var enter = tea.KeyMsg{Type: tea.KeyEnter}

func press(s string) tea.KeyMsg { return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)} }

func TestEpisodesOpenAlongTheRoad(t *testing.T) {
	m := theatreWith(t, 0) // working on tribulation 1: episode 1 only
	uitest.Golden(t, "list", m.View())
	m.cursor = 1
	m, _ = m.Update(enter)
	if m.screen != screenList || !strings.Contains(m.note, "opens at tribulation") {
		t.Fatalf("episode 2 opened early: screen %d, note %q", m.screen, m.note)
	}
}

func TestScenesPayOnceLockOnAMissAndFinishPays(t *testing.T) {
	m := theatreWith(t, 0)
	m, _ = m.Update(enter)
	if len(m.scenes) != 2 {
		t.Fatalf("%d scenes from 30 lines, want 2", len(m.scenes))
	}
	m, _ = m.Update(enter)
	if m.screen != screenScene || strings.Contains(m.View(), "Episode 1, line") && !strings.Contains(m.View(), "1  Episode") {
		t.Fatalf("scene question:\n%s", m.View())
	}
	wrong := (m.scenes[0].Right + 1) % len(m.scenes[0].Choices)
	m, _ = m.Update(press(fmt.Sprint(wrong + 1)))
	m, _ = m.Update(enter)
	m, _ = m.Update(enter)
	if m.screen != screenEpisode || !strings.Contains(m.note, "tomorrow") {
		t.Fatalf("a missed scene reopened: screen %d, note %q", m.screen, m.note)
	}
	m.scene = 1
	m, _ = m.Update(enter)
	m, _ = m.Update(press(fmt.Sprint(m.scenes[1].Right + 1)))
	if !strings.Contains(m.View(), "Episode 1, line") {
		t.Errorf("the English isn't shown after answering:\n%s", m.View())
	}
	w, err := m.store.Wallet()
	if err != nil || w.Tongqian != episode.ScenePay {
		t.Fatalf("one scene right: purse %+v %v", w, err)
	}
	// Tomorrow the missed scene opens again; answering it finishes the episode.
	m.now = func() time.Time { return day.AddDate(0, 0, 1) }
	m, _ = m.Update(enter)
	m = m.Refresh()
	m.scene = 0
	m, _ = m.Update(enter)
	m, _ = m.Update(press(fmt.Sprint(m.scenes[0].Right + 1)))
	w, err = m.store.Wallet()
	if err != nil || w.Tongqian != 2*episode.ScenePay+episode.FinishCopper || w.Yuanbao != episode.FinishIngots {
		t.Errorf("finishing the episode: purse %+v %v", w, err)
	}
}

func TestListeningDrillPaysThroughASession(t *testing.T) {
	m := theatreWith(t, 0)
	var played []time.Duration
	playClip = func(_ string, start, _ time.Duration) error { played = append(played, start); return nil }
	m, _ = m.Update(enter)
	m, cmd := m.Update(press("l"))
	if m.screen != screenListen || cmd == nil {
		t.Fatalf("listening didn't start: screen %d", m.screen)
	}
	runAll(cmd)
	if len(played) == 0 {
		t.Fatal("no clip played")
	}
	line := m.eps[0].Lines[m.listen.drill.Line]
	if m.listen.drill.Type {
		m.listen.input.SetValue(line.Chinese)
		m, _ = m.Update(enter)
	} else {
		m, _ = m.Update(press(fmt.Sprint(m.listen.drill.Right + 1)))
	}
	if !m.listen.right {
		t.Fatalf("the right answer wasn't taken:\n%s", m.View())
	}
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if w, err := m.store.Wallet(); err != nil || w.Tongqian != 1 {
		t.Errorf("a line caught paid %+v %v, want 1 copper", w, err)
	}
}

// runAll runs a command and any batch inside it.
func runAll(cmd tea.Cmd) {
	if cmd == nil {
		return
	}
	if batch, ok := cmd().(tea.BatchMsg); ok {
		for _, c := range batch {
			if c != nil {
				done := make(chan struct{})
				go func() { c(); close(done) }()
				select {
				case <-done:
				case <-time.After(100 * time.Millisecond):
				}
			}
		}
	}
}
