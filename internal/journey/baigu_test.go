package journey

import (
	"slices"
	"testing"
	"unicode"
)

// baiguHan reports whether s holds any Chinese character.
func baiguHan(s string) bool {
	return slices.ContainsFunc([]rune(s), func(r rune) bool { return unicode.Is(unicode.Han, r) })
}

func TestBaiguIsThreeDisguisesInTheNovelsOrder(t *testing.T) {
	b := BaiguTrial()
	if b.Angry == "" || b.AngryEnglish == "" || b.Won == "" || b.WonEnglish == "" {
		t.Errorf("the trial is missing its anger or ending lines: %+v", b)
	}
	skills := []string{"reading", "listening", "writing"}
	if len(b.Disguises) != len(skills) {
		t.Fatalf("%d disguises, want 3", len(b.Disguises))
	}
	ids := map[string]bool{}
	for i, d := range b.Disguises {
		if d.Skill != skills[i] || d.Name == "" || d.English == "" {
			t.Errorf("disguise %d is %+v, want a named %s disguise", i+1, d, skills[i])
		}
		if len(d.Rounds) < 3 {
			t.Errorf("%s has %d rounds, want several so a rematch differs", d.Name, len(d.Rounds))
		}
		for _, r := range d.Rounds {
			if ids[r.ID] || r.ID == "" {
				t.Errorf("round id %q is empty or used twice", r.ID)
			}
			ids[r.ID] = true
			if !baiguHan(r.Text) || !baiguHan(r.Question) || r.English == "" || r.QuestionEnglish == "" {
				t.Errorf("round %s needs Chinese text and question with their English", r.ID)
			}
			if len(r.Choices) != 3 || len(r.EnglishChoices) != 3 {
				t.Errorf("round %s has %d choices and %d English, want 3 of each", r.ID, len(r.Choices), len(r.EnglishChoices))
				continue
			}
			for j, c := range r.Choices {
				if !baiguHan(c) || r.EnglishChoices[j] == "" || slices.Index(r.Choices, c) != j {
					t.Errorf("round %s choice %d %q is empty, not Chinese or repeated", r.ID, j+1, c)
				}
			}
			if (d.Skill == "writing") != (r.Criteria != "") {
				t.Errorf("round %s: criteria %q, but only writing rounds are graded by Claude", r.ID, r.Criteria)
			}
		}
	}
}

func TestBaiguDemonIsABoss(t *testing.T) {
	for _, b := range content.Bosses {
		if b.Name == BaiguDemon {
			if tr := TribulationAt(b.Stage); tr == nil || tr.Demon != BaiguDemon {
				t.Errorf("boss %s is at stage %d, whose tribulation isn't hers", b.Name, b.Stage)
			}
			return
		}
	}
	t.Errorf("no boss %s in bosses.yaml", BaiguDemon)
}
