package imagen

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"
)

const apiBase = "https://generativelanguage.googleapis.com/v1beta"

// Result holds the output of a single image generation.
type Result struct {
	ImagePath string
	MimeType  string
	Text      string // model commentary
	Prompt    string
	CreatedAt time.Time
}

// Client is a Google Gemini image generation client.
type Client struct {
	apiKey     string
	httpClient *http.Client
	model      string
	outputDir  string
}

// New creates an imagen client. Reads GOOGLE_API_KEY from env.
func New(outputDir string) *Client {
	return &Client{
		apiKey:     os.Getenv("GOOGLE_API_KEY"),
		httpClient: &http.Client{Timeout: 120 * time.Second},
		model:      "gemini-2.5-flash-image",
		outputDir:  outputDir,
	}
}

// Available reports whether the client has an API key configured.
func (c *Client) Available() bool {
	return c.apiKey != ""
}

// Model returns the current model name.
func (c *Client) Model() string { return c.model }

// Generate creates an image from a text prompt.
func (c *Client) Generate(ctx context.Context, prompt string) (*Result, error) {
	if !c.Available() {
		return nil, fmt.Errorf("GOOGLE_API_KEY not set")
	}

	// Ensure output directory exists.
	if err := os.MkdirAll(c.outputDir, 0o755); err != nil {
		return nil, fmt.Errorf("create output dir: %w", err)
	}

	// Build request body.
	reqBody := geminiRequest{
		Contents: []content{{
			Parts: []requestPart{{Text: prompt}},
		}},
		GenerationConfig: genConfig{
			ResponseModalities: []string{"IMAGE", "TEXT"},
		},
	}
	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/models/%s:generateContent", apiBase, c.model)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-goog-api-key", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("api call: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	var gemResp geminiResponse
	if err := json.Unmarshal(respBody, &gemResp); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	if gemResp.Error != nil {
		return nil, fmt.Errorf("api error %d: %s", gemResp.Error.Code, gemResp.Error.Message)
	}

	// Extract image data and text from response.
	var imageData []byte
	var mimeType, text string
	for _, cand := range gemResp.Candidates {
		for _, p := range cand.Content.Parts {
			if p.InlineData != nil && p.InlineData.Data != "" {
				decoded, err := base64.StdEncoding.DecodeString(p.InlineData.Data)
				if err != nil {
					return nil, fmt.Errorf("decode base64: %w", err)
				}
				imageData = decoded
				mimeType = p.InlineData.MimeType
			}
			if p.Text != "" {
				text = strings.TrimSpace(p.Text)
			}
		}
	}

	if imageData == nil {
		return nil, fmt.Errorf("no image data in response")
	}

	// Determine file extension.
	ext := ".png"
	switch {
	case strings.Contains(mimeType, "jpeg"), strings.Contains(mimeType, "jpg"):
		ext = ".jpg"
	case strings.Contains(mimeType, "webp"):
		ext = ".webp"
	}

	// Write to file.
	slug := slugify(prompt)
	timestamp := time.Now().Format("20060102-150405")
	filename := fmt.Sprintf("%s-%s%s", slug, timestamp, ext)
	outPath := filepath.Join(c.outputDir, filename)

	if err := os.WriteFile(outPath, imageData, 0o644); err != nil {
		return nil, fmt.Errorf("write image: %w", err)
	}

	return &Result{
		ImagePath: outPath,
		MimeType:  mimeType,
		Text:      text,
		Prompt:    prompt,
		CreatedAt: time.Now(),
	}, nil
}

// OpenImage opens an image file with the system viewer.
func OpenImage(path string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", path)
	case "linux":
		cmd = exec.Command("xdg-open", path)
	default:
		return fmt.Errorf("image viewer not supported on %s", runtime.GOOS)
	}
	return cmd.Start()
}

// -- Request types --

type geminiRequest struct {
	Contents         []content `json:"contents"`
	GenerationConfig genConfig `json:"generationConfig"`
}

type content struct {
	Parts []requestPart `json:"parts"`
}

type requestPart struct {
	Text string `json:"text"`
}

type genConfig struct {
	ResponseModalities []string `json:"responseModalities"`
}

// -- Response types --

type geminiResponse struct {
	Candidates []candidate `json:"candidates"`
	Error      *apiError   `json:"error,omitempty"`
}

type candidate struct {
	Content responseContent `json:"content"`
}

type responseContent struct {
	Parts []responsePart `json:"parts"`
}

type responsePart struct {
	Text       string      `json:"text,omitempty"`
	InlineData *inlineData `json:"inlineData,omitempty"`
}

type inlineData struct {
	MimeType string `json:"mimeType"`
	Data     string `json:"data"`
}

type apiError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// slugify converts a prompt to a short filename-safe slug.
var nonAlnum = regexp.MustCompile(`[^a-z0-9]+`)

func slugify(s string) string {
	s = strings.ToLower(s)
	s = nonAlnum.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	if len(s) > 40 {
		s = s[:40]
	}
	if s == "" {
		s = "image"
	}
	return s
}
