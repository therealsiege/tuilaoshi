package wiki

import (
	"os"
	"path/filepath"
	"testing"
)

func TestUpdateSectionReplacesOnlyThatSection(t *testing.T) {
	store := NewStore(t.TempDir())
	page := "# Overview\n\n## Current Level Assessment\n_Pending_\n\n## Roadmap\n\nKeep me.\n"
	path := filepath.Join(store.Dir, "overview.md")
	if err := os.WriteFile(path, []byte(page), 0644); err != nil {
		t.Fatal(err)
	}
	var gotOld string
	err := store.UpdateSection("overview.md", "Current Level Assessment", func(old string) string {
		gotOld = old
		return "| Skill | Level |"
	})
	if err != nil {
		t.Fatal(err)
	}
	if gotOld != "_Pending_\n" {
		t.Errorf("old body %q", gotOld)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	want := "# Overview\n\n## Current Level Assessment\n| Skill | Level |\n\n## Roadmap\n\nKeep me.\n"
	if string(data) != want {
		t.Errorf("page:\n%s\nwant:\n%s", data, want)
	}
	if err := store.UpdateSection("overview.md", "Missing", func(string) string { return "" }); err == nil {
		t.Error("missing section returned no error")
	}
}
