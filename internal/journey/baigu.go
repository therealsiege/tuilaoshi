package journey

// 三打白骨精 (content/baigu.yaml): the White Bone Demon's tribulation is
// won by seeing through her three disguises, as in chapter 27 of the
// novel. Each disguise tests a skill (reading, listening, writing) and has
// several rounds, so a rematch differs.

// BaiguDemon is the boss whose tribulation is the three disguises.
const BaiguDemon = "白骨精"

// BaiguMorale is the party's morale lost when Wukong unmasks the wrong
// thing and the master is angry with him.
const BaiguMorale = 10

// Baigu is the whole trial: the three disguises in order, the master's
// anger at a wrong unmasking, and the ending once all three are seen through.
type Baigu struct {
	Angry        string     `yaml:"angry"`
	AngryEnglish string     `yaml:"angry_english"`
	Won          string     `yaml:"won"`
	WonEnglish   string     `yaml:"won_english"`
	Disguises    []Disguise `yaml:"disguises"`
}

// Disguise is one shape the demon takes, and the skill that sees through it:
// "reading", "listening" or "writing".
type Disguise struct {
	Name    string       `yaml:"name"`
	English string       `yaml:"english"`
	Skill   string       `yaml:"skill"`
	Rounds  []BaiguRound `yaml:"rounds"`
}

// BaiguRound is one telling of a disguise: what is read or heard, a
// question about it, and three answers, the first one right. A writing
// round also has the criteria Claude grades a written answer against.
type BaiguRound struct {
	ID              string   `yaml:"id"`
	Text            string   `yaml:"text"`
	English         string   `yaml:"english"`
	Question        string   `yaml:"question"`
	QuestionEnglish string   `yaml:"question_english"`
	Choices         []string `yaml:"choices"`
	EnglishChoices  []string `yaml:"english_choices"`
	Criteria        string   `yaml:"criteria,omitempty"`
}

// BaiguTrial returns the trial of the White Bone Demon.
func BaiguTrial() Baigu { return content.Baigu }
