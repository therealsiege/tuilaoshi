package journey

import (
	"math/rand"
	"slices"
	"testing"
)

func TestEventsAreCompleteAndFitEveryParty(t *testing.T) {
	terrains := map[string]bool{"mountain": true, "forest": true, "river": true, "desert": true, "fire": true, "palace": true, "cave": true}
	members := map[string]bool{Master: true}
	for _, c := range content.Companions {
		members[c.Name] = true
	}
	for _, e := range content.Events {
		if e.ID == "" || e.Text == "" || e.English == "" || len(e.Choices) < 2 {
			t.Errorf("event %s is incomplete", e.ID)
		}
		for _, ter := range e.Terrains {
			if !terrains[ter] {
				t.Errorf("event %s names terrain %q", e.ID, ter)
			}
		}
		open := 0
		for _, c := range e.Choices {
			if c.Text == "" || c.English == "" || c.Result == "" || c.ResultEnglish == "" {
				t.Errorf("event %s has an incomplete choice %q", e.ID, c.Text)
			}
			if c.Needs != "" && !members[c.Needs] || e.Needs != "" && !members[e.Needs] {
				t.Errorf("event %s needs an unknown member", e.ID)
			}
			if c.Needs == "" {
				open++
			}
		}
		if open < 2 && e.Needs == "" {
			t.Errorf("event %s leaves a lone master fewer than two choices", e.ID)
		}
	}
	for ter := range terrains {
		if _, ok := PickEvent(ter, []string{Master}, rand.New(rand.NewSource(1))); !ok {
			t.Errorf("no event can happen in %s to the master alone", ter)
		}
	}
}

func TestPickingAndApplyingAnEvent(t *testing.T) {
	rng := rand.New(rand.NewSource(7))
	for i := 0; i < 50; i++ {
		e, ok := PickEvent("mountain", []string{Master}, rng)
		if !ok {
			t.Fatal("no mountain event")
		}
		if e.Needs != "" {
			t.Fatalf("event %s needs %s, who isn't in the party", e.ID, e.Needs)
		}
		for _, c := range e.Choices {
			if c.Needs != "" {
				t.Fatalf("event %s offers %q, which needs %s", e.ID, c.Text, c.Needs)
			}
		}
	}
	tr := NewTrail()
	party := []string{Master, "孙悟空"}
	tr.Apply(Outcome{Food: -100, Water: 50, Health: -20, Who: Master}, party)
	if tr.Food != 0 || tr.Water != MaxWater || tr.Health[Master] != 80 || tr.Health["孙悟空"] != 100 {
		t.Errorf("after the outcome: %+v", tr)
	}
}

func TestACrossingsConditionDecidesWhichChoiceGoesWrong(t *testing.T) {
	c := CrossingFor("river")
	if c == nil || CrossingFor("palace") != nil {
		t.Fatal("rivers should have a crossing and palaces none")
	}
	for seed := int64(0); seed < 20; seed++ {
		e := c.Event(rand.New(rand.NewSource(seed)))
		for _, cond := range c.Conditions {
			if e.Text != c.Name+"："+cond.Text {
				continue
			}
			for i, ch := range e.Choices {
				bad := ch.Result == c.Bad
				if bad != slices.Contains(cond.Danger, i) {
					t.Errorf("%q: choice %q bad %v", cond.Text, ch.Text, bad)
				}
				if bad && ch.Outcome.Health >= c.Choices[i].Outcome.Health {
					t.Errorf("%q: a dangerous choice doesn't hurt", cond.Text)
				}
			}
		}
	}
	for _, cr := range content.Crossings {
		for _, cond := range cr.Conditions {
			if len(cond.Danger) >= len(cr.Choices) {
				t.Errorf("%s: %q leaves no safe way across", cr.Kind, cond.Text)
			}
		}
	}
}
