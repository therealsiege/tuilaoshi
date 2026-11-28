package claude

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/anthropics/anthropic-sdk-go/packages/ssestream"
)

// Through the game's server (web/, POST /api/claude) a signed-in device
// reaches Claude with its session token instead of an API key. The SDK is
// pointed at the server and a middleware turns each Messages call into the
// server's request: the path becomes /api/claude, the body keeps only
// system, messages and max_tokens (the server picks the model), and the
// server's stream of Claude's own events is gathered back into the one
// message the SDK expects. So every prompt in this package works unchanged,
// and every call still goes through send and its meter.

// ErrServerCap is returned when the game's server says today's Claude
// calls are used up (429).
var ErrServerCap = errors.New("today's Claude limit on the game's server is reached")

// ErrSignedOut is returned when the game's server refuses this device's
// session (401).
var ErrSignedOut = errors.New("the game's server didn't accept this device's sign-in; sign in again")

// NewServer creates a client that reaches Claude through the game's server
// at base, signed in with token. A nil httpClient uses the SDK's own.
func NewServer(base, token string, httpClient *http.Client) *Client {
	opts := []option.RequestOption{
		option.WithBaseURL(strings.TrimRight(base, "/") + "/"),
		option.WithAuthToken(token),
		option.WithMiddleware(throughServer),
		// Each try is a call against the day's cap, and a 429 won't pass
		// on a retry: report failures at once.
		option.WithMaxRetries(0),
	}
	if httpClient != nil {
		opts = append(opts, option.WithHTTPClient(httpClient))
	}
	return &Client{client: anthropic.NewClient(opts...), server: true}
}

// serverRequest is the body /api/claude accepts.
type serverRequest struct {
	System    string          `json:"system,omitempty"`
	Messages  []serverMessage `json:"messages"`
	MaxTokens int64           `json:"max_tokens"`
}

type serverMessage struct {
	Role    string      `json:"role"`
	Content []textBlock `json:"content"`
}

type textBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// throughServer is the middleware: a Messages request rewritten for the
// game's server, and the server's event stream gathered into a message.
func throughServer(req *http.Request, next option.MiddlewareNext) (*http.Response, error) {
	if !strings.HasSuffix(req.URL.Path, "/v1/messages") || req.Method != http.MethodPost {
		return nil, fmt.Errorf("the game's server relays only messages, not %s %s", req.Method, req.URL.Path)
	}
	body, err := serverBody(req)
	if err != nil {
		return nil, err
	}
	out := req.Clone(req.Context())
	out.URL.Path = strings.TrimSuffix(req.URL.Path, "v1/messages") + "api/claude"
	out.Body = io.NopCloser(bytes.NewReader(body))
	out.GetBody = func() (io.ReadCloser, error) { return io.NopCloser(bytes.NewReader(body)), nil }
	out.ContentLength = int64(len(body))
	out.Header.Del("X-Api-Key") // the server holds the key; never send one
	out.Header.Set("Content-Type", "application/json")
	out.Header.Set("Accept", "text/event-stream")
	resp, err := next(out)
	if err != nil || resp.StatusCode != http.StatusOK {
		return resp, err
	}
	return gather(resp)
}

// serverBody reads the SDK's request body into the server's.
func serverBody(req *http.Request) ([]byte, error) {
	if req.Body == nil {
		return nil, errors.New("a Claude request with no body")
	}
	data, err := io.ReadAll(req.Body)
	req.Body.Close() //nolint:errcheck // read in full
	if err != nil {
		return nil, err
	}
	var in struct {
		System    []textBlock `json:"system"`
		MaxTokens int64       `json:"max_tokens"`
		Messages  []struct {
			Role    string          `json:"role"`
			Content json.RawMessage `json:"content"`
		} `json:"messages"`
	}
	if err := json.Unmarshal(data, &in); err != nil {
		return nil, fmt.Errorf("reading a Claude request: %w", err)
	}
	var system []string
	for _, b := range in.System {
		system = append(system, b.Text)
	}
	out := serverRequest{System: strings.Join(system, "\n\n"), MaxTokens: in.MaxTokens}
	for _, m := range in.Messages {
		var blocks []textBlock
		var s string
		if json.Unmarshal(m.Content, &s) == nil {
			blocks = []textBlock{{"text", s}}
		} else if err := json.Unmarshal(m.Content, &blocks); err != nil {
			return nil, fmt.Errorf("reading a Claude message: %w", err)
		}
		for _, b := range blocks {
			if b.Type != "text" {
				return nil, fmt.Errorf("the game's server relays only text, not %q blocks", b.Type)
			}
		}
		out.Messages = append(out.Messages, serverMessage{Role: m.Role, Content: blocks})
	}
	return json.Marshal(out)
}

// gather reads the server's stream of Claude's events into the message
// they make, as the JSON reply the SDK reads. An error event mid-reply
// becomes a 502 with Claude's error.
func gather(resp *http.Response) (*http.Response, error) {
	defer resp.Body.Close() //nolint:errcheck // read in full
	dec := ssestream.NewDecoder(resp)
	var msg anthropic.Message
	for dec.Next() {
		ev := dec.Event()
		switch ev.Type {
		case "", "ping":
			continue
		case "error":
			return reply(resp, http.StatusBadGateway, ev.Data), nil
		}
		var e anthropic.MessageStreamEventUnion
		if err := json.Unmarshal(ev.Data, &e); err != nil {
			return nil, fmt.Errorf("reading Claude's reply: %w", err)
		}
		if err := msg.Accumulate(e); err != nil {
			return nil, fmt.Errorf("reading Claude's reply: %w", err)
		}
	}
	if err := dec.Err(); err != nil {
		return nil, fmt.Errorf("reading Claude's reply: %w", err)
	}
	if msg.ID == "" {
		return nil, errors.New("the game's server sent no reply from Claude")
	}
	type block struct {
		Type string `json:"type"`
		Text string `json:"text"`
	}
	out := struct {
		ID         string  `json:"id"`
		Type       string  `json:"type"`
		Role       string  `json:"role"`
		Model      string  `json:"model"`
		Content    []block `json:"content"`
		StopReason string  `json:"stop_reason,omitempty"`
		Usage      struct {
			InputTokens  int64 `json:"input_tokens"`
			OutputTokens int64 `json:"output_tokens"`
		} `json:"usage"`
	}{ID: msg.ID, Type: "message", Role: "assistant", Model: string(msg.Model), Content: []block{}, StopReason: string(msg.StopReason)}
	for _, b := range msg.Content {
		if b.Type == "text" { // thinking and the like are the server's business
			out.Content = append(out.Content, block{"text", b.Text})
		}
	}
	out.Usage.InputTokens, out.Usage.OutputTokens = msg.Usage.InputTokens, msg.Usage.OutputTokens
	data, err := json.Marshal(out)
	if err != nil {
		return nil, err
	}
	return reply(resp, http.StatusOK, data), nil
}

// reply is a JSON response in place of the server's stream.
func reply(orig *http.Response, status int, body []byte) *http.Response {
	h := orig.Header.Clone()
	h.Set("Content-Type", "application/json")
	h.Del("Content-Length")
	return &http.Response{
		Status: fmt.Sprintf("%d %s", status, http.StatusText(status)), StatusCode: status,
		Proto: orig.Proto, ProtoMajor: orig.ProtoMajor, ProtoMinor: orig.ProtoMinor,
		Header: h, Body: io.NopCloser(bytes.NewReader(body)), ContentLength: int64(len(body)),
		Request: orig.Request,
	}
}

// serverError turns the server's refusals into errors that say what to do.
func serverError(err error) error {
	var apiErr *anthropic.Error
	if !errors.As(err, &apiErr) {
		return err
	}
	var body struct {
		Error any `json:"error"`
	}
	msg := ""
	if json.Unmarshal([]byte(apiErr.RawJSON()), &body) == nil {
		switch e := body.Error.(type) {
		case string:
			msg = e
		case map[string]any:
			msg, _ = e["message"].(string)
		}
	}
	switch apiErr.StatusCode {
	case http.StatusTooManyRequests:
		if msg != "" {
			return fmt.Errorf("%w (%s)", ErrServerCap, msg)
		}
		return ErrServerCap
	case http.StatusUnauthorized:
		return ErrSignedOut
	}
	if msg != "" {
		return fmt.Errorf("the game's server: %s (%d)", msg, apiErr.StatusCode)
	}
	return err
}
