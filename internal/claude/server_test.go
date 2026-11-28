package claude

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/therealsiege/tuilaoshi/internal/claude/claudetest"
	"github.com/therealsiege/tuilaoshi/internal/model"
)

type countingMeter struct {
	calls         int
	input, output int64
}

func (m *countingMeter) Allow() bool { return true }
func (m *countingMeter) Used(_ string, in, out int64) {
	m.calls++
	m.input, m.output = m.input+in, m.output+out
}

// A drill graded through the game's server: the SDK's request reaches
// /api/claude with the session token, no key and no model, and the
// server's stream comes back as the reply.
func TestEvaluateDrillThroughTheServer(t *testing.T) {
	t.Setenv("ANTHROPIC_API_KEY", "sk-must-not-leave-this-machine")
	srv := claudetest.New(t, "CORRECT: true\nSCORE: 4\nFEEDBACK: 对！Right.")
	c := NewServer(srv.URL, "session-token", nil)
	meter := &countingMeter{}
	c.Meter = meter
	if !c.Ready() {
		t.Fatal("a server client should be ready without a key")
	}
	res, err := c.EvaluateDrill(t.Context(), model.Card{Chinese: "书", Pinyin: "shū", English: "book"}, "a book", model.ModeRecognition)
	if err != nil {
		t.Fatal(err)
	}
	if !res.Correct || res.Score != 4 || res.Feedback != "对！Right." {
		t.Errorf("graded %+v", res)
	}
	reqs := srv.Requests()
	if len(reqs) != 1 {
		t.Fatalf("%d requests", len(reqs))
	}
	r := reqs[0]
	if r.Path != "/api/claude" || r.Auth != "Bearer session-token" || r.APIKey != "" {
		t.Errorf("sent %s auth %q key %q", r.Path, r.Auth, r.APIKey)
	}
	for k := range r.Body {
		if k != "system" && k != "messages" && k != "max_tokens" {
			t.Errorf("the body carries %q, which the server refuses", k)
		}
	}
	if sys, _ := r.Body["system"].(string); !strings.Contains(sys, "flashcard drill") {
		t.Errorf("system %q", sys)
	}
	if r.Body["max_tokens"] != float64(300) {
		t.Errorf("max_tokens %v", r.Body["max_tokens"])
	}
	msgs, _ := r.Body["messages"].([]any)
	if len(msgs) != 1 || !strings.Contains(fmt.Sprint(msgs[0]), "a book") {
		t.Errorf("messages %v", msgs)
	}
	if meter.calls != 1 || meter.input != 12 || meter.output != 34 {
		t.Errorf("metered %+v", meter)
	}
}

// Every method works through the server unchanged: a chat history, a
// chapter and a gazette.
func TestOtherCallsThroughTheServer(t *testing.T) {
	srv := claudetest.New(t, "好的。",
		"TITLE: 石猴出世\nTEXT: 花果山上有一块石头。\n石头里出来一只猴子。\nENGLISH: A monkey is born from a stone.\nQ: 猴子从哪里来？ | 石头 | 水 | 天",
		"TEXT: 本报讯：一行继续西行。\nENGLISH: Our report: the party went on west.")
	c := NewServer(srv.URL+"/", "tok", nil)
	reply, err := c.Chat(t.Context(), []ChatMessage{{"user", "你好"}, {"assistant", "你好！"}}, "我们练习吧")
	if err != nil || reply != "好的。" {
		t.Fatalf("chat %q %v", reply, err)
	}
	if msgs, _ := srv.Requests()[0].Body["messages"].([]any); len(msgs) != 3 {
		t.Errorf("history sent as %d messages", len(msgs))
	}
	ch, err := c.TellChapter(t.Context(), 1, []string{"石头"})
	if err != nil || ch.Title != "石猴出世" || !strings.HasSuffix(ch.Text, StoryEnding) || len(ch.Questions) != 1 {
		t.Fatalf("chapter %+v %v", ch, err)
	}
	zh, en, err := c.WriteGazette(t.Context(), "facts")
	if err != nil || zh != "本报讯：一行继续西行。" || en == "" {
		t.Fatalf("gazette %q %q %v", zh, en, err)
	}
	if p := srv.Requests()[2].Path; p != "/api/claude" {
		t.Errorf("a base with a slash posted to %s", p)
	}
}

func TestTheServersRefusalsSayWhatToDo(t *testing.T) {
	srv := claudetest.New(t, "x")
	c := NewServer(srv.URL, "tok", nil)
	srv.Refuse(http.StatusTooManyRequests)
	_, err := c.Explain(t.Context(), "了")
	if !errors.Is(err, ErrServerCap) || !strings.Contains(err.Error(), "refused with 429") {
		t.Errorf("over the cap: %v", err)
	}
	if n := len(srv.Requests()); n != 1 {
		t.Errorf("a refused call was tried %d times", n)
	}
	srv.Refuse(http.StatusUnauthorized)
	if _, err := c.Explain(t.Context(), "了"); !errors.Is(err, ErrSignedOut) {
		t.Errorf("signed out: %v", err)
	}
	srv.Refuse(http.StatusBadGateway)
	if _, err := c.Explain(t.Context(), "了"); err == nil || !strings.Contains(err.Error(), "refused with 502") {
		t.Errorf("server error: %v", err)
	}
}

// A failure after the stream began is an error event; it is an error, not
// a short reply.
func TestAnErrorMidReplyIsAnError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream; charset=utf-8")
		for _, e := range claudetest.Events("half")[:3] {
			_, _ = fmt.Fprintf(w, "event: %s\ndata: %s\n\n", e[0], e[1])
		}
		_, _ = fmt.Fprint(w, "event: error\ndata: {\"type\":\"error\",\"error\":{\"type\":\"api_error\",\"message\":\"Claude is overloaded\"}}\n\n")
	}))
	defer srv.Close()
	_, err := NewServer(srv.URL, "tok", nil).Explain(t.Context(), "了")
	if err == nil || !strings.Contains(err.Error(), "Claude is overloaded") {
		t.Errorf("got %v", err)
	}
}
