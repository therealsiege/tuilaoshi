// Command freqlist imports a frequency word list (mylingua's xlsx format:
// word, count, pinyin, HSK level, a name flag, then translations) into the
// learner's private wiki, wiki/private/frequency.yaml, from which the game
// introduces the most frequent unknown words a few a day. The list stays
// private (wiki/private is git-ignored).
//
//	go run ./cmd/freqlist -in list.xlsx [-wiki wiki]
package main

import (
	"archive/zip"
	"encoding/xml"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/therealsiege/tuilaoshi/internal/wiki"
)

func main() {
	in := flag.String("in", "", "the frequency list (.xlsx)")
	dir := flag.String("wiki", "wiki", "the wiki directory")
	sheet := flag.Int("sheet", 2, "the worksheet holding the words")
	flag.Parse()
	if *in == "" {
		fmt.Fprintln(os.Stderr, "freqlist: give the list with -in")
		os.Exit(2)
	}
	words, err := read(*in, *sheet)
	if err == nil {
		err = write(filepath.Join(*dir, wiki.FrequencyFile), words)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "freqlist:", err)
		os.Exit(1)
	}
	fmt.Printf("imported %d words into %s\n", len(words), filepath.Join(*dir, wiki.FrequencyFile))
}

func write(path string, words []wiki.FreqWord) error {
	data, err := yaml.Marshal(words)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

// read takes the words from a worksheet, most frequent first as listed.
func read(path string, sheet int) ([]wiki.FreqWord, error) {
	z, err := zip.OpenReader(path)
	if err != nil {
		return nil, err
	}
	defer z.Close() //nolint:errcheck // read-only
	files := map[string]*zip.File{}
	for _, f := range z.File {
		files[f.Name] = f
	}
	strs, err := sharedStrings(files["xl/sharedStrings.xml"])
	if err != nil {
		return nil, err
	}
	ws := files[fmt.Sprintf("xl/worksheets/sheet%d.xml", sheet)]
	if ws == nil {
		return nil, fmt.Errorf("no worksheet %d", sheet)
	}
	rows, err := sheetRows(ws, strs)
	if err != nil {
		return nil, err
	}
	return parseRows(rows), nil
}

// parseRows reads rows keyed by column letter: B word, C count, D pinyin,
// E HSK level, F name flag, H onward translations. The header row and
// anything without a word and count are skipped.
func parseRows(rows []map[string]string) []wiki.FreqWord {
	var words []wiki.FreqWord
	for _, r := range rows {
		n, err := strconv.Atoi(r["C"])
		if err != nil || r["B"] == "" {
			continue
		}
		var en []string
		for _, col := range []string{"H", "I", "J"} {
			if t := strings.TrimSpace(r[col]); t != "" {
				en = append(en, t)
			}
		}
		hsk, _ := strconv.Atoi(r["E"]) //nolint:errcheck // blank: not an HSK word
		words = append(words, wiki.FreqWord{Word: r["B"], Count: n, Pinyin: r["D"], HSK: hsk, Name: r["F"] == "1", Meaning: strings.Join(en, "; ")})
	}
	return words
}

type xmlCell struct {
	Ref  string `xml:"r,attr"`
	Type string `xml:"t,attr"`
	V    string `xml:"v"`
	Is   struct {
		T string `xml:"t"`
	} `xml:"is"`
}

var colLetters = regexp.MustCompile(`^[A-Z]+`)

func sheetRows(f *zip.File, strs []string) ([]map[string]string, error) {
	var doc struct {
		Rows []struct {
			Cells []xmlCell `xml:"c"`
		} `xml:"sheetData>row"`
	}
	if err := decode(f, &doc); err != nil {
		return nil, err
	}
	rows := make([]map[string]string, 0, len(doc.Rows))
	for _, r := range doc.Rows {
		row := map[string]string{}
		for _, c := range r.Cells {
			v := c.V
			switch c.Type {
			case "s":
				if i, err := strconv.Atoi(v); err == nil && i < len(strs) {
					v = strs[i]
				}
			case "inlineStr":
				v = c.Is.T
			}
			row[colLetters.FindString(c.Ref)] = v
		}
		rows = append(rows, row)
	}
	return rows, nil
}

func sharedStrings(f *zip.File) ([]string, error) {
	if f == nil {
		return nil, nil
	}
	var doc struct {
		Items []struct {
			T    string `xml:"t"`
			Runs []struct {
				T string `xml:"t"`
			} `xml:"r"`
		} `xml:"si"`
	}
	if err := decode(f, &doc); err != nil {
		return nil, err
	}
	out := make([]string, len(doc.Items))
	for i, it := range doc.Items {
		out[i] = it.T
		for _, r := range it.Runs {
			out[i] += r.T
		}
	}
	return out, nil
}

func decode(f *zip.File, into any) error {
	rc, err := f.Open()
	if err != nil {
		return err
	}
	defer rc.Close() //nolint:errcheck // read-only
	data, err := io.ReadAll(rc)
	if err != nil {
		return err
	}
	return xml.Unmarshal(data, into)
}
