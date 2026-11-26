package assessment

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/therealsiege/tuilaoshi/internal/assess"
	"github.com/therealsiege/tuilaoshi/internal/ui/uitest"
)

type fakeSpeaker struct{}

func (fakeSpeaker) Available() bool         { return true }
func (fakeSpeaker) Speak(_, _ string) error { return nil }

type fakeGrader struct {
	score int
	err   error
}

func (g fakeGrader) GradeWriting(_ context.Context, _, _, _ string) (int, string, error) {
	return g.score, "Good: 很自然。", g.err
}

// clockAt is a settable clock for the mock test's timer.
type clockAt struct{ t time.Time }

func (c *clockAt) now() time.Time { return c.t }

const overview = "# Overview\n\n## Current Level Assessment\n\n| Skill | Estimated Level | Confidence | Last Assessed |\n|---|---|---|---|\n" +
	"| Vocabulary (passive) | HSK4 | 2/3 right at HSK4 | 2026-11-01 |\n| Grammar | HSK3 | 2/3 right at HSK3 | 2026-11-01 |\n" +
	"| Reading comprehension | HSK4 | 2/3 right at HSK4 | 2026-11-01 |\n\n## Roadmap\n"

func mockModel(t *testing.T, w, h int, voice bool) (Model, *clockAt) {
	t.Helper()
	pages := uitest.MockTestPages()
	pages["overview.md"] = overview
	m := New(uitest.Wiki(t, pages)).SetSize(w, h).Refresh()
	clock := &clockAt{t: time.Date(2026, 11, 26, 19, 0, 0, 0, time.UTC)}
	m.now, m.rng, m.grader = clock.now, rand.New(rand.NewSource(7)), fakeGrader{score: 2}
	if voice {
		m.tts = fakeSpeaker{}
	}
	return m, clock
}

func press(m Model, keys ...string) Model {
	for _, k := range keys {
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(k)}
		switch k {
		case "enter":
			msg = tea.KeyMsg{Type: tea.KeyEnter}
		case "backspace":
			msg = tea.KeyMsg{Type: tea.KeyBackspace}
		}
		m, _ = m.Update(msg)
	}
	return m
}

// answerRight answers the current item correctly, returning the command
// it gave (the grading, for a written sentence).
func answerRight(t *testing.T, m Model) (Model, tea.Cmd) {
	t.Helper()
	it := m.current()
	switch it.Kind {
	case assess.Tiles:
		used := map[int]bool{}
		rest := it.Answer
		for rest != "" {
			found := false
			for i, tile := range it.Tiles {
				if !used[i] && strings.HasPrefix(rest, tile) {
					used[i], rest, found = true, strings.TrimPrefix(rest, tile), true
					m = press(m, string(rune('1'+i)))
					break
				}
			}
			if !found {
				t.Fatalf("tiles %v do not make %s", it.Tiles, it.Answer)
			}
		}
		return m, nil
	case assess.Compose:
		m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("我推荐这本书。")})
		return m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	}
	return press(m, string(rune('1'+it.Correct))), nil
}

func TestMockTestAllRightPassesAndIsSaved(t *testing.T) {
	m, _ := mockModel(t, 80, 24, true)
	m = press(m, "m", "4")
	if m.phase != phaseMock || m.mock.section != assess.Listening {
		t.Fatalf("phase %v section %d", m.phase, m.mock.section)
	}
	for i := 0; m.phase == phaseMock && i < 100; i++ {
		var cmd tea.Cmd
		m, cmd = answerRight(t, m)
		if m.mock.grading {
			m, _ = m.Update(cmd())
		}
	}
	if m.phase != phaseMockDone {
		t.Fatalf("still in phase %v", m.phase)
	}
	for s, sc := range m.mock.scores {
		if sc.Skipped || sc.Score() != 100 {
			t.Errorf("section %d scored %+v", s, sc)
		}
	}
	if w := m.mock.scores[assess.Writing]; w.Max != assess.TileItems+composePoints {
		t.Errorf("writing out of %d, want the tiles and the graded sentence", w.Max)
	}
	if !m.mock.passed || m.mock.total != 300 || m.mock.estimate != 4 {
		t.Errorf("total %d passed %v estimate %d", m.mock.total, m.mock.passed, m.mock.estimate)
	}
	view := m.View()
	for _, want := range []string{"Total 300 / 300", "合格 pass", "bears it out", "2026-11-26  HSK4"} {
		if !strings.Contains(view, want) {
			t.Errorf("results missing %q:\n%s", want, view)
		}
	}
	saved, err := m.store.MockSittings()
	if err != nil || len(saved) != 1 || saved[0].Level != 4 || saved[0].Sections[1].Score != 100 || !saved[0].Passed {
		t.Errorf("saved %+v %v", saved, err)
	}
}

func TestMockTimerEndsTheSection(t *testing.T) {
	m, clock := mockModel(t, 80, 24, true)
	m = press(m, "m", "5")
	m, _ = answerRight(t, m)
	gen := m.mock.timer
	clock.t = clock.t.Add(assess.SectionTimes[assess.Listening] - time.Second)
	if m, _ = m.Update(tickMsg{gen: gen}); m.mock.section != assess.Listening || !strings.Contains(m.View(), "0:01 left") {
		t.Fatalf("ended a second early:\n%s", m.View())
	}
	clock.t = clock.t.Add(time.Second)
	if m, _ = m.Update(tickMsg{gen: gen - 1}); m.mock.section != assess.Listening {
		t.Fatal("an old timer's tick ended the section")
	}
	m, _ = m.Update(tickMsg{gen: gen})
	if m.mock.section != assess.Reading || m.mock.item != 0 {
		t.Fatalf("section %d item %d after time ran out", m.mock.section, m.mock.item)
	}
	if sc := m.mock.scores[assess.Listening]; sc.Points != 1 || sc.Score() != 10 {
		t.Errorf("listening %+v: the unanswered nine should score nothing", sc)
	}
	if !strings.Contains(m.View(), "7:00 left") {
		t.Errorf("reading should start with its own time:\n%s", m.View())
	}
}

func TestMockWithoutAudioSkipsListening(t *testing.T) {
	m, _ := mockModel(t, 80, 24, false)
	m.grader = nil
	m = press(m, "m")
	if !strings.Contains(m.View(), "listening section will be skipped") {
		t.Errorf("no warning:\n%s", m.View())
	}
	m = press(m, "4")
	if m.mock.section != assess.Reading || !m.mock.scores[assess.Listening].Skipped {
		t.Fatalf("section %d", m.mock.section)
	}
	for i := 0; m.phase == phaseMock && i < 100; i++ {
		if m.current().Kind == assess.Tiles {
			m, _ = answerRight(t, m)
		} else {
			m = press(m, "1")
		}
	}
	if m.phase != phaseMockDone || m.mock.most != 200 || m.mock.scores[assess.Writing].Max != assess.TileItems {
		t.Errorf("phase %v out of %d, writing %+v", m.phase, m.mock.most, m.mock.scores[assess.Writing])
	}
	if !strings.Contains(m.View(), "skipped: no audio") {
		t.Errorf("results don't say listening was skipped:\n%s", m.View())
	}
}

func TestMockGradingErrorIsNotHeldAgainstYou(t *testing.T) {
	m, _ := mockModel(t, 80, 24, false)
	m.grader = fakeGrader{err: errors.New("offline")}
	m = press(m, "m", "4")
	for m.mock.section != assess.Writing || m.current().Kind != assess.Compose {
		m, _ = answerRight(t, m)
	}
	m, cmd := answerRight(t, m)
	m, _ = m.Update(cmd())
	if m.phase != phaseMockDone || m.mock.scores[assess.Writing].Max != assess.TileItems || m.mock.scores[assess.Writing].Score() != 100 {
		t.Errorf("writing %+v", m.mock.scores[assess.Writing])
	}
}

func TestMockTilesUndo(t *testing.T) {
	m, _ := mockModel(t, 80, 24, false)
	m = press(m, "m", "4")
	for m.mock.section != assess.Writing {
		m, _ = answerRight(t, m)
	}
	m = press(m, "1", "1", "2")
	if len(m.mock.placed) != 2 {
		t.Fatalf("placed %v: a tile is placed once", m.mock.placed)
	}
	m = press(m, "backspace")
	if len(m.mock.placed) != 1 {
		t.Errorf("placed %v after undo", m.mock.placed)
	}
}

// screens walks a whole sitting, calling see at each screen.
func screens(t *testing.T, m Model, see func(name string, m Model)) {
	t.Helper()
	see("intro", m)
	m = press(m, "m")
	see("pick", m)
	m = press(m, "4")
	seen := map[assess.ItemKind]bool{}
	for i := 0; m.phase == phaseMock && i < 100; i++ {
		if it := m.current(); !seen[it.Kind] {
			seen[it.Kind] = true
			see(map[assess.ItemKind]string{assess.HeardWord: "heard", assess.HeardMeaning: "meaning", assess.Gap: "gap",
				assess.Passage: "passage", assess.Tiles: "tiles", assess.Compose: "compose"}[it.Kind], m)
		} else {
			see("", m)
		}
		var cmd tea.Cmd
		if i%3 == 0 && m.current().Kind != assess.Tiles && m.current().Kind != assess.Compose {
			m = press(m, "1") // some wrong, so the last answer shows
		} else {
			m, cmd = answerRight(t, m)
		}
		if m.mock.grading {
			m, _ = m.Update(cmd())
		}
		if m.phase == phaseMock && m.current().Kind == assess.Tiles && len(m.mock.placed) == 0 {
			m = press(m, "1")
			see("", m)
			m = press(m, "backspace")
		}
	}
	see("done", m)
}

func TestMockScreensGolden(t *testing.T) {
	for _, size := range []struct{ w, h int }{{80, 24}, {120, 45}} {
		m, _ := mockModel(t, size.w, size.h, true)
		screens(t, m, func(name string, m Model) {
			if name != "" && name != "intro" {
				uitest.Golden(t, fmt.Sprintf("mock-%s-%dx%d", name, size.w, size.h), m.View())
			}
		})
	}
}

func TestMockScreensFitEightyColumns(t *testing.T) {
	m, _ := mockModel(t, 80, 24, true)
	check := func(name string, m Model) {
		t.Helper()
		out := m.View()
		for _, l := range strings.Split(out, "\n") {
			if lipgloss.Width(l) > 80 {
				t.Errorf("%s: a line %d wide: %s", name, lipgloss.Width(l), l)
			}
		}
		if n := strings.Count(out, "\n") + 1; n > 24 {
			t.Errorf("%s takes %d lines:\n%s", name, n, out)
		}
	}
	screens(t, m, check)
	// The diagnostic's results, with the known-words offer and after it.
	d := assessWiki(t).SetSize(80, 24)
	d, _ = d.Update(tea.KeyMsg{Type: tea.KeyEnter})
	for i := 0; d.phase == phaseAsking && i < 50; i++ {
		d.selected = d.question.Correct
		d, _ = d.Update(tea.KeyMsg{Type: tea.KeyEnter})
	}
	check("results", d)
	check("marked", press(d, "k"))
	check("results pick", press(d, "m"))
}
