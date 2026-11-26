package assessment

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/therealsiege/tuilaoshi/internal/assess"
	"github.com/therealsiege/tuilaoshi/internal/audio"
	"github.com/therealsiege/tuilaoshi/internal/claude"
	"github.com/therealsiege/tuilaoshi/internal/ui"
	"github.com/therealsiege/tuilaoshi/internal/wiki"
)

// gradeTimeout bounds how long Claude may take over the written sentence.
const gradeTimeout = 30 * time.Second

// composePoints are what the sentence Claude grades is worth (its 0-2 grade).
const composePoints = 2

// speaker says Chinese aloud; *audio.Client is one.
type speaker interface {
	Available() bool
	Speak(text, lang string) error
}

// grader grades a written sentence 0-2; *claude.Client is one.
type grader interface {
	GradeWriting(ctx context.Context, task, criteria, answer string) (int, string, error)
}

// mockState is a mock test in progress or just finished.
type mockState struct {
	test     assess.Mock
	items    [assess.NumSections][]assess.Item
	scores   [assess.NumSections]assess.SectionScore
	why      [assess.NumSections]string // why a section was skipped
	section  int
	item     int
	selected int
	placed   []int // tile indexes in the order placed
	input    textinput.Model
	grading  bool
	deadline time.Time
	timer    int // which timer is running; older ticks are ignored
	last     string
	audioErr error

	total, most int
	passed      bool
	estimate    int
	sittings    []wiki.MockSitting
	saveErr     error
}

// tickMsg is the mock test's clock ticking for timer number gen.
type tickMsg struct{ gen int }

// gradedMsg is Claude's grade for the written sentence of a section's item.
type gradedMsg struct {
	section, item, score int
	feedback             string
	err                  error
}

// WithTools gives the screen a voice for the listening section and Claude
// for grading the written sentence; either may be missing.
func (m Model) WithTools(tts *audio.Client, ai *claude.Client) Model {
	if tts != nil {
		m.tts = tts
	}
	if ai.Ready() {
		m.grader = ai
	}
	return m
}

func (m Model) clock() time.Time {
	if m.now != nil {
		return m.now()
	}
	return time.Now()
}

// audible reports whether the listening section can be heard.
func (m Model) audible() bool { return m.tts != nil && m.tts.Available() && ui.Prefs.Speech }

// openMock shows the level choice and past sittings.
func (m Model) openMock() Model {
	m.phase = phaseMockPick
	m.mock = mockState{}
	m.mock.sittings, m.mock.saveErr = m.store.MockSittings()
	return m
}

// startMock builds a paper at level and starts its first section.
func (m Model) startMock(level int) (Model, tea.Cmd) {
	pages, pagesErr := m.store.ListPages()
	cards, cardsErr := m.store.DrillableCards()
	m.loadErr = errors.Join(pagesErr, cardsErr)
	rng := m.rng
	if rng == nil {
		rng = rand.New(rand.NewSource(time.Now().UnixNano()))
	}
	old := m.mock.sittings
	m.mock = mockState{test: assess.BuildMock(level, pages, cards, rng), sittings: old}
	t := &m.mock
	t.items = t.test.Sections
	if t.test.Compose != nil && m.grader != nil {
		t.items[assess.Writing] = append(append([]assess.Item{}, t.items[assess.Writing]...), *t.test.Compose)
	}
	for s := range t.items {
		for _, it := range t.items[s] {
			if it.Kind == assess.Compose {
				t.scores[s].Max += composePoints
			} else {
				t.scores[s].Max++
			}
		}
		if len(t.items[s]) == 0 {
			t.scores[s].Skipped, t.why[s] = true, "not enough pages at this level"
		}
	}
	if !m.audible() {
		t.scores[assess.Listening].Skipped, t.why[assess.Listening] = true, "no audio (needs a voice and speech on)"
	}
	t.input = textinput.New()
	t.input.CharLimit = 60
	t.input.Placeholder = "写一个句子"
	m.phase, t.section = phaseMock, -1
	return m.nextSection()
}

// nextSection moves to the next section not skipped, starting its timer,
// or finishes the test after the last.
func (m Model) nextSection() (Model, tea.Cmd) {
	t := &m.mock
	t.section++
	for t.section < assess.NumSections && t.scores[t.section].Skipped {
		t.section++
	}
	if t.section >= assess.NumSections {
		return m.finishMock(), nil
	}
	t.item, t.selected, t.placed, t.last, t.grading = 0, 0, nil, "", false
	t.deadline = m.clock().Add(assess.SectionTimes[t.section])
	t.timer++
	cmd := m.itemCmd() // before m is returned: it may focus the sentence box
	return m, tea.Batch(tick(t.timer), cmd)
}

func tick(gen int) tea.Cmd {
	return tea.Tick(time.Second, func(time.Time) tea.Msg { return tickMsg{gen: gen} })
}

// current is the item being answered.
func (m Model) current() assess.Item { return m.mock.items[m.mock.section][m.mock.item] }

// itemCmd speaks a listening item, or focuses the box for a sentence.
func (m *Model) itemCmd() tea.Cmd {
	it := m.current()
	switch {
	case it.Audio != "" && m.audible():
		tts := m.tts
		return func() tea.Msg { return ui.SpeechDoneMsg{Err: tts.Speak(it.Audio, "zh")} }
	case it.Kind == assess.Compose:
		return m.mock.input.Focus()
	}
	return nil
}

// score records an answer and moves to the next item or section.
func (m Model) score(points int, last string) (Model, tea.Cmd) {
	t := &m.mock
	t.scores[t.section].Points += points
	t.last = last
	t.item++
	t.selected, t.placed = 0, nil
	if t.item >= len(t.items[t.section]) {
		return m.nextSection()
	}
	cmd := m.itemCmd()
	return m, cmd
}

// choose answers a multiple-choice item.
func (m Model) choose(i int) (Model, tea.Cmd) {
	it := m.current()
	heard := ""
	if it.Audio != "" {
		heard = "  " + it.Audio
	}
	if i == it.Correct {
		return m.score(1, "✓"+heard)
	}
	return m.score(0, "✗ "+it.Choices[it.Correct]+heard)
}

// place puts tile i next in the sentence, judging it once all are placed.
func (m Model) place(i int) (Model, tea.Cmd) {
	it := m.current()
	for _, p := range m.mock.placed {
		if p == i {
			return m, nil
		}
	}
	m.mock.placed = append(m.mock.placed, i)
	if len(m.mock.placed) < len(it.Tiles) {
		return m, nil
	}
	var b strings.Builder
	for _, p := range m.mock.placed {
		b.WriteString(it.Tiles[p])
	}
	if b.String() == it.Answer {
		return m.score(1, "✓ "+it.Answer)
	}
	return m.score(0, "✗ "+it.Answer)
}

// submit sends the written sentence to Claude.
func (m Model) submit() (Model, tea.Cmd) {
	answer := strings.TrimSpace(m.mock.input.Value())
	if answer == "" || m.mock.grading {
		return m, nil
	}
	m.mock.grading = true
	it, g, section, item := m.current(), m.grader, m.mock.section, m.mock.item
	return m, func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), gradeTimeout)
		defer cancel()
		score, feedback, err := g.GradeWriting(ctx, it.Prompt,
			"A complete, grammatical Chinese sentence that uses "+it.Answer+" correctly.", answer)
		return gradedMsg{section: section, item: item, score: score, feedback: feedback, err: err}
	}
}

// finishMock scores the sitting, compares it with the level check and
// saves it.
func (m Model) finishMock() Model {
	t := &m.mock
	m.phase = phaseMockDone
	t.total, t.most = assess.MockTotal(t.scores)
	t.passed = assess.MockPassed(t.total, t.most)
	t.estimate = assess.Estimate(m.results)
	if t.estimate == 0 {
		if data, err := os.ReadFile(filepath.Join(m.store.Dir, overviewPage)); err == nil {
			t.estimate = assess.EstimateFromSection(section(string(data), overviewSection))
		}
	}
	if t.most == 0 {
		return m // nothing was sat, so nothing is saved
	}
	sitting := wiki.MockSitting{Date: m.clock().Format("2006-01-02"), Level: t.test.Level,
		Total: t.total, Max: t.most, Passed: t.passed}
	for s, sc := range t.scores {
		sitting.Sections = append(sitting.Sections, wiki.MockSection{Name: assess.SectionNames[s], Score: sc.Score(), Skipped: sc.Skipped})
	}
	t.saveErr = m.store.SaveMockSitting(sitting)
	if t.saveErr == nil {
		t.sittings = append(t.sittings, sitting)
	}
	return m
}

// section is the body of a page's "## heading" section.
func section(page, heading string) string {
	_, rest, ok := strings.Cut(page, "## "+heading+"\n")
	if !ok {
		return ""
	}
	if i := strings.Index(rest, "\n## "); i >= 0 {
		rest = rest[:i]
	}
	return rest
}

// updateMock handles keys and clock ticks during a mock test.
func (m Model) updateMock(msg tea.Msg) (Model, tea.Cmd) {
	t := &m.mock
	switch msg := msg.(type) {
	case tickMsg:
		if m.phase != phaseMock || msg.gen != t.timer {
			return m, nil
		}
		if !m.clock().Before(t.deadline) {
			return m.nextSection() // time is up: what's unanswered scores nothing
		}
		return m, tick(t.timer)
	case ui.SpeechDoneMsg:
		t.audioErr = msg.Err
		return m, nil
	case gradedMsg:
		if m.phase != phaseMock || msg.section != t.section || msg.item != t.item {
			return m, nil
		}
		t.grading = false
		t.input.Reset()
		if msg.err != nil {
			t.scores[t.section].Max -= composePoints // not held against you
			t.scores[t.section].Skipped = t.scores[t.section].Max == 0
			return m.score(0, "Could not grade the sentence: "+msg.err.Error())
		}
		return m.score(min(msg.score, composePoints), fmt.Sprintf("%d/%d %s", msg.score, composePoints, msg.feedback))
	case tea.KeyMsg:
		return m.mockKey(msg)
	}
	return m, nil
}

// mockKey handles a key in the level choice, a section or the results.
func (m Model) mockKey(k tea.KeyMsg) (Model, tea.Cmd) {
	t := &m.mock
	switch m.phase {
	case phaseMockPick:
		switch k.String() {
		case "4", "5":
			return m.startMock(int(k.String()[0] - '0'))
		}
		return m, nil
	case phaseMockDone:
		switch {
		case k.String() == "m":
			return m.openMock(), nil
		case key.Matches(k, m.gkeys.Enter):
			return m, ui.SwitchTo(ui.ViewDashboard)
		}
		return m, nil
	}
	it := m.current()
	switch it.Kind {
	case assess.Compose:
		if t.grading {
			return m, nil
		}
		if key.Matches(k, m.gkeys.Enter) {
			return m.submit()
		}
		var cmd tea.Cmd
		t.input, cmd = t.input.Update(k)
		return m, cmd
	case assess.Tiles:
		switch s := k.String(); {
		case k.Type == tea.KeyBackspace && len(t.placed) > 0:
			t.placed = t.placed[:len(t.placed)-1]
		case len(s) == 1 && s[0] >= '1' && int(s[0]-'1') < len(it.Tiles):
			return m.place(int(s[0] - '1'))
		}
		return m, nil
	}
	n := len(it.Choices)
	switch s := k.String(); {
	case k.Type == tea.KeyTab && it.Audio != "":
		cmd := m.itemCmd()
		return m, cmd
	case key.Matches(k, m.gkeys.Up) && t.selected > 0:
		t.selected--
	case key.Matches(k, m.gkeys.Down) && t.selected < n-1:
		t.selected++
	case key.Matches(k, m.gkeys.Enter):
		return m.choose(t.selected)
	case len(s) == 1 && s[0] >= '1' && int(s[0]-'1') < n:
		return m.choose(int(s[0] - '1'))
	}
	return m, nil
}
