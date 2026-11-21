package journey

// Dialogue is a contextual line spoken by a Journey to the West companion.
type Dialogue struct {
	Speaker     string `yaml:"speaker,omitempty"`     // companion name in Chinese
	English     string `yaml:"english,omitempty"`     // companion name in English
	Line        string `yaml:"line,omitempty"`        // the Chinese dialogue line
	Pinyin      string `yaml:"pinyin,omitempty"`      // pinyin for the line
	Translation string `yaml:"translation,omitempty"` // English translation
	Trigger     string `yaml:"trigger,omitempty"`     // when this shows: "streak", "rank_up", "boss_defeat", "mistake", "review_complete", "idle"
}

// RandomDialogue returns a contextual dialogue line based on the trigger event.
// Uses a simple rotation based on the count parameter to avoid randomness.
func RandomDialogue(trigger string, count int) *Dialogue {
	matches := dialogueByTrigger(trigger)
	if len(matches) == 0 {
		return nil
	}
	return &matches[count%len(matches)]
}

// dialogueByTrigger filters the dialogue pool to lines matching the given trigger.
func dialogueByTrigger(trigger string) []Dialogue {
	all := dialoguePool()
	var matches []Dialogue
	for _, d := range all {
		if d.Trigger == trigger {
			matches = append(matches, d)
		}
	}
	return matches
}

// dialoguePool returns all companion dialogue lines.
func dialoguePool() []Dialogue {
	return content.Dialogue
}
