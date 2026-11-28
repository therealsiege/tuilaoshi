// Package claudetest fakes the game's server's /api/claude for tests: it
// streams Claude's own events for each reply, as the server does, and
// keeps the requests it was sent.
package claudetest

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

// Request is a call the fake server took.
type Request struct {
	Path, Auth, APIKey string
	Body               map[string]any
}

// Server is a fake game server. Replies are given in turn; once they run
// out the last is repeated.
type Server struct {
	*httptest.Server

	mu       sync.Mutex
	status   int
	replies  []string
	requests []Request
}

// New starts a fake server answering with the given texts, closed when the
// test ends.
func New(t testing.TB, replies ...string) *Server {
	s := &Server{replies: replies}
	s.Server = httptest.NewServer(http.HandlerFunc(s.serve))
	t.Cleanup(s.Close)
	return s
}

// Refuse makes every later call fail with status and a JSON {error}, as
// the server's own refusals are; 0 answers again.
func (s *Server) Refuse(status int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.status = status
}

// Requests returns the calls taken so far.
func (s *Server) Requests() []Request {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]Request(nil), s.requests...)
}

func (s *Server) serve(w http.ResponseWriter, r *http.Request) {
	req := Request{Path: r.URL.Path, Auth: r.Header.Get("Authorization"), APIKey: r.Header.Get("X-Api-Key")}
	data, err := io.ReadAll(r.Body)
	if err == nil {
		err = json.Unmarshal(data, &req.Body)
	}
	if err != nil {
		http.Error(w, `{"error":"body must be a JSON object"}`, http.StatusBadRequest)
		return
	}
	s.mu.Lock()
	n := len(s.requests)
	s.requests = append(s.requests, req)
	status := s.status
	text := ""
	if len(s.replies) > 0 {
		text = s.replies[min(n, len(s.replies)-1)]
	}
	s.mu.Unlock()
	if status != 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		fmt.Fprintf(w, `{"error":"refused with %d"}`, status) //nolint:errcheck // a test server's reply
		return
	}
	w.Header().Set("Content-Type", "text/event-stream; charset=utf-8")
	for _, ev := range Events(text) {
		fmt.Fprintf(w, "event: %s\ndata: %s\n\n", ev[0], ev[1]) //nolint:errcheck // a test server's reply
	}
}

// Events are the stream events of a reply of text, as [type, json] pairs,
// the way the server passes Claude's own along.
func Events(text string) [][2]string {
	delta, _ := json.Marshal(text) //nolint:errcheck // a string always marshals
	return [][2]string{
		{"message_start", `{"type":"message_start","message":{"id":"msg_fake","type":"message","role":"assistant","model":"claude-fake","content":[],"stop_reason":null,"usage":{"input_tokens":12,"output_tokens":1}}}`},
		{"content_block_start", `{"type":"content_block_start","index":0,"content_block":{"type":"text","text":""}}`},
		{"ping", `{"type":"ping"}`},
		{"content_block_delta", `{"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":` + string(delta) + `}}`},
		{"content_block_stop", `{"type":"content_block_stop","index":0}`},
		{"message_delta", `{"type":"message_delta","delta":{"stop_reason":"end_turn"},"usage":{"output_tokens":34}}`},
		{"message_stop", `{"type":"message_stop"}`},
	}
}
