// Command demonart paints a portrait of each demon in the journey bestiary,
// and of each foot soldier, with Google's Gemini Flash image model and saves
// it where the game looks for it (assets/images/demons/<name>.png, or
// soldiers/). Needs GEMINI_API_KEY (or
// GOOGLE_API_KEY). Existing portraits are kept unless -force is given.
//
//	go run ./cmd/demonart [-only 白骨精] [-force] [-model gemini-2.5-flash-image]
package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/therealsiege/tuilaoshi/internal/journey"
)

func main() {
	model := flag.String("model", envOr("GEMINI_IMAGE_MODEL", "gemini-2.5-flash-image"), "Gemini image model")
	only := flag.String("only", "", "paint just this demon (Chinese name)")
	force := flag.Bool("force", false, "repaint portraits that already exist")
	size := flag.Int("size", 320, "longest side of the saved portrait, in pixels")
	flag.Parse()

	key := envOr("GEMINI_API_KEY", os.Getenv("GOOGLE_API_KEY"))
	if key == "" {
		fmt.Fprintln(os.Stderr, "demonart: set GEMINI_API_KEY (e.g. in ~/.config/tuilaoshi/env)")
		os.Exit(1)
	}
	client := newClient(key, *model)
	failed := 0
	enemies := journey.Bestiary()
	for _, s := range journey.Soldiers() {
		enemies = append(enemies, s.BestiaryEntry)
	}
	for _, d := range enemies {
		if *only != "" && d.Name != *only {
			continue
		}
		if _, err := os.Stat(d.ImagePath); err == nil && !*force {
			fmt.Printf("  %s %s: already painted\n", d.Name, d.English)
			continue
		}
		fmt.Printf("  %s %s: painting...\n", d.Name, d.English)
		png, err := client.paint(promptFor(d))
		if err == nil {
			err = save(d.ImagePath, png, *size)
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "    %v\n", err)
			failed++
		}
	}
	if failed > 0 {
		os.Exit(1)
	}
}

// save writes a portrait and shrinks it with sips (macOS) to keep the repo small.
func save(path string, png []byte, size int) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	if err := os.WriteFile(path, png, 0644); err != nil {
		return err
	}
	if sips, err := exec.LookPath("sips"); err == nil {
		if out, err := exec.Command(sips, "-Z", fmt.Sprint(size), path).CombinedOutput(); err != nil {
			return fmt.Errorf("resize %s: %v %s", path, err, out)
		}
	}
	return nil
}

func envOr(name, fallback string) string {
	if v := os.Getenv(name); v != "" {
		return v
	}
	return fallback
}
