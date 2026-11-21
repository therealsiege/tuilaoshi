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

func TestListPagesSeesChangedNewAndDeletedFiles(t *testing.T) {
	store := NewStore(t.TempDir())
	write := func(name, title string) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(store.Dir, name), []byte("---\ntype: reference\n---\n\n# "+title+"\n"), 0644); err != nil {
			t.Fatal(err)
		}
	}
	titles := func() map[string]string {
		t.Helper()
		pages, err := store.ListPages()
		if err != nil {
			t.Fatal(err)
		}
		got := map[string]string{}
		for _, p := range pages {
			got[p.Path] = p.Title
		}
		return got
	}

	write("a.md", "First")
	write("b.md", "Gone soon")
	titles() // fill the cache

	write("a.md", "Edited outside the app")
	write("c.md", "New")
	if err := os.Remove(filepath.Join(store.Dir, "b.md")); err != nil {
		t.Fatal(err)
	}
	got := titles()
	want := map[string]string{"a.md": "Edited outside the app", "c.md": "New"}
	if len(got) != len(want) || got["a.md"] != want["a.md"] || got["c.md"] != want["c.md"] {
		t.Errorf("pages %v, want %v", got, want)
	}
}
