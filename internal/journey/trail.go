package journey

import (
	"fmt"

	"github.com/therealsiege/tuilaoshi/internal/wiki"
)

// The trail (Oregon Trail-style) runs the walk on days and supplies:
// walking spends days at the chosen pace, each day the party eats 斋饭 by
// the chosen rations and drinks water, and pace and rations move each
// member's health. Running out hurts. A disciple at zero health can't help
// in battle until rested; the master collapsing sends the party back to
// the last town.

// Paces and rations, gentlest first.
var (
	Paces   = []string{"steady", "brisk", "forced"}
	Rations = []string{"filling", "meager", "bare"}
)

// Trail is the party's state on the road, saved with the journey.
type Trail struct {
	Day     int            `yaml:"day"`
	Steps   int            `yaml:"steps"` // steps walked today
	Food    int            `yaml:"food"`  // 斋饭 portions
	Water   int            `yaml:"water"`
	Pace    string         `yaml:"pace"`
	Rations string         `yaml:"rations"`
	Health  map[string]int `yaml:"health"`            // member -> 0..100
	Morale  map[string]int `yaml:"morale"`            // member -> 0..100
	Notes   []string       `yaml:"-"`                 // what the last day brought, for the map
	Crossed []string       `yaml:"crossed,omitempty"` // regions whose obstacle has been crossed
}

// Supplies a journey starts with, and the most water the party carries.
const (
	StartFood  = 40
	StartWater = 20
	MaxWater   = 30
	Master     = "唐僧"
)

// NewTrail is a fresh journey's trail: day 1, full health and spirits.
func NewTrail() Trail {
	return Trail{Day: 1, Food: StartFood, Water: StartWater, Pace: "steady", Rations: "filling",
		Health: map[string]int{}, Morale: map[string]int{}}
}

// stepsPerDay: how far a day's walking goes at each pace.
var stepsPerDay = map[string]int{"steady": 6, "brisk": 9, "forced": 12}

var (
	foodPerMember = map[string]int{"filling": 3, "meager": 2, "bare": 1}
	rationHealth  = map[string]int{"filling": 3, "meager": 0, "bare": -4}
	paceHealth    = map[string]int{"steady": 2, "brisk": -2, "forced": -6}
	paceMorale    = map[string]int{"steady": 1, "brisk": 0, "forced": -3}
)

// Members are the party walking: the master and the disciples who have joined.
func Members(cleared map[int]string) []string {
	out := []string{Master}
	for _, c := range Party(cleared) {
		if c.Unlocked && c.Name != Master {
			out = append(out, c.Name)
		}
	}
	return out
}

// Join makes sure every member has health and morale, new members at full.
func (t *Trail) Join(members []string) {
	if t.Health == nil {
		t.Health = map[string]int{}
	}
	if t.Morale == nil {
		t.Morale = map[string]int{}
	}
	for _, m := range members {
		if _, ok := t.Health[m]; !ok {
			t.Health[m], t.Morale[m] = 100, 80
		}
	}
}

// Step walks one cell. When the day's walking is done it passes a day and
// reports so, with what the day brought in t.Notes.
func (t *Trail) Step(members []string, terrain string) bool {
	t.Steps++
	if t.Steps < stepsPerDay[t.Pace] {
		return false
	}
	t.PassDay(members, terrain, false)
	return true
}

// PassDay ends a day: eating, drinking, and health and spirits by pace and
// rations. A resting day (rest true) walks nowhere and heals.
func (t *Trail) PassDay(members []string, terrain string, rest bool) {
	t.Join(members)
	t.Day++
	t.Steps = 0
	t.Notes = nil
	n := len(members)
	eat := foodPerMember[t.Rations] * n
	hungry := t.Food < eat
	t.Food = max(t.Food-eat, 0)
	drink := n
	if terrain == "desert" || terrain == "fire" {
		drink *= 2
	}
	thirsty := t.Water < drink
	t.Water = max(t.Water-drink, 0)
	if terrain == "river" {
		t.Water = MaxWater // rivers refill the gourds
	}
	health, morale := rationHealth[t.Rations]+paceHealth[t.Pace], paceMorale[t.Pace]
	if rest {
		health, morale = 10, 5
	}
	if hungry {
		health, morale = health-10, morale-8
		t.Notes = append(t.Notes, "The 斋饭 has run out: the party goes hungry.")
	}
	if thirsty {
		health, morale = health-15, morale-8
		t.Notes = append(t.Notes, "The water is gone: the party thirsts.")
	}
	for _, m := range members {
		before := t.Health[m]
		t.Health[m] = clamp(t.Health[m] + health)
		t.Morale[m] = clamp(t.Morale[m] + morale)
		if before > 0 && t.Health[m] == 0 {
			t.Notes = append(t.Notes, fmt.Sprintf("%s has fallen ill.", m))
		}
	}
	if !hungry && t.Food < eat*2 {
		t.Notes = append(t.Notes, "The 斋饭 runs low: buy more in a town, or ask for alms.")
	}
}

// Ill reports whether a member is too sick to help (health 0).
func (t Trail) Ill(member string) bool { h, ok := t.Health[member]; return ok && h == 0 }

// Collapsed reports whether the master has collapsed.
func (t Trail) Collapsed() bool { return t.Ill(Master) }

// Restore heals the party fully and lifts spirits: a night at an inn.
func (t *Trail) Restore(members []string) {
	t.Join(members)
	for _, m := range members {
		t.Health[m] = 100
		t.Morale[m] = clamp(t.Morale[m] + 20)
	}
	t.Water = MaxWater
}

// Next cycles a setting to its next choice.
func Next(choices []string, cur string) string {
	for i, c := range choices {
		if c == cur {
			return choices[(i+1)%len(choices)]
		}
	}
	return choices[0]
}

func clamp(v int) int { return min(max(v, 0), 100) }

// LoadTrail reads a journey's trail (a fresh one for a new journey) and
// the party walking it, joined at full health.
func LoadTrail(store *wiki.Store, cleared map[int]string) (Trail, []string, error) {
	var t Trail
	ok, err := store.Trail(&t)
	if !ok {
		t = NewTrail()
	}
	members := Members(cleared)
	t.Join(members)
	return t, members, err
}
