package audio

import (
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// PlayClip plays part of an audio file, such as one line of an episode,
// with ffplay. It blocks until the clip ends.
func PlayClip(path string, start, end time.Duration) error {
	player, err := exec.LookPath("ffplay")
	if err != nil {
		return fmt.Errorf("playing original audio needs ffplay (brew install ffmpeg)")
	}
	out, err := exec.Command(player, "-nodisp", "-autoexit", "-loglevel", "error",
		"-ss", fmt.Sprintf("%.3f", start.Seconds()), "-t", fmt.Sprintf("%.3f", (end-start).Seconds()), path).CombinedOutput()
	if err != nil {
		return fmt.Errorf("ffplay: %v %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}
