// Command proofsheet writes every Chinese line of the game's content (the
// embedded YAML in internal/journey/content) to one markdown review sheet,
// each beside its English and a blank column for a native speaker's notes.
//
//	go run ./cmd/proofsheet -out ~/Desktop/proofsheet.md
package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode"

	"gopkg.in/yaml.v3"
)

func main() {
	dir := flag.String("content", "internal/journey/content", "the content folder")
	out := flag.String("out", "", "where to write the sheet (default: standard output)")
	flag.Parse()
	log.SetFlags(0)
	var b strings.Builder
	n, err := write(&b, *dir)
	if err != nil {
		fail(err)
	}
	if *out == "" {
		if _, err := io.WriteString(os.Stdout, b.String()); err != nil {
			fail(err)
		}
		return
	}
	if err := os.WriteFile(*out, []byte(b.String()), 0644); err != nil {
		fail(err)
	}
	log.Printf("%d lines written to %s", n, *out)
}

func fail(err error) { log.Fatal("proofsheet: ", err) }

// line is one piece of Chinese to check: where it is, and its English.
type line struct {
	where, chinese, english string
}

// write writes the sheet for every YAML file in dir and returns how many
// lines it holds.
func write(w *strings.Builder, dir string) (int, error) {
	files, err := filepath.Glob(filepath.Join(dir, "*.yaml"))
	if err != nil {
		return 0, err
	}
	sort.Strings(files)
	fmt.Fprintln(w, "# TuiLaoShi: Chinese to proofread")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Every Chinese line in the game's content, by file. Mark anything unnatural, wrong or odd for the scene in the last column.")
	total := 0
	for _, path := range files {
		data, err := os.ReadFile(path)
		if err != nil {
			return total, err
		}
		var root yaml.Node
		if err := yaml.Unmarshal(data, &root); err != nil {
			return total, fmt.Errorf("%s: %w", path, err)
		}
		var lines []line
		collect(&root, "", &lines)
		if len(lines) == 0 {
			continue
		}
		fmt.Fprintf(w, "\n## %s (%d)\n\n| # | Where | 中文 | English | Note |\n|---|---|---|---|---|\n", filepath.Base(path), len(lines))
		for i, l := range lines {
			fmt.Fprintf(w, "| %d | %s | %s | %s | |\n", i+1, l.where, cell(l.chinese), cell(l.english))
		}
		total += len(lines)
	}
	return total, nil
}

func cell(s string) string {
	return strings.ReplaceAll(strings.ReplaceAll(strings.TrimSpace(s), "|", "\\|"), "\n", "<br>")
}

// hasHan reports whether s is Chinese text: it has Chinese characters, and
// isn't an English sentence quoting a few (an origin that names 盘丝洞).
func hasHan(s string) bool {
	han, latin := 0, 0
	for _, r := range s {
		switch {
		case unicode.Is(unicode.Han, r):
			han++
		case r < unicode.MaxASCII && unicode.IsLetter(r):
			latin++
		}
	}
	return han > 0 && han*3 >= latin
}

// collect walks a YAML node, gathering each Chinese string with the
// English its siblings give it: key_english, or english for text, name,
// line and question, or the same item of english_choices for choices.
func collect(n *yaml.Node, where string, out *[]line) {
	switch n.Kind {
	case yaml.DocumentNode:
		for _, c := range n.Content {
			collect(c, where, out)
		}
	case yaml.SequenceNode:
		for i, c := range n.Content {
			collect(c, fmt.Sprintf("%s[%d]", where, i), out)
		}
	case yaml.MappingNode:
		fields := map[string]*yaml.Node{}
		for i := 0; i+1 < len(n.Content); i += 2 {
			fields[n.Content[i].Value] = n.Content[i+1]
		}
		for i := 0; i+1 < len(n.Content); i += 2 {
			k, v := n.Content[i].Value, n.Content[i+1]
			at := strings.TrimPrefix(where+"."+k, ".")
			if strings.HasSuffix(k, "english") || k == "art" {
				continue
			}
			switch {
			case v.Kind == yaml.ScalarNode && hasHan(v.Value):
				*out = append(*out, line{at, v.Value, english(fields, k)})
			case v.Kind == yaml.SequenceNode && k == "choices" && fields["english_choices"] != nil:
				en := fields["english_choices"].Content
				for j, c := range v.Content {
					if c.Kind == yaml.ScalarNode && hasHan(c.Value) {
						e := ""
						if j < len(en) {
							e = en[j].Value
						}
						*out = append(*out, line{fmt.Sprintf("%s[%d]", at, j), c.Value, e})
					}
				}
			default:
				collect(v, at, out)
			}
		}
	}
}

func english(fields map[string]*yaml.Node, key string) string {
	if e, ok := fields[key+"_english"]; ok && e.Kind == yaml.ScalarNode {
		return e.Value
	}
	switch key {
	case "text", "name", "line", "question", "title", "desc":
		if e, ok := fields["english"]; ok && e.Kind == yaml.ScalarNode {
			return e.Value
		}
	}
	return ""
}
