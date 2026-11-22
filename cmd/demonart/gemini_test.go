package main

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/therealsiege/tuilaoshi/internal/journey"
)

func TestPaintSendsPromptAndReturnsImage(t *testing.T) {
	var got genRequest
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("x-goog-api-key") != "k" || !strings.HasSuffix(r.URL.Path, "/models/m:generateContent") {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Error(err)
		}
		if _, err := w.Write([]byte(`{"candidates":[{"content":{"parts":[{"text":"here"},{"inlineData":{"mimeType":"image/png","data":"` +
			base64.StdEncoding.EncodeToString([]byte("PNG!")) + `"}}]}}]}`)); err != nil {
			t.Error(err)
		}
	}))
	defer srv.Close()

	c := newClient("k", "m")
	c.base = srv.URL
	img, err := c.paint("a demon")
	if err != nil || string(img) != "PNG!" {
		t.Fatalf("paint = %q, %v", img, err)
	}
	if got.Contents[0].Parts[0].Text != "a demon" || got.GenerationConfig.ResponseModalities[0] != "IMAGE" {
		t.Errorf("request %+v", got)
	}
}

func TestPaintReportsErrors(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/models/blocked") {
			if _, err := w.Write([]byte(`{"candidates":[{"content":{"parts":[{"text":"no image for you"}]}}]}`)); err != nil {
				t.Error(err)
			}
			return
		}
		http.Error(w, `{"error":"quota"}`, http.StatusTooManyRequests)
	}))
	defer srv.Close()
	for model, want := range map[string]string{"m": "429", "blocked": "no image"} {
		c := newClient("k", model)
		c.base = srv.URL
		if _, err := c.paint("x"); err == nil || !strings.Contains(err.Error(), want) {
			t.Errorf("model %s: err %v, want %q", model, err, want)
		}
	}
}

func TestPromptsDescribeEveryDemon(t *testing.T) {
	for _, d := range journey.Bestiary() {
		p := promptFor(d)
		if !strings.Contains(p, d.Name) || !strings.Contains(p, "No text") {
			t.Errorf("prompt for %s: %q", d.Name, p)
		}
	}
}
