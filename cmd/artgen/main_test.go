package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// testPNG is a w by h picture of one colour.
func testPNG(t *testing.T, w, h int, c color.RGBA) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, c)
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func writeFile(t *testing.T, path string, data []byte) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
}

// Every companion has a model sheet to paint: wide, kept to their approved
// picture, and carrying the weapon details their sheet states.
func TestSheetsDrawEveryCompanion(t *testing.T) {
	p := testComposer(t)
	names := p.all("sheet")
	if len(names) != len(p.c.Companions) {
		t.Fatalf("sheets %v, want one per companion", names)
	}
	j, err := p.job("sheet", "猪八戒")
	if err != nil {
		t.Fatal(err)
	}
	if j.Ratio != "16:9" || j.Subject != "猪八戒" || j.Name != "猪八戒" {
		t.Errorf("sheet job = %+v", j)
	}
	for _, must := range []string{p.b.Kinds["sheet"].Frame, "exactly nine long, straight iron teeth", "from the back"} {
		if !strings.Contains(j.Prompt, strings.TrimSpace(must)) {
			t.Errorf("Pigsy's sheet is missing %q", must)
		}
	}
	if _, err := p.job("sheet", "nobody"); err == nil {
		t.Error("a sheet for someone with no character sheet should be an error")
	}
}

// A model sheet is painted from the approved picture as its source of truth.
func TestSheetsAttachTheApprovedPicture(t *testing.T) {
	p := testComposer(t)
	out := t.TempDir()
	ref := testPNG(t, 4, 4, color.RGBA{200, 30, 30, 255})
	writeFile(t, filepath.Join(out, "refs", "沙悟净.png"), ref)
	j, err := p.job("sheet", "沙悟净")
	if err != nil {
		t.Fatal(err)
	}
	prompt, refs := attach(j, out, "clean")
	if len(refs) != 1 || !bytes.Equal(refs[0], ref) {
		t.Fatalf("got %d pictures, want the approved one", len(refs))
	}
	if !strings.Contains(prompt, sheetRefNote) || strings.Contains(prompt, refNote) {
		t.Errorf("a sheet should carry the sheet note, not the portrait one:\n%s", prompt)
	}
}

// A scene attaches its cast's approved pictures, then the model sheets of
// those who have one, shrunk, each in cast order and named in its note.
func TestScenesAttachModelSheets(t *testing.T) {
	p := testComposer(t)
	out := t.TempDir()
	tang := testPNG(t, 4, 4, color.RGBA{200, 30, 30, 255})
	writeFile(t, filepath.Join(out, "refs", "唐僧.png"), tang)
	writeFile(t, filepath.Join(out, "clean", "sheet", "孙悟空.png"), testPNG(t, 2*sheetWidth, 20, color.RGBA{220, 170, 40, 255}))
	writeFile(t, filepath.Join(out, "clean", "sheet", "唐僧.png"), testPNG(t, 16, 9, color.RGBA{150, 20, 20, 255}))
	writeFile(t, filepath.Join(out, "real", "sheet", "白骨精.png"), testPNG(t, 16, 9, color.RGBA{1, 2, 3, 255}))

	j, err := p.job("scene", "baigu-strike") // cast: 孙悟空, 白骨精, 唐僧
	if err != nil {
		t.Fatal(err)
	}
	prompt, refs := attach(j, out, "clean")
	if len(refs) != 3 {
		t.Fatalf("got %d pictures, want Tang's ref then Wukong's and Tang's sheets", len(refs))
	}
	if !bytes.Equal(refs[0], tang) {
		t.Error("the approved pictures should come first")
	}
	wukong, err := png.DecodeConfig(bytes.NewReader(refs[1]))
	if err != nil {
		t.Fatal(err)
	}
	if wukong.Width != sheetWidth || wukong.Height != 10 {
		t.Errorf("Wukong's sheet is %dx%d, want it shrunk to %dx10", wukong.Width, wukong.Height, sheetWidth)
	}
	if small, _ := png.DecodeConfig(bytes.NewReader(refs[2])); small.Width != 16 {
		t.Errorf("a small sheet should be attached as it is, got width %d", small.Width)
	}
	for _, must := range []string{fmt.Sprintf(keepNote, "Tang Sanzang"), fmt.Sprintf(sheetNote, "Sun Wukong, Tang Sanzang")} {
		if !strings.Contains(prompt, must) {
			t.Errorf("prompt is missing %q", must)
		}
	}
	if strings.Index(prompt, "approved designs of Tang") > strings.Index(prompt, "model sheets, in order") {
		t.Error("the sheet note should follow the note for the approved pictures")
	}

	// With no sheets painted, nothing is said about them.
	prompt, refs = attach(j, t.TempDir(), "clean")
	if len(refs) != 0 || strings.Contains(prompt, "model sheet") {
		t.Errorf("no pictures should mean no notes, got %d and:\n%s", len(refs), prompt)
	}
}

func TestShrinkAveragesAndKeepsColour(t *testing.T) {
	big := testPNG(t, 40, 20, color.RGBA{10, 120, 240, 255})
	small, err := shrink(big, 10)
	if err != nil {
		t.Fatal(err)
	}
	img, err := png.Decode(bytes.NewReader(small))
	if err != nil {
		t.Fatal(err)
	}
	if b := img.Bounds(); b.Dx() != 10 || b.Dy() != 5 {
		t.Fatalf("shrunk to %v, want 10x5", b)
	}
	if r, g, b, _ := img.At(3, 2).RGBA(); r>>8 != 10 || g>>8 != 120 || b>>8 != 240 {
		t.Errorf("colour changed to %d,%d,%d", r>>8, g>>8, b>>8)
	}
	if same, _ := shrink(big, 100); !bytes.Equal(same, big) {
		t.Error("a picture already narrow enough should be kept as it is")
	}
	if _, err := shrink([]byte("not a png"), 10); err == nil {
		t.Error("a broken picture should be an error")
	}
}
