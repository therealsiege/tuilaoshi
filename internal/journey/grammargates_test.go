package journey

import (
	"slices"
	"strings"
	"testing"
)

func TestEveryRegionHasAGrammarGate(t *testing.T) {
	if len(content.GrammarGates) != len(content.Regions) {
		t.Errorf("%d grammar gates for %d regions", len(content.GrammarGates), len(content.Regions))
	}
	patterns := map[string]bool{}
	for _, r := range content.Regions {
		g, ok := GrammarGateFor(r.ID)
		if !ok {
			t.Errorf("region %s has no grammar gate", r.ID)
			continue
		}
		if patterns[g.Pattern] {
			t.Errorf("pattern %s guards two regions", g.Pattern)
		}
		patterns[g.Pattern] = true
		if g.Pattern == "" || g.Pinyin == "" || g.Explain == "" || len(g.Markers) == 0 {
			t.Errorf("gate %s is incomplete: %+v", r.ID, g)
		}
		if n := len(g.Examples); n < 2 || n > 3 {
			t.Errorf("gate %s has %d examples, want 2 or 3", r.ID, n)
		}
		for _, e := range g.Examples {
			if e.English == "" || !g.Uses(e.Chinese) {
				t.Errorf("gate %s: example %q lacks English or doesn't pass its own rule", r.ID, e.Chinese)
			}
		}
		if len(g.Sentences) < 6 {
			t.Errorf("gate %s has %d tile sentences, want 6 or more", r.ID, len(g.Sentences))
		}
		for _, s := range g.Sentences {
			checkTileSentence(t, g, s)
		}
	}
}

// checkTileSentence: a tile sentence has 3 to 9 tiles (one number key
// each), English, passes the pattern's rule, and its other orders use the
// same words and pass too.
func checkTileSentence(t *testing.T, g GrammarGate, s TileSentence) {
	t.Helper()
	text := strings.Join(s.Words, "")
	if len(s.Words) < 3 || len(s.Words) > 9 || s.English == "" {
		t.Errorf("gate %s: sentence %s has %d tiles or no English", g.Region, text, len(s.Words))
	}
	if !g.Uses(text) {
		t.Errorf("gate %s: sentence %s doesn't pass its pattern's rule", g.Region, text)
	}
	for _, a := range s.Also {
		if !sameRunes(a, text) || !g.Uses(a) {
			t.Errorf("gate %s: other order %s of %s doesn't use the same words or the pattern", g.Region, a, text)
		}
	}
}

func sameRunes(a, b string) bool {
	x, y := []rune(a), []rune(b)
	slices.Sort(x)
	slices.Sort(y)
	return slices.Equal(x, y)
}

func TestGrammarGateRules(t *testing.T) {
	ba, _ := GrammarGateFor("lingshan-origins")
	for s, want := range map[string]bool{
		"悟空把金箍棒放进了耳朵里。": true,
		"我把。":      false, // nothing done after 把
		"悟空拿着金箍棒。": false, // no 把
		"把书":       false, // too short
	} {
		if got := ba.Uses(s); got != want {
			t.Errorf("把 %q: %v, want %v", s, got, want)
		}
	}
	yue, _ := GrammarGateFor("pingding")
	if !yue.Uses("雨越下越大") || yue.Uses("雨越越下大") || yue.Uses("天气越来很冷") {
		t.Error("越…越 needs two 越 with something between")
	}
	de, _ := GrammarGateFor("lingshan")
	if !de.Uses("难道你不知道吗？") || de.Uses("你不知道吗？") {
		t.Error("难道…吗 may end on 吗 but needs 难道")
	}
	s := TileSentence{Words: []string{"师父", "一边", "走路", "一边", "念经"}, Also: []string{"师父一边念经一边走路"}}
	if !s.Fits(s.Words) || !s.Fits([]string{"师父", "一边", "念经", "一边", "走路"}) || s.Fits([]string{"一边", "师父", "走路", "一边", "念经"}) {
		t.Error("Fits takes the sentence and its other orders only")
	}
	if q := (TileSentence{Words: []string{"难道", "你", "吗"}, English: "Really?"}); q.Text() != "难道你吗？" {
		t.Errorf("a question ends %q", q.Text())
	}
}
