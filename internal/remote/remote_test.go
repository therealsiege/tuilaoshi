package remote

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/therealsiege/tuilaoshi/internal/model"
	"github.com/therealsiege/tuilaoshi/internal/wiki"
)

// fakeServer keeps exports per device the way web/app/api/sync does, for
// tokens of the form "token-<device>".
type fakeServer struct {
	mu      sync.Mutex
	exports map[string]map[string]string
}

func (f *fakeServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()
	device := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer token-")
	switch {
	case r.Method == "GET" && r.URL.Path == "/api/sync":
		type exp struct {
			Device string            `json:"device"`
			Files  map[string]string `json:"files"`
		}
		out := []exp{}
		for d, files := range f.exports {
			if d != device {
				out = append(out, exp{d, files})
			}
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"devices": out}) //nolint:errcheck // test server
	case r.Method == "PUT" && strings.HasPrefix(r.URL.Path, "/api/sync/"):
		if strings.TrimPrefix(r.URL.Path, "/api/sync/") != device {
			w.WriteHeader(403)
			_, _ = w.Write([]byte(`{"error":"a device may only write its own export"}`)) //nolint:errcheck // test server
			return
		}
		var body struct {
			Files map[string]string `json:"files"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			w.WriteHeader(400)
			return
		}
		f.exports[device] = body.Files
		_, _ = w.Write([]byte(`{}`)) //nolint:errcheck // test server
	case r.Method == "POST" && r.URL.Path == "/api/pair/claim":
		_, _ = w.Write([]byte(`{"token":"token-mac","device":"mac","expiresAt":"2027-11-27T00:00:00Z"}`)) //nolint:errcheck // test server
	default:
		w.WriteHeader(404)
	}
}

func oneCardWiki(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	page := "---\ntype: character\nlevel: HSK1\nreview_interval: 1\nnext_review: 2020-01-02\n---\n\n# 书\n\n- **Character**: 书\n- **Pinyin**: shū\n- **Meaning**: book\n"
	if err := os.MkdirAll(filepath.Join(dir, "characters"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "characters", "shu1.md"), []byte(page), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

func due(t *testing.T, s *wiki.Store) int {
	t.Helper()
	items, err := wiki.DueItems(s, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	return len(items)
}

// A review on the phone reaches the Mac through the server.
func TestSyncThroughTheServer(t *testing.T) {
	srv := httptest.NewServer(&fakeServer{exports: map[string]map[string]string{}})
	defer srv.Close()
	ctx := context.Background()
	wikiDir := oneCardWiki(t)

	phone := wiki.NewStore(wikiDir)
	phone.Progress = t.TempDir()
	defer phone.Close() //nolint:errcheck // test
	cards, err := phone.DrillableCards()
	if err != nil || len(cards) != 1 {
		t.Fatalf("cards = %v, %v", cards, err)
	}
	if err := phone.RecordReview(&cards[0], model.DrillDifficulty(2), time.Now()); err != nil {
		t.Fatal(err)
	}
	if err := (Client{Base: srv.URL, Token: "token-iphone"}).Sync(ctx, phone, "iphone"); err != nil {
		t.Fatal(err)
	}

	session, err := ClaimCode(ctx, srv.URL, "123456", "mac")
	if err != nil || session.Token != "token-mac" {
		t.Fatalf("claim = %+v, %v", session, err)
	}
	mac := wiki.NewStore(wikiDir)
	mac.Progress = t.TempDir()
	defer mac.Close() //nolint:errcheck // test
	if due(t, mac) != 1 {
		t.Fatal("the Mac should start with the card due")
	}
	if err := (Client{Base: srv.URL, Token: session.Token}).Sync(ctx, mac, "mac"); err != nil {
		t.Fatal(err)
	}
	if due(t, mac) != 0 {
		t.Error("the phone's review didn't reach the Mac through the server")
	}
}

// The server's errors come back as errors, and a device can't write
// another's export.
func TestServerErrorsComeBack(t *testing.T) {
	srv := httptest.NewServer(&fakeServer{exports: map[string]map[string]string{}})
	defer srv.Close()
	s := wiki.NewStore(oneCardWiki(t))
	s.Progress = t.TempDir()
	defer s.Close() //nolint:errcheck // test
	err := (Client{Base: srv.URL, Token: "token-iphone"}).Sync(context.Background(), s, "mac")
	if err == nil || !strings.Contains(err.Error(), "only write its own") {
		t.Errorf("err = %v; want the server's refusal", err)
	}
}

func TestOnlyExportFilesTravel(t *testing.T) {
	for rel, want := range map[string]bool{
		"schedules.yaml": true, "saves/abc.yaml": true, "reviews.csv": true,
		"private/frequency.yaml": false, "../x.yaml": false, "/etc/x.yaml": false, "notes.md": false,
	} {
		if got := exportFile(rel); got != want {
			t.Errorf("exportFile(%q) = %v, want %v", rel, got, want)
		}
	}
}
