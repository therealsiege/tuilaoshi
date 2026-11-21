// Command mandarin runs the tuilaoshi terminal tutor over a wiki directory.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/cj/mandarin-wiki/internal/app"
)

func main() {
	wikiDir := flag.String("wiki", "wiki", "path to the wiki directory")
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

	if _, err := tea.NewProgram(app.New(dir)).Run(); err != nil {
		fmt.Fprintf(os.Stderr, "mandarin: %v\n", err)
		os.Exit(1)
	}
}
