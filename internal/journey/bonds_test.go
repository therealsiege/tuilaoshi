package journey

import "testing"

func TestBondsCoverTheDisciples(t *testing.T) {
	joined := map[string]bool{}
	for _, c := range content.Companions {
		joined[c.Name] = true
	}
	if len(content.Bonds) != 4 {
		t.Fatalf("%d bonds, want the four disciples", len(content.Bonds))
	}
	for _, b := range content.Bonds {
		if !joined[b.Companion] {
			t.Errorf("bond for %s, who is not a companion", b.Companion)
		}
		if b.Voice == "" || b.Upgrade == "" || b.UpgradeDesc == "" {
			t.Errorf("bond %s is missing its voice or upgrade", b.Companion)
		}
		if len(b.Scenes) != len(BondRanks)-1 {
			t.Errorf("bond %s has %d scenes, want one per rank above the first", b.Companion, len(b.Scenes))
		}
		for _, s := range b.Scenes {
			if s.Title == "" || s.EnglishTitle == "" || s.Text == "" || s.English == "" {
				t.Errorf("bond %s scene %q is incomplete", b.Companion, s.Title)
			}
		}
	}
	if BondRankFor(0) != 0 || BondRankFor(15) != BondUpgradeRank || BondRankFor(999) != len(BondRanks)-1 {
		t.Error("bond ranks don't follow points")
	}
}
