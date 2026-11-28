package mobile

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/therealsiege/tuilaoshi/internal/journey"
	"github.com/therealsiege/tuilaoshi/internal/road"
)

// The trail on the phone, by the Mac's rules (internal/road): days and
// supplies, the day's events and the region's crossing, and demons
// overheard plotting. An event, crossing or plot met on a step stays open
// until chosen or answered and then walked on from (WalkOn); no step is
// taken meanwhile.

// The draws the trail's chances are rolled against, 0-99; nil rolls at
// random. Tests fix them.
var (
	eventRoll, plotRoll, encounterRoll func() int
	termWeather                        = func() string { return road.Today().Weather }
)

// errOpen refuses a step while something on the road waits to be met.
var errOpen = errors.New("finish what the road brought first, then walk on")

type trailMember struct {
	Name   string `json:"name"`
	Health int    `json:"health"` // 0-100
	Morale int    `json:"morale"` // 0-100
	Ill    bool   `json:"ill,omitempty"`
}

type trailTerm struct {
	Name    string `json:"name"`
	English string `json:"english"`
	Weather string `json:"weather"` // cold, heat or rain
	Text    string `json:"text"`    // what it does on the road
}

type trailState struct {
	Day         int           `json:"day"`
	Steps       int           `json:"steps"` // walked today
	Food        int           `json:"food"`  // 斋饭 portions
	Water       int           `json:"water"`
	FoodLow     bool          `json:"foodLow,omitempty"`
	WaterLow    bool          `json:"waterLow,omitempty"`
	Coins       int           `json:"coins"` // 铜钱 in the purse
	Pace        string        `json:"pace"`
	PaceNote    string        `json:"paceNote"`
	Rations     string        `json:"rations"`
	RationNote  string        `json:"rationNote"`
	Term        *trailTerm    `json:"term,omitempty"`
	SandalDays  int           `json:"sandalDays"` // days of walking left on the sandals
	SandalsWorn bool          `json:"sandalsWorn,omitempty"`
	Forewarned  bool          `json:"forewarned,omitempty"`
	Party       []trailMember `json:"party"`
	HP          int           `json:"hp"`
	MaxHP       int           `json:"maxHP"`
	Note        string        `json:"note,omitempty"` // what the last change did
}

// Trail is the party's state on the road as JSON: the day, the supplies,
// the purse, the pace and rations, the season's weather, the sandals, the
// party's health and spirits, and the pilgrim's HP.
func (c *Core) Trail() (string, error) {
	return c.trail(func(*road.Walk) (string, error) { return "", nil })
}

// NextPace changes to the next pace (steady, brisk, forced) and returns the trail.
func (c *Core) NextPace() (string, error) {
	return c.trail(func(w *road.Walk) (string, error) { return w.NextPace() })
}

// NextRations changes to the next rations (filling, meager, bare) and returns the trail.
func (c *Core) NextRations() (string, error) {
	return c.trail(func(w *road.Walk) (string, error) { return w.NextRations() })
}

// Rest passes a day without walking and returns the trail.
func (c *Core) Rest() (string, error) {
	return c.trail(func(w *road.Walk) (string, error) { return w.Rest(termWeather()) })
}

// trail does change to the loaded road and returns the trail after it, with
// what it said.
func (c *Core) trail(change func(*road.Walk) (string, error)) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	w, err := c.loadedWalk()
	if err != nil {
		return "", err
	}
	note, err := change(w)
	if err != nil {
		return "", err
	}
	out, err := c.toTrail(*w)
	if err != nil {
		return "", err
	}
	out.Note = note
	data, err := json.Marshal(out)
	return string(data), err
}

func (c *Core) toTrail(w road.Walk) (trailState, error) {
	t := w.Trail
	out := trailState{Day: t.Day, Steps: t.Steps, Food: t.Food, Water: t.Water, FoodLow: w.FoodLow(), WaterLow: w.WaterLow(),
		Pace: t.Pace, PaceNote: road.PaceNote[t.Pace], Rations: t.Rations, RationNote: road.RationNote[t.Rations],
		SandalDays: w.SandalDaysLeft(), SandalsWorn: t.SandalsWorn(), Forewarned: t.Forewarned, Party: []trailMember{}}
	if term := road.Today(); term.Name != "" && term.Weather != "" {
		out.Term = &trailTerm{term.Name, term.English, term.Weather, road.WeatherShort[term.Weather]}
	}
	for _, name := range w.Members {
		h := t.Health[name]
		out.Party = append(out.Party, trailMember{name, h, t.Morale[name], h == 0})
	}
	purse, err := c.store.Wallet()
	if err != nil {
		return out, err
	}
	out.Coins = purse.Tongqian
	out.HP, out.MaxHP, err = road.PilgrimHP(c.store)
	return out, err
}

// loadedWalk is the road as last loaded, loading it if need be. The
// caller holds c.mu.
func (c *Core) loadedWalk() (*road.Walk, error) {
	if c.walk == nil {
		w, err := road.Load(c.store)
		if err != nil {
			return nil, err
		}
		c.walk = &w
	}
	return c.walk, nil
}

// openEvent is an event or crossing being met.
type openEvent struct {
	journey.Event
	chose  int    // -1 until chosen
	effect string // what the choice changed
	cross  bool   // the region's obstacle: crossed on walking on
}

type eventChoice struct {
	Text    string `json:"text"`
	English string `json:"english,omitempty"` // once chosen
}

type roadEvent struct {
	ID            string        `json:"id"`
	Crossing      bool          `json:"crossing,omitempty"`
	Day           int           `json:"day"`
	Text          string        `json:"text"`
	English       string        `json:"english,omitempty"` // once chosen, as on the Mac
	Choices       []eventChoice `json:"choices"`
	Chose         int           `json:"chose"` // -1 until chosen
	Result        string        `json:"result,omitempty"`
	ResultEnglish string        `json:"resultEnglish,omitempty"`
	Effect        string        `json:"effect,omitempty"`
	Forewarned    bool          `json:"forewarned,omitempty"` // harm will be halved
}

func (c *Core) toEvent() *roadEvent {
	e := c.event
	if e == nil {
		return nil
	}
	out := &roadEvent{ID: e.ID, Crossing: e.cross, Day: c.walk.Trail.Day, Text: e.Text, Chose: e.chose, Effect: e.effect,
		Forewarned: e.chose < 0 && c.walk.Trail.Forewarned}
	for _, ch := range e.Choices {
		out.Choices = append(out.Choices, eventChoice{Text: ch.Text})
	}
	if e.chose >= 0 {
		out.English = e.English
		for i, ch := range e.Choices {
			out.Choices[i].English = ch.English
		}
		out.Result, out.ResultEnglish = e.Choices[e.chose].Result, e.Choices[e.chose].ResultEnglish
	}
	return out
}

// Choose meets the open event or crossing by its choice i (from 0) and
// returns it as JSON with what came of it: the result, the English, and
// what it changed. Walk on (WalkOn) to cross or go on.
func (c *Core) Choose(i int) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	e := c.event
	switch {
	case e == nil:
		return "", errors.New("no event is open")
	case e.chose >= 0:
		return "", errors.New("the event is chosen already: walk on")
	case i < 0 || i >= len(e.Choices):
		return "", fmt.Errorf("no choice %d of %d", i, len(e.Choices))
	}
	effect, err := c.walk.Choose(e.Event, i)
	if err != nil {
		return "", err
	}
	e.chose, e.effect = i, effect
	data, err := json.Marshal(c.toEvent())
	return string(data), err
}

// openPlot is demons overheard plotting.
type openPlot struct {
	journey.Plot
	order []int // the choices as shown: indices into Choices, 0 the right one
	chose int   // the shown choice picked, -1 until answered
}

type plotLine struct {
	Who     string `json:"who"`
	Text    string `json:"text"`
	English string `json:"english,omitempty"` // once answered
}

type plotChoice struct {
	Text    string `json:"text"`
	English string `json:"english,omitempty"` // once answered: the right one's and the one chosen
	Right   bool   `json:"right,omitempty"`   // once answered
}

type roadPlot struct {
	Night           int          `json:"night"`
	Lines           []plotLine   `json:"lines"`
	Speech          string       `json:"speech"` // the lines' Chinese, to read aloud
	Question        string       `json:"question"`
	QuestionEnglish string       `json:"questionEnglish,omitempty"`
	Choices         []plotChoice `json:"choices"`
	Chose           int          `json:"chose"` // -1 until answered
	Forewarned      bool         `json:"forewarned,omitempty"`
	Warning         string       `json:"warning,omitempty"` // answered rightly: the demons' plan, forewarned
	WarningEnglish  string       `json:"warningEnglish,omitempty"`
}

func (c *Core) toPlot() *roadPlot {
	p := c.plot
	if p == nil {
		return nil
	}
	answered := p.chose >= 0
	out := &roadPlot{Night: c.walk.Trail.Day, Speech: road.PlotText(p.Plot), Question: p.Question, Chose: p.chose}
	for _, l := range p.Lines {
		line := plotLine{Who: l.Who, Text: l.Text}
		if answered {
			line.English = l.English
		}
		out.Lines = append(out.Lines, line)
	}
	for i, ch := range p.order {
		pc := plotChoice{Text: p.Choices[ch]}
		if answered {
			pc.Right = ch == 0
			if ch == 0 || i == p.chose {
				pc.English = p.EnglishChoices[ch]
			}
		}
		out.Choices = append(out.Choices, pc)
	}
	if answered {
		out.QuestionEnglish = p.QuestionEnglish
		if road.Right(p.order, p.chose) {
			out.Forewarned, out.Warning, out.WarningEnglish = true, p.Warning, p.WarningEnglish
		}
	}
	return out
}

// Answer answers the overheard demons with the shown choice i (from 0)
// and returns the plot as JSON, with the translations: answered rightly,
// the party is forewarned and the next danger does half its harm.
func (c *Core) Answer(i int) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	p := c.plot
	switch {
	case p == nil:
		return "", errors.New("no demons are overheard")
	case p.chose >= 0:
		return "", errors.New("answered already: walk on")
	case i < 0 || i >= len(p.order):
		return "", fmt.Errorf("no choice %d of %d", i, len(p.order))
	}
	if _, err := c.walk.Answer(p.order, i); err != nil {
		return "", err
	}
	p.chose = i
	data, err := json.Marshal(c.toPlot())
	return string(data), err
}

// WalkOn closes the event, crossing, plot or gate open, as the Mac's
// enter does: a crossing chosen puts the party on the far side, and a
// gate is left (passed or not). It returns where the pilgrim stands, as a
// step's outcome.
func (c *Core) WalkOn() (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.walk == nil {
		return "", errors.New("nothing is open")
	}
	w := c.walk
	out := stepOutcome{Kind: string(road.Moved)}
	switch {
	case c.event != nil && c.event.chose < 0:
		return "", errors.New("choose first")
	case c.event != nil:
		if c.event.cross {
			if err := w.Cross(); err != nil {
				return "", err
			}
			out.Text = "The party is across " + w.Region.Crossing.Name + " " + w.Region.Crossing.English + "."
		}
		c.event = nil
	case c.plot != nil && c.plot.chose < 0:
		return "", errors.New("answer first")
	case c.plot != nil:
		c.plot = nil
	case c.gate != nil:
		out.Text = "The grammar gate waits before the lair."
		if c.gate.Passed() {
			out.Text = "The gate stands open: walk into the lair to face the demon."
		}
		c.gate = nil
	default:
		return "", errors.New("nothing is open")
	}
	out.Pos = point{w.Pos.X, w.Pos.Y}
	data, err := json.Marshal(out)
	return string(data), err
}
