package journey

import (
	"regexp"
	"strconv"
	"testing"
)

// The journey content files are validated as a whole, so a new demon, boss
// or lore passage that is incomplete or points at something missing fails
// here rather than in the game.

var firstNumber = regexp.MustCompile(`\d+`)

func TestContentLoads(t *testing.T) {
	if _, err := LoadContent(); err != nil {
		t.Fatal(err)
	}
}

func TestTribulationsAreTheEightyOne(t *testing.T) {
	if len(content.Tribulations) != 81 {
		t.Fatalf("%d tribulations, want 81", len(content.Tribulations))
	}
	for i, tr := range content.Tribulations {
		if tr.Number != i+1 {
			t.Errorf("tribulation %d is numbered %d", i+1, tr.Number)
		}
		if tr.Name == "" || tr.English == "" || tr.Desc == "" || tr.Location == "" {
			t.Errorf("tribulation %d is missing a field: %+v", tr.Number, tr)
		}
	}
}

func TestBestiaryEntriesAreComplete(t *testing.T) {
	seen := map[string]bool{}
	for _, d := range content.Bestiary {
		if seen[d.Name] {
			t.Errorf("demon %s listed twice", d.Name)
		}
		seen[d.Name] = true
		if d.English == "" || d.Origin == "" || d.Powers == "" || d.Weakness == "" || d.Art == "" {
			t.Errorf("demon %s is missing a field", d.Name)
		}
		n, err := strconv.Atoi(firstNumber.FindString(d.Chapter))
		if err != nil || n < 1 || n > 100 {
			t.Errorf("demon %s has chapter %q, want one in 1-100", d.Name, d.Chapter)
		}
	}
}

func TestBossesResolveAndEscalate(t *testing.T) {
	prev := 0
	for _, b := range content.Bosses {
		if b.Required <= prev {
			t.Errorf("boss %s needs %d characters, not more than the boss before (%d)", b.Name, b.Required, prev)
		}
		prev = b.Required
		if e := BossEntry(b); e.Art == "" {
			t.Errorf("boss %s %s has no bestiary entry with art", b.Name, b.English)
		}
	}
}

func TestCompanionsUnlockInOrder(t *testing.T) {
	if len(content.Companions) == 0 || content.Companions[0].UnlockAt != 0 {
		t.Fatal("the first companion must travel from the start")
	}
	for i := 1; i < len(content.Companions); i++ {
		if content.Companions[i].UnlockAt <= content.Companions[i-1].UnlockAt {
			t.Errorf("companion %s unlocks no later than %s", content.Companions[i].Name, content.Companions[i-1].Name)
		}
	}
}

func TestLoreAndDialogueReferencesResolve(t *testing.T) {
	for _, l := range content.Lore {
		if TribulationAt(l.Tribulation) == nil {
			t.Errorf("lore %q unlocks at tribulation %d, which doesn't exist", l.Title, l.Tribulation)
		}
		if l.Passage == "" || l.Pinyin == "" || l.English == "" {
			t.Errorf("lore %q is missing its passage, pinyin or translation", l.Title)
		}
	}
	companions := map[string]bool{}
	for _, c := range content.Companions {
		companions[c.Name] = true
	}
	triggers := map[string]bool{"streak": true, "rank_up": true, "boss_defeat": true, "mistake": true, "review_complete": true, "idle": true}
	for _, d := range content.Dialogue {
		if !companions[d.Speaker] {
			t.Errorf("dialogue %q is spoken by %s, who is not a companion", d.Line, d.Speaker)
		}
		if !triggers[d.Trigger] {
			t.Errorf("dialogue %q has unknown trigger %q", d.Line, d.Trigger)
		}
	}
}
