package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"
)

// Char holds all data needed to generate a character wiki page.
type Char struct {
	Hanzi      string
	Pinyin     string // diacritical: dà
	PinyinNum  string // numbered: da4
	Meaning    string // short English
	MeaningCap string // capitalized for heading
	Radical    string // e.g. 大 (big)
	Strokes    int
	Level      string // HSK1–HSK6
	Tags       string // comma-separated inside brackets
	Compounds  []Compound
	Sentences  []Sentence
	Notes      string
	Filename   string // override filename if needed (for disambiguation)
}

// Compound is a word containing the character.
type Compound struct {
	Hanzi   string
	Pinyin  string
	Meaning string
}

// Sentence is an example sentence.
type Sentence struct {
	Chinese string
	Pinyin  string
	English string
}

const pageTmpl = `---
type: character
level: {{.Level}}
status: active
tags: [{{.Tags}}]
introduced: {{.Today}}
review_interval: 1
next_review: {{.Tomorrow}}
---

# {{.Hanzi}} {{.MeaningCap}}

- **Character**: {{.Hanzi}}
- **Pinyin**: {{.Pinyin}}
- **Meaning**: {{.Meaning}}
- **Radical**: {{.Radical}}
- **Stroke count**: {{.Strokes}}

## Compounds
{{range .Compounds}}
- {{.Hanzi}} ({{.Pinyin}}) — {{.Meaning}}
{{- end}}

## Example Sentences
{{range .Sentences}}
- {{.Chinese}}({{.Pinyin}}) — {{.English}}
{{- end}}

## Notes

{{.Notes}}
`

type tmplData struct {
	Char
	Today    string
	Tomorrow string
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// meaningSlug returns the first word of the meaning, lowercased, for filename disambiguation.
func meaningSlug(meaning string) string {
	m := strings.ToLower(meaning)
	// Strip leading "to " for verbs
	m = strings.TrimPrefix(m, "to ")
	// Take first word only
	if idx := strings.IndexAny(m, " ,;/"); idx > 0 {
		m = m[:idx]
	}
	// Remove non-alphanumeric
	var sb strings.Builder
	for _, r := range m {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			sb.WriteRune(r)
		}
	}
	return sb.String()
}

func main() {
	outDir := filepath.Join("wiki", "characters")
	if len(os.Args) > 1 {
		outDir = os.Args[1]
	}

	if err := os.MkdirAll(outDir, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "cannot create output dir: %v\n", err)
		os.Exit(1)
	}

	t := template.Must(template.New("page").Parse(pageTmpl))

	today := time.Now().Format("2006-01-02")
	tomorrow := time.Now().Add(24 * time.Hour).Format("2006-01-02")

	var generated, skipped, errored int

	allChars := AllCharacters()
	fmt.Printf("Total characters in data: %d\n", len(allChars))

	// Track used filenames to disambiguate collisions
	used := make(map[string]bool)

	for _, ch := range allChars {
		fname := ch.Filename
		if fname == "" {
			fname = ch.PinyinNum + ".md"
		}
		if !strings.HasSuffix(fname, ".md") {
			fname += ".md"
		}
		// Strip slashes from filenames (some entries have "yan1/yan4" style PinyinNum)
		fname = strings.ReplaceAll(fname, "/", "-")

		// Disambiguate in-run collisions only
		if used[fname] {
			base := strings.TrimSuffix(fname, ".md")
			slug := meaningSlug(ch.Meaning)
			if slug != "" {
				fname = base + "-" + slug + ".md"
			}
			// Still colliding? append counter
			if used[fname] {
				for n := 2; n < 100; n++ {
					var candidate string
					if slug != "" {
						candidate = fmt.Sprintf("%s-%s-%d.md", base, slug, n)
					} else {
						candidate = fmt.Sprintf("%s-%d.md", base, n)
					}
					if !used[candidate] {
						fname = candidate
						break
					}
				}
			}
		}
		used[fname] = true
		outPath := filepath.Join(outDir, fname)

		if fileExists(outPath) {
			skipped++
			continue
		}

		data := tmplData{
			Char:     ch,
			Today:    today,
			Tomorrow: tomorrow,
		}

		f, err := os.Create(outPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error creating %s: %v\n", outPath, err)
			errored++
			continue
		}

		if err := t.Execute(f, data); err != nil {
			fmt.Fprintf(os.Stderr, "error writing %s: %v\n", outPath, err)
			f.Close()
			errored++
			continue
		}
		f.Close()
		generated++
	}

	fmt.Printf("\nResults:\n")
	fmt.Printf("  Generated: %d\n", generated)
	fmt.Printf("  Skipped (existing): %d\n", skipped)
	fmt.Printf("  Errors: %d\n", errored)
	fmt.Printf("  Total files now: %d\n", generated+skipped)
}
