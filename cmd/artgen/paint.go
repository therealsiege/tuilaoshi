package main

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/therealsiege/tuilaoshi/internal/journey"
)

// painter calls Gemini's generateContent for images, with any reference
// pictures attached after the prompt.
type painter struct {
	key, model, size string
	http             *http.Client
}

type part struct {
	Text       string      `json:"text,omitempty"`
	InlineData *inlineData `json:"inlineData,omitempty"`
}

type inlineData struct {
	MimeType string `json:"mimeType"`
	Data     string `json:"data"`
}

func (p painter) paint(prompt, ratio string, refs [][]byte) ([]byte, error) {
	parts := []part{{Text: prompt}}
	for _, r := range refs {
		parts = append(parts, part{InlineData: &inlineData{MimeType: "image/png", Data: base64.StdEncoding.EncodeToString(r)}})
	}
	body, err := json.Marshal(map[string]any{
		"contents": []any{map[string]any{"parts": parts}},
		"generationConfig": map[string]any{
			"responseModalities": []string{"IMAGE"},
			"imageConfig":        map[string]any{"aspectRatio": ratio, "imageSize": p.size},
		},
	})
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", "https://generativelanguage.googleapis.com/v1beta/models/"+p.model+":generateContent", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-goog-api-key", p.key)
	resp, err := p.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close() //nolint:errcheck // read-only body
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s: %.300s", resp.Status, strings.TrimSpace(string(data)))
	}
	var gr struct {
		Candidates []struct {
			Content struct {
				Parts []part `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}
	if err := json.Unmarshal(data, &gr); err != nil {
		return nil, err
	}
	for _, c := range gr.Candidates {
		for _, pt := range c.Content.Parts {
			if pt.InlineData != nil {
				return base64.StdEncoding.DecodeString(pt.InlineData.Data)
			}
		}
	}
	return nil, fmt.Errorf("no image in the reply")
}

func newPainter(key, model, size string) painter {
	return painter{key: key, model: model, size: size, http: &http.Client{Timeout: 5 * time.Minute}}
}

// layoutColours are the colours the map legend in bible.yaml names.
var layoutColours = map[journey.Tile]color.RGBA{
	journey.TileBlock: {70, 80, 72, 255}, journey.TilePath: {214, 190, 140, 255}, journey.TileWild: {120, 170, 90, 255},
	journey.TileTown: {200, 50, 40, 255}, journey.TileStage: {240, 200, 40, 255}, journey.TileEast: {214, 190, 140, 255},
	journey.TileWest: {214, 190, 140, 255}, journey.TileCrossing: {60, 120, 210, 255},
}

// layoutPNG draws a region map as flat colours, cell pixels to a tile.
func layoutPNG(m journey.RegionMap, cell int) ([]byte, error) {
	img := image.NewRGBA(image.Rect(0, 0, m.W*cell, m.H*cell))
	for y := 0; y < m.H*cell; y++ {
		for x := 0; x < m.W*cell; x++ {
			img.Set(x, y, layoutColours[m.Tiles[y/cell][x/cell]])
		}
	}
	var buf bytes.Buffer
	err := png.Encode(&buf, img)
	return buf.Bytes(), err
}

// readKeys reads only the named keys from a KEY=value env file, so the
// rest of a shared file never enters this process.
func readKeys(path string, names ...string) (map[string]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close() //nolint:errcheck // read-only
	want := map[string]bool{}
	for _, n := range names {
		want[n] = true
	}
	out := map[string]string{}
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimPrefix(strings.TrimSpace(sc.Text()), "export ")
		k, v, ok := strings.Cut(line, "=")
		if !ok || !want[k] {
			continue
		}
		out[k] = strings.Trim(strings.TrimSpace(v), `"'`)
	}
	return out, sc.Err()
}
