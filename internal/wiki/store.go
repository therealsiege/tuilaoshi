package wiki

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/therealsiege/tuilaoshi/internal/model"
)

// Store provides read/write access to wiki markdown files.
type Store struct {
	Dir string // absolute path to the wiki/ directory
}

// NewStore creates a Store rooted at the given wiki directory.
func NewStore(dir string) *Store {
	return &Store{Dir: dir}
}

// ReadPage reads and parses a single wiki page by its relative path.
func (s *Store) ReadPage(relPath string) (model.Page, error) {
	data, err := os.ReadFile(filepath.Join(s.Dir, relPath))
	if err != nil {
		return model.Page{}, fmt.Errorf("read page %s: %w", relPath, err)
	}
	return ParsePage(relPath, data)
}

// ListPages returns all markdown pages under the wiki directory.
func (s *Store) ListPages() ([]model.Page, error) {
	var pages []model.Page
	err := filepath.Walk(s.Dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || !strings.HasSuffix(info.Name(), ".md") {
			return nil
		}
		relPath, err := filepath.Rel(s.Dir, path)
		if err != nil {
			return err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return nil // skip unreadable files
		}
		page, err := ParsePage(relPath, data)
		if err != nil {
			return nil // skip unparseable files
		}
		pages = append(pages, page)
		return nil
	})
	return pages, err
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

// writeFrontmatter updates the frontmatter of a page on disk without changing
// the body. Callers outside this package save reviews with RecordReview.
func (s *Store) writeFrontmatter(page model.Page) error {
	path := filepath.Join(s.Dir, page.Path)
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read for update %s: %w", page.Path, err)
	}

	_, body := splitFrontmatter(data)

	var b strings.Builder
	b.WriteString("---\n")
	fmt.Fprintf(&b, "type: %s\n", page.Type)
	if page.Level != "" {
		fmt.Fprintf(&b, "level: %s\n", page.Level)
	}
	if page.Status != "" {
		fmt.Fprintf(&b, "status: %s\n", page.Status)
	}
	if !page.LastReviewed.IsZero() {
		fmt.Fprintf(&b, "last_reviewed: %s\n", FormatDate(page.LastReviewed))
	}
	if len(page.Tags) > 0 {
		fmt.Fprintf(&b, "tags: [%s]\n", strings.Join(page.Tags, ", "))
	}
	if !page.Introduced.IsZero() {
		fmt.Fprintf(&b, "introduced: %s\n", FormatDate(page.Introduced))
	}
	if page.ReviewInterval > 0 {
		fmt.Fprintf(&b, "review_interval: %d\n", page.ReviewInterval)
	}
	if !page.NextReview.IsZero() {
		fmt.Fprintf(&b, "next_review: %s\n", FormatDate(page.NextReview))
	}
	b.WriteString("---\n\n")
	b.WriteString(body)

	return os.WriteFile(path, []byte(b.String()), 0644)
}
