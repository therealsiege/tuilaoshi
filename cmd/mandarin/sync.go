package main

import (
	"os"
	"strings"

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

// syncProgress merges other Macs' progress in and writes this one's out.
func syncProgress(dir, root string) error {
	if root == "" {
		return nil
	}
	store := wiki.NewStore(dir)
	defer store.Close() //nolint:errcheck // the sync error is the one to report
	return store.Sync(root, syncName())
}
