// Command artgen paints TuiLaoShi's art from the art bible (bible.yaml) and
// the game's own content, so every picture is accurate to the game and
// consistent with the rest:
//
//	artgen -kind portrait -only 孙悟空 -print   show the prompt, paint nothing
//	artgen -kind map -only baihu                paint White Tiger Ridge from its tile map
//	artgen -kind vista                          paint every region's title card
//	artgen -kind scene -only baigu-strike       paint one battle cut scene
//	artgen -kind sheet -only 猪八戒              paint one companion's model sheet
//
// Pictures go to OUT/DIRECTION/KIND/NAME.png with the prompt beside them as
// NAME.txt; existing pictures are kept unless -force. A character with an
// approved picture in OUT/refs/NAME.png is painted with it attached, so they
// keep one design; in a scene, their model sheet in OUT/DIRECTION/sheet/NAME.png
// is attached too, when there is one. Keys are read from the env file named by
// TUILAOSHI_ART_ENV, set in the environment or in ~/.config/tuilaoshi/env
// (else that file itself): only GOOGLE_API_KEY, and IMAGE_MODEL if set. A
// GOOGLE_API_KEY in ~/.config/tuilaoshi/env wins: the game's key pays. Nothing painted belongs in git until it is approved.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/therealsiege/tuilaoshi/internal/journey"
)

const castNote = "The attached pictures, in order, are the approved designs of %s. Paint them small but recognisable: the same costume, colours and weapon."

const peerNote = "The other attached pictures are approved portraits of other characters, for style only: match their painting style, level of realism, adult proportions, lighting and framing, but do not copy their faces or costumes."

const keepNote = "The attached pictures, in order, are the approved designs of %s. Keep each one's face, build, colours and weapon exactly the same, and their costume too unless the scene says what they wear at this point in the story; change only the pose, place and light."

const refNote = "The attached picture is the approved design of this character. Keep their face, build, costume, colours and weapon exactly the same; change only the pose, place and light this prompt asks for."

const sheetRefNote = "The attached picture is the approved design of this character, the source of truth for this model sheet. Every view must match it exactly: the same face, build, costume, colours and weapon. Where this prompt states a detail precisely (how many teeth or rings, where the bands are, the shape of the ends), paint it as the prompt says."

const sheetNote = "After those, the attached model sheets, in order, show %s from the front, three-quarter, side and back, with close-ups of the face and weapon. Keep each one's design exactly as on their model sheet: the same face, build, colours and weapon, every costume piece, and the weapon's shape, ends, bands and count of teeth or rings; their costume too unless the scene says what they wear at this point in the story. Where a sheet and a portrait differ, follow the sheet."

// sheetWidth is the width model sheets are shrunk to before they are
// attached, so a scene with the whole party stays under the request limit.
const sheetWidth = 1024

func main() {
	home, err := os.UserHomeDir()
	check(err)
	kind := flag.String("kind", "", "portrait, vista, map, town, scene or sheet")
	only := flag.String("only", "", "comma-separated names or ids (default: every one of the kind)")
	direction := flag.String("direction", "", "real, clean or blend (default: the bible's)")
	dry := flag.Bool("print", false, "print the prompts and paint nothing")
	out := flag.String("out", filepath.Join(home, "TuiLaoShi art"), "folder for the pictures, outside git")
	size := flag.String("size", "2K", "image size: 1K, 2K or 4K")
	force := flag.Bool("force", false, "paint again even if the picture exists")
	flag.Parse()

	b, err := loadBible()
	check(err)
	c, err := journey.LoadContent()
	check(err)
	if *direction == "" {
		*direction = b.Direction
	}
	if _, ok := b.Directions[*direction]; !ok {
		check(fmt.Errorf("no direction %q in bible.yaml", *direction))
	}
	p := composer{b: b, c: c, direction: *direction}

	names := p.all(*kind)
	if *only != "" {
		names = strings.Split(*only, ",")
	}
	if len(names) == 0 {
		_, err := p.job(*kind, "")
		check(err)
	}
	var jobs []Job
	for _, n := range names {
		j, err := p.job(*kind, strings.TrimSpace(n))
		check(err)
		jobs = append(jobs, j)
	}
	if *dry {
		for _, j := range jobs {
			fmt.Printf("=== %s %s (%s)\n%s\n\n", j.Kind, j.Name, j.Ratio, j.Prompt)
		}
		return
	}

	// The keys live in TUILAOSHI_ART_ENV, which the environment or the
	// game's own settings file may name; else in that settings file.
	envFile := os.Getenv("TUILAOSHI_ART_ENV")
	settings := filepath.Join(home, ".config", "tuilaoshi", "env")
	if envFile == "" {
		if s, err := readKeys(settings, "TUILAOSHI_ART_ENV"); err == nil {
			envFile = s["TUILAOSHI_ART_ENV"]
		}
	}
	if envFile == "" {
		envFile = settings
	}
	keys, err := readKeys(envFile, "GOOGLE_API_KEY", "IMAGE_MODEL")
	check(err)
	// The game's own Google key, when it has one, pays for painting.
	if own, err := readKeys(settings, "GOOGLE_API_KEY"); err == nil && own["GOOGLE_API_KEY"] != "" {
		keys["GOOGLE_API_KEY"] = own["GOOGLE_API_KEY"]
	}
	if keys["GOOGLE_API_KEY"] == "" {
		check(fmt.Errorf("no GOOGLE_API_KEY in %s or %s", settings, envFile))
	}
	model := keys["IMAGE_MODEL"]
	if model == "" {
		model = "gemini-3-pro-image"
	}
	pt := newPainter(keys["GOOGLE_API_KEY"], model, *size)

	dir := filepath.Join(*out, *direction, *kind)
	check(os.MkdirAll(dir, 0o755))
	failed := 0
	for _, j := range jobs {
		path := filepath.Join(dir, j.Name+".png")
		if _, err := os.Stat(path); err == nil && !*force {
			fmt.Printf("%s: kept (use -force to paint again)\n", j.Name)
			continue
		}
		prompt, refs := attach(j, *out, *direction)
		start := time.Now()
		img, err := pt.paint(prompt, j.Ratio, refs)
		if err != nil {
			fmt.Printf("%s: %v\n", j.Name, err)
			failed++
			continue
		}
		check(os.WriteFile(path, img, 0o644))
		check(os.WriteFile(strings.TrimSuffix(path, ".png")+".txt", []byte(prompt+"\n"), 0o644))
		fmt.Printf("%s: %d KB in %s\n", j.Name, len(img)/1024, time.Since(start).Round(time.Second))
	}
	if failed > 0 {
		os.Exit(1)
	}
}

// attach adds the pictures a job is painted with, after its prompt, and the
// notes that say what each is for: the map layout, the subject's approved
// picture, style peers, the cast of a vista, and the cast of a scene with
// their approved pictures and then their model sheets.
func attach(j Job, out, direction string) (string, [][]byte) {
	prompt, refs := j.Prompt, [][]byte(nil)
	ref := func(name string) ([]byte, error) { return os.ReadFile(filepath.Join(out, "refs", name+".png")) }
	if j.Layout != nil {
		png, err := layoutPNG(*j.Layout, 24)
		check(err)
		refs = append(refs, png)
	}
	if j.Subject != "" {
		if r, err := ref(j.Subject); err == nil {
			note := refNote
			if j.Kind == "sheet" {
				note = sheetRefNote
			}
			prompt += "\n\n" + note
			refs = append(refs, r)
		}
	}
	var peers []string
	for _, n := range j.Peers {
		if len(peers) == 2 {
			break
		}
		if r, err := ref(n); err == nil {
			peers = append(peers, n)
			refs = append(refs, r)
		}
	}
	if len(peers) > 0 {
		prompt += "\n\n" + peerNote
	}
	var shown []string
	for _, c := range j.Cast {
		if r, err := ref(c.Name); err == nil {
			shown = append(shown, c.English)
			refs = append(refs, r)
		}
	}
	if len(shown) > 0 {
		prompt += "\n\n" + fmt.Sprintf(castNote, strings.Join(shown, ", "))
	}
	var kept []string
	for _, f := range j.Keep {
		if r, err := ref(f.Name); err == nil {
			kept = append(kept, f.English)
			refs = append(refs, r)
		}
	}
	if len(kept) > 0 {
		prompt += "\n\n" + fmt.Sprintf(keepNote, strings.Join(kept, ", "))
	}
	var sheets []string
	for _, f := range j.Keep {
		raw, err := os.ReadFile(filepath.Join(out, direction, "sheet", f.Name+".png"))
		if err != nil {
			continue
		}
		small, err := shrink(raw, sheetWidth)
		if err != nil {
			fmt.Fprintf(os.Stderr, "artgen: model sheet of %s: %v\n", f.Name, err)
			continue
		}
		sheets = append(sheets, f.English)
		refs = append(refs, small)
	}
	if len(sheets) > 0 {
		prompt += "\n\n" + fmt.Sprintf(sheetNote, strings.Join(sheets, ", "))
	}
	return prompt, refs
}

// shrink re-encodes a PNG at most width pixels wide, averaging each block
// of pixels, so large pictures can be attached without overflowing the
// request.
func shrink(data []byte, width int) ([]byte, error) {
	src, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	b := src.Bounds()
	if b.Dx() <= width {
		return data, nil
	}
	height := b.Dy() * width / b.Dx()
	dst := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		y0, y1 := b.Min.Y+y*b.Dy()/height, b.Min.Y+(y+1)*b.Dy()/height
		for x := 0; x < width; x++ {
			x0, x1 := b.Min.X+x*b.Dx()/width, b.Min.X+(x+1)*b.Dx()/width
			var r, g, bl, a, n uint64
			for sy := y0; sy < y1; sy++ {
				for sx := x0; sx < x1; sx++ {
					cr, cg, cb, ca := src.At(sx, sy).RGBA()
					r, g, bl, a, n = r+uint64(cr), g+uint64(cg), bl+uint64(cb), a+uint64(ca), n+1
				}
			}
			i := dst.PixOffset(x, y)
			dst.Pix[i], dst.Pix[i+1], dst.Pix[i+2], dst.Pix[i+3] = uint8(r/n>>8), uint8(g/n>>8), uint8(bl/n>>8), uint8(a/n>>8)
		}
	}
	var buf bytes.Buffer
	err = png.Encode(&buf, dst)
	return buf.Bytes(), err
}

func check(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, "artgen:", err)
		os.Exit(1)
	}
}
