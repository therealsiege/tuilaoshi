// Command mandarin runs the tuilaoshi terminal tutor over a wiki directory.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/therealsiege/tuilaoshi/internal/app"
	"github.com/therealsiege/tuilaoshi/internal/wiki"
)

func main() {
	wikiDir := flag.String("wiki", "wiki", "path to the wiki directory")
	export := flag.String("export", "", "write your progress as readable files into this folder, then exit")
	importDir := flag.String("import", "", "read progress exported with -export (e.g. from another Mac), then exit")
	flag.Parse()

	dir, err := filepath.Abs(*wikiDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "mandarin: %v\n", err)
		os.Exit(1)
	}
	if info, err := os.Stat(dir); err != nil || !info.IsDir() {
		fmt.Fprintf(os.Stderr, "mandarin: wiki directory %s not found; run from the repo root or pass -wiki <dir>\n", dir)
		os.Exit(1)
	}

	if *export != "" || *importDir != "" {
		if err := moveProgress(dir, *export, *importDir); err != nil {
			fmt.Fprintf(os.Stderr, "mandarin: %v\n", err)
			os.Exit(1)
		}
		return
	}

	m, err := app.New(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "mandarin: %v\n", err)
		os.Exit(1)
	}
	if _, err := tea.NewProgram(m).Run(); err != nil {
		fmt.Fprintf(os.Stderr, "mandarin: %v\n", err)
		os.Exit(1)
	}
}

// moveProgress exports or imports the learner's progress.
func moveProgress(dir, export, importDir string) error {
	store := wiki.NewStore(dir)
	defer store.Close() //nolint:errcheck // exiting
	if export != "" {
		if err := store.Export(export); err != nil {
			return err
		}
		fmt.Println("Progress written to", export)
		return nil
	}
	if err := store.Import(importDir); err != nil {
		return err
	}
	fmt.Println("Progress read from", importDir, "(today's backup is in", filepath.Join(dir, "backups")+")")
	return nil
}
