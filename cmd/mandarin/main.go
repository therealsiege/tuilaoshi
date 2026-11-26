// Command mandarin runs the tuilaoshi terminal tutor over a wiki directory.
package main

import (
	"errors"
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
	deck := flag.String("deck", "", "write your studied and hard words to this CSV file for Anki or Pleco, then exit")
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

	if *deck != "" {
		if err := writeDeck(dir, *deck); err != nil {
			fmt.Fprintf(os.Stderr, "mandarin: %v\n", err)
			os.Exit(1)
		}
		return
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
	stop := logStderr(dir)
	_, err = tea.NewProgram(m).Run()
	stop()
	if errors.Is(err, tea.ErrProgramPanic) {
		fmt.Fprintf(os.Stderr, "\nTuiLaoShi crashed. The details are in %s; your progress up to the last answer is saved.\n", filepath.Join(dir, crashLog))
		os.Exit(1)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "mandarin: %v\n", err)
		os.Exit(1)
	}
}

// writeDeck writes the learner's words as a CSV deck for Anki or Pleco.
func writeDeck(dir, path string) (err error) {
	store := wiki.NewStore(dir)
	defer store.Close() //nolint:errcheck // exiting
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := f.Close(); err == nil {
			err = cerr
		}
	}()
	n, err := store.WriteDeck(f)
	if err != nil {
		return err
	}
	fmt.Printf("%d words written to %s\n", n, path)
	return nil
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
