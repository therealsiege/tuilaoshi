package journey

import (
	"math/rand"
	"slices"
)

// Event is something that befalls the party on the trail (content/events.yaml).
type Event struct {
	ID       string        `yaml:"id"`
	Terrains []string      `yaml:"terrains,omitempty"` // where it can happen; empty: anywhere
	Needs    string        `yaml:"needs,omitempty"`    // a member it needs in the party
	Text     string        `yaml:"text"`
	English  string        `yaml:"english"`
	Choices  []EventChoice `yaml:"choices"`
}

// EventChoice is one way to meet an event.
type EventChoice struct {
	Text          string  `yaml:"text"`
	English       string  `yaml:"english"`
	Needs         string  `yaml:"needs,omitempty"`
	Outcome       Outcome `yaml:"outcome"`
	Result        string  `yaml:"result"`
	ResultEnglish string  `yaml:"result_english"`
}

// Outcome is what a choice does to the trail.
type Outcome struct {
	Food   int    `yaml:"food,omitempty"`
	Water  int    `yaml:"water,omitempty"`
	Coins  int    `yaml:"coins,omitempty"` // 铜钱, paid by the caller
	Health int    `yaml:"health,omitempty"`
	Morale int    `yaml:"morale,omitempty"`
	Who    string `yaml:"who,omitempty"` // health and morale for this member only
}

// EventChance is the chance, in percent, that a day brings an event.
const EventChance = 30

// PickEvent chooses an event that can happen on terrain to this party,
// with only the choices the party can make; ok is false if none fits.
func PickEvent(terrain string, members []string, rng *rand.Rand) (Event, bool) {
	var fit []Event
	for _, e := range content.Events {
		if (len(e.Terrains) > 0 && !slices.Contains(e.Terrains, terrain)) || (e.Needs != "" && !slices.Contains(members, e.Needs)) {
			continue
		}
		var choices []EventChoice
		for _, c := range e.Choices {
			if c.Needs == "" || slices.Contains(members, c.Needs) {
				choices = append(choices, c)
			}
		}
		if len(choices) >= 2 {
			e.Choices = choices
			fit = append(fit, e)
		}
	}
	if len(fit) == 0 {
		return Event{}, false
	}
	return fit[rng.Intn(len(fit))], true
}

// Apply puts an outcome's food, water, health and morale into the trail;
// coins are the caller's to pay.
func (t *Trail) Apply(o Outcome, members []string) {
	t.Join(members)
	t.Food = max(t.Food+o.Food, 0)
	t.Water = min(max(t.Water+o.Water, 0), MaxWater)
	for _, m := range members {
		if o.Who == "" || o.Who == m {
			t.Health[m] = clamp(t.Health[m] + o.Health)
			t.Morale[m] = clamp(t.Morale[m] + o.Morale)
		}
	}
}

// Events returns every trail event.
func Events() []Event { return content.Events }
