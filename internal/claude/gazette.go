package claude

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
)

const gazetteSystemPrompt = `You write 邸报, the Tang court gazette, a short newspaper-style report on the pilgrim monk 唐僧 and his disciples travelling west for the scriptures, for a learner of Mandarin who reads simplified characters at about HSK4-5.

Write one issue from the facts you are given about the past week: about 200 to 350 characters of simplified Chinese in the plain, slightly formal voice of a court bulletin (本报讯, 据悉, 一行). Report only what the facts say; you may add a little colour, but invent no battles, places or deeds. Use HSK4-5 words and grammar, with short paragraphs. If the week was quiet, say so briefly and look ahead. Keep the learner's diary words as a quotation if one is given.

Answer in exactly this format and nothing else:
TEXT: the report in Chinese, which may run over several lines
ENGLISH: its English, line for line`

// WriteGazette asks for the week's issue of the court gazette from facts
// about the journey, returning its Chinese and English.
func (c *Client) WriteGazette(ctx context.Context, facts string) (string, string, error) {
	msg, err := c.send(ctx, anthropic.MessageNewParams{
		Model:     chatModel,
		MaxTokens: 2048, // room to think before answering; small limits return empty text
		System:    []anthropic.TextBlockParam{{Text: gazetteSystemPrompt}},
		Messages:  []anthropic.MessageParam{anthropic.NewUserMessage(anthropic.NewTextBlock(facts))},
	})
	if err != nil {
		return "", "", fmt.Errorf("claude gazette: %w", err)
	}
	return ParseGazette(extractText(msg))
}

// ParseGazette reads the TEXT and ENGLISH of an issue, each of which may
// run over several lines.
func ParseGazette(reply string) (zh, en string, err error) {
	var text, english []string
	field := ""
	for _, line := range strings.Split(reply, "\n") {
		line = strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(line, "TEXT:"):
			field, line = "text", strings.TrimSpace(strings.TrimPrefix(line, "TEXT:"))
		case strings.HasPrefix(line, "ENGLISH:"):
			field, line = "english", strings.TrimSpace(strings.TrimPrefix(line, "ENGLISH:"))
		}
		switch {
		case line == "":
		case field == "text":
			text = append(text, line)
		case field == "english":
			english = append(english, line)
		}
	}
	if len(text) == 0 {
		return "", "", errors.New("the gazette came out blank")
	}
	return strings.Join(text, "\n"), strings.Join(english, "\n"), nil
}
