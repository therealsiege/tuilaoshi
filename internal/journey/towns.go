package journey

// Town is a stop on the road west with a notice board of side quests.
type Town struct {
	ID          string      `yaml:"id"`
	Name        string      `yaml:"name"`
	English     string      `yaml:"english"`
	Stage       int         `yaml:"stage"` // reached once this stage is being worked on or cleared
	Desc        string      `yaml:"desc"`
	DescEnglish string      `yaml:"desc_english"`
	Quests      []SideQuest `yaml:"quests"`
}

// SideQuest is a townsperson's request: read their notice or letter and answer
// a question about it. Choices are written right answer first.
type SideQuest struct {
	ID           string   `yaml:"id"`
	Giver        string   `yaml:"giver"`
	GiverEnglish string   `yaml:"giver_english"`
	Title        string   `yaml:"title"`
	Text         string   `yaml:"text"`
	English      string   `yaml:"english"`
	Question     string   `yaml:"question"`
	Choices      []string `yaml:"choices"`
}

// QuestReward is the copper a quest pays, once.
const QuestReward = 20

// Towns returns the towns in road order.
func Towns() []Town { return content.Towns }

// Reached reports whether the learner has come as far as the town: its
// stage is the one being worked on, or behind it (current 0: all cleared).
func (t Town) Reached(current int) bool { return current == 0 || t.Stage <= current }
