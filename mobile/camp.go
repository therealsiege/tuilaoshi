package mobile

import (
	"encoding/json"
	"regexp"
	"time"

	"github.com/therealsiege/tuilaoshi/internal/journey"
	"github.com/therealsiege/tuilaoshi/internal/wiki"
)

// keyHint is a Mac key in a bounty's text, " (r)", which a phone has no use for.
var keyHint = regexp.MustCompile(`\s*\([a-zA-Z?]\)$`)

type bounty struct {
	Text   string `json:"text"`
	Have   int    `json:"have"`
	Target int    `json:"target"`
	Pay    int    `json:"pay"`
	Done   bool   `json:"done"`
}

type today struct {
	Reviewed  int      `json:"reviewed"`
	Correct   int      `json:"correct"`
	DailyGoal int      `json:"dailyGoal"`
	Bounties  []bounty `json:"bounties"`
	Shields   int      `json:"shields"`
	Festival  *struct {
		Name    string `json:"name"`
		English string `json:"english"`
		On      bool   `json:"on"`
		Days    int    `json:"days"`
	} `json:"festival,omitempty"`
	Term struct {
		Name    string `json:"name"`
		English string `json:"english"`
		Text    string `json:"text"`
		Starts  bool   `json:"starts"`
		Weather string `json:"weather"`
	} `json:"term"`
	Missed     int `json:"missed"`
	DemonWakes int `json:"demonWakes"`
}

// CampToday is camp's Today panel as JSON: the day's reviews against the
// daily goal, the bounties and shields, the festival on or coming, the
// solar term, and how near the week's missed words are to waking the
// Inner Demon.
func (c *Core) CampToday() (string, error) {
	now := time.Now()
	daily := c.store.TodayStats()
	bounties, shields, err := c.store.Bounties(now)
	if err != nil {
		return "", err
	}
	out := today{Reviewed: daily.CardsReviewed, Correct: daily.Correct, DailyGoal: wiki.DailyGoal,
		Bounties: []bounty{}, Shields: shields, DemonWakes: journey.InnerDemonWakes}
	for _, b := range bounties {
		out.Bounties = append(out.Bounties, bounty{keyHint.ReplaceAllString(b.Text, ""), b.Have, b.Target, b.Pay, b.Done})
	}
	if f, on, ok := journey.FestivalNow(now); ok {
		out.Festival = &struct {
			Name    string `json:"name"`
			English string `json:"english"`
			On      bool   `json:"on"`
			Days    int    `json:"days"`
		}{f.Name, f.English, on, f.DaysUntil(now)}
	}
	t, starts := journey.TermOn(now)
	out.Term.Name, out.Term.English, out.Term.Text, out.Term.Starts, out.Term.Weather = t.Name, t.English, t.Text, starts, t.Weather
	cards, err := c.store.DrillableCards()
	if err != nil {
		return "", err
	}
	known := map[string]bool{}
	for _, cd := range cards {
		if cd.Chinese != "" {
			known[cd.ID] = true
		}
	}
	for _, id := range c.store.MissedSince(journey.WeekStart(now)) {
		if known[id] {
			out.Missed++
		}
	}
	data, err := json.Marshal(out)
	return string(data), err
}

// settings are the learner's preferences as the app reads and writes them.
// They live in the progress file, so they sync with the Mac.
type settings struct {
	Speech      bool   `json:"speech"`
	Effects     bool   `json:"effects"`
	ToneColours bool   `json:"toneColours"`
	ReviewChunk int    `json:"reviewChunk"`
	NewWords    int    `json:"newWords"`
	Difficulty  string `json:"difficulty"`
	Scheduler   string `json:"scheduler"`
}

// Settings returns the learner's settings as JSON.
func (c *Core) Settings() (string, error) {
	s, err := c.store.Settings()
	if err != nil {
		return "", err
	}
	data, err := json.Marshal(settings{s.Speech, s.Effects, s.ToneColours, s.ReviewChunk, s.NewWords, s.Difficulty, s.Scheduler})
	return string(data), err
}

// SaveSettings writes the settings the app shows, keeping the Mac-only
// ones (input, encounters, portraits, the Claude cap) as they were, and
// applies the review scheduler at once.
func (c *Core) SaveSettings(settingsJSON string) error {
	var in settings
	if err := json.Unmarshal([]byte(settingsJSON), &in); err != nil {
		return err
	}
	s, err := c.store.Settings()
	if err != nil {
		return err
	}
	s.Speech, s.Effects, s.ToneColours, s.ReviewChunk, s.NewWords, s.Difficulty, s.Scheduler =
		in.Speech, in.Effects, in.ToneColours, in.ReviewChunk, in.NewWords, in.Difficulty, in.Scheduler
	if err := c.store.SaveSettings(s); err != nil {
		return err
	}
	saved, err := c.store.Settings() // normalised
	c.store.Scheduler = saved.Scheduler
	return err
}
