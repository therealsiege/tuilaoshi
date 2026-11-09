package wiki

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cj/mandarin-wiki/internal/model"
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
		relPath, _ := filepath.Rel(s.Dir, path)
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

// WriteFrontmatter updates the frontmatter of a page on disk without changing the body.
func (s *Store) WriteFrontmatter(page model.Page) error {
	path := filepath.Join(s.Dir, page.Path)
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read for update %s: %w", page.Path, err)
	}

	_, body := splitFrontmatter(data)

	var b strings.Builder
	b.WriteString("---\n")
	b.WriteString(fmt.Sprintf("type: %s\n", page.Type))
	if page.Level != "" {
		b.WriteString(fmt.Sprintf("level: %s\n", page.Level))
	}
	if page.Status != "" {
		b.WriteString(fmt.Sprintf("status: %s\n", page.Status))
	}
	if !page.LastReviewed.IsZero() {
		b.WriteString(fmt.Sprintf("last_reviewed: %s\n", FormatDate(page.LastReviewed)))
	}
	if len(page.Tags) > 0 {
		b.WriteString(fmt.Sprintf("tags: [%s]\n", strings.Join(page.Tags, ", ")))
	}
	if !page.Introduced.IsZero() {
		b.WriteString(fmt.Sprintf("introduced: %s\n", FormatDate(page.Introduced)))
	}
	if page.ReviewInterval > 0 {
		b.WriteString(fmt.Sprintf("review_interval: %d\n", page.ReviewInterval))
	}
	if !page.NextReview.IsZero() {
		b.WriteString(fmt.Sprintf("next_review: %s\n", FormatDate(page.NextReview)))
	}
	b.WriteString("---\n\n")
	b.WriteString(body)

	return os.WriteFile(path, []byte(b.String()), 0644)
}

// PageCount returns the total number of markdown files.
func (s *Store) PageCount() int {
	pages, _ := s.ListPages()
	return len(pages)
}

// WriteSession saves a drill/review session as a markdown file in wiki/sessions/.
func (s *Store) WriteSession(sess model.Session) error {
	sessDir := filepath.Join(s.Dir, "sessions")
	if err := os.MkdirAll(sessDir, 0755); err != nil {
		return fmt.Errorf("create sessions dir: %w", err)
	}

	filename := sess.StartedAt.Format("2006-01-02-150405") + ".md"
	path := filepath.Join(sessDir, filename)

	var b strings.Builder
	b.WriteString("---\n")
	b.WriteString("type: session\n")
	b.WriteString(fmt.Sprintf("status: active\n"))
	b.WriteString(fmt.Sprintf("last_reviewed: %s\n", FormatDate(time.Now())))
	b.WriteString(fmt.Sprintf("tags: [%s]\n", sess.Type))
	b.WriteString("---\n\n")

	b.WriteString(fmt.Sprintf("# Session — %s\n\n", sess.StartedAt.Format("2006-01-02 15:04")))
	b.WriteString(fmt.Sprintf("- **Type**: %s\n", sess.Type))
	b.WriteString(fmt.Sprintf("- **Duration**: %s\n", sess.EndedAt.Sub(sess.StartedAt).Truncate(time.Second)))
	b.WriteString(fmt.Sprintf("- **Cards**: %d\n", sess.CardsTotal))
	b.WriteString(fmt.Sprintf("- **Correct**: %d\n", sess.CardsCorrect))
	if sess.CardsTotal > 0 {
		pct := float64(sess.CardsCorrect) / float64(sess.CardsTotal) * 100
		b.WriteString(fmt.Sprintf("- **Accuracy**: %.0f%%\n", pct))
	}
	b.WriteString("\n## Results\n\n")
	b.WriteString("| Card | Answer | Result | Difficulty |\n")
	b.WriteString("|------|--------|--------|------------|\n")
	for _, item := range sess.Items {
		result := "incorrect"
		if item.Correct {
			result = "correct"
		}
		diff := difficultyLabel(item.Difficulty)
		b.WriteString(fmt.Sprintf("| %s | %s | %s | %s |\n",
			item.CardID, item.UserAnswer, result, diff))
	}

	return os.WriteFile(path, []byte(b.String()), 0644)
}

func difficultyLabel(d model.DrillDifficulty) string {
	switch d {
	case model.DiffAgain:
		return "Again"
	case model.DiffHard:
		return "Hard"
	case model.DiffGood:
		return "Good"
	case model.DiffEasy:
		return "Easy"
	default:
		return "Good"
	}
}
