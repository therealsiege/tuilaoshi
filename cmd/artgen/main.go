// Command artgen paints TuiLaoShi's art from the art bible (bible.yaml) and
// the game's own content, so every picture is accurate to the game and
// consistent with the rest:
//
//	artgen -kind portrait -only 孙悟空 -print   show the prompt, paint nothing
//	artgen -kind map -only baihu                paint White Tiger Ridge from its tile map
//	artgen -kind vista                          paint every region's title card
//
// Pictures go to OUT/DIRECTION/KIND/NAME.png with the prompt beside them as
// NAME.txt; existing pictures are kept unless -force. A character with an
// approved picture in OUT/refs/NAME.png is painted with it attached, so they
// keep one design. Keys are read from the env file named by
// TUILAOSHI_ART_ENV, set in the environment or in ~/.config/tuilaoshi/env
// (else that file itself): only GOOGLE_API_KEY, and IMAGE_MODEL if set. A
// GOOGLE_API_KEY in ~/.config/tuilaoshi/env wins: the game's key pays. Nothing painted belongs in git until it is approved.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/therealsiege/tuilaoshi/internal/journey"
)

const castNote = "The attached pictures, in order, are the approved designs of %s. Paint them small but recognisable: the same costume, colours and weapon."

const refNote = "The attached picture is the approved design of this character. Keep their face, build, costume, colours and weapon exactly the same; change only the pose, place and light this prompt asks for."

func main() {
	home, err := os.UserHomeDir()
	check(err)
	kind := flag.String("kind", "", "portrait, vista, map or town")
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
		prompt, refs := j.Prompt, [][]byte(nil)
		if j.Layout != nil {
			png, err := layoutPNG(*j.Layout, 24)
			check(err)
			refs = append(refs, png)
		}
		if j.Subject != "" {
			if ref, err := os.ReadFile(filepath.Join(*out, "refs", j.Subject+".png")); err == nil {
				prompt += "\n\n" + refNote
				refs = append(refs, ref)
			}
		}
		var shown []string
		for _, c := range j.Cast {
			if ref, err := os.ReadFile(filepath.Join(*out, "refs", c.Name+".png")); err == nil {
				shown = append(shown, c.English)
				refs = append(refs, ref)
			}
		}
		if len(shown) > 0 {
			prompt += "\n\n" + fmt.Sprintf(castNote, strings.Join(shown, ", "))
		}
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

func check(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, "artgen:", err)
		os.Exit(1)
	}
}
