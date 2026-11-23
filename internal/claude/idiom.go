package claude

import (
	"context"
	"fmt"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
)

// Idiom is a 成语 with its reading and meaning.
type Idiom struct {
	Text, Pinyin, Meaning string
}

// IdiomTurn is the judge's answer to one move of an idiom chain (成语接龙):
// whether the learner's idiom stands, why, and the companion's reply.
type IdiomTurn struct {
	Valid  bool
	Note   string // short English: why it stands or not, with a fix
	Played Idiom  // the learner's idiom, read and glossed
	Reply  Idiom  // the companion's next link; empty when stumped
}

const idiomSystemPrompt = `You referee and play 成语接龙 (idiom chain) with a learner of Mandarin (around HSK3-4), as the companion %s. Each idiom must be a real, standard four-character 成语 whose first character is the last character of the previous idiom, or a homophone of it (same syllable, tones may differ). Prefer common idioms a learner can learn from.

Answer in exactly these lines, nothing else:
VALID: yes or no (whether the learner's idiom is a real 成语 and chains correctly; "yes" for the opening turn)
NOTE: one short English sentence: why, and for a mistake a correct idiom that would have chained
PLAYED: the learner's idiom | pinyin with tone marks | short English meaning (or "none")
REPLY: your next idiom, chaining from the learner's if valid, else from the previous idiom | pinyin | meaning (or "none" if you can't continue)`

// IdiomMove asks for the companion's move: an opening when chain is empty,
// else the judgement of attempt after chain and the companion's reply. A
// hint asks instead for an idiom the learner could play (as REPLY).
func (c *Client) IdiomMove(ctx context.Context, companion string, chain []string, attempt string, hint bool) (IdiomTurn, error) {
	var prompt string
	switch {
	case len(chain) == 0:
		prompt = "Open the game: VALID yes, PLAYED none, REPLY with a common opening idiom, and a NOTE naming the character (the last of your idiom) the learner's idiom must start with."
	case hint:
		prompt = fmt.Sprintf("The chain so far: %s. The learner asks for a hint: VALID yes, PLAYED none, and REPLY with a common idiom the learner could play next.", strings.Join(chain, " → "))
	default:
		prompt = fmt.Sprintf("The chain so far: %s. The learner plays: %s", strings.Join(chain, " → "), attempt)
	}
	msg, err := c.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     chatModel,
		MaxTokens: 2048, // the model thinks before answering; a small budget ends mid-thought
		System:    []anthropic.TextBlockParam{{Text: fmt.Sprintf(idiomSystemPrompt, companion)}},
		Messages:  []anthropic.MessageParam{anthropic.NewUserMessage(anthropic.NewTextBlock(prompt))},
	})
	if err != nil {
		return IdiomTurn{}, fmt.Errorf("claude idiom: %w", err)
	}
	return ParseIdiomTurn(extractText(msg)), nil
}

// ParseIdiomTurn reads the judge's lines.
func ParseIdiomTurn(response string) IdiomTurn {
	var t IdiomTurn
	for _, line := range strings.Split(response, "\n") {
		key, val, ok := strings.Cut(strings.TrimSpace(line), ":")
		if !ok {
			continue
		}
		val = strings.TrimSpace(val)
		switch strings.ToUpper(strings.TrimSpace(key)) {
		case "VALID":
			t.Valid = strings.HasPrefix(strings.ToLower(val), "y")
		case "NOTE":
			t.Note = val
		case "PLAYED":
			t.Played = parseIdiom(val)
		case "REPLY":
			t.Reply = parseIdiom(val)
		}
	}
	return t
}

func parseIdiom(s string) Idiom {
	if strings.EqualFold(strings.TrimSpace(s), "none") {
		return Idiom{}
	}
	parts := strings.Split(s, "|")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	for len(parts) < 3 {
		parts = append(parts, "")
	}
	return Idiom{Text: parts[0], Pinyin: parts[1], Meaning: parts[2]}
}
