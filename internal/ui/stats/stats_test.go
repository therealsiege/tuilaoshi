package stats

import (
	"strings"
	"testing"
	"time"

	"github.com/therealsiege/tuilaoshi/internal/journey"
	"github.com/therealsiege/tuilaoshi/internal/model"
	"github.com/therealsiege/tuilaoshi/internal/ui/uitest"
	"github.com/therealsiege/tuilaoshi/internal/wiki"
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

func TestHardWordsAndTheWeekAhead(t *testing.T) {
	store := uitest.Wiki(t, map[string]string{"characters/shu1.md": "---\ntype: character\nreview_interval: 3\n---\n\n# 书\n\n- **Character**: 书\n- **Pinyin**: shū\n- **Meaning**: book\n"})
	now := time.Date(2026, 11, 25, 9, 0, 0, 0, time.Local)
	cards, err := store.DrillableCards()
	if err != nil || len(cards) != 1 {
		t.Fatalf("cards %v %v", cards, err)
	}
	for range wiki.HardWordLapses {
		if err := store.RecordReview(&cards[0], model.DiffAgain, now); err != nil {
			t.Fatal(err)
		}
	}
	m := New(store)
	m.now = func() time.Time { return now }
	out := m.Refresh().View()
	for _, want := range []string{"Hard Words", "书", "×6", "Retention: 0% of 6 reviews", "Tomorrow"} {
		if want == "Tomorrow" {
			want = now.AddDate(0, 0, 1).Format("Mon") // the missed word is due tomorrow
		}
		if !strings.Contains(out, want) {
			t.Errorf("progress lacks %q:\n%s", want, out)
		}
	}
}

func TestKnownByLevel(t *testing.T) {
	m := Model{pages: []model.Page{
		{Type: "character", Level: "HSK1", ReviewInterval: journey.LearnedInterval()},
		{Type: "character", Level: "HSK1"},
		{Type: "vocabulary", Level: "HSK4", ReviewInterval: 90},
		{Type: "grammar", Level: "HSK4", ReviewInterval: 90},
	}}
	out := strings.Join(m.renderKnownByLevel(), "\n")
	if !strings.Contains(out, "HSK1") || !strings.Contains(out, "1/2") || !strings.Contains(out, "1/1") || strings.Contains(out, "HSK2") {
		t.Errorf("known by level:\n%s", out)
	}
}
