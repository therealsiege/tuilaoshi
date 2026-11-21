package stats

import (
	"strings"
	"testing"

	"github.com/therealsiege/tuilaoshi/internal/journey"
	"github.com/therealsiege/tuilaoshi/internal/model"
)

func TestReviewLevel(t *testing.T) {
	tests := []struct {
		interval int
		want     string
	}{
		{0, "New"}, {1, "Learning"}, {3, "Learning"}, {4, "Familiar"},
		{14, "Familiar"}, {15, "Confident"}, {30, "Confident"}, {31, "Long-term"},
	}
	for _, tt := range tests {
		if got := reviewLevel(tt.interval); got != tt.want {
			t.Errorf("reviewLevel(%d) = %q, want %q", tt.interval, got, tt.want)
		}
	}
}

func TestMasteredCountMatchesJourney(t *testing.T) {
	pages := []model.Page{
		{Type: "character", ReviewInterval: 7},
		{Type: "character", ReviewInterval: 40},
		{Type: "character", ReviewInterval: 6},
		{Type: "vocabulary", ReviewInterval: 90},
	}
	m := Model{pages: pages}
	out := m.renderMasteryLevels()
	want := "Characters mastered: 2"
	if journey.CountMastered(pages) != 2 || !strings.Contains(out, want) {
		t.Errorf("stats should show %q (journey count %d); got:\n%s", want, journey.CountMastered(pages), out)
	}
	if strings.Contains(out, "Mastered ") {
		t.Errorf("review levels should not use the word Mastered:\n%s", out)
	}
}
