package journey

import "math/rand"

// Alms (化缘) is a townsperson met while asking for food, and three things
// to say to them. The first choice is the fitting, polite request.
type Alms struct {
	Text           string   `yaml:"text"`
	English        string   `yaml:"english"`
	Choices        []string `yaml:"choices"`
	EnglishChoices []string `yaml:"english_choices"`
	Thanks         string   `yaml:"thanks"`
	ThanksEnglish  string   `yaml:"thanks_english"`
}

// What asking well brings: 斋饭 and the party's spirits.
const (
	AlmsFood   = 8
	AlmsMorale = 5
)

// PickAlms chooses a townsperson to ask.
func PickAlms(rng *rand.Rand) Alms { return content.Alms[rng.Intn(len(content.Alms))] }

// Barter is what a town's store gives in 斋饭 for one material.
const Barter = 6
