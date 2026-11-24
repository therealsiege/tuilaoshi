package wiki

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"gopkg.in/yaml.v3"
)

func TestFrequencyWordsJoinReviewsAFewADay(t *testing.T) {
	store := NewStore(t.TempDir())
	list := []FreqWord{
		{Word: "的", Count: 900, Pinyin: "de", Meaning: "of"},
		{Word: "孙悟空", Count: 800, Pinyin: "Sūn Wùkōng", Meaning: "Sun Wukong", Name: true},
		{Word: "我们", Count: 750, Pinyin: "wǒmen", Meaning: "we", HSK: 1},
		{Word: "他", Count: 740, Pinyin: "tā", Meaning: "he", HSK: 1},
		{Word: "们", Count: 730, Pinyin: "men", Meaning: "plural", HSK: 1},
		{Word: "他们", Count: 720, Pinyin: "tāmen", Meaning: "they"},
		{Word: "妖怪", Count: 700, Pinyin: "yāoguài", Meaning: "demon; monster; goblin"},
		{Word: "法宝", Count: 600, Pinyin: "fǎbǎo", Meaning: "magic weapon"},
		{Word: "修行", Count: 500, Pinyin: "xiūxíng", Meaning: "to practise"},
	}
	data, err := yaml.Marshal(list)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(store.Dir, "private"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(store.Dir, FrequencyFile), data, 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(store.Dir, "vocabulary"), 0755); err != nil {
		t.Fatal(err)
	}
	page := "---\ntype: vocabulary\n---\n\n| Word | Pinyin | Meaning |\n|---|---|---|\n| 的 | de | of |\n"
	if err := os.WriteFile(filepath.Join(store.Dir, "vocabulary", "basic.md"), []byte(page), 0644); err != nil {
		t.Fatal(err)
	}
	now := time.Now()
	n, err := store.IntroduceFrequency(2, now)
	if err != nil || n != 2 {
		t.Fatalf("first day added %d, %v; want 2", n, err)
	}
	if n, _ := store.IntroduceFrequency(2, now); n != 0 { //nolint:errcheck // checked above
		t.Errorf("the same day added %d more", n)
	}
	cards, err := store.DrillableCards()
	if err != nil {
		t.Fatal(err)
	}
	var deck []string
	for _, c := range cards {
		if FrequencyPage(c.ID) {
			deck = append(deck, c.Chinese)
		}
	}
	if len(deck) != 2 || deck[0] != "妖怪" || deck[1] != "法宝" {
		t.Errorf("deck %v, want 妖怪 法宝 (known words, names and easy HSK words skipped)", deck)
	}
	if n, err := store.IntroduceFrequency(2, now.AddDate(0, 0, 1)); err != nil || n != 1 {
		t.Errorf("next day added %d, %v; want the last 1", n, err)
	}
}
