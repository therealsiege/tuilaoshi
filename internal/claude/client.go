package claude

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/therealsiege/tuilaoshi/internal/model"
)

// Use SDK constants for model names.
var (
	drillModel = anthropic.ModelClaudeHaiku4_5
	chatModel  = anthropic.ModelClaudeSonnet5_5
)

const tutorSystemPrompt = `You are a Mandarin Chinese tutor helping a learner reactivate their university-level Mandarin.

Learner profile:
- 4 years university Mandarin (reading, writing, speaking, listening)
- Simplified characters (简体字)
- Estimated peak: HSK4-5 range
- Current state: rusty — recognition intact, active recall degraded

Your responses are spoken aloud via text-to-speech. The learner hears your Chinese text through a native Chinese voice. You CAN speak — your text output IS your voice.

Guidelines:
- Respond in a mix of Chinese and English appropriate to the learner's level
- Always include pinyin for new or complex vocabulary in parentheses after the Chinese
- Be encouraging but accurate — correct mistakes clearly
- Use context-rich examples, not isolated words
- Progressively increase Chinese-only interaction as proficiency returns
- Since your responses are spoken aloud, keep Chinese portions natural and conversational`

const drillSystemPrompt = `You are evaluating a Mandarin flashcard drill answer. Be concise.

Respond in this exact format:
CORRECT: true/false
SCORE: 0-5 (0=no answer, 1=wrong, 2=partially correct, 3=correct with issues, 4=good, 5=perfect)
FEEDBACK: Brief feedback in Chinese with English translation. If wrong, give the correct answer.`

// Client wraps the Anthropic API for drill evaluation and chat.
type Client struct {
	client anthropic.Client
	server bool // through the game's server (see server.go), not an API key

	// Meter, when set, is asked before each call and told what each call
	// used (see usage.go).
	Meter Meter
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
// With no API key the answer is checked locally (see LocalGrade).
func (c *Client) EvaluateDrill(ctx context.Context, card model.Card, userAnswer string, mode model.DrillMode) (model.DrillResult, error) {
	if !c.Ready() {
		return LocalGrade(card, userAnswer, mode), nil
	}
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

	msg, err := c.send(ctx, anthropic.MessageNewParams{
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
	return c.ChatAs(ctx, tutorSystemPrompt, history, userMessage)
}

// ChatAs is Chat with another system prompt, such as a companion's voice.
func (c *Client) ChatAs(ctx context.Context, system string, history []ChatMessage, userMessage string) (string, error) {
	var messages []anthropic.MessageParam
	for _, h := range history {
		if h.Role == "user" {
			messages = append(messages, anthropic.NewUserMessage(anthropic.NewTextBlock(h.Content)))
		} else {
			messages = append(messages, anthropic.NewAssistantMessage(anthropic.NewTextBlock(h.Content)))
		}
	}
	messages = append(messages, anthropic.NewUserMessage(anthropic.NewTextBlock(userMessage)))

	msg, err := c.send(ctx, anthropic.MessageNewParams{
		Model:     chatModel,
		MaxTokens: 1024,
		System: []anthropic.TextBlockParam{
			{Text: system},
		},
		Messages: messages,
	})
	if err != nil {
		return "", fmt.Errorf("claude chat: %w", err)
	}

	return extractText(msg), nil
}

const writingSystemPrompt = `You grade a short piece of Chinese writing by a learner reactivating rusty university Mandarin (around HSK3-4). Be fair and encouraging, not picky about style.

Respond in this exact format:
SCORE: 0-2 (2 = does the task in understandable Chinese; 1 = mostly does it, or has errors that obscure meaning; 0 = doesn't do the task, or isn't Chinese)
FEEDBACK: one or two short sentences in English, with a corrected or more natural version in Chinese if there were mistakes.`

// GradeWriting grades a learner's Chinese for a task against criteria,
// returning a 0-2 score and feedback.
func (c *Client) GradeWriting(ctx context.Context, task, criteria, answer string) (int, string, error) {
	prompt := fmt.Sprintf("Task given to the learner: %s\nWhat the answer must do: %s\nLearner's answer: %s", task, criteria, answer)
	msg, err := c.send(ctx, anthropic.MessageNewParams{
		Model:     drillModel,
		MaxTokens: 300,
		System:    []anthropic.TextBlockParam{{Text: writingSystemPrompt}},
		Messages:  []anthropic.MessageParam{anthropic.NewUserMessage(anthropic.NewTextBlock(prompt))},
	})
	if err != nil {
		return 0, "", fmt.Errorf("claude writing grade: %w", err)
	}
	score, feedback := ParseWritingGrade(extractText(msg))
	return score, feedback, nil
}

// ParseWritingGrade reads the SCORE and FEEDBACK lines of a writing grade.
func ParseWritingGrade(response string) (score int, feedback string) {
	for _, line := range strings.Split(response, "\n") {
		line = strings.TrimSpace(line)
		if v, ok := strings.CutPrefix(line, "SCORE:"); ok {
			if n, err := strconv.Atoi(strings.TrimSpace(v)); err == nil && n >= 0 && n <= 2 {
				score = n
			}
		} else if v, ok := strings.CutPrefix(line, "FEEDBACK:"); ok {
			feedback = strings.TrimSpace(v)
		} else if feedback != "" && line != "" {
			feedback += " " + line
		}
	}
	return score, feedback
}

const decomposeSystemPrompt = `You split simplified Chinese characters into their top-level written components: the two or three parts a learner sees first, each itself a character or radical, which together are exactly the whole character. Use simplified forms only. Examples: 河 氵 可; 想 相 心; 好 女 子; 森 木 木 木; 语 讠 吾; 草 艹 早; 闻 门 耳; 宫 宀 吕; 选 辶 先; 邻 令 阝; 爷 父 卩; 器 吅 犬 吅; 睡 目 垂; 赛 宀 共 贝.
For each character given, answer one line: the character, a space, then its components separated by spaces. A character that is a single component (人, 水, 山) gets just itself. No other text.`

// DecomposeCharacters asks for each character's main written components.
func (c *Client) DecomposeCharacters(ctx context.Context, chars []string) (map[string][]string, error) {
	msg, err := c.send(ctx, anthropic.MessageNewParams{
		Model:     chatModel, // the small model splits characters wrongly too often
		MaxTokens: 2048,      // room to think before answering
		System:    []anthropic.TextBlockParam{{Text: decomposeSystemPrompt}},
		Messages:  []anthropic.MessageParam{anthropic.NewUserMessage(anthropic.NewTextBlock(strings.Join(chars, "\n")))},
	})
	if err != nil {
		return nil, fmt.Errorf("claude decompose: %w", err)
	}
	return ParseComponents(extractText(msg), chars), nil
}

// ParseComponents reads "河 氵 可" lines for the characters asked about.
func ParseComponents(response string, asked []string) map[string][]string {
	want := map[string]bool{}
	for _, c := range asked {
		want[c] = true
	}
	out := map[string][]string{}
	for _, line := range strings.Split(response, "\n") {
		f := strings.Fields(line)
		if len(f) >= 2 && want[f[0]] {
			out[f[0]] = f[1:]
		}
	}
	return out
}

// Explain requests a grammar or usage explanation.
func (c *Client) Explain(ctx context.Context, topic string) (string, error) {
	prompt := fmt.Sprintf(
		"Explain this Mandarin grammar point or usage clearly with examples:\n\n%s",
		topic,
	)

	msg, err := c.send(ctx, anthropic.MessageNewParams{
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
			// An unparseable score stays 0; Correct and Feedback still stand.
			var score int
			if _, err := fmt.Sscanf(val, "%d", &score); err == nil {
				result.Score = score
			}
		}
		if strings.HasPrefix(upper, "FEEDBACK:") {
			result.Feedback = strings.TrimSpace(line[len("FEEDBACK:"):])
		}
	}

	return result
}
