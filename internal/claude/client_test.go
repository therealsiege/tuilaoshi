package claude

import "testing"

func TestParseWritingGrade(t *testing.T) {
	score, fb := ParseWritingGrade("SCORE: 2\nFEEDBACK: Clear and natural.\nNice work.")
	if score != 2 || fb != "Clear and natural. Nice work." {
		t.Errorf("got %d %q", score, fb)
	}
	if score, _ := ParseWritingGrade("SCORE: 7\nFEEDBACK: x"); score != 0 {
		t.Errorf("an out-of-range score should read as 0, got %d", score)
	}
}

func TestParseComponents(t *testing.T) {
	got := ParseComponents("河 氵 可\n想 相 心\n人 人\nnoise\n猫 犭 苗", []string{"河", "想", "人"})
	if len(got) != 3 || len(got["河"]) != 2 || got["河"][0] != "氵" || len(got["人"]) != 1 || got["猫"] != nil {
		t.Errorf("parsed %v", got)
	}
}
