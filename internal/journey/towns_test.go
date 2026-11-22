package journey

import "testing"

func TestTownsAreCompleteAndInRoadOrder(t *testing.T) {
	ids := map[string]bool{}
	prev := 0
	for _, town := range content.Towns {
		if town.Stage < prev || TribulationAt(town.Stage) == nil {
			t.Errorf("town %s at stage %d is out of order or off the road", town.Name, town.Stage)
		}
		prev = town.Stage
		if town.Name == "" || town.English == "" || town.Desc == "" || town.DescEnglish == "" || len(town.Quests) == 0 {
			t.Errorf("town %s is incomplete", town.ID)
		}
		for _, q := range town.Quests {
			if ids[q.ID] {
				t.Errorf("quest id %s used twice", q.ID)
			}
			ids[q.ID] = true
			if q.Giver == "" || q.GiverEnglish == "" || q.Title == "" || q.Text == "" || q.English == "" {
				t.Errorf("quest %s is incomplete", q.ID)
			}
			if q.Hunting() {
				r := content.Regions[RegionOf(town.Stage)]
				s := EnemyLookup(q.Target)
				served := false
				for _, sol := range content.Soldiers {
					if sol.Name == q.Target && sol.From <= r.To && r.From <= sol.To {
						served = true
					}
				}
				if s == nil || !served || q.Count < 1 || q.Count > 3 {
					t.Errorf("hunt %s targets %q x%d, not a soldier of %s", q.ID, q.Target, q.Count, r.ID)
				}
				continue
			}
			if q.Writing() {
				if q.Criteria == "" || q.Question != "" || len(q.Choices) > 0 {
					t.Errorf("writing quest %s needs criteria and no question", q.ID)
				}
				continue
			}
			if q.Question == "" || len(q.Choices) < 3 {
				t.Errorf("quest %s has %d choices", q.ID, len(q.Choices))
			}
			seen := map[string]bool{}
			for _, c := range q.Choices {
				if seen[c] {
					t.Errorf("quest %s repeats choice %s", q.ID, c)
				}
				seen[c] = true
			}
		}
	}
	for _, town := range content.Towns {
		s := town.Stone
		if s == nil {
			continue
		}
		if len([]rune(s.Idiom)) != 4 || s.Pinyin == "" || s.Meaning == "" || s.Clue == "" {
			t.Errorf("town %s has an incomplete stone %+v", town.ID, s)
		}
		if !s.Opens(" "+s.Idiom+"。") || s.Opens(string([]rune(s.Idiom)[:3])) {
			t.Errorf("stone %s opens wrongly", s.Idiom)
		}
	}
	if len(content.Towns) < 10 {
		t.Errorf("only %d towns", len(content.Towns))
	}
}
