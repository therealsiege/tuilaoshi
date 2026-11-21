// Package uitest compares rendered screens with golden files, so a layout
// change shows up as a reviewable diff.
package uitest

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/therealsiege/tuilaoshi/internal/wiki"
)

var ansi = regexp.MustCompile(`\x1b\[[0-9;?]*[A-Za-z]`)

// Plain strips colors and trailing spaces so golden files are readable text.
func Plain(screen string) string {
	lines := strings.Split(ansi.ReplaceAllString(screen, ""), "\n")
	for i, l := range lines {
		lines[i] = strings.TrimRight(l, " ")
	}
	return strings.Join(lines, "\n")
}

// Golden compares screen with testdata/<name>.golden. Set UPDATE_GOLDEN=1
// (make golden) to rewrite the files, then review the diff.
func Golden(t *testing.T, name, screen string) {
	t.Helper()
	got := Plain(screen)
	path := filepath.Join("testdata", name+".golden")
	if os.Getenv("UPDATE_GOLDEN") != "" {
		if err := os.MkdirAll("testdata", 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(got), 0644); err != nil {
			t.Fatal(err)
		}
		return
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("%v (run make golden to create it)", err)
	}
	if got != string(want) {
		t.Errorf("screen %s changed; if intended, run make golden and review the diff.\n--- want\n%s\n--- got\n%s", name, want, got)
	}
}

// Wiki writes pages (path relative to the wiki root -> content) to a
// temporary wiki and returns its store.
func Wiki(t *testing.T, pages map[string]string) *wiki.Store {
	t.Helper()
	dir := t.TempDir()
	for name, content := range pages {
		path := filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}
	return wiki.NewStore(dir)
}

// OverdueCharacter is a character page due for review long ago, so it is
// due on any test date.
const OverdueCharacter = `---
type: character
level: HSK1
review_interval: 1
last_reviewed: 2020-01-01
next_review: 2020-01-02
---

# 书 (shū)

- **Character**: 书
- **Pinyin**: shū
- **Meaning**: book
`
