package audio

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPlayFindsTheEffectsSound(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "Tink.aiff"), nil, 0644); err != nil {
		t.Fatal(err)
	}
	defer func(d string, p func(string)) { soundDir, playSound = d, p }(soundDir, playSound)
	soundDir = dir
	var played []string
	playSound = func(path string) { played = append(played, filepath.Base(path)) }
	Play(Right)
	Play(Wrong)          // its sound isn't there: nothing plays
	Play(Effect("nope")) // no such effect
	if len(played) != 1 || played[0] != "Tink.aiff" {
		t.Errorf("played %v, want just Tink.aiff", played)
	}
}
