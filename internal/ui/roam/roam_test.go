package roam

import (
	"fmt"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/therealsiege/tuilaoshi/internal/journey"
	"github.com/therealsiege/tuilaoshi/internal/ui/uitest"
)

// roadWith is a journey with stages 1..cleared won and learned characters.
func roadWith(t *testing.T, cleared, learned int) Model {
	t.Helper()
	pages := map[string]string{}
	for i := 0; i < learned; i++ {
		ch := string(rune(0x4E00 + i))
		pages[fmt.Sprintf("characters/c%d.md", i)] = fmt.Sprintf("---\ntype: character\nreview_interval: 2\n---\n\n# %s\n\n- **Character**: %s\n", ch, ch)
	}
	store := uitest.Wiki(t, pages)
	if _, err := store.NewSave("test", time.Now()); err != nil {
		t.Fatal(err)
	}
	for n := 1; n <= cleared; n++ {
		if err := store.RecordStageCleared(n, time.Now()); err != nil {
			t.Fatal(err)
		}
	}
	encounterRoll = func() int { return 99 } // no wild encounters unless a test asks
	return New(store).SetSize(100, 30).Refresh()
}

// walkTo moves the pilgrim along the road until next to p (or onto it,
// for a town), using the map's own walking.
func walkTo(t *testing.T, m Model, target journey.Point) (Model, tea.Cmd) {
	t.Helper()
	for steps := 0; steps < 200; steps++ {
		dx := 0
		switch {
		case m.pos.X > target.X:
			dx = -1
		case m.pos.X < target.X:
			dx = 1
		}
		next := journey.Point{X: m.pos.X + dx, Y: m.pos.Y}
		dy := 0
		if dx == 0 || m.region.At(next) == journey.TileBlock {
			dx = 0
			// follow the road up or down to the column's road cell
			for _, d := range []int{-1, 1} {
				if m.region.At(journey.Point{X: m.pos.X, Y: m.pos.Y + d}) != journey.TileBlock {
					if (d < 0) == (m.pos.Y > target.Y) || m.pos.Y == target.Y {
						dy = d
					}
				}
			}
			if dy == 0 {
				dy = map[bool]int{true: -1, false: 1}[m.pos.Y > target.Y]
			}
		}
		var cmd tea.Cmd
		before := m.pos
		m, cmd = m.step(dx, dy)
		if cmd != nil || m.note != "" {
			return m, cmd
		}
		if m.pos == before {
			t.Fatalf("stuck at %v heading for %v", m.pos, target)
		}
	}
	t.Fatalf("never reached %v", target)
	return m, nil
}

func TestTheRoadStartsInTheFirstRegion(t *testing.T) {
	m := roadWith(t, 0, 0)
	uitest.Golden(t, "first-region", m.View())
	if m.region.Index != 0 || m.pos != m.region.Start {
		t.Fatalf("new journey at region %d %v", m.region.Index, m.pos)
	}
	place, err := m.store.Place()
	if err != nil || place.Region != m.region.Region.ID {
		t.Errorf("place not saved: %+v %v", place, err)
	}
}

func TestATribulationBarsTheRoadUntilWon(t *testing.T) {
	m := roadWith(t, 0, 0) // stage 1 opens with no characters
	var first journey.Point
	for p, n := range m.region.Stages {
		if n == 1 {
			first = p
		}
	}
	m, cmd := walkTo(t, m, first)
	if cmd == nil {
		t.Fatalf("walking into the open tribulation didn't start its battle: %q", m.note)
	}
	if _, ok := cmd().(FightMsg); !ok {
		t.Fatalf("got %T, want FightMsg", cmd())
	}
	if m.pos == first {
		t.Fatal("the pilgrim walked onto an uncleared tribulation")
	}
	// Won: the road past it opens, and the next is sealed until learned.
	if err := m.store.RecordStageCleared(1, time.Now()); err != nil {
		t.Fatal(err)
	}
	m = m.Refresh()
	var second journey.Point
	for p, n := range m.region.Stages {
		if n == 2 {
			second = p
		}
	}
	m, cmd = walkTo(t, m, second)
	if cmd != nil || !strings.Contains(m.note, "sealed") {
		t.Errorf("tribulation 2 with nothing learned: cmd %v, note %q", cmd, m.note)
	}
}

func TestDoorsLeadBetweenRegions(t *testing.T) {
	m := roadWith(t, 4, 0) // region 1 (stages 1-4) cleared
	m.region = journey.BuildMap(0)
	m.pos = m.region.End
	m, _ = m.step(-1, 0)
	if m.region.Index != 1 || m.pos != m.region.Start {
		t.Fatalf("west door led to region %d %v", m.region.Index, m.pos)
	}
	reloaded := New(m.store).Refresh()
	if reloaded.region.Index != 1 || reloaded.pos != m.pos {
		t.Errorf("the place wasn't restored: region %d %v", reloaded.region.Index, reloaded.pos)
	}
	m, _ = m.step(1, 0)
	if m.region.Index != 0 || m.pos != m.region.End {
		t.Errorf("east door led to region %d %v", m.region.Index, m.pos)
	}
}

func TestWalkingIntoATownEntersIt(t *testing.T) {
	m := roadWith(t, 4, 0)
	m.region, m.pos = journey.BuildMap(1), journey.BuildMap(1).Start // Chang'an's region
	var town journey.Point
	for p := range m.region.Towns {
		town = p
	}
	m, cmd := walkTo(t, m, town)
	msg, ok := cmd().(EnterTownMsg)
	if !ok || msg.Town != "changan" {
		t.Fatalf("got %#v, want to enter Chang'an", cmd())
	}
}

func TestWildGroundMeetsTheRegionsClans(t *testing.T) {
	m := roadWith(t, 0, 0)
	var wild, road journey.Point
	for y := 0; y < m.region.H; y++ {
		for x := 0; x < m.region.W; x++ {
			p := journey.Point{X: x, Y: y}
			switch m.region.At(p) {
			case journey.TileWild:
				wild = p
			case journey.TilePath:
				road = p
			}
		}
	}
	defer func(roll func() int) { encounterRoll = roll }(encounterRoll)
	encounterRoll = func() int { return 0 }
	m.pos = journey.Point{X: road.X + 1, Y: road.Y}
	if _, cmd := m.step(-1, 0); cmd != nil {
		t.Error("the road itself drew an encounter")
	}
	m.pos = journey.Point{X: wild.X + 1, Y: wild.Y}
	m, cmd := m.step(-1, 0)
	if m.pos != wild || cmd == nil {
		t.Fatalf("stepping onto wild ground at %v: pos %v, cmd %v", wild, m.pos, cmd)
	}
	msg, ok := cmd().(EncounterMsg)
	if !ok || msg.Soldier.From > m.region.Region.To {
		t.Fatalf("got %#v, want a soldier of the region", cmd())
	}
	if !strings.Contains(m.View(), "HP") {
		t.Error("the map doesn't show HP")
	}
}
