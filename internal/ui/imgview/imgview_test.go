package imgview

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

func testImage(t *testing.T) string {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 64, 64))
	for x := 0; x < 64; x++ {
		for y := 0; y < 64; y++ {
			img.Set(x, y, color.RGBA{uint8(x * 4), 80, uint8(y * 4), 255})
		}
	}
	path := filepath.Join(t.TempDir(), "demon.png")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := png.Encode(f, img); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestRenderUsesChafaAndCaches(t *testing.T) {
	if chafaPath == "" {
		t.Skip("chafa not installed")
	}
	path := testImage(t)
	out := Render(path, 20, 8)
	if out == "" || strings.Contains(out, "\x1b[?25l") {
		t.Fatalf("chafa render %q", out)
	}
	if lines := strings.Count(out, "\n") + 1; lines > 8 {
		t.Errorf("%d lines, want at most 8", lines)
	}
	if again := Render(path, 20, 8); again != out {
		t.Error("second render differs; want the cached art")
	}
}

func TestRenderFallsBackWithoutChafa(t *testing.T) {
	orig := chafaPath
	t.Cleanup(func() { chafaPath = orig })
	chafaPath = ""
	if out := Render(testImage(t), 20, 8); out == "" {
		t.Error("no render without chafa; want the go-termimg fallback")
	}
	if got := RenderWithFallback("/no/such.png", 20, 8, "ART"); got != "ART" {
		t.Errorf("missing file: %q, want the fallback art", got)
	}
}

func TestChafaLinesResetColors(t *testing.T) {
	if chafaPath == "" {
		t.Skip("chafa not installed")
	}
	for _, l := range strings.Split(Render(testImage(t), 12, 4), "\n") {
		if !strings.HasSuffix(l, "\x1b[0m") {
			t.Fatalf("line %q doesn't end with a color reset", l)
		}
	}
}

func TestChafaArtStaysWithinItsWidth(t *testing.T) {
	if chafaPath == "" {
		t.Skip("chafa not installed")
	}
	path := filepath.Join("..", "..", "..", "assets", "images", "demons", "红孩儿.png")
	if _, err := os.Stat(path); err != nil {
		t.Skip("no portraits")
	}
	for _, l := range strings.Split(Render(path, 44, 22), "\n") {
		if w := lipgloss.Width(l); w > 44 {
			t.Fatalf("line is %d columns, wider than 44: %q", w, l)
		}
		if strings.IndexFunc(ansi.Strip(l), func(r rune) bool { return r < 128 && r != ' ' }) >= 0 {
			t.Fatalf("line has ASCII characters: %q", ansi.Strip(l))
		}
	}
}
