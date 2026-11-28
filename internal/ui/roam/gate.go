package roam

import (
	"context"
	"fmt"
	"math/rand"
	"slices"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/therealsiege/tuilaoshi/internal/claude"
	"github.com/therealsiege/tuilaoshi/internal/journey"
	"github.com/therealsiege/tuilaoshi/internal/ui/input"
)

// Grammar gates (语法关): before a region's demon's lair can be fought,
// the learner passes the grammar pattern guarding it by building five
// sentences with it. The first rounds lay word tiles in order (number keys
// choose, backspace undoes); with Claude the last is a sentence of their
// own, typed with the pinyin input, which passes when it keeps the
// pattern's simple rules and Claude judges it right. Without a key the
// five rounds are all tiles.

// Grader grades Chinese written for a task against criteria, 0-2, with feedback.
type Grader func(ctx context.Context, task, criteria, answer string) (int, string, error)

// gateRounds is how many sentences pass a gate.
const gateRounds = 5

// gateRand seeds a gate's choice of sentences and its shuffles; tests fix it.
var gateRand = func() *rand.Rand { return rand.New(rand.NewSource(rand.Int63())) }

// gateGradeMsg carries Claude's judgement of a typed sentence.
type gateGradeMsg struct {
	score    int
	feedback string
	err      error
}

// gateRun is a gate being taken.
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

// passed reports whether every round is done.
func (g *gateRun) passed() bool { return g.round >= gateRounds }

// typed reports whether this round is the sentence of the learner's own.
func (g *gateRun) typed() bool { return !g.passed() && g.round >= len(g.rounds) }

// sentence is the tile round's sentence.
func (g *gateRun) sentence() journey.TileSentence { return g.Sentences[g.rounds[g.round]] }

// shuffle lays the round's tiles out afresh, never in the right order.
func (g *gateRun) shuffle() {
	s := g.sentence()
	g.picked = nil
	for {
		g.tiles = g.rng.Perm(len(s.Words))
		if len(s.Words) < 2 || !s.Fits(g.chosen(g.tiles)) {
			return
		}
	}
}

// chosen is the words of the shown tiles at positions, in that order.
func (g *gateRun) chosen(positions []int) []string {
	s := g.sentence()
	words := make([]string, len(positions))
	for i, p := range positions {
		words[i] = s.Words[g.tiles[p]]
	}
	return words
}

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
	g, _ := journey.GrammarGateFor(m.region.Region.ID)
	tiles := gateRounds
	if m.grader != nil {
		tiles-- // the last round is typed
	}
	rng := gateRand()
	run := &gateRun{GrammarGate: g, rng: rng, rounds: rng.Perm(len(g.Sentences))[:min(tiles, len(g.Sentences))]}
	run.shuffle()
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
		if len(g.picked) > 0 {
			g.picked = g.picked[:len(g.picked)-1]
		}
		return m, nil
	}
	s := k.String()
	if len(s) != 1 || s[0] < '1' || s[0] > '9' {
		return m, nil
	}
	p := int(s[0] - '1')
	if p >= len(g.tiles) || slices.Contains(g.picked, p) {
		return m, nil
	}
	g.picked = append(g.picked, p)
	if len(g.picked) < len(g.tiles) {
		return m, nil
	}
	sentence := g.sentence()
	if sentence.Fits(g.chosen(g.picked)) {
		g.note = "对！" + strings.Join(g.chosen(g.picked), "") + " " + sentence.English
		return m.nextGateRound()
	}
	g.misses++
	g.note = "不对. The tiles are shuffled again: try once more."
	if g.misses >= 3 {
		g.note += " It reads: " + sentence.Text()
	}
	g.shuffle()
	return m, nil
}

// nextGateRound moves on a round: the next tiles, the typed round, or
// through the gate.
func (m Model) nextGateRound() (Model, tea.Cmd) {
	g := m.gate
	g.round++
	g.misses = 0
	switch {
	case g.passed():
		m.passGate()
	case g.typed():
		m.gateIn.Reset()
		return m, m.gateIn.Focus()
	default:
		g.shuffle()
	}
	return m, nil
}

// passGate records the gate passed for this journey.
func (m *Model) passGate() {
	id := m.region.Region.ID
	if err := m.store.PassGrammarGate(id); err != nil {
		m.loadErr = err
	}
	if !slices.Contains(m.gates, id) {
		m.gates = append(m.gates, id)
	}
	m.gate.round = gateRounds
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
	if !g.Uses(text) {
		g.note, g.feedback = fmt.Sprintf("That doesn't use %s yet. %s", g.Pattern, g.Explain), ""
		return m, nil
	}
	g.grading, g.note, g.feedback = true, "Claude is reading your sentence...", ""
	grader := m.grader
	task := fmt.Sprintf("Write one Chinese sentence of your own using the grammar pattern %s (%s): %s", g.Pattern, g.Pinyin, g.Explain)
	criteria := fmt.Sprintf("A complete, grammatical sentence that uses %s correctly and naturally, at about HSK3-4; any topic. Score 0 if %s is missing or misused.", g.Pattern, g.Pattern)
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
