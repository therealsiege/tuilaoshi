package wiki

import "testing"

func TestGazetteIssuesAreKeptPerJourneyAndWeek(t *testing.T) {
	store := NewStore(t.TempDir())
	if err := store.SaveGazette(Gazette{Week: "2026-11-23", Text: "一", English: "one"}); err != nil {
		t.Fatal(err)
	}
	if err := store.SaveGazette(Gazette{Week: "2026-11-23", Text: "二", English: "two"}); err != nil {
		t.Fatal(err)
	}
	if err := store.SaveGazette(Gazette{Week: "2026-11-30", Text: "三", English: "three"}); err != nil {
		t.Fatal(err)
	}
	issues, err := store.Gazettes()
	if err != nil || len(issues) != 2 || issues[0].Text != "二" || issues[1].Week != "2026-11-30" {
		t.Errorf("issues %+v %v", issues, err)
	}
	store.Save = "other"
	if issues, err := store.Gazettes(); err != nil || len(issues) != 0 {
		t.Errorf("another journey sees this one's gazette: %+v %v", issues, err)
	}
}
