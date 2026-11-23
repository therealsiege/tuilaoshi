package journey

import (
	"fmt"
	"sort"
	"time"
)

// Festival is a Chinese festival on the real calendar (content/festivals.yaml).
type Festival struct {
	ID           string             `yaml:"id"`
	Name         string             `yaml:"name"`
	English      string             `yaml:"english"`
	Dates        []string           `yaml:"dates"` // YYYY-MM-DD, one per year
	WindowBefore int                `yaml:"window_before"`
	WindowAfter  int                `yaml:"window_after"`
	Desc         string             `yaml:"desc"`
	Words        []FestivalWord     `yaml:"words"`
	Text         string             `yaml:"text"`
	TextEnglish  string             `yaml:"text_english"`
	Questions    []FestivalQuestion `yaml:"questions"`
}

// FestivalWord is a word a festival teaches.
type FestivalWord struct {
	Chinese string `yaml:"chinese"`
	Pinyin  string `yaml:"pinyin"`
	Meaning string `yaml:"meaning"`
}

// FestivalQuestion asks about the festival's text; the right choice is first.
type FestivalQuestion struct {
	Question string   `yaml:"question"`
	Choices  []string `yaml:"choices"`
}

// Occasion is one year's festival: the festival and its date.
type Occasion struct {
	Festival
	Date time.Time
}

// Year names the occasion, for its quests and reward ("chunjie:2027").
func (o Occasion) Year() string { return fmt.Sprintf("%s:%d", o.ID, o.Date.Year()) }

// On reports whether the occasion's window includes now's day.
func (o Occasion) On(now time.Time) bool {
	day := dayOf(now)
	return !day.Before(o.Date.AddDate(0, 0, -o.WindowBefore)) && !day.After(o.Date.AddDate(0, 0, o.WindowAfter))
}

// DaysUntil counts days from now's day to the occasion's date.
func (o Occasion) DaysUntil(now time.Time) int { return int(o.Date.Sub(dayOf(now)).Hours() / 24) }

func dayOf(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.Local)
}

// Occasions lists every festival's dates, in order.
func Occasions() []Occasion {
	var out []Occasion
	for _, f := range content.Festivals {
		for _, d := range f.Dates {
			if t, err := time.ParseInLocation("2006-01-02", d, time.Local); err == nil {
				out = append(out, Occasion{Festival: f, Date: t})
			}
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Date.Before(out[j].Date) })
	return out
}

// FestivalNow returns the occasion on now, or else the next one, and
// whether it is on; ok is false when the table has run out.
func FestivalNow(now time.Time) (o Occasion, on, ok bool) {
	for _, occ := range Occasions() {
		if occ.On(now) {
			return occ, true, true
		}
	}
	for _, occ := range Occasions() {
		if occ.Date.After(dayOf(now)) {
			return occ, false, true
		}
	}
	return Occasion{}, false, false
}
