package roam

import (
	"context"
	"errors"
	"math/rand"
	"slices"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/therealsiege/tuilaoshi/internal/journey"
	"github.com/therealsiege/tuilaoshi/internal/ui/uitest"
)

// fixGateRand makes gates choose and shuffle the same way every run.
func fixGateRand(t *testing.T) {
	t.Helper()
	old := gateRand
	gateRand = func() *rand.Rand { return rand.New(rand.NewSource(1)) }
	t.Cleanup(func() { gateRand = old })
}

// rightOrder is the shown tiles' numbers, as keys, that lay the round's
// sentence in order.
func rightOrder(g *gateRun) []string {
	var keys []string
	for w := range g.sentence().Words {
		p := slices.Index(g.tiles, w)
		keys = append(keys, string(rune('1'+p)))
	}
	return keys
}

func pressAll(m Model, keys []string) Model {
	for _, k := range keys {
		m, _ = m.Update(press(k))
	}
	return m
}

// lairOpen is a journey with the first region's first three won and its
// lair open to fight, the pilgrim standing beside it.
func lairOpen(t *testing.T) (Model, journey.Point) {
	t.Helper()
	m := roadWith(t, 3, 0)
	lair := m.region.Lair()
	stages := journey.Stages(m.counts, m.cleared)
	m.stage = &stages[m.region.Region.To-1]
	m.stage.Status = journey.StageOpen // as if enough were learned
	for _, d := range []journey.Point{{X: 1}, {X: -1}, {Y: 1}, {Y: -1}} {
		p := journey.Point{X: lair.X + d.X, Y: lair.Y + d.Y}
		if t := m.region.At(p); t == journey.TilePath || t == journey.TileWild {
			m.pos = p
			m.save()
			return m, lair
		}
	}
	t.Fatal("no ground beside the lair")
	return m, lair
}

func TestTheLairWaitsForItsGrammarGate(t *testing.T) {
	m, lair := lairOpen(t)
	m, cmd := m.step(lair.X-m.pos.X, lair.Y-m.pos.Y)
	if cmd != nil || m.gate == nil || m.pos == lair {
		t.Fatalf("walking into the lair with the gate unpassed: cmd %v, gate %v, at %v", cmd, m.gate, m.pos)
	}
	if !strings.Contains(m.View(), "语法关") || !strings.Contains(m.View(), "把") {
		t.Errorf("the gate screen doesn't show its pattern:\n%s", m.View())
	}
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if m.gate != nil {
		t.Fatal("esc should step back from the gate")
	}
	if err := m.store.PassGrammarGate(m.region.Region.ID); err != nil {
		t.Fatal(err)
	}
	stage := m.stage
	m = m.Refresh()
	m.stage = stage
	if _, cmd = m.act(); cmd == nil {
		t.Fatal("the lair with its gate passed doesn't start the fight")
	}
	if _, ok := cmd().(FightMsg); !ok {
		t.Fatalf("got %T, want FightMsg", cmd())
	}
}

func TestTheTileRoundsPassTheGateWithoutClaude(t *testing.T) {
	m, _ := lairOpen(t)
	m = m.openGate()
	if len(m.gate.rounds) != gateRounds {
		t.Fatalf("without Claude %d tile rounds, want %d", len(m.gate.rounds), gateRounds)
	}
	m, _ = m.Update(press("1"))
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	if len(m.gate.picked) != 0 {
		t.Fatalf("backspace didn't take the tile back: %v", m.gate.picked)
	}
	for round := 0; round < gateRounds; round++ {
		if m.Typing() {
			t.Fatal("a typed round without Claude")
		}
		wrong := rightOrder(m.gate)
		wrong[0], wrong[1] = wrong[1], wrong[0]
		m = pressAll(m, wrong)
		if m.gate.round != round || len(m.gate.picked) != 0 || !strings.Contains(m.gate.note, "不对") {
			t.Fatalf("round %d, a wrong order: round %d, picked %v, note %q", round, m.gate.round, m.gate.picked, m.gate.note)
		}
		if m.gate.sentence().Fits(m.gate.chosen([]int{0, 1, 2, 3, 4, 5, 6, 7, 8}[:len(m.gate.tiles)])) {
			t.Fatal("the tiles were reshuffled into the right order")
		}
		m = pressAll(m, rightOrder(m.gate))
		if m.gate.round != round+1 || !strings.HasPrefix(m.gate.note, "对") {
			t.Fatalf("round %d, the right order: round %d, note %q", round, m.gate.round, m.gate.note)
		}
	}
	if !m.gate.passed() || !m.gatePassed() {
		t.Fatal("five sentences laid and the gate isn't passed")
	}
	if saved, err := m.store.GrammarGates(); err != nil || !slices.Contains(saved, m.region.Region.ID) {
		t.Errorf("the passed gate wasn't saved: %v %v", saved, err)
	}
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if m.gate != nil || !New(m.store, nil).Refresh().gatePassed() {
		t.Error("enter should walk on, and the gate stay passed on reload")
	}
}

func TestTheTypedRoundTakesTheRulesAndClaude(t *testing.T) {
	m, _ := lairOpen(t)
	calls, score, fail := 0, 0, error(nil)
	m.grader = func(_ context.Context, task, criteria, answer string) (int, string, error) {
		calls++
		if !strings.Contains(task, "把") || !strings.Contains(criteria, "把") {
			t.Errorf("the task doesn't name the pattern: %q / %q", task, criteria)
		}
		return score, "说得对。", fail
	}
	m = m.openGate()
	for round := 0; round < gateRounds-1; round++ {
		m = pressAll(m, rightOrder(m.gate))
	}
	if !m.Typing() {
		t.Fatal("with Claude the last round should be typed")
	}
	send := func(text string) tea.Cmd {
		m.gateIn.SetValue(text)
		var cmd tea.Cmd
		m, cmd = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
		return cmd
	}
	if cmd := send("我把。"); cmd != nil || calls != 0 || !strings.Contains(m.gate.note, "doesn't use 把") {
		t.Fatalf("a sentence without the pattern went to Claude: calls %d, note %q", calls, m.gate.note)
	}
	fail = errors.New("offline")
	cmd := send("悟空把金箍棒放进了耳朵里。")
	m, _ = m.Update(cmd())
	if m.gate.passed() || !strings.Contains(m.gate.note, "Couldn't reach Claude") {
		t.Fatalf("a failed call: passed %v, note %q", m.gate.passed(), m.gate.note)
	}
	fail = nil
	cmd = send("悟空把金箍棒放进了耳朵里。")
	m, _ = m.Update(cmd())
	if m.gate.passed() || m.gate.feedback == "" {
		t.Fatalf("Claude's 0 passed the gate or lost its feedback: %+v", m.gate)
	}
	score = 2
	cmd = send("悟空把金箍棒放进了耳朵里。")
	m, _ = m.Update(cmd())
	if !m.gate.passed() || calls != 3 {
		t.Fatalf("Claude's 2: passed %v after %d calls", m.gate.passed(), calls)
	}
	if saved, err := m.store.GrammarGates(); err != nil || !slices.Contains(saved, m.region.Region.ID) {
		t.Errorf("the gate passed by writing wasn't saved: %v %v", saved, err)
	}
}

func TestGateScreens(t *testing.T) {
	fixGateRand(t)
	m, _ := lairOpen(t)
	m = m.openGate()
	m, _ = m.Update(press("1"))
	uitest.Golden(t, "gate-120x45", m.View())
	small := m.SetSize(80, 24)
	uitest.Golden(t, "gate-80x24", small.View())
}

func TestEveryGateScreenFitsASmallWindow(t *testing.T) {
	m := roadWith(t, 0, 0).SetSize(80, 24)
	m.grader = func(context.Context, string, string, string) (int, string, error) { return 0, "", nil }
	long := strings.Repeat("这个句子的语法有一点儿问题，请再想一想。", 8)
	for i := range journey.Regions() {
		m.region = journey.BuildMap(i)
		m = m.openGate()
		g := m.gate
		check := func(what string) {
			out := m.View()
			if n := strings.Count(out, "\n") + 1; n > 24 {
				t.Errorf("%s %s: %d lines:\n%s", g.Region, what, n, out)
			}
			for _, l := range strings.Split(out, "\n") {
				if lipgloss.Width(l) > 80 {
					t.Errorf("%s %s: a line %d wide:\n%s", g.Region, what, lipgloss.Width(l), l)
				}
			}
		}
		rounds := g.rounds
		for s := range g.Sentences {
			g.rounds, g.round, g.misses = []int{s}, 0, 3
			g.shuffle()
			g.picked = []int{0, 1}
			g.note = "不对. The tiles are shuffled again: try once more. It reads: " + g.sentence().Text()
			check("tiles " + g.sentence().Text())
		}
		g.rounds, g.round = rounds, len(rounds)
		g.note, g.feedback = "That doesn't use "+g.Pattern+" yet. "+g.Explain, long
		check("typed")
		g.round, g.note = gateRounds, "好！ The guardian bows."
		check("passed")
	}
}
