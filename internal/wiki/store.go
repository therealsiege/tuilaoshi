package wiki

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/therealsiege/tuilaoshi/internal/model"
)

// Store provides read/write access to wiki markdown files.
type Store struct {
	Dir string // absolute path to the wiki/ directory

	// Save is the id of the journey being played (see saves.go); empty means
	// the single journey of wiki/journey.yaml.
	Save string

	// Parsed pages by relative path, reused while a file's modification time
	// and size are unchanged, so screens can list the whole wiki cheaply.
	mu    sync.Mutex
	pages map[string]cachedPage

	// Progress is the folder for the progress file and its backups; empty
	// means Dir. Tests reading the real wiki point it elsewhere.
	Progress string

	// Pages is the folder for the pages the game writes itself: sessions
	// and the vocabulary it collects (idioms, festival words, sentence
	// cards, the frequency deck). Empty means Dir. On the phone the wiki
	// is read-only, so these pages live beside the progress file and are
	// read over the wiki's own.
	Pages string

	// Scheduler is how reviews are spaced: SchedulerClassic (or empty)
	// or SchedulerFSRS. The app sets it from the learner's settings.
	Scheduler string

	// The progress file (see progress.go), opened on first use.
	dbOnce sync.Once
	db     *sql.DB
	dbErr  error
}

type cachedPage struct {
	mod  time.Time
	size int64
	page model.Page
}

// NewStore creates a Store rooted at the given wiki directory.
func NewStore(dir string) *Store {
	return &Store{Dir: dir}
}

// pagesDir is the folder the game writes its own pages to.
func (s *Store) pagesDir() string {
	if s.Pages != "" {
		return s.Pages
	}
	return s.Dir
}

// readPath is where a page is read from: the game's own copy if it has
// written one, else the wiki's.
func (s *Store) readPath(relPath string) string {
	if s.pagesDir() != s.Dir {
		if p := filepath.Join(s.pagesDir(), relPath); fileExists(p) {
			return p
		}
	}
	return filepath.Join(s.Dir, relPath)
}

// writePath is where the game writes a page, making its folder.
func (s *Store) writePath(relPath string) (string, error) {
	path := filepath.Join(s.pagesDir(), relPath)
	return path, os.MkdirAll(filepath.Dir(path), 0755)
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// roots are the folders pages are listed from, the game's own first so
// its copy of a page wins.
func (s *Store) roots() []string {
	if s.pagesDir() != s.Dir {
		return []string{s.pagesDir(), s.Dir}
	}
	return []string{s.Dir}
}

// ReadPage reads and parses a single wiki page by its relative path.
func (s *Store) ReadPage(relPath string) (model.Page, error) {
	data, err := os.ReadFile(s.readPath(relPath))
	if err != nil {
		return model.Page{}, fmt.Errorf("read page %s: %w", relPath, err)
	}
	page, err := ParsePage(relPath, data)
	if err != nil {
		return page, err
	}
	schedules, err := s.readSchedules()
	if sch, ok := schedules[relPath]; ok {
		withSchedule(&page, sch)
	}
	return page, err
}

// ListPages returns all markdown pages under the wiki directory. Files
// unchanged since the last call are not read or parsed again.
func (s *Store) ListPages() ([]model.Page, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var pages []model.Page
	fresh := map[string]cachedPage{}
	var err error
	for _, root := range s.roots() {
		if err = s.walkPages(root, fresh, &pages); err != nil {
			break
		}
	}
	if err != nil {
		return pages, err
	}
	s.pages = fresh // drops deleted files
	schedules, err := s.readSchedules()
	if err != nil {
		return pages, err
	}
	for i := range pages {
		if sch, ok := schedules[pages[i].Path]; ok {
			withSchedule(&pages[i], sch)
		}
	}
	return pages, nil
}

// walkPages adds the pages under root not already in fresh, reusing
// cached ones whose file hasn't changed.
func (s *Store) walkPages(root string, fresh map[string]cachedPage, pages *[]model.Page) error {
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			if root != s.Dir && os.IsNotExist(err) {
				return nil // the game hasn't written any pages yet
			}
			return err
		}
		if info.IsDir() || !strings.HasSuffix(info.Name(), ".md") {
			return nil
		}
		relPath, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		if _, done := fresh[relPath]; done {
			return nil
		}
		if c, ok := s.pages[relPath]; ok && c.mod.Equal(info.ModTime()) && c.size == info.Size() {
			fresh[relPath] = c
			*pages = append(*pages, c.page)
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return nil // skip unreadable files
		}
		page, err := ParsePage(relPath, data)
		if err != nil {
			return nil // skip unparseable files
		}
		fresh[relPath] = cachedPage{mod: info.ModTime(), size: info.Size(), page: page}
		*pages = append(*pages, page)
		return nil
	})
	return err
}

// withSchedule puts a page's saved review schedule over its frontmatter:
// pages are content, and a reviewed page's schedule lives in the progress
// file.
func withSchedule(p *model.Page, sch wordSchedule) {
	p.Introduced = parseDate(sch.Introduced)
	p.LastReviewed = parseDate(sch.LastReviewed)
	p.ReviewInterval = sch.ReviewInterval
	p.NextReview = parseDate(sch.NextReview)
}

// ListPagesByType returns pages matching the given type.
func (s *Store) ListPagesByType(pageType string) ([]model.Page, error) {
	all, err := s.ListPages()
	if err != nil {
		return nil, err
	}
	var result []model.Page
	for _, p := range all {
		if p.Type == pageType {
			result = append(result, p)
		}
	}
	return result, nil
}

// ListDirs returns immediate subdirectory names of the wiki directory.
func (s *Store) ListDirs() ([]string, error) {
	entries, err := os.ReadDir(s.Dir)
	if err != nil {
		return nil, err
	}
	var dirs []string
	for _, e := range entries {
		if e.IsDir() {
			dirs = append(dirs, e.Name())
		}
	}
	return dirs, nil
}

// ListDirPages returns all pages in a specific subdirectory.
func (s *Store) ListDirPages(subdir string) ([]model.Page, error) {
	dir := filepath.Join(s.Dir, subdir)
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var pages []model.Page
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		relPath := filepath.Join(subdir, e.Name())
		data, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			continue
		}
		page, err := ParsePage(relPath, data)
		if err != nil {
			continue
		}
		pages = append(pages, page)
	}
	return pages, nil
}

// UpdateSection rewrites the body of a page's "## heading" section, up to
// the next "## " heading, with update(old body). It fails if the page has
// no such section.
func (s *Store) UpdateSection(relPath, heading string, update func(old string) string) error {
	data, err := os.ReadFile(s.readPath(relPath))
	if err != nil {
		return err
	}
	lines := strings.Split(string(data), "\n")
	start := -1
	for i, l := range lines {
		if strings.TrimSpace(l) == "## "+heading {
			start = i
			break
		}
	}
	if start < 0 {
		return fmt.Errorf("%s has no %q section", relPath, heading)
	}
	end := len(lines)
	for i := start + 1; i < len(lines); i++ {
		if strings.HasPrefix(lines[i], "## ") {
			end = i
			break
		}
	}
	body := strings.Trim(update(strings.Join(lines[start+1:end], "\n")), "\n")
	out := append(append(append([]string{}, lines[:start+1]...), body, ""), lines[end:]...)
	path, err := s.writePath(relPath)
	if err != nil {
		return err
	}
	return os.WriteFile(path, []byte(strings.Join(out, "\n")), 0644)
}
