package journey

import (
	"fmt"
	"math/rand"
	"strings"
)

// Mystery (火眼金睛 Fiery Eyes) is a town's puzzle: something has gone
// wrong, a few facts are known, and of three villagers one is a demon in
// disguise. Each villager has a true story and a lie; the demon tells the
// lie, whose flaw contradicts the facts.
type Mystery struct {
	ID               string     `yaml:"id"`
	Title            string     `yaml:"title"`
	English          string     `yaml:"english"`
	Situation        string     `yaml:"situation"`
	SituationEnglish string     `yaml:"situation_english"`
	Facts            []string   `yaml:"facts"`
	FactsEnglish     []string   `yaml:"facts_english"`
	Villagers        []Villager `yaml:"villagers"`
}

// Villager is someone questioned in a mystery, played by Claude.
type Villager struct {
	Name        string `yaml:"name"`
	Role        string `yaml:"role"`
	English     string `yaml:"english"`
	Persona     string `yaml:"persona"`
	Story       string `yaml:"story"`        // what happened to them, true to the facts
	Lie         string `yaml:"lie"`          // what they say when they are the demon
	Flaw        string `yaml:"flaw"`         // why the lie can't be true
	FlawEnglish string `yaml:"flaw_english"` // the same, in English
}

// What a mystery allows and what it pays or costs.
const (
	MysteryQuestions = 8  // questions across all villagers
	MysteryBounty    = 30 // 铜钱 for unmasking the demon
	MysteryMorale    = 10 // the party's spirits lost when the demon slips away
)

// Mysteries returns every mystery setup.
func Mysteries() []Mystery { return content.Mysteries }

// PickMystery chooses a mystery and which villager is the demon.
func PickMystery(rng *rand.Rand) (Mystery, int) {
	m := content.Mysteries[rng.Intn(len(content.Mysteries))]
	return m, rng.Intn(len(m.Villagers))
}

// VillagerPrompt is the system prompt for playing villager i, who is the
// demon when i == demon.
func (m Mystery) VillagerPrompt(i, demon int) string {
	v := m.Villagers[i]
	var b strings.Builder
	fmt.Fprintf(&b, "You are %s (%s), %s, in a small town on the road west in the world of Journey to the West. %s\n\n", v.Name, v.Role, v.English, v.Persona)
	fmt.Fprintf(&b, "What happened: %s\nWhat everyone in town knows:\n", m.Situation)
	for _, f := range m.Facts {
		b.WriteString("- " + f + "\n")
	}
	if i == demon {
		fmt.Fprintf(&b, "\nSecretly you are a demon in disguise, and you caused this. Your story is: %s\n", v.Lie)
		b.WriteString("Keep to your story and its details even when they don't fit the facts; never correct them. Never say or hint that you are a demon. If pressed, be friendly and a little evasive.\n")
	} else {
		fmt.Fprintf(&b, "\nYou are an ordinary villager and innocent. What you saw and did: %s\n", v.Story)
		b.WriteString("Everything you say must agree with the facts above. If you don't know something, say so.\n")
	}
	b.WriteString("\nA travelling monk's disciple, a learner of Chinese (around HSK3-4), is questioning you. Stay in character. Answer in 1-3 short, simple sentences of simplified Chinese (HSK3-4 words), with no pinyin. Then, on a separate last line, write EN: followed by an English translation of your answer.")
	return b.String()
}

// SplitReply separates a villager's Chinese answer from its English line
// (marked EN:).
func SplitReply(reply string) (chinese, english string) {
	var zh []string
	for _, line := range strings.Split(strings.TrimSpace(reply), "\n") {
		line = strings.TrimSpace(line)
		if v, ok := strings.CutPrefix(line, "EN:"); ok {
			english = strings.TrimSpace(v)
		} else if line != "" {
			zh = append(zh, line)
		}
	}
	return strings.Join(zh, ""), english
}
