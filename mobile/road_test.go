package mobile

import (
	"encoding/json"
	"strings"
	"testing"
)

func openRoad(t *testing.T) (*Core, roadMap) {
	t.Helper()
	return openRoadIn(t, t.TempDir(), t.TempDir())
}

func openRoadIn(t *testing.T, wikiDir, progress string) (*Core, roadMap) {
	t.Helper()
	c, err := Open(wikiDir, progress)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { c.Close() }) //nolint:errcheck // test
	data, err := c.Road()
	var rm roadMap
	if err != nil || json.Unmarshal([]byte(data), &rm) != nil {
		t.Fatalf("road %s %v", data, err)
	}
	return c, rm
}

func step(t *testing.T, c *Core, dx, dy int) stepOutcome {
	t.Helper()
	data, err := c.Step(dx, dy)
	var o stepOutcome
	if err != nil || json.Unmarshal([]byte(data), &o) != nil {
		t.Fatalf("step %s %v", data, err)
	}
	return o
}

func TestTheRoadIsTheRegionWalked(t *testing.T) {
	_, rm := openRoad(t)
	if rm.Region.ID != "lingshan-origins" || rm.Region.Name == "" || rm.Region.English == "" || rm.Region.Terrain != "river" {
		t.Errorf("region %+v", rm.Region)
	}
	if rm.Width != 46 || rm.Height != 13 || len(rm.Tiles) != rm.Height || len(rm.Tiles[0]) != rm.Width {
		t.Fatalf("map %dx%d with %d rows", rm.Width, rm.Height, len(rm.Tiles))
	}
	if got := rm.Tiles[rm.Pos.Y][rm.Pos.X]; got != 'R' {
		t.Errorf("the pilgrim starts on %q, want the road", got)
	}
	if strings.Trim(strings.Join(rm.Tiles, ""), "BRGTSEWC") != "" {
		t.Errorf("unknown tile letters in %v", rm.Tiles)
	}
	if len(rm.Stages) != 4 || rm.Stages[0].Number != 1 || rm.Stages[0].State != "open" || rm.Stages[1].State != "sealed" {
		t.Errorf("stages %+v", rm.Stages)
	}
	for _, s := range rm.Stages {
		if rm.Tiles[s.Y][s.X] != 'S' {
			t.Errorf("stage %d at %d,%d is %q", s.Number, s.X, s.Y, rm.Tiles[s.Y][s.X])
		}
	}
	if cr := rm.Crossing; cr == nil || cr.Crossed || rm.Tiles[cr.Y][cr.X] != 'C' || cr.Name == "" {
		t.Errorf("crossing %+v", cr)
	}
}

func TestStepsWalkTheRoad(t *testing.T) {
	wikiDir, progress := t.TempDir(), t.TempDir()
	c, rm := openRoadIn(t, wikiDir, progress)
	if o := step(t, c, 1, 0); o.Kind != "blocked" || o.Text == "" || o.Pos != rm.Pos {
		t.Errorf("east of the first region's start: %+v", o)
	}
	o := step(t, c, -1, 0)
	if o.Kind != "moved" || o.Pos.X != rm.Pos.X-1 {
		t.Fatalf("west along the road: %+v", o)
	}
	if _, err := c.Step(1, 1); err == nil {
		t.Error("a diagonal step should be refused")
	}
	c.Close() //nolint:errcheck // test
	if _, again := openRoadIn(t, wikiDir, progress); again.Pos != o.Pos {
		t.Errorf("reopened at %+v, want %+v", again.Pos, o.Pos)
	}
}

func TestAStepThroughTheWestDoorComesToTheNextRegion(t *testing.T) {
	c, _ := openRoad(t)
	c.walk.Pos = c.walk.Region.End
	o := step(t, c, -1, 0)
	if o.Kind != "region" || o.Region == nil || o.Region.ID != "changan" || !strings.HasPrefix(o.Text, "You come to 长安城外") {
		t.Fatalf("through the west door: %+v", o)
	}
	data, err := c.Road()
	var rm roadMap
	if err != nil || json.Unmarshal([]byte(data), &rm) != nil || rm.Region.ID != "changan" || rm.Pos != o.Pos || len(rm.Towns) == 0 {
		t.Errorf("the next region's road: %s %v", data, err)
	}
	town := rm.Towns[0]
	c.walk.Pos.X, c.walk.Pos.Y = town.X+1, town.Y
	if o := step(t, c, -1, 0); o.Kind != "town" || o.Town == nil || o.Town.ID != town.ID || !strings.HasPrefix(o.Text, "You enter ") {
		t.Errorf("into %s: %+v", town.ID, o)
	}
}

func TestAStepMeetsWhatBarsTheRoad(t *testing.T) {
	c, rm := openRoad(t)
	first := rm.Stages[0]
	c.walk.Pos.X, c.walk.Pos.Y = first.X+1, first.Y
	o := step(t, c, -1, 0)
	if o.Kind != "tribulation" || o.Stage != 1 || !o.Fight || !strings.Contains(o.Text, "第1难") || o.Chinese == "" {
		t.Errorf("the open tribulation: %+v", o)
	}
	sealed := rm.Stages[2]
	c.walk.Pos.X, c.walk.Pos.Y = sealed.X+1, sealed.Y
	if o := step(t, c, -1, 0); o.Kind != "tribulation" || o.Fight || o.Text != "An earlier tribulation still waits on the road behind you." {
		t.Errorf("a later tribulation: %+v", o)
	}
	c.walk.Pos.X, c.walk.Pos.Y = rm.Crossing.X+1, rm.Crossing.Y
	if o := step(t, c, -1, 0); o.Kind != "blocked" || o.Text == "" || !strings.Contains(o.Chinese, rm.Crossing.Name) {
		t.Errorf("the river: %+v", o)
	}
}
