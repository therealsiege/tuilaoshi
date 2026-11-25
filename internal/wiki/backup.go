package wiki

import (
	"database/sql"
	"encoding/csv"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Backups: the first time the progress file opens each day, a copy is
// written to wiki/backups/progress-YYYY-MM-DD.db, and the newest
// keepBackups are kept. Export writes the progress out as readable files
// (the docs under their old YAML names, schedules.yaml, reviews.csv), and
// Import reads such a folder back in, e.g. on another Mac.
const (
	backupsDir  = "backups"
	keepBackups = 7
)

// backupDaily copies the progress file once a day and prunes old copies.
func (s *Store) backupDaily(db *sql.DB, now time.Time) error {
	dir := filepath.Join(s.Dir, backupsDir)
	path := filepath.Join(dir, "progress-"+now.Format(dateFmt)+".db")
	if _, err := os.Stat(path); err == nil {
		return nil
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("backup: %w", err)
	}
	if _, err := db.Exec(`VACUUM INTO ?`, path); err != nil {
		return fmt.Errorf("backup: %w", err)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("backup: %w", err)
	}
	var copies []string
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), "progress-") && strings.HasSuffix(e.Name(), ".db") {
			copies = append(copies, e.Name())
		}
	}
	sort.Strings(copies) // dated names sort oldest first
	for len(copies) > keepBackups {
		if err := os.Remove(filepath.Join(dir, copies[0])); err != nil {
			return fmt.Errorf("backup: %w", err)
		}
		copies = copies[1:]
	}
	return nil
}

// Export files: every doc under its old name, plus these two.
const (
	exportSchedules = "schedules.yaml"
	exportReviews   = "reviews.csv"
)

// exportedSchedule is a card's schedule as exported.
type exportedSchedule struct {
	wordSchedule `yaml:",inline"`
	Reps         int `yaml:"reps,omitempty"`
	Lapses       int `yaml:"lapses,omitempty"`
}

// Export writes the learner's progress into dir as readable files.
func (s *Store) Export(dir string) error {
	db, err := s.progress()
	if err != nil {
		return err
	}
	names, err := s.docNames("")
	if err != nil {
		return err
	}
	for _, name := range names {
		body, _, err := s.readDoc(name)
		if err != nil {
			return err
		}
		if err := writeExport(dir, name, body); err != nil {
			return err
		}
	}
	rows, err := db.Query(`SELECT card_id, introduced, last_reviewed, review_interval, next_review, reps, lapses FROM schedules ORDER BY card_id`)
	if err != nil {
		return err
	}
	schedules := map[string]exportedSchedule{}
	for rows.Next() {
		var id string
		var e exportedSchedule
		if err := rows.Scan(&id, &e.Introduced, &e.LastReviewed, &e.ReviewInterval, &e.NextReview, &e.Reps, &e.Lapses); err != nil {
			_ = rows.Close() //nolint:errcheck // the scan error is the one to report
			return err
		}
		schedules[id] = e
	}
	if err := errors.Join(rows.Err(), rows.Close()); err != nil {
		return err
	}
	data, err := yaml.Marshal(schedules)
	if err != nil {
		return err
	}
	if err := writeExport(dir, exportSchedules, data); err != nil {
		return err
	}
	return s.exportReviews(db, filepath.Join(dir, exportReviews))
}

func writeExport(dir, name string, body []byte) error {
	path := filepath.Join(dir, filepath.FromSlash(name))
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("export %s: %w", name, err)
	}
	if err := os.WriteFile(path, body, 0644); err != nil {
		return fmt.Errorf("export %s: %w", name, err)
	}
	return nil
}

func (s *Store) exportReviews(db *sql.DB, path string) (err error) {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("export reviews: %w", err)
	}
	defer func() { err = errors.Join(err, f.Close()) }()
	w := csv.NewWriter(f)
	if err := w.Write([]string{"card_id", "at", "rating", "interval_was", "interval_now"}); err != nil {
		return err
	}
	rows, err := db.Query(`SELECT card_id, at, rating, interval_was, interval_now FROM reviews ORDER BY id`)
	if err != nil {
		return err
	}
	defer rows.Close() //nolint:errcheck // read-only
	for rows.Next() {
		var id, at string
		var rating, was, now int
		if err := rows.Scan(&id, &at, &rating, &was, &now); err != nil {
			return err
		}
		if err := w.Write([]string{id, at, strconv.Itoa(rating), strconv.Itoa(was), strconv.Itoa(now)}); err != nil {
			return err
		}
	}
	w.Flush()
	return errors.Join(rows.Err(), w.Error())
}

// Import reads an exported folder into the progress file, replacing the
// docs and schedules it holds and adding reviews not already logged. The
// day's backup is made first.
func (s *Store) Import(dir string) error {
	db, err := s.progress()
	if err != nil {
		return err
	}
	if err := s.backupDaily(db, time.Now()); err != nil {
		return err
	}
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck // a no-op after Commit
	err = filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		name := filepath.ToSlash(rel)
		if name == exportSchedules || name == exportReviews || !strings.HasSuffix(name, ".yaml") {
			return nil
		}
		body, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		_, err = tx.Exec(`INSERT INTO docs (name, body, updated) VALUES (?, ?, ?)
			ON CONFLICT(name) DO UPDATE SET body = excluded.body, updated = excluded.updated`,
			name, string(body), time.Now().Format(time.RFC3339))
		return err
	})
	if err != nil {
		return fmt.Errorf("import: %w", err)
	}
	if data, err := os.ReadFile(filepath.Join(dir, exportSchedules)); err == nil {
		var schedules map[string]exportedSchedule
		if err := yaml.Unmarshal(data, &schedules); err != nil {
			return fmt.Errorf("import %s: %w", exportSchedules, err)
		}
		for id, e := range schedules {
			if _, err := tx.Exec(`INSERT OR REPLACE INTO schedules (card_id, introduced, last_reviewed, review_interval, next_review, reps, lapses) VALUES (?, ?, ?, ?, ?, ?, ?)`,
				id, e.Introduced, e.LastReviewed, e.ReviewInterval, e.NextReview, e.Reps, e.Lapses); err != nil {
				return fmt.Errorf("import %s: %w", exportSchedules, err)
			}
		}
	} else if !os.IsNotExist(err) {
		return err
	}
	if err := importReviews(tx, filepath.Join(dir, exportReviews)); err != nil {
		return err
	}
	return tx.Commit()
}

func importReviews(tx *sql.Tx, path string) error {
	f, err := os.Open(path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	defer f.Close() //nolint:errcheck // read-only
	records, err := csv.NewReader(f).ReadAll()
	if err != nil {
		return fmt.Errorf("import %s: %w", exportReviews, err)
	}
	for i, r := range records {
		if i == 0 || len(r) != 5 {
			continue // the header
		}
		var seen int
		if err := tx.QueryRow(`SELECT count(*) FROM reviews WHERE card_id = ? AND at = ?`, r[0], r[1]).Scan(&seen); err != nil {
			return err
		}
		if seen > 0 {
			continue
		}
		if _, err := tx.Exec(`INSERT INTO reviews (card_id, at, rating, interval_was, interval_now) VALUES (?, ?, ?, ?, ?)`, r[0], r[1], r[2], r[3], r[4]); err != nil {
			return fmt.Errorf("import %s: %w", exportReviews, err)
		}
	}
	return nil
}
