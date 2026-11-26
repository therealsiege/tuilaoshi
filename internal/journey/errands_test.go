package journey

import (
	"testing"
	"time"
)

func TestErrandsAreComplete(t *testing.T) {
	if len(content.Errands) < 10 {
		t.Fatalf("only %d errands", len(content.Errands))
	}
	ids := map[string]bool{}
	for _, e := range content.Errands {
		if ids[e.ID] {
			t.Errorf("errand %s listed twice", e.ID)
		}
		ids[e.ID] = true
		if e.Role == "" || e.Name == "" || e.RoleEnglish == "" || e.Persona == "" || e.Opening == "" || e.OpeningEnglish == "" ||
			e.Goal == "" || e.GoalEnglish == "" || e.Reward <= 0 {
			t.Errorf("errand %s is missing a field: %+v", e.ID, e)
		}
		if len(e.Points) < 2 || len(e.Points) > 4 {
			t.Errorf("errand %s has %d points, want 2-4", e.ID, len(e.Points))
		}
		points := map[string]bool{}
		for _, p := range e.Points {
			if p.ID == "" || p.Chinese == "" || p.English == "" || p.Listen == "" || points[p.ID] {
				t.Errorf("errand %s: point %+v is incomplete or repeated", e.ID, p)
			}
			points[p.ID] = true
		}
	}
}

func TestATownOffersOneErrandADay(t *testing.T) {
	day := time.Date(2026, 11, 22, 9, 0, 0, 0, time.Local)
	if ErrandFor("changan", day).ID != ErrandFor("changan", day.Add(8*time.Hour)).ID {
		t.Error("the errand changed within the day")
	}
	seen := map[string]bool{}
	for i := range 30 {
		seen[ErrandFor("changan", day.AddDate(0, 0, i)).ID] = true
	}
	if len(seen) < 4 {
		t.Errorf("a month in one town offers only %d errands", len(seen))
	}
}
