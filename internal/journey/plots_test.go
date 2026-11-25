package journey

import "testing"

func TestPlotsAreComplete(t *testing.T) {
	if len(content.Plots) < 10 {
		t.Fatalf("only %d plots", len(content.Plots))
	}
	ids := map[string]bool{}
	for _, p := range content.Plots {
		if p.ID == "" || ids[p.ID] {
			t.Errorf("plot %q: missing or repeated id", p.ID)
		}
		ids[p.ID] = true
		if n := len(p.Lines); n < 2 || n > 4 {
			t.Errorf("plot %s: %d lines, want 2-4", p.ID, n)
		}
		speakers := map[string]bool{}
		for _, l := range p.Lines {
			if l.Who == "" || l.Text == "" || l.English == "" {
				t.Errorf("plot %s has an incomplete line %q", p.ID, l.Text)
			}
			speakers[l.Who] = true
		}
		if len(speakers) != 2 {
			t.Errorf("plot %s: %d speakers, want two demons", p.ID, len(speakers))
		}
		if p.Question == "" || p.QuestionEnglish == "" || p.Warning == "" || p.WarningEnglish == "" {
			t.Errorf("plot %s is missing its question or warning, or their English", p.ID)
		}
		if len(p.Choices) != 3 || len(p.EnglishChoices) != len(p.Choices) {
			t.Errorf("plot %s: want 3 choices, each with English; got %d and %d", p.ID, len(p.Choices), len(p.EnglishChoices))
		}
	}
}

func TestForewarningHalvesTheHarm(t *testing.T) {
	o, spared := Forewarn(Outcome{Food: -10, Water: 4, Coins: -6, Health: -15, Morale: -5, Who: Master})
	if !spared || o != (Outcome{Food: -5, Water: 4, Coins: -6, Health: -7, Morale: -2, Who: Master}) {
		t.Errorf("forewarned outcome %+v, spared %v", o, spared)
	}
	if _, spared := Forewarn(Outcome{Food: 12, Coins: -10}); spared {
		t.Error("a harmless outcome was spared")
	}
}
