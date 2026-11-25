package roam

import (
	"math/rand"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/therealsiege/tuilaoshi/internal/journey"
	"github.com/therealsiege/tuilaoshi/internal/ui/uitest"
)

// nightFalls walks the step that ends a steady day, with plotting demons
// sure to be overheard and no event.
func nightFalls(t *testing.T, m Model) Model {
	t.Helper()
	defer func(roll func() int) { plotRoll = roll }(plotRoll)
	plotRoll = func() int { return 0 }
	m.trail.Steps = 5
	m.pos = m.region.Start
	m, _ = m.step(-1, 0)
	return m
}

// answer presses the shown choice that is right (or a wrong one).
func answer(m Model, right bool) Model {
	for i, c := range m.plot.order {
		if (c == 0) == right {
			m, _ = m.Update(press(string(rune('1' + i))))
			return m
		}
	}
	return m
}

// opening is an English line's first few words, which wrapping keeps together.
func opening(english string) string { return strings.Join(strings.Fields(english)[:3], " ") }

// harm opens an event whose only choices hurt the party.
func harm(m Model) Model {
	hurt := journey.EventChoice{Text: "走", English: "Walk", Outcome: journey.Outcome{Food: -10, Health: -20}, Result: "好疼。", ResultEnglish: "Ouch."}
	m.event = &trailEvent{Event: journey.Event{ID: "test", Text: "有埋伏！", English: "An ambush!", Choices: []journey.EventChoice{hurt, hurt}}, chose: -1}
	return m
}

func TestOverheardDemonsForewarnTheParty(t *testing.T) {
	m := nightFalls(t, roadWith(t, 0, 0))
	if m.plot == nil {
		t.Fatal("no demons overheard at nightfall")
	}
	if m.plot.heard || !strings.Contains(m.View(), m.plot.Lines[0].Text) {
		t.Errorf("without audio the demons' lines should be shown to read:\n%s", m.View())
	}
	if !strings.Contains(m.Hear(), m.plot.Lines[0].Text) {
		t.Errorf("ctrl+r should read the demons' lines: %q", m.Hear())
	}
	if strings.Contains(m.View(), opening(m.plot.Lines[0].English)) {
		t.Error("the English shows before answering")
	}
	m = answer(m, true)
	if !m.trail.Forewarned || !strings.Contains(m.View(), "Forewarned") || !strings.Contains(m.View(), opening(m.plot.Lines[0].English)) {
		t.Fatalf("after the right answer:\n%s", m.View())
	}
	saved, _, err := journey.LoadTrail(m.store, m.cleared)
	if err != nil || !saved.Forewarned {
		t.Fatalf("the forewarning wasn't saved: %v", err)
	}
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if m.plot != nil {
		t.Fatal("enter should walk on")
	}
	if again := nightFalls(t, m); again.plot != nil {
		t.Error("a forewarned party overheard more demons")
	}

	food, health := m.trail.Food, m.trail.Health[journey.Master]
	m = harm(m)
	m, _ = m.Update(press("1"))
	if m.trail.Food != food-5 || m.trail.Health[journey.Master] != health-10 {
		t.Errorf("forewarned harm: food %d->%d, health %d->%d; want it halved", food, m.trail.Food, health, m.trail.Health[journey.Master])
	}
	if m.trail.Forewarned || !strings.Contains(m.event.effect, "forewarned") {
		t.Errorf("the warning should be spent and said: %q, still forewarned %v", m.event.effect, m.trail.Forewarned)
	}
}

func TestAWrongAnswerLeavesTheRoadAsItWas(t *testing.T) {
	m := nightFalls(t, roadWith(t, 0, 0))
	m = answer(m, false)
	if m.trail.Forewarned || !strings.Contains(m.View(), "slip away") {
		t.Fatalf("after a wrong answer:\n%s", m.View())
	}
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	health := m.trail.Health[journey.Master]
	m = harm(m)
	m, _ = m.Update(press("1"))
	if m.trail.Health[journey.Master] != health-20 {
		t.Errorf("unwarned harm: health %d->%d, want -20", health, m.trail.Health[journey.Master])
	}
}

func TestNoDemonsOnANightWithAnEvent(t *testing.T) {
	m := roadWith(t, 0, 0)
	defer func(roll func() int) { eventRoll = roll }(eventRoll)
	eventRoll = func() int { return 0 }
	if m = nightFalls(t, m); m.event == nil || m.plot != nil {
		t.Errorf("an event night: event %v, plot %v", m.event != nil, m.plot != nil)
	}
}

func TestTheOverheardDemonsScreen(t *testing.T) {
	m := roadWith(t, 0, 0)
	m = m.overhear(journey.Plots()[0], rand.New(rand.NewSource(2)))
	m.plot.heard = true // listening: the Chinese is hidden until answered
	out := m.View()
	uitest.Golden(t, "overheard-listening", out)
	if strings.Contains(out, m.plot.Lines[0].Text) {
		t.Error("while listening the lines should be hidden")
	}
	m = answer(m, true)
	uitest.Golden(t, "overheard-forewarned", m.View())
}
