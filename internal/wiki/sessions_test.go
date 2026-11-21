package wiki

import (
	"testing"
	"time"

	"github.com/cj/mandarin-wiki/internal/model"
)

func writeTestSession(t *testing.T, store *Store, start time.Time, items ...model.SessionItem) {
	t.Helper()
	sess := model.Session{StartedAt: start, EndedAt: start.Add(time.Minute), Type: "drill", Items: items}
	if err := store.WriteSession(sess); err != nil {
		t.Fatal(err)
	}
}

func TestTodayStatsCountsOnlyCorrectAnswers(t *testing.T) {
	store := newTestStore(t)
	writeTestSession(t, store, time.Now(),
		model.SessionItem{CardID: "a", UserAnswer: "book", Correct: true, Difficulty: model.DiffGood},
		model.SessionItem{CardID: "b", UserAnswer: "water", Correct: false, Difficulty: model.DiffAgain},
		model.SessionItem{CardID: "c", UserAnswer: "fire", Correct: false, Difficulty: model.DiffEasy},
	)

	got := store.TodayStats()

	if got.CardsReviewed != 3 || got.Correct != 1 {
		t.Errorf("TodayStats() = %d reviewed, %d correct; want 3 and 1", got.CardsReviewed, got.Correct)
	}
}

func TestParseResultRow(t *testing.T) {
	tests := []struct {
		line        string
		wantCorrect bool
		wantOK      bool
	}{
		{"| characters/a.md | book | correct | Good |", true, true},
		{"| characters/a.md | book | incorrect | Good |", false, true},
		{"| vocabulary/x.md#好 | either | or | correct | Hard |", true, true},
		{resultHeader, false, false},
		{"|------|--------|--------|------------|", false, false},
		{"| Word | Pinyin | English |", false, false},
		{"- **Cards**: 3", false, false},
	}
	for _, tt := range tests {
		correct, ok := parseResultRow(tt.line)
		if correct != tt.wantCorrect || ok != tt.wantOK {
			t.Errorf("parseResultRow(%q) = %v, %v; want %v, %v", tt.line, correct, ok, tt.wantCorrect, tt.wantOK)
		}
	}
}

func TestStreakAndWeekStats(t *testing.T) {
	store := newTestStore(t)
	now := time.Now()
	for _, daysAgo := range []int{0, 1, 2, 4} {
		writeTestSession(t, store, now.AddDate(0, 0, -daysAgo),
			model.SessionItem{CardID: "a", Correct: true})
	}

	if got := store.Streak(); got != 3 {
		t.Errorf("Streak() = %d, want 3", got)
	}

	week := store.WeekStats()
	wantReviewed := [7]int{0, 0, 1, 0, 1, 1, 1}
	for i, day := range week {
		if day.CardsReviewed != wantReviewed[i] || day.Correct != wantReviewed[i] {
			t.Errorf("day %d (%s): %d reviewed, %d correct; want %d", i, day.Date, day.CardsReviewed, day.Correct, wantReviewed[i])
		}
	}
}

func TestStreakStartsYesterdayWhenNothingToday(t *testing.T) {
	store := newTestStore(t)
	writeTestSession(t, store, time.Now().AddDate(0, 0, -1), model.SessionItem{CardID: "a"})

	if got := store.Streak(); got != 1 {
		t.Errorf("Streak() = %d, want 1", got)
	}
}

func TestParseSessionSkipsMalformedXP(t *testing.T) {
	rec := parseSession("2026-11-21", xpLinePrefix+" 12\n"+xpLinePrefix+" lots\n")
	if rec.XP != 12 {
		t.Errorf("XP = %d, want 12 (malformed line ignored)", rec.XP)
	}
}

func TestStatsReportsUnreadableWiki(t *testing.T) {
	store := NewStore(t.TempDir() + "/missing")
	if _, err := store.Stats(); err == nil {
		t.Error("Stats on a missing wiki returned no error")
	}
}
