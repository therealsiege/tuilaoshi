// Package imgview renders images in the TUI. It prefers chafa, which turns
// an image into colored Unicode art that works in any terminal and inside
// Bubble Tea views, and falls back to go-termimg. Results are cached, since
// views redraw often. All functions are safe to call with missing files;
// they return empty strings on failure.
package imgview

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/blacktop/go-termimg"
)

var (
	detectedProto     termimg.Protocol
	detectedProtoOnce sync.Once

	cacheMu sync.Mutex
	cache   = map[cacheKey]string{}
)

// cacheKey identifies a rendering: the file as last modified, at a size.
type cacheKey struct {
	path          string
	mod           time.Time
	width, height int
}

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

// chafaPath is the chafa binary, or "" when it isn't installed; tests replace it.
var chafaPath = func() string {
	p, err := exec.LookPath("chafa")
	if err != nil {
		return ""
	}
	return p
}()

// Render renders an image file to a string for use in a Bubble Tea View(),
// at most width columns by height rows. Returns empty string if the file
// doesn't exist or rendering fails.
func Render(path string, width, height int) string {
	if path == "" {
		return ""
	}
	info, err := os.Stat(path)
	if err != nil {
		return ""
	}
	key := cacheKey{path, info.ModTime(), width, height}
	cacheMu.Lock()
	out, ok := cache[key]
	cacheMu.Unlock()
	if ok {
		return out
	}
	if chafaPath != "" {
		out = renderChafa(path, width, height)
	}
	if out == "" {
		out = renderTermimg(path, width, height)
	}
	cacheMu.Lock()
	cache[key] = out
	cacheMu.Unlock()
	return out
}

// renderChafa draws the image as colored Unicode symbols, which Bubble Tea
// can lay out like any text.
func renderChafa(path string, width, height int) string {
	// Name the symbols: chafa guesses from TERM and falls back to ASCII
	// under terminals it doesn't know (tmux, mosh). Every shape except
	// letters, digits and double-width glyphs gives the sharpest picture
	// that still lines up; work 9 is chafa's most careful matching.
	out, err := exec.Command(chafaPath, "--format", "symbols", "--symbols", "all-ascii-latin-alnum-wide",
		"--work", "9", "--colors", "full", "--size", fmt.Sprintf("%dx%d", width, height), "--animate", "off", path).Output()
	if err != nil {
		return ""
	}
	// chafa may wrap its output in cursor show/hide codes; views only want the art.
	s := strings.NewReplacer("\x1b[?25l", "", "\x1b[?25h", "").Replace(string(out))
	// End each line with a reset so the art's colors don't bleed into the
	// next line of the view.
	lines := strings.Split(strings.TrimRight(s, "\n"), "\n")
	for i := range lines {
		lines[i] += "\x1b[0m"
	}
	return strings.Join(lines, "\n")
}

func renderTermimg(path string, width, height int) string {
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
