package wiki

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
	_ "modernc.org/sqlite" // the pure-Go SQLite driver, registered as "sqlite"
)

// The learner's progress lives in one SQLite file, wiki/progress.db,
// ignored by git; the pages are content only. Each piece of progress that
// used to be a YAML file (a journey's save, study bookkeeping, settings,
// the forge cache) is a row of docs under its old file name, so
// `sqlite3 wiki/progress.db "select body from docs"` reads as it did.
// Review schedules and every answer given have tables of their own.
const progressFile = "progress.db"

const schema = `
CREATE TABLE IF NOT EXISTS docs (
	name    TEXT PRIMARY KEY, -- the old file name, e.g. study.yaml, saves/<id>.yaml
	body    TEXT NOT NULL,    -- its YAML
	updated TEXT NOT NULL     -- RFC 3339
);
CREATE TABLE IF NOT EXISTS schedules (
	card_id         TEXT PRIMARY KEY, -- a page path, or path#word for a vocabulary row
	introduced      TEXT NOT NULL DEFAULT '',
	last_reviewed   TEXT NOT NULL DEFAULT '',
	review_interval INTEGER NOT NULL DEFAULT 0, -- days
	next_review     TEXT NOT NULL DEFAULT '',
	reps            INTEGER NOT NULL DEFAULT 0, -- reviews given
	lapses          INTEGER NOT NULL DEFAULT 0  -- reviews answered "again"
);
CREATE TABLE IF NOT EXISTS reviews (
	id            INTEGER PRIMARY KEY,
	card_id       TEXT NOT NULL,
	at            TEXT NOT NULL, -- RFC 3339
	rating        INTEGER NOT NULL, -- model.DrillDifficulty
	interval_was  INTEGER NOT NULL,
	interval_now  INTEGER NOT NULL
);
CREATE INDEX IF NOT EXISTS reviews_at ON reviews(at);
CREATE INDEX IF NOT EXISTS reviews_card ON reviews(card_id);
`

// progress opens the progress file, creating it on first use and bringing
// in the YAML files kept before it existed.
func (s *Store) progress() (*sql.DB, error) {
	s.dbOnce.Do(func() {
		path := filepath.Join(s.progressDir(), progressFile)
		_, statErr := os.Stat(path)
		fresh := os.IsNotExist(statErr)
		db, err := sql.Open("sqlite", "file:"+path+"?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)&_pragma=synchronous(NORMAL)")
		if err == nil {
			db.SetMaxOpenConns(1) // one writer; keeps reads and writes in order
			_, err = db.Exec(schema)
		}
		if err == nil && fresh && s.progressDir() == s.Dir {
			err = s.importFiles(db)
		}
		if err == nil {
			err = s.moveGeneratedPages(db)
		}
		if err == nil && !fresh {
			err = s.backupDaily(db, time.Now())
		}
		if err != nil {
			s.dbErr = fmt.Errorf("open %s: %w", progressFile, err)
			if db != nil {
				_ = db.Close() //nolint:errcheck // the open error is the one to report
			}
			return
		}
		s.db = db
	})
	return s.db, s.dbErr
}

// progressDir is the folder holding the progress file.
func (s *Store) progressDir() string {
	if s.Progress != "" {
		return s.Progress
	}
	return s.Dir
}

// Close closes the progress file.
func (s *Store) Close() error {
	if s.db == nil {
		return nil
	}
	return s.db.Close()
}

// readDoc returns a piece of progress by its old file name; ok is false
// when there is none yet.
func (s *Store) readDoc(name string) (body []byte, ok bool, err error) {
	db, err := s.progress()
	if err != nil {
		return nil, false, err
	}
	var text string
	switch err := db.QueryRow(`SELECT body FROM docs WHERE name = ?`, name).Scan(&text); {
	case errors.Is(err, sql.ErrNoRows):
		return nil, false, nil
	case err != nil:
		return nil, false, fmt.Errorf("read %s: %w", name, err)
	}
	return []byte(text), true, nil
}

// readYAML decodes a piece of progress into out; a missing one leaves out alone.
func (s *Store) readYAML(name string, out any) error {
	data, ok, err := s.readDoc(name)
	if err != nil || !ok {
		return err
	}
	if err := yaml.Unmarshal(data, out); err != nil {
		return fmt.Errorf("parse %s: %w", name, err)
	}
	return nil
}

// writeYAML saves a piece of progress under its old file name.
func (s *Store) writeYAML(name string, v any) error {
	data, err := yaml.Marshal(v)
	if err != nil {
		return fmt.Errorf("encode %s: %w", name, err)
	}
	return s.writeDoc(name, data)
}

func (s *Store) writeDoc(name string, body []byte) error {
	db, err := s.progress()
	if err != nil {
		return err
	}
	_, err = db.Exec(`INSERT INTO docs (name, body, updated) VALUES (?, ?, ?)
		ON CONFLICT(name) DO UPDATE SET body = excluded.body, updated = excluded.updated`,
		name, string(body), time.Now().Format(time.RFC3339))
	if err != nil {
		return fmt.Errorf("save %s: %w", name, err)
	}
	return nil
}

func (s *Store) deleteDoc(name string) error {
	db, err := s.progress()
	if err != nil {
		return err
	}
	if _, err := db.Exec(`DELETE FROM docs WHERE name = ?`, name); err != nil {
		return fmt.Errorf("delete %s: %w", name, err)
	}
	return nil
}

// docUpdated is when a piece of progress was last saved.
func (s *Store) docUpdated(name string) (time.Time, error) {
	db, err := s.progress()
	if err != nil {
		return time.Time{}, err
	}
	var at string
	if err := db.QueryRow(`SELECT updated FROM docs WHERE name = ?`, name).Scan(&at); err != nil {
		return time.Time{}, fmt.Errorf("read %s: %w", name, err)
	}
	t, _ := time.Parse(time.RFC3339, at) //nolint:errcheck // a bad stamp reads as zero
	return t, nil
}

// docNames lists the pieces of progress under a prefix, e.g. "saves/".
func (s *Store) docNames(prefix string) ([]string, error) {
	db, err := s.progress()
	if err != nil {
		return nil, err
	}
	rows, err := db.Query(`SELECT name FROM docs WHERE substr(name, 1, ?) = ? ORDER BY name`, len(prefix), prefix)
	if err != nil {
		return nil, fmt.Errorf("list %s: %w", prefix, err)
	}
	defer rows.Close() //nolint:errcheck // read-only
	var names []string
	for rows.Next() {
		var n string
		if err := rows.Scan(&n); err != nil {
			return nil, err
		}
		names = append(names, n)
	}
	return names, rows.Err()
}

// importedFiles are the YAML files kept before the progress file; on its
// first open each is read in and renamed <name>.imported.
var importedFiles = []string{journeyFile, studyFile, settingsFile, forgeFile}

// importFiles brings the old YAML files into a new progress file, in one
// transaction, then renames them so nothing is lost.
func (s *Store) importFiles(db *sql.DB) error {
	names := append([]string{}, importedFiles...)
	if entries, err := os.ReadDir(filepath.Join(s.Dir, savesDir)); err == nil {
		for _, e := range entries {
			if !e.IsDir() && strings.HasSuffix(e.Name(), ".yaml") {
				names = append(names, savesDir+"/"+e.Name())
			}
		}
	}
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck // a no-op after Commit
	var done []string
	for _, name := range names {
		path := filepath.Join(s.Dir, filepath.FromSlash(name))
		data, err := os.ReadFile(path)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return fmt.Errorf("import %s: %w", name, err)
		}
		stamp := time.Now()
		if info, err := os.Stat(path); err == nil {
			stamp = info.ModTime()
		}
		if _, err := tx.Exec(`INSERT INTO docs (name, body, updated) VALUES (?, ?, ?)`, name, string(data), stamp.Format(time.RFC3339)); err != nil {
			return fmt.Errorf("import %s: %w", name, err)
		}
		done = append(done, path)
	}
	words, err := os.ReadFile(filepath.Join(s.Dir, wordScheduleFile))
	switch {
	case err == nil:
		var schedules map[string]wordSchedule
		if err := yaml.Unmarshal(words, &schedules); err != nil {
			return fmt.Errorf("import %s: %w", wordScheduleFile, err)
		}
		for id, sch := range schedules {
			if _, err := tx.Exec(`INSERT INTO schedules (card_id, introduced, last_reviewed, review_interval, next_review) VALUES (?, ?, ?, ?, ?)`,
				id, sch.Introduced, sch.LastReviewed, sch.ReviewInterval, sch.NextReview); err != nil {
				return fmt.Errorf("import %s: %w", wordScheduleFile, err)
			}
		}
		done = append(done, filepath.Join(s.Dir, wordScheduleFile))
	case !os.IsNotExist(err):
		return fmt.Errorf("import %s: %w", wordScheduleFile, err)
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	for _, path := range done {
		if err := os.Rename(path, path+".imported"); err != nil {
			return err
		}
	}
	return nil
}

// moveGeneratedPages moves the word lists the app writes (idiom duels,
// festivals) from the tracked vocabulary folder into private/, carrying
// their words' schedules to the new card ids.
func (s *Store) moveGeneratedPages(db *sql.DB) error {
	for _, p := range []struct{ old, now string }{
		{"vocabulary/chengyu-duels.md", idiomPage},
		{"vocabulary/festivals.md", festivalPage},
	} {
		from := filepath.Join(s.Dir, p.old)
		if _, err := os.Stat(from); os.IsNotExist(err) {
			continue
		}
		to := filepath.Join(s.Dir, p.now)
		if err := os.MkdirAll(filepath.Dir(to), 0755); err != nil {
			return err
		}
		if err := os.Rename(from, to); err != nil {
			return err
		}
		if _, err := db.Exec(`UPDATE schedules SET card_id = ? || substr(card_id, ?) WHERE substr(card_id, 1, ?) = ?`,
			p.now, len(p.old)+1, len(p.old)+1, p.old+"#"); err != nil {
			return err
		}
	}
	return nil
}
