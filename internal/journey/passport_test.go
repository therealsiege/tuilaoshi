package journey

import "testing"

func TestEveryGateAsksTheFourQuestions(t *testing.T) {
	if len(content.Gates) != 10 {
		t.Fatalf("%d gates, want the 10 kingdoms", len(content.Gates))
	}
	towns := map[string]bool{}
	for _, tw := range content.Towns {
		towns[tw.ID] = true
	}
	for _, g := range content.Gates {
		if !towns[g.Town] || g.Officer == "" || g.Greeting == "" || g.GreetingEnglish == "" {
			t.Errorf("gate %s is incomplete or not a town", g.Town)
		}
		kinds := map[string]bool{}
		for _, q := range g.Questions {
			kinds[q.Kind] = true
			if q.Ask == "" || q.English == "" || !q.Accepts(q.Sample()) {
				t.Errorf("%s: question %q incomplete, or its sample doesn't pass", g.Town, q.Ask)
			}
		}
		if len(kinds) != 4 || len(g.Questions) != 4 {
			t.Errorf("%s asks %v", g.Town, kinds)
		}
	}
	q := GateQuestion{Kind: "to"}
	if !q.Accepts("我们要去西天。") || q.Accepts("我们要去东边。") {
		t.Error("Accepts")
	}
}
