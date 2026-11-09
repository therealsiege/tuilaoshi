package claude

import (
	"context"
	"fmt"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/cj/mandarin-wiki/internal/model"
)

// Use SDK constants for model names.
var (
	drillModel = anthropic.ModelClaudeHaiku4_5
	chatModel  = anthropic.ModelClaudeSonnet4_5
)

const tutorSystemPrompt = `You are a Mandarin Chinese tutor helping a learner reactivate their university-level Mandarin.

Learner profile:
- 4 years university Mandarin (reading, writing, speaking, listening)
- Simplified characters (简体字)
- Estimated peak: HSK4-5 range
- Current state: rusty — recognition intact, active recall degraded

Guidelines:
- Respond in a mix of Chinese and English appropriate to the learner's level
- Always include pinyin for new or complex vocabulary
- Be encouraging but accurate — correct mistakes clearly
- Use context-rich examples, not isolated words
- Progressively increase Chinese-only interaction as proficiency returns`

const drillSystemPrompt = `You are evaluating a Mandarin flashcard drill answer. Be concise.

Respond in this exact format:
CORRECT: true/false
SCORE: 0-5 (0=no answer, 1=wrong, 2=partially correct, 3=correct with issues, 4=good, 5=perfect)
FEEDBACK: Brief feedback in Chinese with English translation. If wrong, give the correct answer.`

// Client wraps the Anthropic API for drill evaluation and chat.
type Client struct {
	client anthropic.Client
}

// New creates a Claude API client. Reads ANTHROPIC_API_KEY from environment.
func New() *Client {
	return &Client{
		client: anthropic.NewClient(),
	}
}

// ChatMessage represents a single message in a conversation.
type ChatMessage struct {
	Role    string // "user" or "assistant"
	Content string
}

// EvaluateDrill sends a card + user answer to Claude for evaluation.
func (c *Client) EvaluateDrill(ctx context.Context, card model.Card, userAnswer string, mode model.DrillMode) (model.DrillResult, error) {
	var modeDesc string
	switch mode {
	case model.ModeRecognition:
		modeDesc = "recognize the meaning of a Chinese word/phrase"
	case model.ModeProduction:
		modeDesc = "produce the Chinese characters for an English meaning"
	case model.ModeListening:
		modeDesc = "write the Chinese characters after hearing them spoken"
	case model.ModeSentence:
		modeDesc = "fill in the missing word in a Chinese sentence"
	default:
		modeDesc = "recall the meaning"
	}

	prompt := fmt.Sprintf(
		"Drill mode: The student was asked to %s.\nCard type: %s\nChinese: %s\nPinyin: %s\nExpected meaning: %s\nUser's answer: %s",
		modeDesc, card.Type, card.Chinese, card.Pinyin, card.English, userAnswer,
	)

	msg, err := c.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     drillModel,
		MaxTokens: 300,
		System: []anthropic.TextBlockParam{
			{Text: drillSystemPrompt},
		},
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(prompt)),
		},
	})
	if err != nil {
		return model.DrillResult{}, fmt.Errorf("claude drill eval: %w", err)
	}

	response := extractText(msg)
	return parseDrillResponse(response), nil
}

// Chat sends a message in a conversational Mandarin practice session.
func (c *Client) Chat(ctx context.Context, history []ChatMessage, userMessage string) (string, error) {
	var messages []anthropic.MessageParam
	for _, h := range history {
		if h.Role == "user" {
			messages = append(messages, anthropic.NewUserMessage(anthropic.NewTextBlock(h.Content)))
		} else {
			messages = append(messages, anthropic.NewAssistantMessage(anthropic.NewTextBlock(h.Content)))
		}
	}
	messages = append(messages, anthropic.NewUserMessage(anthropic.NewTextBlock(userMessage)))

	msg, err := c.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     chatModel,
		MaxTokens: 1024,
		System: []anthropic.TextBlockParam{
			{Text: tutorSystemPrompt},
		},
		Messages: messages,
	})
	if err != nil {
		return "", fmt.Errorf("claude chat: %w", err)
	}

	return extractText(msg), nil
}

// Explain requests a grammar or usage explanation.
func (c *Client) Explain(ctx context.Context, topic string) (string, error) {
	prompt := fmt.Sprintf(
		"Explain this Mandarin grammar point or usage clearly with examples:\n\n%s",
		topic,
	)

	msg, err := c.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     chatModel,
		MaxTokens: 1024,
		System: []anthropic.TextBlockParam{
			{Text: tutorSystemPrompt},
		},
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(prompt)),
		},
	})
	if err != nil {
		return "", fmt.Errorf("claude explain: %w", err)
	}

	return extractText(msg), nil
}

func extractText(msg *anthropic.Message) string {
	var parts []string
	for _, block := range msg.Content {
		if block.Type == "text" {
			parts = append(parts, block.Text)
		}
	}
	return strings.Join(parts, "\n")
}

func parseDrillResponse(response string) model.DrillResult {
	result := model.DrillResult{Feedback: response}

	for _, line := range strings.Split(response, "\n") {
		line = strings.TrimSpace(line)
		upper := strings.ToUpper(line)
		if strings.HasPrefix(upper, "CORRECT:") {
			val := strings.TrimSpace(strings.TrimPrefix(upper, "CORRECT:"))
			result.Correct = val == "TRUE"
		}
		if strings.HasPrefix(upper, "SCORE:") {
			val := strings.TrimSpace(strings.TrimPrefix(upper, "SCORE:"))
			fmt.Sscanf(val, "%d", &result.Score)
		}
		if strings.HasPrefix(upper, "FEEDBACK:") {
			result.Feedback = strings.TrimSpace(line[len("FEEDBACK:"):])
		}
	}

	return result
}
