package journey

import "time"

// Errand (任务对话) is a task done by talking: a townsperson, played by
// Claude, and a goal whose points the learner has to get across in Chinese.
type Errand struct {
	ID             string        `yaml:"id"`
	Role           string        `yaml:"role"` // who they are, in Chinese (客栈掌柜)
	Name           string        `yaml:"name"` // what they are called (王掌柜)
	RoleEnglish    string        `yaml:"role_english"`
	Persona        string        `yaml:"persona"` // their character, in English, for Claude
	Opening        string        `yaml:"opening"` // their first words
	OpeningEnglish string        `yaml:"opening_english"`
	Goal           string        `yaml:"goal"`
	GoalEnglish    string        `yaml:"goal_english"`
	Points         []ErrandPoint `yaml:"points"`
	Reward         int           `yaml:"reward"` // 铜钱 when every point is done
}

// ErrandPoint is one thing the learner has to get across.
type ErrandPoint struct {
	ID      string `yaml:"id"`
	Chinese string `yaml:"zh"`
	English string `yaml:"en"`
	Listen  string `yaml:"listen"` // the fact Claude listens for, in English
}

// What an errand allows and what finishing one lifts.
const (
	ErrandTurns  = 8 // things the learner can say before the person moves on
	ErrandMorale = 5
)

// Errands returns every errand.
func Errands() []Errand { return content.Errands }

// ErrandFor is the errand a town offers on a day: the same one all day,
// and a different mix of towns and errands from day to day.
func ErrandFor(town string, now time.Time) Errand {
	h := uint32(2166136261) // FNV-1a
	for _, c := range []byte(town + "@" + now.Format("2006-01-02")) {
		h = (h ^ uint32(c)) * 16777619
	}
	return content.Errands[h%uint32(len(content.Errands))]
}
