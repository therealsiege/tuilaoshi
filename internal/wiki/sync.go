package wiki

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Sync between Macs: each Mac exports its progress, as readable files, to
// its own folder in a shared one (iCloud Drive, say), and merges every
// other Mac's folder in: the newer copy of each save and doc wins, a
// card's schedule is the one reviewed last, and reviews are added once.
// The live database never sits in the shared folder.

// Sync merges in every other Mac's folder under root, then writes this
// Mac's progress to root/name.
func (s *Store) Sync(root, name string) error {
	if name == "" || strings.ContainsAny(name, `/\`) {
		return fmt.Errorf("bad sync name %q", name)
	}
	if err := os.MkdirAll(root, 0700); err != nil {
		return fmt.Errorf("sync: %w", err)
	}
	entries, err := os.ReadDir(root)
	if err != nil {
		return fmt.Errorf("sync: %w", err)
	}
	for _, e := range entries {
		if e.IsDir() && e.Name() != name && !strings.HasPrefix(e.Name(), ".") {
			if err := s.importFolder(filepath.Join(root, e.Name()), true); err != nil {
				return fmt.Errorf("sync from %s: %w", e.Name(), err)
			}
		}
	}
	// Write to a fresh folder, then swap it in, so another Mac never reads
	// half an export.
	tmp, err := os.MkdirTemp(root, "."+name+"-")
	if err != nil {
		return fmt.Errorf("sync: %w", err)
	}
	if err := s.Export(tmp); err != nil {
		_ = os.RemoveAll(tmp) //nolint:errcheck // the export error is the one to report
		return err
	}
	mine := filepath.Join(root, name)
	if err := os.RemoveAll(mine); err != nil {
		return fmt.Errorf("sync: %w", err)
	}
	return os.Rename(tmp, mine)
}

// newerDoc reports whether a doc saved at stamp is newer than this Mac's.
func newerDoc(tx *sql.Tx, name, stamp string) bool {
	var mine string
	if err := tx.QueryRow(`SELECT updated FROM docs WHERE name = ?`, name).Scan(&mine); err != nil {
		return true // this Mac has none
	}
	theirs, err1 := time.Parse(time.RFC3339, stamp)
	ours, err2 := time.Parse(time.RFC3339, mine)
	return err1 == nil && (err2 != nil || theirs.After(ours))
}

// newerSchedule reports whether a card reviewed on lastReviewed is newer
// than this Mac's schedule for it.
func newerSchedule(tx *sql.Tx, id, lastReviewed string) bool {
	var mine string
	if err := tx.QueryRow(`SELECT last_reviewed FROM schedules WHERE card_id = ?`, id).Scan(&mine); err != nil {
		return true
	}
	return lastReviewed > mine
}
