package journey

import (
	"strings"
	"testing"
)

func TestADayOfWalkingEatsAndTires(t *testing.T) {
	tr := NewTrail()
	party := []string{Master, "孙悟空"}
	for i := 0; i < 5; i++ {
		if tr.Step(party, "mountain") {
			t.Fatalf("a steady day ended after %d steps", i+1)
		}
	}
	if !tr.Step(party, "mountain") || tr.Day != 2 || tr.Food != StartFood-6 || tr.Water != StartWater-2 {
		t.Fatalf("after a steady day: day %d food %d water %d", tr.Day, tr.Food, tr.Water)
	}
	tr.Pace, tr.Rations = "forced", "bare"
	tr.PassDay(party, "desert", false)
	if tr.Health[Master] != 100-10 || tr.Water != StartWater-2-4 {
		t.Errorf("a forced, bare day in the desert: health %d (want 90), water %d", tr.Health[Master], tr.Water)
	}
}

func TestRunningOutHurtsAndRestHeals(t *testing.T) {
	tr := NewTrail()
	party := []string{Master}
	tr.Food, tr.Water = 0, 0
	for i := 0; i < 5; i++ {
		tr.PassDay(party, "mountain", false)
	}
	if !tr.Collapsed() || !strings.Contains(strings.Join(tr.Notes, " "), "water") {
		t.Fatalf("five days without food or water: health %d, notes %v", tr.Health[Master], tr.Notes)
	}
	tr.PassDay(party, "river", false)
	if tr.Water != MaxWater-0 && tr.Water != MaxWater {
		t.Errorf("a river day should refill the water: %d", tr.Water)
	}
	tr.Restore(party)
	if tr.Collapsed() || tr.Health[Master] != 100 {
		t.Errorf("an inn should restore the master: %d", tr.Health[Master])
	}
	if Next(Paces, "forced") != "steady" || Next(Rations, "x") != "filling" {
		t.Error("settings don't cycle")
	}
}

func TestSandalsWearOutAndPricesRiseGoingWest(t *testing.T) {
	members := []string{Master}
	tr := NewTrail()
	tr.Food, tr.Water = 1000, MaxWater
	for range SandalLife {
		tr.Water = MaxWater
		tr.PassDay(members, "mountain", false)
	}
	if tr.SandalsWorn() || tr.stepsToday() != stepsPerDay["steady"] {
		t.Fatalf("sandals worn after %d days", SandalLife)
	}
	before := tr.Health[Master]
	tr.PassDay(members, "mountain", false)
	if !tr.SandalsWorn() || tr.stepsToday() != stepsPerDay["steady"]-wornSteps || tr.Health[Master] != clamp(before+rationHealth["filling"]+paceHealth["steady"]-wornHealth) {
		t.Errorf("worn sandals: worn %v, steps %d, health %d -> %d", tr.SandalsWorn(), tr.stepsToday(), before, tr.Health[Master])
	}
	tr.NewSandals()
	if tr.SandalsWorn() {
		t.Error("new sandals still worn")
	}
	if p := (Town{Stage: 1}).Price(FoodPrice); p != FoodPrice {
		t.Errorf("first town charges %d, want the base %d", p, FoodPrice)
	}
	if p := (Town{Stage: 81}).Price(FoodPrice); p != 4*FoodPrice {
		t.Errorf("the last town charges %d, want four times the base", p)
	}
}
