// Package imgview wraps go-termimg for rendering images in the TUI.
// It auto-detects the best terminal protocol (Kitty, Sixel, iTerm2)
// and falls back to Unicode halfblocks. All functions are safe to call
// with missing files — they return empty strings on failure.
package imgview

import (
	"os"
	"sync"

	"github.com/blacktop/go-termimg"
)

var (
	detectedProto     termimg.Protocol
	detectedProtoOnce sync.Once
)

func detectProto() termimg.Protocol {
	detectedProtoOnce.Do(func() {
		detectedProto = termimg.DetectProtocol()
	})
	return detectedProto
}

// Available returns true if the terminal supports a graphics protocol
// beyond basic halfblock fallback (i.e. Kitty, Sixel, or iTerm2).
func Available() bool {
	p := detectProto()
	return p == termimg.Kitty || p == termimg.ITerm2 || p == termimg.Sixel
}

// Render renders an image file to a string for use in a Bubble Tea View().
// Returns empty string if the file doesn't exist or rendering fails.
func Render(path string, width, height int) string {
	if path == "" {
		return ""
	}
	if _, err := os.Stat(path); err != nil {
		return ""
	}
	img, err := termimg.Open(path)
	if err != nil {
		return ""
	}
	result, err := img.Size(width, height).Scale(termimg.ScaleFit).Render()
	if err != nil {
		return ""
	}
	return result
}

// RenderWithFallback tries to render an image file. If the file doesn't
// exist or rendering fails, it returns the fallback string (typically ASCII art).
func RenderWithFallback(path string, width, height int, fallback string) string {
	rendered := Render(path, width, height)
	if rendered == "" {
		return fallback
	}
	return rendered
}
