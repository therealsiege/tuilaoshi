package journey

import "math/rand"

// Plot is two demons overheard at night plotting an ambush
// (content/plots.yaml). The first choice answers the question rightly.
type Plot struct {
	ID              string     `yaml:"id"`
	Lines           []PlotLine `yaml:"lines"`
	Question        string     `yaml:"question"`
	QuestionEnglish string     `yaml:"question_english"`
	Choices         []string   `yaml:"choices"`
	EnglishChoices  []string   `yaml:"english_choices"`
	Warning         string     `yaml:"warning"`
	WarningEnglish  string     `yaml:"warning_english"`
}

// PlotLine is one demon's line.
type PlotLine struct {
	Who     string `yaml:"who"`
	Text    string `yaml:"text"`
	English string `yaml:"english"`
}

// PlotChance is the chance, in percent, that a day without an event ends
// with demons overheard plotting.
const PlotChance = 20

// PickPlot chooses a plot to overhear.
func PickPlot(rng *rand.Rand) Plot { return content.Plots[rng.Intn(len(content.Plots))] }

// Plots are every overheard plot.
func Plots() []Plot { return content.Plots }

// Forewarn is an outcome met by a forewarned party: every loss of food,
// water, health and morale halved (coins are a price, not harm). spared
// reports whether anything was halved.
func Forewarn(o Outcome) (_ Outcome, spared bool) {
	for _, v := range []*int{&o.Food, &o.Water, &o.Health, &o.Morale} {
		if *v < 0 {
			*v /= 2
			spared = true
		}
	}
	return o, spared
}
