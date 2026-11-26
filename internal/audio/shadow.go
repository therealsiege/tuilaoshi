package audio

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Shadowing records the learner saying a line back and transcribes it on
// this Mac with whisper.cpp (whisper-cli), so nothing leaves the machine.
// Recording uses ffmpeg from the default microphone; macOS asks once for
// microphone access for the terminal.

// WhisperModelEnv names a whisper.cpp model file to use instead of the
// first ggml-*.bin found in ~/.local/share/whisper.
const WhisperModelEnv = "TUILAOSHI_WHISPER_MODEL"

// whisperPrompt steers transcription to simplified Mandarin.
const whisperPrompt = "以下是普通话的句子。"

// WhisperModel returns the whisper.cpp model to use, or "" when none is
// installed. Larger models are preferred: they hear Mandarin far better.
func WhisperModel() string {
	if m := os.Getenv(WhisperModelEnv); m != "" {
		return m
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	found, err := filepath.Glob(filepath.Join(home, ".local", "share", "whisper", "ggml-*.bin"))
	if err != nil || len(found) == 0 {
		return ""
	}
	rank := func(p string) int {
		for i, size := range []string{"large", "medium", "small", "base", "tiny"} {
			if strings.Contains(filepath.Base(p), size) {
				return i
			}
		}
		return 5
	}
	sort.Slice(found, func(i, j int) bool { return rank(found[i]) < rank(found[j]) })
	return found[0]
}

// ShadowReady says what shadowing still needs, or "" when it can run.
func ShadowReady() string {
	switch {
	case !onPath("ffmpeg"):
		return "Shadowing needs ffmpeg to record (brew install ffmpeg)."
	case !onPath("whisper-cli"):
		return "Shadowing needs whisper.cpp to listen (brew install whisper-cpp)."
	case WhisperModel() == "":
		return "Shadowing needs a whisper.cpp model in ~/.local/share/whisper (e.g. ggml-large-v3-turbo-q5_0.bin)."
	}
	return ""
}

func onPath(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

// Record records d of audio from the default microphone to a 16 kHz mono
// WAV file at path, as whisper.cpp wants it.
func Record(path string, d time.Duration) error {
	out, err := exec.Command("ffmpeg", "-hide_banner", "-loglevel", "error", "-y",
		"-f", "avfoundation", "-i", ":default", "-t", fmt.Sprintf("%.2f", d.Seconds()),
		"-ar", "16000", "-ac", "1", path).CombinedOutput()
	if err != nil {
		msg := strings.TrimSpace(string(out))
		if strings.Contains(msg, "Input/output error") {
			msg = "no microphone: allow the terminal in System Settings, Privacy & Security, Microphone"
		}
		return fmt.Errorf("recording: %s", msg)
	}
	return nil
}

// Transcribe returns the Mandarin heard in a WAV file.
func Transcribe(path string) (string, error) {
	model := WhisperModel()
	if model == "" {
		return "", errors.New("no whisper.cpp model installed")
	}
	out, err := exec.Command("whisper-cli", "-m", model, "-l", "zh", "-nt", "-np",
		"--prompt", whisperPrompt, "-f", path).Output()
	if err != nil {
		return "", fmt.Errorf("whisper: %w", err)
	}
	return strings.TrimSpace(strings.ReplaceAll(string(out), "\n", "")), nil
}

// Shadow records d from the microphone and returns what was heard,
// keeping the recording only while it is transcribed.
func Shadow(d time.Duration) (string, error) {
	dir, err := os.MkdirTemp("", "tuilaoshi-shadow-")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(dir) //nolint:errcheck // a temp folder; nothing to do if it lingers
	path := filepath.Join(dir, "say.wav")
	if err := Record(path, d); err != nil {
		return "", err
	}
	return Transcribe(path)
}
