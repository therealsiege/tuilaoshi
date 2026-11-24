package audio

import (
	"os"
	"os/exec"
)

// Sound effects are short cues for game events, played with the system's
// own sounds (macOS's afplay and /System/Library/Sounds), so the repo
// carries no audio. Where either is missing, nothing plays.

// Effect is a game event with a sound.
type Effect string

const (
	Right     Effect = "right"     // a right answer
	Wrong     Effect = "wrong"     // a miss
	Victory   Effect = "victory"   // a battle won
	Defeat    Effect = "defeat"    // a battle lost
	Encounter Effect = "encounter" // a clan leaps out
	LevelUp   Effect = "levelup"   // a level, rank or boss milestone
	Coins     Effect = "coins"     // something bought, crafted or claimed
)

var effectSounds = map[Effect]string{
	Right:     "Tink",
	Wrong:     "Basso",
	Victory:   "Glass",
	Defeat:    "Sosumi",
	Encounter: "Submarine",
	LevelUp:   "Hero",
	Coins:     "Purr",
}

// playSound starts a sound file without waiting for it; tests replace it.
var playSound = func(path string) {
	player, err := exec.LookPath("afplay")
	if err != nil {
		return
	}
	cmd := exec.Command(player, path)
	if cmd.Start() == nil {
		go cmd.Wait() //nolint:errcheck // a sound that fails to finish doesn't matter
	}
}

// soundDir is where the system sounds live.
var soundDir = "/System/Library/Sounds"

// Play plays an effect's sound in the background, if the system has it.
func Play(e Effect) {
	name, ok := effectSounds[e]
	if !ok {
		return
	}
	path := soundDir + "/" + name + ".aiff"
	if _, err := os.Stat(path); err != nil {
		return
	}
	playSound(path)
}
