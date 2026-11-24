package roam

import (
	"fmt"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/therealsiege/tuilaoshi/internal/journey"
	"github.com/therealsiege/tuilaoshi/internal/ui"
	"github.com/therealsiege/tuilaoshi/internal/ui/uitest"
	"github.com/therealsiege/tuilaoshi/internal/wiki"
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
	eventRoll = func() int { return 99 }     // nor trail events
	return New(store).SetSize(100, 30).Refresh()
}

// walkTo moves the pilgrim along the shortest walkable route until next
// to target (or onto it, for a town), using the map's own walking and
// getting past any obstacle or event on the way.
func walkTo(t *testing.T, m Model, target journey.Point) (Model, tea.Cmd) {
	t.Helper()
	for steps := 0; steps < 300; steps++ {
		next, ok := nextStep(m.region, m.pos, target)
		if !ok {
			t.Fatalf("no route from %v to %v", m.pos, target)
		}
		var cmd tea.Cmd
		m, cmd = m.step(next.X-m.pos.X, next.Y-m.pos.Y)
		if m.event != nil {
			m, _ = m.Update(press("1"))
			m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
			continue
		}
		if cmd != nil || (m.note != "" && !strings.HasPrefix(m.note, "Day ")) {
			return m, cmd
		}
	}
	t.Fatalf("never reached %v", target)
	return m, nil
}

// nextStep is the first step of a shortest route from p to target over
// walkable cells (obstacles count as walkable: they are crossed).
func nextStep(rm journey.RegionMap, p, target journey.Point) (journey.Point, bool) {
	prev := map[journey.Point]journey.Point{p: p}
	queue := []journey.Point{p}
	for len(queue) > 0 {
		q := queue[0]
		queue = queue[1:]
		if q == target {
			for prev[q] != p {
				q = prev[q]
			}
			return q, true
		}
		for _, d := range []journey.Point{{X: -1}, {X: 1}, {Y: -1}, {Y: 1}} {
			n := journey.Point{X: q.X + d.X, Y: q.Y + d.Y}
			if _, seen := prev[n]; seen || rm.At(n) == journey.TileBlock || rm.At(n) == journey.TileEast || rm.At(n) == journey.TileWest {
				continue
			}
			prev[n] = q
			queue = append(queue, n)
		}
	}
	return journey.Point{}, false
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

func TestEncounterSettings(t *testing.T) {
	defer func(p wiki.Settings) { ui.Prefs = p }(ui.Prefs)
	for setting, want := range map[string]int{"normal": encounterChance, "fewer": encounterChance / 2, "off": 0} {
		ui.Prefs.Encounters = setting
		if got := encounterRate(); got != want {
			t.Errorf("%s: rate %d, want %d", setting, got, want)
		}
	}
}

func TestTheTrailRunsOnDaysAndSupplies(t *testing.T) {
	m := roadWith(t, 0, 0)
	m, _ = m.Update(press("p")) // brisk
	m, _ = m.Update(press("f")) // meager
	if m.trail.Pace != "brisk" || m.trail.Rations != "meager" {
		t.Fatalf("pace %s rations %s", m.trail.Pace, m.trail.Rations)
	}
	m, _ = m.Update(press("w")) // rest a day
	saved, _, err := journey.LoadTrail(m.store, m.cleared)
	if err != nil || saved.Day != 2 || saved.Pace != "brisk" || saved.Food != journey.StartFood-2 {
		t.Fatalf("after resting: %+v %v", saved, err)
	}
	m.trail.Health[journey.Master], m.trail.Food, m.trail.Water = 5, 0, 0
	m.trail.Steps = 8 // the next step ends a brisk day
	m.pos = m.region.Start
	m, _ = m.step(-1, 0)
	if !m.trail.Collapsed() && m.trail.Health[journey.Master] != 30 {
		t.Fatalf("a starving, thirsty master at 5 health: %d", m.trail.Health[journey.Master])
	}
	if !strings.Contains(m.note, "collapses") || m.pos != m.region.Start {
		t.Errorf("the master should be carried back: %q at %v", m.note, m.pos)
	}
}

func press(s string) tea.KeyMsg { return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)} }

func TestADayCanBringAnEvent(t *testing.T) {
	m := roadWith(t, 0, 0)
	defer func(roll func() int) { eventRoll = roll }(eventRoll)
	eventRoll = func() int { return 0 }
	m.trail.Steps = 5 // the next step ends a steady day
	m.pos = m.region.Start
	m, _ = m.step(-1, 0)
	if m.event == nil {
		t.Fatal("no event at the end of the day")
	}
	opening := strings.Join(strings.Fields(m.event.English)[:3], " ")
	if strings.Contains(m.View(), opening) {
		t.Errorf("the English shows before choosing:\n%s", m.View())
	}
	food, health := m.trail.Food, m.trail.Health[journey.Master]
	o := m.event.Choices[0].Outcome
	m, _ = m.Update(press("1"))
	if m.event.chose != 0 || !strings.Contains(m.View(), opening) {
		t.Fatalf("after choosing:\n%s", m.View())
	}
	if o.Coins >= 0 {
		wantFood := max(food+o.Food, 0)
		if m.trail.Food != wantFood {
			t.Errorf("food %d, want %d", m.trail.Food, wantFood)
		}
		if o.Who == "" || o.Who == journey.Master {
			if want := min(max(health+o.Health, 0), 100); m.trail.Health[journey.Master] != want {
				t.Errorf("health %d, want %d", m.trail.Health[journey.Master], want)
			}
		}
	}
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if m.event != nil {
		t.Error("enter should walk on")
	}
}

func TestCrossingTheRegionsRiver(t *testing.T) {
	m := roadWith(t, 0, 0) // region 1 is a river region
	if m.region.Crossing == nil {
		t.Fatal("no crossing in the first region")
	}
	m.pos = journey.Point{X: m.region.CrossAt.X + 1, Y: m.region.CrossAt.Y}
	m, _ = m.step(-1, 0)
	if m.event == nil || !m.event.cross || m.pos == m.region.CrossAt {
		t.Fatalf("stepping onto the river: event %v, at %v", m.event, m.pos)
	}
	m, _ = m.Update(press("2")) // the ferry
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if m.event != nil || m.pos != m.region.CrossAt || !m.hasCrossed() {
		t.Fatalf("after crossing: event %v, at %v, crossed %v", m.event, m.pos, m.hasCrossed())
	}
	m, _ = m.step(-1, 0)
	if m.event != nil {
		t.Error("a crossed river asked again")
	}
	if reloaded := New(m.store).Refresh(); !reloaded.hasCrossed() {
		t.Error("the crossing wasn't saved")
	}
}
