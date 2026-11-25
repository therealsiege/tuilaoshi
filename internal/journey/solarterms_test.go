package journey

import (
	"testing"
	"time"
)

func TestSolarTermsBeginOnTheirDays(t *testing.T) {
	if len(content.SolarTerms) != 24 {
		t.Fatalf("%d terms", len(content.SolarTerms))
	}
	// Well-known dates (China time) the terms begin.
	for _, tt := range []struct{ day, name string }{
		{"2026-02-04", "立春"}, {"2026-03-20", "春分"}, {"2026-06-21", "夏至"},
		{"2026-09-23", "秋分"}, {"2026-11-22", "小雪"}, {"2026-12-22", "冬至"},
		{"2027-01-05", "小寒"}, {"2027-02-04", "立春"},
	} {
		d, err := time.ParseInLocation("2006-01-02", tt.day, chinaTime)
		if err != nil {
			t.Fatal(err)
		}
		term, starts := TermOn(d.Add(12 * time.Hour))
		if term.Name != tt.name || !starts {
			t.Errorf("%s: %s (starts %v), want %s beginning", tt.day, term.Name, starts, tt.name)
		}
	}
	term, starts := TermOn(time.Date(2026, 11, 25, 9, 0, 0, 0, chinaTime))
	if term.Name != "小雪" || starts {
		t.Errorf("late November is within 小雪: %s %v", term.Name, starts)
	}
}

func TestTheSeasonsWeatherOnTheTrail(t *testing.T) {
	members := []string{Master, "孙悟空"}
	for _, tt := range []struct {
		weather     string
		food, water int
	}{
		{"", StartFood - 6, StartWater - 2},
		{"cold", StartFood - 6 - 2*coldFood, StartWater - 2},
		{"heat", StartFood - 6, StartWater - 2 - 2*heatWater},
		{"rain", StartFood - 6, StartWater - 2 + rainWater},
	} {
		tr := NewTrail()
		tr.Weather = tt.weather
		tr.PassDay(members, "mountain", false)
		if tr.Food != tt.food || tr.Water != tt.water {
			t.Errorf("%q: food %d water %d, want %d and %d", tt.weather, tr.Food, tr.Water, tt.food, tt.water)
		}
	}
}
