package journey

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/therealsiege/tuilaoshi/internal/segment"
	"github.com/therealsiege/tuilaoshi/internal/wiki"
)

func TestScenesAreCompleteAndSpokenByPartyMembers(t *testing.T) {
	joined := map[string]int{}
	for _, c := range content.Companions {
		joined[c.Name] = c.Stage
	}
	for _, sc := range content.Story.Scenes {
		if TribulationAt(sc.Stage) == nil {
			t.Errorf("scene %q is for stage %d, which doesn't exist", sc.Title, sc.Stage)
		}
		if sc.Title == "" || len(sc.Text) == 0 || sc.English == "" || sc.Line == "" || sc.LineEnglish == "" {
			t.Errorf("scene for stage %d is missing a field", sc.Stage)
		}
		if at, ok := joined[sc.Speaker]; !ok || at > sc.Stage {
			t.Errorf("scene %d is closed by %s, who hasn't joined the party by then", sc.Stage, sc.Speaker)
		}
	}
	seen := map[string]bool{}
	for _, g := range content.Story.Glossary {
		if g.Word == "" || g.Pinyin == "" || g.Meaning == "" || seen[g.Word] {
			t.Errorf("glossary entry %+v needs a unique word, pinyin and meaning", g)
		}
		seen[g.Word] = true
	}
}

// Nearly every word in a scene must have a gloss, from the decks or the
// story glossary, or the scene can't be read word by word.
func TestScenesAreGlossedByRealWiki(t *testing.T) {
	dir := filepath.Join("..", "..", "wiki")
	if _, err := os.Stat(dir); err != nil || testing.Short() {
		t.Skip("needs the wiki")
	}
	cards, err := wiki.NewStore(dir).DrillableCards()
	if err != nil {
		t.Fatal(err)
	}
	known := map[string]bool{}
	for _, c := range append(cards, StoryGlossary()...) {
		known[c.Chinese] = true
	}
	for _, sc := range content.Story.Scenes {
		var words, missing []string
		for _, line := range sc.Text {
			for _, tok := range segment.Split(line, known) {
				if tok.Word {
					words = append(words, tok.Text)
					if !known[tok.Text] {
						missing = append(missing, tok.Text)
					}
				}
			}
		}
		if len(missing)*10 > len(words) {
			t.Errorf("scene %d: %d of %d words have no gloss: %s", sc.Stage, len(missing), len(words), strings.Join(missing, " "))
		}
	}
}
