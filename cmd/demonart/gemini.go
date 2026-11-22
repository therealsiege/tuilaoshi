package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/therealsiege/tuilaoshi/internal/journey"
)

// client calls the Gemini API's generateContent for images.
type client struct {
	key, model, base string
	http             *http.Client
}

func newClient(key, model string) *client {
	return &client{key: key, model: model, base: "https://generativelanguage.googleapis.com/v1beta",
		http: &http.Client{Timeout: 2 * time.Minute}}
}

type genRequest struct {
	Contents         []genContent `json:"contents"`
	GenerationConfig genConfig    `json:"generationConfig"`
}

type genContent struct {
	Parts []genPart `json:"parts"`
}

type genPart struct {
	Text       string      `json:"text,omitempty"`
	InlineData *inlineData `json:"inlineData,omitempty"`
}

type inlineData struct {
	MimeType string `json:"mimeType"`
	Data     string `json:"data"`
}

type genConfig struct {
	ResponseModalities []string    `json:"responseModalities"`
	ImageConfig        imageConfig `json:"imageConfig"`
}

type imageConfig struct {
	AspectRatio string `json:"aspectRatio"`
}

type genResponse struct {
	Candidates []struct {
		Content genContent `json:"content"`
	} `json:"candidates"`
}

// paint asks the model for one square image and returns its bytes.
func (c *client) paint(prompt string) (_ []byte, err error) {
	body, err := json.Marshal(genRequest{
		Contents:         []genContent{{Parts: []genPart{{Text: prompt}}}},
		GenerationConfig: genConfig{ResponseModalities: []string{"IMAGE"}, ImageConfig: imageConfig{AspectRatio: "1:1"}},
	})
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/models/%s:generateContent", c.base, c.model), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-goog-api-key", c.key)
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("gemini: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); err == nil && closeErr != nil {
			err = fmt.Errorf("gemini: close response: %w", closeErr)
		}
	}()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("gemini: read response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("gemini: %s: %s", resp.Status, strings.TrimSpace(string(data)))
	}
	var gr genResponse
	if err := json.Unmarshal(data, &gr); err != nil {
		return nil, fmt.Errorf("gemini: parse response: %w", err)
	}
	for _, cand := range gr.Candidates {
		for _, p := range cand.Content.Parts {
			if p.InlineData != nil && strings.HasPrefix(p.InlineData.MimeType, "image/") {
				return base64.StdEncoding.DecodeString(p.InlineData.Data)
			}
		}
	}
	return nil, fmt.Errorf("gemini: no image in the response")
}

// promptFor describes a demon for the painter: who it is and what it can
// do, in an original style that stays readable at terminal size.
func promptFor(d journey.BestiaryEntry) string {
	name := fmt.Sprintf("%s (%s)", d.English, d.Name)
	if d.TitleEng != "" {
		name += ", " + d.TitleEng
	}
	subject := fmt.Sprintf("%s, a demon from the classic Chinese novel Journey to the West. %s. Powers: %s.",
		name, strings.TrimSuffix(d.Origin, "."), d.Powers)
	if d.Name == "如来佛祖" {
		subject = "the Buddha (如来佛祖) as he appears in the classic Chinese novel Journey to the West: serene, majestic and radiant, shown with reverence, seated in golden light."
	}
	return "A single original character portrait of " + subject +
		" Style: vivid traditional Chinese ink-and-watercolor painting with bold outlines and strong colors. " +
		"One full figure, centered, filling most of the frame, high contrast so it reads at small sizes. " +
		"The background is a plain, uniformly dark deep indigo or black that fills the whole square canvas to every edge: no paper, no light margins, no parchment. " +
		"No text, no letters, no seals, no borders, no watermark."
}
