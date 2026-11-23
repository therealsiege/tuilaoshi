package wiki

import (
	"testing"
	"time"
)

func TestIdiomsJoinReviewsOnce(t *testing.T) {
	store := NewStore(t.TempDir())
	words := []Word{{"天长地久", "tiān cháng dì jiǔ", "everlasting"}, {"久别重逢", "jiǔ bié chóng féng", "to meet after | a long parting"}}
	for i, want := range []int{2, 0} {
		n, err := store.AddIdioms(words, time.Now())
		if err != nil || n != want {
			t.Fatalf("round %d: added %d, %v; want %d", i, n, err, want)
		}
	}
	cards, err := store.DrillableCards()
	if err != nil {
		t.Fatal(err)
	}
	found := map[string]bool{}
	for _, c := range cards {
		found[c.Chinese] = c.English != ""
	}
	if !found["天长地久"] || !found["久别重逢"] {
		t.Errorf("idioms aren't review cards: %v", found)
	}
}
