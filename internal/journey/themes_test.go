package journey

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/therealsiege/tuilaoshi/internal/model"
	"github.com/therealsiege/tuilaoshi/internal/wiki"
)

func TestThemeMatches(t *testing.T) {
	fire := Theme{Keywords: []string{"fire", "flame"}, Contains: []string{"火"}, Tags: []string{"weather"}}
	page := model.Page{Level: "HSK2"}
	for _, tt := range []struct {
		card model.Card
		tags []string
		want bool
	}{
		{model.Card{Chinese: "火车", English: "train"}, nil, true},               // contains 火
		{model.Card{Chinese: "焰", English: "flames, blaze"}, nil, true},        // keyword, plural
		{model.Card{Chinese: "雨", English: "rain"}, []string{"weather"}, true}, // tag
		{model.Card{Chinese: "书", English: "book"}, nil, false},
		{model.Card{Chinese: "灭", English: "fireproof"}, nil, false}, // whole words only
	} {
		page.Tags = tt.tags
		if got := fire.Matches(tt.card, page); got != tt.want {
			t.Errorf("Matches(%s %q) = %v, want %v", tt.card.Chinese, tt.card.English, got, tt.want)
		}
	}
	basics := Theme{Levels: []string{"HSK1"}}
	if !basics.Matches(model.Card{}, model.Page{Level: "HSK1"}) || basics.Matches(model.Card{}, model.Page{Level: "HSK2"}) {
		t.Error("level-only theme should match exactly its levels")
	}
}

func TestStageCardsWidenWhenThemeIsThin(t *testing.T) {
	var cards []model.Card
	var pages []model.Page
	for i := 0; i < 20; i++ {
		path := string(rune('a'+i)) + ".md"
		pages = append(pages, model.Page{Path: path, Level: "HSK1"})
		cards = append(cards, model.Card{ID: path, Chinese: string(rune(0x4E00 + i)), English: "plain"})
	}
	s := Stage{Tribulation: Tribulation{Number: 1}, Words: Theme{Keywords: []string{"fire"}}}
	if got := StageCards(s, pages, cards); len(got) != 20 {
		t.Errorf("no fire words: got %d cards, want the 20 capped cards", len(got))
	}
	for i := 0; i < MinStageWords; i++ {
		cards[i].English = "fire"
	}
	if got := StageCards(s, pages, cards); len(got) != MinStageWords {
		t.Errorf("enough fire words: got %d cards, want the %d themed ones", len(got), MinStageWords)
	}
}

// Every theme must find enough words in the real wiki to fill a battle.
func TestThemesResolveAgainstRealWiki(t *testing.T) {
	dir := filepath.Join("..", "..", "wiki")
	if _, err := os.Stat(dir); err != nil || testing.Short() {
		t.Skip("needs the wiki")
	}
	store := wiki.NewStore(dir)
	pages, err := store.ListPages()
	if err != nil {
		t.Fatal(err)
	}
	cards, err := store.DrillableCards()
	if err != nil {
		t.Fatal(err)
	}
	for _, th := range content.Themes {
		s := Stage{Tribulation: Tribulation{Number: 81}, Words: th} // no level cap that late
		n := 0
		for _, c := range StageCards(s, pages, cards) {
			if th.Matches(c, pageFor(pages, c)) {
				n++
			}
		}
		if n < MinStageWords {
			t.Errorf("theme %s finds %d words, want at least %d", th.ID, n, MinStageWords)
		}
	}
}

func pageFor(pages []model.Page, c model.Card) model.Page {
	for _, p := range pages {
		if p.Path == c.ID || (len(c.ID) > len(p.Path) && c.ID[:len(p.Path)+1] == p.Path+"#") {
			return p
		}
	}
	return model.Page{}
}

func TestThemeMatchesRadical(t *testing.T) {
	water := Theme{Radicals: []string{"氵"}}
	for radical, want := range map[string]bool{"氵 (water)": true, "氵(water)": true, "木 (wood)": false, "": false} {
		if got := water.Matches(model.Card{Chinese: "x"}, model.Page{Radical: radical}); got != want {
			t.Errorf("radical %q: Matches = %v, want %v", radical, got, want)
		}
	}
}

func TestRadicalIndexNormalizes(t *testing.T) {
	idx := RadicalIndex([]model.Page{
		{Type: "character", Character: "灯", Radical: "火 (fire)"},
		{Type: "character", Character: "打", Radical: "扌(hand)"},
		{Type: "vocabulary", Character: "x", Radical: "y"},
	})
	if idx["灯"] != "火 (fire)" || idx["打"] != "扌 (hand)" || len(idx) != 2 {
		t.Errorf("index %v", idx)
	}
}

func TestStageProgressCountsLearnedWords(t *testing.T) {
	var pages []model.Page
	var cards []model.Card
	for i := 0; i < MinStageWords; i++ {
		path := string(rune('a'+i)) + ".md"
		pages = append(pages, model.Page{Path: path, Level: "HSK1"})
		c := model.Card{ID: path, Chinese: string(rune(0x706B + i)), English: "fire"}
		if i < 3 {
			c.ReviewInterval = content.Progression.LearningInterval
		}
		cards = append(cards, c)
	}
	s := Stage{Tribulation: Tribulation{Number: 1}, Words: Theme{Keywords: []string{"fire"}}}
	if got := StageProgress(s, pages, cards); got != (WordProgress{Learned: 3, Total: MinStageWords}) {
		t.Errorf("progress %+v", got)
	}
}
