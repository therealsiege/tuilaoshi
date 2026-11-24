package journey

import (
	"math/rand"
	"slices"
)

// Crossing is a kind of obstacle on the road (content/crossings.yaml).
type Crossing struct {
	Kind       string              `yaml:"kind"`
	Name       string              `yaml:"name"`
	English    string              `yaml:"english"`
	Terrains   []string            `yaml:"terrains"`
	Choices    []CrossingChoice    `yaml:"choices"`
	Conditions []CrossingCondition `yaml:"conditions"`
	Harm       Outcome             `yaml:"harm"`
	Bad        string              `yaml:"bad"`
	BadEnglish string              `yaml:"bad_english"`
}

// CrossingChoice is one way across, with what it costs when it goes well.
type CrossingChoice struct {
	Text      string  `yaml:"text"`
	English   string  `yaml:"english"`
	Outcome   Outcome `yaml:"outcome"`
	OK        string  `yaml:"ok"`
	OKEnglish string  `yaml:"ok_english"`
}

// CrossingCondition is how the obstacle stands today; danger lists the
// choices it makes go wrong.
type CrossingCondition struct {
	Text    string `yaml:"text"`
	English string `yaml:"english"`
	Danger  []int  `yaml:"danger"`
}

// CrossingFor returns the obstacle a terrain puts on the road, if any.
func CrossingFor(terrain string) *Crossing {
	for i := range content.Crossings {
		if slices.Contains(content.Crossings[i].Terrains, terrain) {
			return &content.Crossings[i]
		}
	}
	return nil
}

// Event makes a crossing into an event to meet: today's condition as the
// text, and each choice's outcome and result by whether the condition
// makes it go wrong.
func (c Crossing) Event(rng *rand.Rand) Event {
	cond := c.Conditions[rng.Intn(len(c.Conditions))]
	e := Event{ID: "crossing:" + c.Kind, Text: c.Name + "：" + cond.Text, English: c.English + ": " + cond.English}
	for i, ch := range c.Choices {
		choice := EventChoice{Text: ch.Text, English: ch.English, Outcome: ch.Outcome, Result: ch.OK, ResultEnglish: ch.OKEnglish}
		if slices.Contains(cond.Danger, i) {
			choice.Outcome = sum(ch.Outcome, c.Harm)
			choice.Result, choice.ResultEnglish = c.Bad, c.BadEnglish
		}
		e.Choices = append(e.Choices, choice)
	}
	return e
}

func sum(a, b Outcome) Outcome {
	return Outcome{Food: a.Food + b.Food, Water: a.Water + b.Water, Coins: a.Coins + b.Coins,
		Health: a.Health + b.Health, Morale: a.Morale + b.Morale, Who: a.Who}
}
