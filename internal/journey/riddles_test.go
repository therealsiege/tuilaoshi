package journey

import (
	"testing"
	"unicode/utf8"
)

func TestRiddlesAreComplete(t *testing.T) {
	if len(content.Riddles) < 15 {
		t.Fatalf("only %d riddles", len(content.Riddles))
	}
	seen := map[string]bool{}
	for i, r := range content.Riddles {
		if r.Riddle == "" || r.Pinyin == "" || r.English == "" || r.Explain == "" || utf8.RuneCountInString(r.Answer) != 1 {
			t.Errorf("riddle %d is incomplete or its answer isn't one character: %+v", i, r)
		}
		if seen[r.Answer] {
			t.Errorf("two riddles answer %s", r.Answer)
		}
		seen[r.Answer] = true
	}
}

func TestRiddleForIsStableAndVaries(t *testing.T) {
	a, b := RiddleFor("changan", "2026-11-25", 0), RiddleFor("changan", "2026-11-25", 0)
	if a != b {
		t.Error("the same town and day gave two riddles")
	}
	if RiddleFor("changan", "2026-11-25", 1) == a {
		t.Error("the day's second riddle repeats the first")
	}
	if !a.Solves(" "+a.Answer+" ") || a.Solves("x") {
		t.Error("Solves")
	}
}
