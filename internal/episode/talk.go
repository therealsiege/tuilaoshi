package episode

import (
	"fmt"
	"strings"

	"github.com/therealsiege/tuilaoshi/internal/journey"
	"github.com/therealsiege/tuilaoshi/internal/model"
)

// Talker is the companion who talks an episode over: the Monkey King, whose
// story the episodes tell.
const Talker = "孙悟空"

// talkLines is how much of the dialogue the companion is given to recall.
const talkLines = 80

// TalkWords picks up to n of an episode's theme words to build a talk on.
func TalkWords(words []model.Card, n int) []model.Card { return words[:min(n, len(words))] }

// TalkPrompt is the companion's usual prompt, about this episode: they are
// given its dialogue to recall, and asked to talk about what happened in
// it using its words. The dialogue is sent only to the model, at runtime.
func TalkPrompt(c journey.Companion, b journey.Bond, title string, lines []model.DialogueLine, words []model.Card) string {
	var dialogue strings.Builder
	for _, l := range lines[:min(talkLines, len(lines))] {
		dialogue.WriteString(l.Chinese)
		dialogue.WriteString("\n")
	}
	return journey.TalkPrompt(c, b, words) + fmt.Sprintf(`

Today, talk with the learner about the episode they have just watched, 《%s》, as if remembering it together: what happened, who said what, how they felt. Ask about its events. Here is the start of its dialogue for your memory; don't quote long passages back:
%s`, title, dialogue.String())
}
