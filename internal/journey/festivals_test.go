package journey

import (
	"testing"
	"time"
)

func TestFestivalsAreCompleteAndCoverNextYear(t *testing.T) {
	nextYear := time.Now().Year() + 1
	for _, f := range content.Festivals {
		if f.Name == "" || f.English == "" || f.Desc == "" || f.Text == "" || f.TextEnglish == "" || len(f.Words) < 6 || len(f.Questions) < 2 {
			t.Errorf("festival %s is incomplete", f.ID)
		}
		for _, q := range f.Questions {
			if len(q.Choices) < 3 {
				t.Errorf("festival %s question %q has %d choices", f.ID, q.Question, len(q.Choices))
			}
		}
		covered := false
		for _, d := range f.Dates {
			day, err := time.Parse("2006-01-02", d)
			if err != nil {
				t.Errorf("festival %s date %q: %v", f.ID, d, err)
			}
			covered = covered || day.Year() == nextYear || (day.Year() == nextYear+1 && day.Month() == time.January)
		}
		if !covered {
			t.Errorf("festival %s has no date for %d: extend its dates", f.ID, nextYear)
		}
	}
}

func TestFestivalNowAndNext(t *testing.T) {
	at := func(s string) time.Time {
		d, _ := time.ParseInLocation("2006-01-02", s, time.Local) //nolint:errcheck // fixed dates
		return d.Add(10 * time.Hour)
	}
	if o, on, ok := FestivalNow(at("2026-11-23")); !ok || on || o.ID != "dongzhi" || o.DaysUntil(at("2026-11-23")) != 29 {
		t.Errorf("late November: %s on %v ok %v, in %d days", o.ID, on, ok, o.DaysUntil(at("2026-11-23")))
	}
	if o, on, _ := FestivalNow(at("2027-02-15")); !on || o.ID != "chunjie" {
		t.Errorf("nine days after the New Year: %s on %v", o.ID, on)
	}
	if o, on, _ := FestivalNow(at("2026-12-24")); !on || o.Year() != "dongzhi:2026" {
		t.Errorf("two days after the solstice: %s on %v", o.Year(), on)
	}
}

func TestWeeksStartOnMonday(t *testing.T) {
	sun := time.Date(2026, 11, 29, 20, 0, 0, 0, time.Local)
	if got := WeekStart(sun); got.Weekday() != time.Monday || got.Day() != 23 {
		t.Errorf("week of Sunday 29 Nov starts %v", got)
	}
	if WeekID(sun) != "2026-W48" {
		t.Errorf("week id %s", WeekID(sun))
	}
}
