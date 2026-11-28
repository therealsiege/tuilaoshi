// Package remote talks to the game's server (web/): signing in, pairing a
// device, and syncing progress. Sync carries each device's exported
// progress (wiki.Store.Export) up and every other device's down; the merge
// itself is the store's own (wiki.Store.Sync), run here on this device.
package remote

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/therealsiege/tuilaoshi/internal/wiki"
)

// Client is a signed-in device: the server's base URL and this device's token.
type Client struct {
	Base  string
	Token string
	HTTP  *http.Client
}

// Session is what signing in or claiming a pairing code returns.
type Session struct {
	Token     string `json:"token"`
	Device    string `json:"device"`
	ExpiresAt string `json:"expiresAt"`
}

func (c Client) http() *http.Client {
	if c.HTTP != nil {
		return c.HTTP
	}
	return &http.Client{Timeout: time.Minute}
}

// call sends a JSON request and decodes a JSON reply, turning the server's
// {error} replies into errors.
func (c Client) call(ctx context.Context, method, path string, in, out any) error {
	var body io.Reader
	if in != nil {
		data, err := json.Marshal(in)
		if err != nil {
			return err
		}
		body = bytes.NewReader(data)
	}
	req, err := http.NewRequestWithContext(ctx, method, strings.TrimRight(c.Base, "/")+path, body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}
	resp, err := c.http().Do(req)
	if err != nil {
		return fmt.Errorf("server: %w", err)
	}
	defer resp.Body.Close() //nolint:errcheck // read-only body
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode >= 300 {
		var e struct {
			Error string `json:"error"`
		}
		if json.Unmarshal(data, &e) == nil && e.Error != "" {
			return fmt.Errorf("server: %s", e.Error)
		}
		return fmt.Errorf("server: %s", resp.Status)
	}
	if out == nil {
		return nil
	}
	return json.Unmarshal(data, out)
}

// SignInWithApple trades an Apple identity token for this device's session.
func SignInWithApple(ctx context.Context, base, identityToken, device string) (Session, error) {
	var s Session
	err := Client{Base: base}.call(ctx, "POST", "/api/auth/apple", map[string]string{"identityToken": identityToken, "device": device}, &s)
	return s, err
}

// PairingCode asks for a short code another device can claim.
func (c Client) PairingCode(ctx context.Context) (code, expiresAt string, err error) {
	var r struct {
		Code      string `json:"code"`
		ExpiresAt string `json:"expiresAt"`
	}
	err = c.call(ctx, "POST", "/api/pair", nil, &r)
	return r.Code, r.ExpiresAt, err
}

// ClaimCode trades a pairing code for this device's session.
func ClaimCode(ctx context.Context, base, code, device string) (Session, error) {
	var s Session
	err := Client{Base: base}.call(ctx, "POST", "/api/pair/claim", map[string]string{"code": code, "device": device}, &s)
	return s, err
}

// exportFile reports whether a path in an export travels: the export's
// YAML and reviews.csv, never anything under private/.
func exportFile(rel string) bool {
	rel = filepath.ToSlash(rel)
	return filepath.IsLocal(rel) && !strings.HasPrefix(rel, "private/") &&
		(rel == "reviews.csv" || strings.HasSuffix(rel, ".yaml"))
}

// Sync merges every other device's progress from the server into store,
// then uploads this device's, as device.
func (c Client) Sync(ctx context.Context, store *wiki.Store, device string) error {
	var others struct {
		Devices []struct {
			Device string            `json:"device"`
			Files  map[string]string `json:"files"`
		} `json:"devices"`
	}
	if err := c.call(ctx, "GET", "/api/sync", nil, &others); err != nil {
		return err
	}
	root, err := os.MkdirTemp("", "tuilaoshi-sync-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(root) //nolint:errcheck // a temp folder
	for _, d := range others.Devices {
		if d.Device == device || !filepath.IsLocal(d.Device) || strings.ContainsAny(d.Device, `/\`) {
			continue
		}
		for rel, body := range d.Files {
			if !exportFile(rel) {
				return fmt.Errorf("sync: %s sent a file it shouldn't: %q", d.Device, rel)
			}
			path := filepath.Join(root, d.Device, filepath.FromSlash(rel))
			if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
				return err
			}
			if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
				return err
			}
		}
	}
	if err := store.Sync(root, device); err != nil {
		return err
	}
	mine := filepath.Join(root, device)
	files := map[string]string{}
	err = filepath.WalkDir(mine, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		rel, err := filepath.Rel(mine, path)
		if err != nil || !exportFile(rel) {
			return err
		}
		data, err := os.ReadFile(path)
		files[filepath.ToSlash(rel)] = string(data)
		return err
	})
	if err != nil {
		return err
	}
	return c.call(ctx, "PUT", "/api/sync/"+device, map[string]any{"files": files}, nil)
}
