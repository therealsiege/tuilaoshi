package claude

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
)

// StoryEnding is how every chapter of a Chinese storyteller's tale ends.
const StoryEnding = "欲知后事如何，且听下回分解。"

const storytellerSystemPrompt = `You are a 说书人, a storyteller in a Chinese teahouse, retelling 西游记 (Journey to the West) one chapter (回) a day to a learner of Mandarin who reads simplified characters at about HSK3.

Retell the events of the chapter you are given in your own words: never quote the novel's text or any translation. Write about 300 to 450 characters of simplified Chinese in short, plain sentences, the lively voice of a teahouse storyteller. Use the learner's known words wherever you can, so that about 95 to 98 percent of the words are ones they know; names of people and places from the novel are fine. End the chapter with exactly: 欲知后事如何，且听下回分解。

Answer in exactly this format and nothing else:
TITLE: a one-line chapter title in Chinese
TEXT: the retelling, which may run over several lines
ENGLISH: a one or two sentence summary in English`

// Chapter is one chapter as the storyteller tells it.
type Chapter struct {
	Title, Text, English string
}

// TellChapter asks the storyteller for chapter n of the novel, retold
// mostly in the words the learner knows.
func (c *Client) TellChapter(ctx context.Context, n int, known []string) (Chapter, error) {
	prompt := fmt.Sprintf("Chapter: 第%d回 of 100.\nWords the learner knows: %s", n, strings.Join(known, " "))
	msg, err := c.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     chatModel,
		MaxTokens: 4096, // room to think before a long answer; small limits return empty text
		System:    []anthropic.TextBlockParam{{Text: storytellerSystemPrompt}},
		Messages:  []anthropic.MessageParam{anthropic.NewUserMessage(anthropic.NewTextBlock(prompt))},
	})
	if err != nil {
		return Chapter{}, fmt.Errorf("claude storyteller: %w", err)
	}
	return ParseChapter(extractText(msg))
}

// ParseChapter reads the TITLE, TEXT and ENGLISH of a told chapter. TEXT
// may run over several lines; a chapter without the storyteller's closing
// words gets them, so every chapter ends the same way.
func ParseChapter(reply string) (Chapter, error) {
	var ch Chapter
	var text []string
	field := ""
	for _, line := range strings.Split(reply, "\n") {
		line = strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(line, "TITLE:"):
			field, ch.Title = "title", strings.TrimSpace(strings.TrimPrefix(line, "TITLE:"))
		case strings.HasPrefix(line, "TEXT:"):
			field = "text"
			if v := strings.TrimSpace(strings.TrimPrefix(line, "TEXT:")); v != "" {
				text = append(text, v)
			}
		case strings.HasPrefix(line, "ENGLISH:"):
			field, ch.English = "english", strings.TrimSpace(strings.TrimPrefix(line, "ENGLISH:"))
		case line == "":
		case field == "text":
			text = append(text, line)
		case field == "english":
			ch.English += " " + line
		}
	}
	ch.Text = strings.Join(text, "\n")
	if ch.Text == "" {
		return Chapter{}, errors.New("the storyteller said nothing")
	}
	if !strings.HasSuffix(strings.TrimRight(ch.Text, "。！ "), "且听下回分解") {
		ch.Text += "\n" + StoryEnding
	}
	return ch, nil
}
