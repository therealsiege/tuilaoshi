package wiki

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/therealsiege/tuilaoshi/internal/model"
)

// readOnlyWiki is a wiki like the one bundled in the iPhone app: it can't
// be written, so the game's own pages and progress go elsewhere.
func readOnlyWiki(t *testing.T) *Store {
	t.Helper()
	dir := t.TempDir()
	write := func(rel, body string) {
		p := filepath.Join(dir, rel)
		if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), 0644); err != nil {
			t.Fatal(err)
		}
	}
	write("characters/shu1.md", "---\ntype: character\nlevel: HSK1\n---\n\n# 书\n\n- **Character**: 书\n- **Pinyin**: shū\n- **Meaning**: book\n")
	write("overview.md", "# Overview\n\n## Current Level Assessment\n_Pending_\n")
	for _, d := range []string{"characters", "."} {
		if err := os.Chmod(filepath.Join(dir, d), 0555); err != nil {
			t.Fatal(err)
		}
	}
	t.Cleanup(func() {
		for _, d := range []string{"characters", "."} {
			_ = os.Chmod(filepath.Join(dir, d), 0755) //nolint:errcheck // so the temp dir can be removed
		}
	})
	s := NewStore(dir)
	s.Progress = t.TempDir()
	s.Pages = filepath.Join(s.Progress, "pages")
	return s
}

// Everything the game writes lands in Pages, is read back over the wiki,
// and leaves the read-only wiki as it was.
func TestTheGameWritesItsOwnPagesBesideAReadOnlyWiki(t *testing.T) {
	s := readOnlyWiki(t)
	now := time.Date(2026, 11, 27, 9, 0, 0, 0, time.UTC)

	if n, err := s.AddIdioms([]Word{{Chinese: "一路平安", Pinyin: "yí lù píng ān", Meaning: "safe travels"}}, now); err != nil || n != 1 {
		t.Fatalf("AddIdioms = %d, %v", n, err)
	}
	if ok, err := s.AddSentenceCard(Word{Chinese: "书", Pinyin: "shū", Meaning: "book"}, "这是我的书。", now); err != nil || !ok {
		t.Fatalf("AddSentenceCard = %v, %v", ok, err)
	}
	if err := s.UpdateSection("overview.md", "Current Level Assessment", func(string) string { return "HSK3" }); err != nil {
		t.Fatalf("UpdateSection: %v", err)
	}
	sess := model.Session{Type: "review", StartedAt: now, EndedAt: now.Add(time.Minute), CardsTotal: 1, CardsCorrect: 1,
		Items: []model.SessionItem{{CardID: "characters/shu1.md", Correct: true}}}
	if err := s.WriteSession(sess); err != nil {
		t.Fatalf("WriteSession: %v", err)
	}

	pages, err := s.ListPages()
	if err != nil {
		t.Fatal(err)
	}
	found := map[string]bool{}
	for _, p := range pages {
		found[p.Path] = true
	}
	for _, want := range []string{"characters/shu1.md", idiomPage, sentencePage, "overview.md"} {
		if !found[want] {
			t.Errorf("ListPages is missing %s; got %v", want, found)
		}
	}
	page, err := s.ReadPage("overview.md")
	if err != nil || !strings.Contains(page.Body, "HSK3") {
		t.Errorf("overview.md should read back the game's copy, got %q, %v", page.Body, err)
	}
	if n := len(s.readSessions()); n != 1 {
		t.Errorf("read back %d sessions from Pages, want 1", n)
	}
	orig, err := os.ReadFile(filepath.Join(s.Dir, "overview.md"))
	if err != nil || strings.Contains(string(orig), "HSK3") {
		t.Errorf("the wiki's own overview.md changed: %q", orig)
	}
}
