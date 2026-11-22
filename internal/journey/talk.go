package journey

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/therealsiege/tuilaoshi/internal/model"
)

// TalkOpener is the hidden first message that asks a companion to begin.
const TalkOpener = "（请你先开口。）"

// TalkPrompt is the system prompt for talking with a companion: their own
// voice, the learner's level, the words to work into the talk, and the
// grade line that turns real Chinese into bond points.
func TalkPrompt(c Companion, b Bond, words []model.Card) string {
	var list []string
	for _, w := range words {
		list = append(list, fmt.Sprintf("%s (%s, %s)", w.Chinese, w.Pinyin, w.English))
	}
	return fmt.Sprintf(`You are %s (%s) from Journey to the West, travelling west with the learner, a fellow pilgrim. Speak as this character: %s

The learner is reactivating rusty university Mandarin (around HSK3-4, simplified characters). Keep each reply to two or three short sentences of natural Chinese at that level, in character, about life on the road west, and end with a question the learner can answer. Put pinyin in parentheses after any word above HSK4. No English except a short gloss in parentheses when a word is hard.

Work these words the learner is studying into the conversation, and invite the learner to use them: %s.

After your reply, on a last line by itself, grade the learner's most recent message (not your own) as "BOND: n":
2 = understandable Chinese that uses one of the study words correctly;
1 = understandable Chinese;
0 = English, nonsense, or the conversation's opening.
If the learner made a mistake, correct it gently in character before you go on.`, c.Name, c.English, b.Voice, strings.Join(list, ", "))
}

var gradeLine = regexp.MustCompile(`(?m)^\s*BOND:\s*([0-2])\s*$`)

// ParseTalkReply splits a companion's reply into what they say and the
// grade of the learner's message (0 when missing).
func ParseTalkReply(reply string) (text string, grade int) {
	m := gradeLine.FindStringSubmatchIndex(reply)
	if m == nil {
		return strings.TrimSpace(reply), 0
	}
	grade, _ = strconv.Atoi(reply[m[2]:m[3]]) //nolint:errcheck // the pattern admits only 0-2
	return strings.TrimSpace(reply[:m[0]] + reply[m[1]:]), grade
}
