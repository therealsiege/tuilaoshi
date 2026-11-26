package wiki

import (
	"bytes"
	"encoding/csv"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/therealsiege/tuilaoshi/internal/model"
)

const leveledPage = `---
type: vocabulary
level: HSK2
---

# Greetings

| Word | Pinyin | English | Example |
|------|--------|---------|---------|
| 你好 | nǐ hǎo | hello, "hi" | 你好！ |
| 再见 | zàijiàn | goodbye | 再见！ |
`

func TestWriteDeckListsStudiedAndHardWords(t *testing.T) {
	store := newTestStore(t)
	if err := os.WriteFile(filepath.Join(store.Dir, "vocabulary/greetings.md"), []byte(leveledPage), 0644); err != nil {
		t.Fatal(err)
	}
	cards, err := store.DrillableCards()
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 11, 20, 9, 0, 0, 0, time.UTC)
	hello := findCard(t, cards, "vocabulary/greetings.md#你好")
	for range HardWordLapses {
		if err := store.RecordReview(&hello, model.DiffAgain, now); err != nil {
			t.Fatal(err)
		}
	}
	book := findCard(t, cards, "characters/shu1.md")
	if err := store.RecordReview(&book, model.DiffGood, now); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	n, err := store.WriteDeck(&buf)
	if err != nil {
		t.Fatal(err)
	}
	rows, err := csv.NewReader(&buf).ReadAll()
	if err != nil {
		t.Fatalf("the deck isn't valid CSV: %v\n%s", err, buf.String())
	}
	want := [][]string{
		{"Chinese", "Pinyin", "Meaning", "Tags"},
		{"书", "shū", "book", ""},
		{"你好", "nǐ hǎo", `hello, "hi"`, "hard HSK2"},
	}
	if n != 2 || !reflect.DeepEqual(sortedAfterHeader(rows), want) {
		t.Errorf("deck of %d words:\n%q\nwant\n%q", n, rows, want)
	}
}

// sortedAfterHeader puts the rows after the header in Chinese order, as
// the page walk's order is not the test's concern.
func sortedAfterHeader(rows [][]string) [][]string {
	if len(rows) < 2 {
		return rows
	}
	slices.SortFunc(rows[1:], func(a, b []string) int { return strings.Compare(a[0], b[0]) })
	return rows
}
