package mobile

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDeleteAccountCallsTheServer(t *testing.T) {
	var got string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r.Method + " " + r.URL.Path + " " + r.Header.Get("Authorization")
		if r.Header.Get("Authorization") != "Bearer good" {
			w.WriteHeader(401)
			_, _ = w.Write([]byte(`{"error":"invalid or expired token; sign in again"}`)) //nolint:errcheck // test server
			return
		}
		w.WriteHeader(204)
	}))
	defer srv.Close()

	if err := DeleteAccount(srv.URL, "good"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if got != "DELETE /api/account Bearer good" {
		t.Fatalf("server saw %q", got)
	}
	if err := DeleteAccount(srv.URL, "stale"); err == nil || !strings.Contains(err.Error(), "sign in again") {
		t.Fatalf("a refused token should come back as an error, got %v", err)
	}
}
