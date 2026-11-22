package journey

import (
	"strings"
	"unicode"
)

// Town is a stop on the road west with a notice board of side quests.
type Town struct {
	ID          string      `yaml:"id"`
	Name        string      `yaml:"name"`
	English     string      `yaml:"english"`
	Stage       int         `yaml:"stage"` // reached once this stage is being worked on or cleared
	Desc        string      `yaml:"desc"`
	DescEnglish string      `yaml:"desc_english"`
	Stone       *Stone      `yaml:"stone,omitempty"`
	Quests      []SideQuest `yaml:"quests"`
}

// SideQuest is a townsperson's request: read their notice or letter and
// answer a question about it (choices written right answer first), or,
// for kind "write", write what they ask for in Chinese, graded by Claude
// against the criteria.
type SideQuest struct {
	ID           string   `yaml:"id"`
	Kind         string   `yaml:"kind,omitempty"` // "" (read) or "write"
	Giver        string   `yaml:"giver"`
	GiverEnglish string   `yaml:"giver_english"`
	Title        string   `yaml:"title"`
	Text         string   `yaml:"text"`
	English      string   `yaml:"english"`
	Question     string   `yaml:"question"`
	Choices      []string `yaml:"choices"`
	Criteria     string   `yaml:"criteria,omitempty"` // what a written answer must do
}

// Writing reports whether the quest asks for writing rather than an answer.
func (q SideQuest) Writing() bool { return q.Kind == "write" }

// writePay is the copper a written request pays per grade point (1-2).
const writePay = 20

// Pay is what the quest pays for an answer graded grade (reading: 1 if right).
func (q SideQuest) Pay(grade int) int {
	if q.Writing() {
		return writePay * grade
	}
	return QuestReward * min(grade, 1)
}

// Stone is a town's secret: an inscription giving a 成语's meaning, a clue
// and its first character, opened by typing the whole idiom from memory.
type Stone struct {
	Idiom   string `yaml:"idiom"`
	Pinyin  string `yaml:"pinyin"`
	Meaning string `yaml:"meaning"`
	Clue    string `yaml:"clue"`
}

// Opens reports whether typed is the idiom, ignoring spaces and punctuation.
func (s Stone) Opens(typed string) bool {
	var b strings.Builder
	for _, r := range typed {
		if unicode.Is(unicode.Han, r) {
			b.WriteRune(r)
		}
	}
	return b.String() == s.Idiom
}

// Inscription shows the first character and a box for each one hidden.
func (s Stone) Inscription() string {
	r := []rune(s.Idiom)
	return string(r[0]) + strings.Repeat("□", len(r)-1)
}

// QuestReward is the copper a quest pays, once.
const QuestReward = 20

// Towns returns the towns in road order.
func Towns() []Town { return content.Towns }

// Reached reports whether the learner has come as far as the town: its
// stage is the one being worked on, or behind it (current 0: all cleared).
func (t Town) Reached(current int) bool { return current == 0 || t.Stage <= current }
