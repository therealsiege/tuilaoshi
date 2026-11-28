package roam

import (
	"context"
	"math/rand"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/therealsiege/tuilaoshi/internal/claude"
	"github.com/therealsiege/tuilaoshi/internal/journey"
	"github.com/therealsiege/tuilaoshi/internal/road"
	"github.com/therealsiege/tuilaoshi/internal/ui/input"
)

// Grammar gates (语法关): before a region's demon's lair can be fought,
// the learner passes the grammar pattern guarding it by building five
// sentences with it (the road package's gate.go has the rules). The first
// rounds lay word tiles in order (number keys choose, backspace undoes);
// with Claude the last is a sentence of their own, typed with the pinyin
// input, which passes when it keeps the pattern's simple rules and Claude
// judges it right. Without a key the five rounds are all tiles.

// Grader grades Chinese written for a task against criteria, 0-2, with feedback.
type Grader func(ctx context.Context, task, criteria, answer string) (int, string, error)

// gateRounds is how many sentences pass a gate.
const gateRounds = road.GateRounds

// gateRand seeds a gate's choice of sentences and its shuffles; tests fix it.
var gateRand = func() *rand.Rand { return rand.New(rand.NewSource(rand.Int63())) }

// gateGradeMsg carries Claude's judgement of a typed sentence.
type gateGradeMsg struct {
	score    int
	feedback string
	err      error
}

// gateRun is a gate being taken: the road's rules (road.Gate) with what
// the screen adds.
type gateRun struct {
	journey.GrammarGate
	rng      *rand.Rand
	rounds   []int  // the tile rounds' sentences, by index into Sentences
	round    int    // the round under way, from 0; gateRounds once passed
	tiles    []int  // the sentence's words in the order shown
	picked   []int  // the shown tiles chosen so far, in order
	misses   int    // wrong orders of this round's sentence
	note     string // what came of the last try
	feedback string // Claude's word on the typed sentence
	grading  bool   // the typed sentence is with Claude
}

// rules is the gate as the road's rules see it.
func (g *gateRun) rules() *road.Gate {
	return &road.Gate{GrammarGate: g.GrammarGate, Rng: g.rng, Rounds: g.rounds, Round: g.round,
		Tiles: g.tiles, Picked: g.picked, Misses: g.misses, Note: g.note}
}

// keep takes back what the rules changed.
func (g *gateRun) keep(r *road.Gate) {
	g.rounds, g.round, g.tiles, g.picked, g.misses, g.note = r.Rounds, r.Round, r.Tiles, r.Picked, r.Misses, r.Note
}

// passed reports whether every round is done.
func (g *gateRun) passed() bool { return g.rules().Passed() }

// typed reports whether this round is the sentence of the learner's own.
func (g *gateRun) typed() bool { return g.rules().Typed() }

// sentence is the tile round's sentence.
func (g *gateRun) sentence() journey.TileSentence { return g.rules().Sentence() }

// shuffle lays the round's tiles out afresh, never in the right order.
func (g *gateRun) shuffle() {
	r := g.rules()
	r.Shuffle()
	g.keep(r)
}

// chosen is the words of the shown tiles at positions, in that order.
func (g *gateRun) chosen(positions []int) []string { return g.rules().Chosen(positions) }

// WithWriting lets the gate's last round be typed (dict for the pinyin
// input) and judged by Claude; client nil keeps the gates to tiles.
func (m Model) WithWriting(client *claude.Client, dict *input.Dict) Model {
	m.gateIn = gateInput(dict)
	if client != nil {
		m.grader = client.GradeWriting
	}
	return m
}

func gateInput(dict *input.Dict) input.Model {
	in := input.NewWithDict(dict)
	in.SetPlaceholder("write in Chinese: type pinyin, then pick characters")
	in.SetCharLimit(80)
	return in
}

// Typing reports whether the gate's typed round has the keyboard.
func (m Model) Typing() bool { return m.gate != nil && m.gate.typed() }

// gatePassed reports whether this region's lair is free to fight: its
// gate is passed, or it has none.
func (m Model) gatePassed() bool { return m.walk().GatePassed() }

// openGate starts the region's grammar gate from its first round.
func (m Model) openGate() Model {
	r, _ := road.OpenGate(m.region.Region.ID, m.grader != nil, gateRand())
	run := &gateRun{GrammarGate: r.GrammarGate, rng: r.Rng}
	run.keep(&r)
	m.gate, m.note = run, ""
	return m
}

// updateGate takes the gate's keys: tiles by number, backspace to undo,
// esc to step back from the gate; or the typed round's input.
func (m Model) updateGate(k tea.KeyMsg) (Model, tea.Cmd) {
	g := m.gate
	switch {
	case g.passed():
		if key.Matches(k, m.gkeys.Enter) || k.Type == tea.KeySpace || key.Matches(k, m.gkeys.Back) {
			m.gate, m.note = nil, "The gate stands open: walk into the & to face the demon."
		}
		return m, nil
	case g.typed():
		return m.updateGateTyping(k)
	case key.Matches(k, m.gkeys.Back):
		m.gate, m.note = nil, "The grammar gate waits before the lair."
		return m, nil
	case k.Type == tea.KeyBackspace:
		r := g.rules()
		r.Undo()
		g.keep(r)
		return m, nil
	}
	s := k.String()
	if len(s) != 1 || s[0] < '1' || s[0] > '9' {
		return m, nil
	}
	r := g.rules()
	laid := r.Lay(int(s[0] - '1'))
	g.keep(r)
	if laid != road.LaidRight {
		return m, nil
	}
	switch {
	case g.passed():
		m.passGate()
	case g.typed():
		m.gateIn.Reset()
		return m, m.gateIn.Focus()
	}
	return m, nil
}

// passGate records the gate passed for this journey.
func (m *Model) passGate() {
	w := m.walk()
	r := m.gate.rules()
	if err := w.PassGate(r); err != nil {
		m.loadErr = err
	}
	m.gates = w.Gates
	m.gate.keep(r)
	m.gateIn.Blur()
}

// updateGateTyping types the sentence; enter sends it, esc steps back.
func (m Model) updateGateTyping(k tea.KeyMsg) (Model, tea.Cmd) {
	g := m.gate
	switch {
	case key.Matches(k, m.gkeys.Back) && !m.gateIn.Composing():
		m.gateIn.Blur()
		m.gate, m.note = nil, "The grammar gate waits before the lair."
		return m, nil
	case k.Type == tea.KeyEnter && !m.gateIn.Composing():
		return m.submitGate()
	case g.grading:
		return m, nil
	}
	var cmd tea.Cmd
	m.gateIn, cmd = m.gateIn.Update(k)
	return m, cmd
}

// submitGate checks the typed sentence by the pattern's rules, then sends
// it to Claude.
func (m Model) submitGate() (Model, tea.Cmd) {
	g := m.gate
	text := strings.TrimSpace(m.gateIn.Value())
	if text == "" || g.grading || m.grader == nil {
		return m, nil
	}
	r := g.rules()
	if note := r.Unused(text); note != "" {
		g.note, g.feedback = note, ""
		return m, nil
	}
	g.grading, g.note, g.feedback = true, "Claude is reading your sentence...", ""
	grader := m.grader
	task, criteria := r.Task(), r.Criteria()
	return m, func() tea.Msg {
		score, fb, err := grader(context.Background(), task, criteria, text)
		return gateGradeMsg{score, fb, err}
	}
}

// onGateGrade takes Claude's judgement: passing goes through the gate, a
// miss can be mended and sent again.
func (m Model) onGateGrade(msg gateGradeMsg) Model {
	g := m.gate
	if g == nil || !g.typed() {
		return m // the gate was left while Claude read
	}
	g.grading = false
	switch {
	case msg.err != nil:
		g.note = "Couldn't reach Claude: " + msg.err.Error()
	case msg.score == 0:
		g.note, g.feedback = "Not yet: mend it and send it again.", msg.feedback
	default:
		g.note, g.feedback = "好！ The guardian bows.", msg.feedback
		m.passGate()
	}
	return m
}
