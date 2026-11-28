package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/therealsiege/tuilaoshi/internal/remote"
	"github.com/therealsiege/tuilaoshi/internal/wiki"
)

// SyncDirEnv names the shared folder (iCloud Drive, say) progress is synced
// through: merged in at start, written out at quit.
const SyncDirEnv = "TUILAOSHI_SYNC_DIR"

// syncName is this Mac's folder name in the shared folder: its host name,
// or TUILAOSHI_SYNC_NAME when set (two wikis on one Mac need two names).
func syncName() string {
	if name := os.Getenv("TUILAOSHI_SYNC_NAME"); name != "" {
		return name
	}
	name, err := os.Hostname()
	if err != nil || name == "" {
		return "mac"
	}
	return strings.TrimSuffix(strings.Split(name, ".")[0], ".local")
}

// ServerEnv names the game's server (web/), for pairing; DefaultServer is
// the deployed one.
const (
	ServerEnv     = "TUILAOSHI_SERVER"
	DefaultServer = "https://tuilaoshi.vercel.app"
)

// serverURL is the server to pair with: TUILAOSHI_SERVER, else the deployed one.
func serverURL() string {
	if s := os.Getenv(ServerEnv); s != "" {
		return s
	}
	return DefaultServer
}

// session is this Mac's pairing with the game's server, kept in
// ~/.config/tuilaoshi/session.json, readable only by the learner.
type session struct {
	Base   string `json:"base"`
	Token  string `json:"token"`
	Device string `json:"device"`
}

func sessionPath() (string, error) {
	home, err := os.UserHomeDir()
	return filepath.Join(home, ".config", "tuilaoshi", "session.json"), err
}

// loadSession reads the pairing, or nil when this Mac isn't paired.
func loadSession() (*session, error) {
	path, err := sessionPath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var s session
	if err := json.Unmarshal(data, &s); err != nil || s.Token == "" || s.Base == "" {
		return nil, fmt.Errorf("%s is damaged; pair again with -pair", path)
	}
	return &s, nil
}

// pair claims a code shown on the phone, so this Mac syncs through the
// game's server from now on.
func pair(base, code string) error {
	if base == "" {
		return fmt.Errorf("no server to pair with: pass -server or set %s", ServerEnv)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	got, err := remote.ClaimCode(ctx, base, strings.TrimSpace(code), syncName())
	if err != nil {
		return err
	}
	path, err := sessionPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	data, err := json.Marshal(session{Base: base, Token: got.Token, Device: got.Device})
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}

// syncProgress merges the other devices' progress in and writes this
// Mac's out: through the game's server once paired, else through the
// shared folder root, if any.
func syncProgress(dir, root string) error {
	s, err := loadSession()
	if err != nil {
		return err
	}
	if s == nil && root == "" {
		return nil
	}
	store := wiki.NewStore(dir)
	defer store.Close() //nolint:errcheck // the sync error is the one to report
	if s != nil {
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()
		return remote.Client{Base: s.Base, Token: s.Token}.Sync(ctx, store, s.Device)
	}
	return store.Sync(root, syncName())
}
