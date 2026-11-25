package journey

import (
	"hash/fnv"
	"strings"
)

// Riddle is a lantern riddle (字谜): a character described by its parts.
type Riddle struct {
	Riddle  string `yaml:"riddle"`
	Answer  string `yaml:"answer"`
	Pinyin  string `yaml:"pinyin"`
	English string `yaml:"english"`
	Explain string `yaml:"explain"`
}

// Riddles returns every lantern riddle.
func Riddles() []Riddle { return content.Riddles }

// Lantern riddle rewards and limits: copper for a riddle guessed, tries
// before the answer is shown, and riddles a town has each day (more at the
// Lantern Festival, 元宵, when they pay double).
const (
	RiddlePay      = 15
	RiddleTries    = 3
	RiddlesADay    = 1
	RiddlesAtFeast = 3
	LanternFeast   = "元宵节"
)

// RiddleFor is the town's nth riddle of a day: the same for everyone on
// that day, different across towns and days.
func RiddleFor(town, day string, n int) Riddle {
	h := fnv.New32a()
	_, _ = h.Write([]byte(town + day)) //nolint:errcheck // hash writes don't fail
	all := content.Riddles
	return all[(int(h.Sum32()%uint32(len(all)))+n)%len(all)]
}

// Solves reports whether a guess is the riddle's character.
func (r Riddle) Solves(guess string) bool { return strings.TrimSpace(guess) == r.Answer }
