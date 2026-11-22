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
