package mobile

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/therealsiege/tuilaoshi/internal/journey"
	"github.com/therealsiege/tuilaoshi/internal/road"
)

// tileCodes are the letters a region's tiles cross as: B blocked, R the
// road, G wild ground, T a town, S a tribulation, E the door east, W the
// door west, C the crossing.
var tileCodes = map[journey.Tile]byte{
	journey.TileBlock: 'B', journey.TilePath: 'R', journey.TileWild: 'G', journey.TileTown: 'T',
	journey.TileStage: 'S', journey.TileEast: 'E', journey.TileWest: 'W', journey.TileCrossing: 'C',
}

type point struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type roadTown struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	English string `json:"english"`
	X       int    `json:"x"`
	Y       int    `json:"y"`
}

type roadStage struct {
	Number int    `json:"number"`
	X      int    `json:"x"`
	Y      int    `json:"y"`
	State  string `json:"state"` // cleared, open or sealed
}

type roadCrossing struct {
	Kind    string `json:"kind"`
	Name    string `json:"name"`
	English string `json:"english"`
	X       int    `json:"x"`
	Y       int    `json:"y"`
	Crossed bool   `json:"crossed"`
}

type roadMap struct {
	Region struct {
		ID      string `json:"id"`
		Name    string `json:"name"`
		English string `json:"english"`
		Terrain string `json:"terrain"`
	} `json:"region"`
	Width    int           `json:"width"`
	Height   int           `json:"height"`
	Tiles    []string      `json:"tiles"` // rows, top first, one letter a tile (see tileCodes)
	Towns    []roadTown    `json:"towns"`
	Stages   []roadStage   `json:"stages"`
	Crossing *roadCrossing `json:"crossing,omitempty"`
	Gate     *roadGateSign `json:"gate,omitempty"` // the lair's grammar gate, while the lair stands
	Pos      point         `json:"pos"`
}

type roadGateSign struct {
	Pattern string `json:"pattern"`
	Pinyin  string `json:"pinyin"`
	Passed  bool   `json:"passed"`
}

// Road is the region the pilgrim walks, as JSON: its names and terrain,
// its tiles as rows of letters, its towns, its tribulations by state, its
// crossing, and where the pilgrim stands. It reloads the journey, so the
// stages count what was studied since.
func (c *Core) Road() (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	w, err := road.Load(c.store)
	if err != nil {
		return "", err
	}
	c.walk = &w
	data, err := json.Marshal(toRoadMap(w))
	return string(data), err
}

func toRoadMap(w road.Walk) roadMap {
	rm := w.Region
	out := roadMap{Width: rm.W, Height: rm.H, Towns: []roadTown{}, Stages: []roadStage{}, Pos: point{w.Pos.X, w.Pos.Y}}
	out.Region.ID, out.Region.Name, out.Region.English, out.Region.Terrain = rm.Region.ID, rm.Region.Name, rm.Region.English, rm.Region.Terrain
	for _, row := range rm.Tiles {
		line := make([]byte, len(row))
		for x, t := range row {
			line[x] = tileCodes[t]
		}
		out.Tiles = append(out.Tiles, string(line))
	}
	for p, t := range rm.Towns {
		out.Towns = append(out.Towns, roadTown{t.ID, t.Name, t.English, p.X, p.Y})
	}
	sort.Slice(out.Towns, func(a, b int) bool { return out.Towns[a].X > out.Towns[b].X }) // east to west, as walked
	for p, n := range rm.Stages {
		state := "sealed"
		switch {
		case w.StageCleared(p):
			state = "cleared"
		case w.Stage != nil && w.Stage.Number == n && w.Stage.Status == journey.StageOpen:
			state = "open"
		}
		out.Stages = append(out.Stages, roadStage{n, p.X, p.Y, state})
	}
	sort.Slice(out.Stages, func(a, b int) bool { return out.Stages[a].Number < out.Stages[b].Number })
	if cr := rm.Crossing; cr != nil {
		out.Crossing = &roadCrossing{cr.Kind, cr.Name, cr.English, rm.CrossAt.X, rm.CrossAt.Y, w.HasCrossed()}
	}
	if g, ok := journey.GrammarGateFor(rm.Region.ID); ok {
		if _, won := w.Cleared[rm.Region.To]; !won {
			out.Gate = &roadGateSign{g.Pattern, g.Pinyin, w.GatePassed()}
		}
	}
	return out
}

type stepName struct {
	ID      string `json:"id,omitempty"`
	Name    string `json:"name"`
	English string `json:"english"`
}

type stepOutcome struct {
	Kind      string     `json:"kind"`              // moved, blocked, town, tribulation, encounter, region, crossing, event or plot
	Text      string     `json:"text,omitempty"`    // what to say: why it's blocked, the tribulation, who leaps out…
	Chinese   string     `json:"chinese,omitempty"` // the Chinese of it, to read aloud
	Stage     int        `json:"stage,omitempty"`   // tribulation: its number
	Fight     bool       `json:"fight,omitempty"`   // tribulation: open, so its battle starts
	Town      *roadTown  `json:"town,omitempty"`
	Soldier   *stepName  `json:"soldier,omitempty"`
	Region    *stepName  `json:"region,omitempty"`    // region: the one come to; reload Road
	Day       string     `json:"day,omitempty"`       // a day ended on this step: "Day 4." and what it brought
	Collapsed bool       `json:"collapsed,omitempty"` // the master collapsed and was carried back
	Event     *roadEvent `json:"event,omitempty"`     // event or crossing: met with Choose, then WalkOn
	Plot      *roadPlot  `json:"plot,omitempty"`      // plot: answered with Answer, then WalkOn
	Gate      *roadGate  `json:"gate,omitempty"`      // tribulation: the lair's grammar gate, taken with LayTile
	Pos       point      `json:"pos"`
}

// Step walks the pilgrim one cell (dx, dy: one of them ±1, the other 0)
// by the Mac's rules and returns what came of it as JSON, with where the
// pilgrim now stands. A step spends part of the trail's day; a day's end
// may bring an event or demons plotting, which open (see trail.go), as do
// the region's crossing and the grammar gate before its lair (gate.go).
// Battles and towns are the Mac's for now (#62, #63): they only say what
// is there.
func (c *Core) Step(dx, dy int) (string, error) {
	if dx*dx+dy*dy != 1 {
		return "", fmt.Errorf("step (%d, %d) is not one cell", dx, dy)
	}
	set, err := c.store.Settings()
	if err != nil {
		return "", err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.event != nil || c.plot != nil || c.gate != nil {
		return "", errOpen
	}
	w, err := c.loadedWalk()
	if err != nil {
		return "", err
	}
	o, err := w.Move(dx, dy)
	if err != nil {
		return "", err
	}
	out := toStep(*w, o)
	switch {
	case o.Kind == road.Crossing:
		c.event = &openEvent{Event: *o.Event, chose: -1, cross: true}
		out.Event = c.toEvent()
	case o.Kind == road.Tribulation && o.Gate:
		out.Gate = c.openGate()
	case o.Kind == road.Moved:
		out, err = c.walked(w, road.Rate(set.Encounters))
		if err != nil {
			return "", err
		}
	}
	data, err := json.Marshal(out)
	return string(data), err
}

// walked passes the trail's step, then says what the cell holds: the
// day's event or plot comes first, as on the Mac. The caller holds c.mu.
func (c *Core) walked(w *road.Walk, rate int) (stepOutcome, error) {
	to := w.Pos
	d, err := w.Walked(termWeather(), eventRoll, plotRoll)
	if err != nil {
		return stepOutcome{}, err
	}
	var out stepOutcome
	switch {
	case d.Event != nil:
		c.event = &openEvent{Event: *d.Event, chose: -1}
		out = stepOutcome{Kind: "event", Event: c.toEvent(), Chinese: d.Event.Text}
	case d.Plot != nil:
		c.plot = &openPlot{Plot: *d.Plot, order: d.Order, chose: -1}
		out = stepOutcome{Kind: "plot", Plot: c.toPlot(), Chinese: road.PlotText(*d.Plot)}
	default:
		out = toStep(*w, w.Arrive(to, rate, encounterRoll))
	}
	out.Day, out.Collapsed, out.Pos = d.Note, d.Collapsed, point{w.Pos.X, w.Pos.Y}
	return out, nil
}

// forgetRoad drops the road loaded, for the next call to read afresh.
func (c *Core) forgetRoad() {
	c.mu.Lock()
	c.walk, c.event, c.plot, c.gate = nil, nil, nil, nil
	c.mu.Unlock()
}

// toStep says an outcome in the phone's words: the Mac's where it has
// them, and a line where the Mac shows a screen instead.
func toStep(w road.Walk, o road.Outcome) stepOutcome {
	out := stepOutcome{Kind: string(o.Kind), Text: o.Note, Pos: point{w.Pos.X, w.Pos.Y}}
	switch o.Kind {
	case road.Blocked:
		out.Text = "The way is blocked."
	case road.Crossing:
		cr := w.Region.Crossing
		out.Text, out.Chinese = fmt.Sprintf("%s %s bars the road: choose a way across.", cr.Name, cr.English), o.Event.Text
	case road.Tribulation:
		t := journey.AllTribulations()[o.Stage-1]
		out.Stage, out.Fight, out.Chinese = o.Stage, o.Fight, t.Name
		switch {
		case o.Gate:
			out.Text = fmt.Sprintf("第%d难 %s %s: a grammar gate stands before the lair.", t.Number, t.Name, t.English)
		case o.Fight:
			out.Text = fmt.Sprintf("第%d难 %s %s bars the road: its battle waits here.", t.Number, t.Name, t.English)
		}
	case road.Town:
		out.Town = &roadTown{o.Town.ID, o.Town.Name, o.Town.English, w.Pos.X, w.Pos.Y}
		out.Text, out.Chinese = fmt.Sprintf("You enter %s %s.", o.Town.Name, o.Town.English), o.Town.Name
	case road.Encounter:
		out.Soldier = &stepName{Name: o.Soldier.Name, English: o.Soldier.English}
		out.Chinese = o.Soldier.Name
	case road.Travelled:
		r := w.Region.Region
		out.Region, out.Chinese = &stepName{r.ID, r.Name, r.English}, r.Name
	}
	return out
}
