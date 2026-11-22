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
	if len(content.Towns) < 10 {
		t.Errorf("only %d towns", len(content.Towns))
	}
}
