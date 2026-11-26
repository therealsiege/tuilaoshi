package wiki

import (
	"slices"
	"testing"
	"time"
)

func TestPassedGrammarGatesAreSavedPerJourney(t *testing.T) {
	store := NewStore(t.TempDir())
	if _, err := store.NewSave("first", time.Now()); err != nil {
		t.Fatal(err)
	}
	for _, r := range []string{"changan", "lingshan-origins", "changan"} {
		if err := store.PassGrammarGate(r); err != nil {
			t.Fatal(err)
		}
	}
	got, err := store.GrammarGates()
	if err != nil || !slices.Equal(got, []string{"changan", "lingshan-origins"}) {
		t.Fatalf("passed %v, err %v; want each once in order", got, err)
	}
	if _, err := store.NewSave("second", time.Now()); err != nil {
		t.Fatal(err)
	}
	if got, err := store.GrammarGates(); err != nil || len(got) != 0 {
		t.Errorf("a new journey starts with gates %v (err %v)", got, err)
	}
}
