package mobile

import (
	"encoding/json"
	"regexp"
	"time"

	campnews "github.com/therealsiege/tuilaoshi/internal/camp"
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

// Letter is Guanyin's latest letter as JSON (week, text, english), writing
// this week's first if it is due; "null" before any week with reviews.
func (c *Core) Letter() (string, error) {
	l, err := campnews.LatestLetter(c.store, time.Now())
	if err != nil {
		return "", err
	}
	if l == nil {
		return "null", nil
	}
	data, err := json.Marshal(map[string]string{"week": l.Week, "text": l.Text, "english": l.English})
	return string(data), err
}

// Gazette is this week's court gazette as JSON (week, text, english,
// byHand). Until the phone reaches Claude through the game's server, the
// court's clerk sets it from the template, as on a Mac without a key.
func (c *Core) Gazette() (string, error) {
	now := time.Now()
	g, err := campnews.WeekGazette(c.store, now)
	if err != nil {
		return "", err
	}
	if g == nil {
		clerk, _, err := campnews.ClerkGazette(c.store, now)
		if err != nil {
			return "", err
		}
		if err := c.store.SaveGazette(clerk); err != nil {
			return "", err
		}
		g = &clerk
	}
	data, err := json.Marshal(map[string]any{"week": g.Week, "text": g.Text, "english": g.English, "byHand": g.ByHand})
	return string(data), err
}

// Diary is the travel diary's entries as JSON, newest first. Writing an
// entry needs Claude to grade it, so on the phone it waits for the server.
func (c *Core) Diary() (string, error) {
	entries, err := c.store.Diary()
	if err != nil {
		return "", err
	}
	type entry struct {
		Date     string `json:"date"`
		Day      int    `json:"day"`
		Text     string `json:"text"`
		Score    int    `json:"score"`
		Feedback string `json:"feedback,omitempty"`
	}
	out := []entry{}
	for i := len(entries) - 1; i >= 0; i-- {
		e := entries[i]
		out = append(out, entry{e.Date, e.Day, e.Text, e.Score, e.Feedback})
	}
	data, err := json.Marshal(out)
	return string(data), err
}

// Progress is My progress as JSON: the last seven days' reviews, the
// reviews coming in the next seven, the hardest words, and the wiki's size.
func (c *Core) Progress() (string, error) {
	now := time.Now()
	type day struct {
		Date     string `json:"date"`
		Reviewed int    `json:"reviewed"`
		Correct  int    `json:"correct"`
	}
	type hard struct {
		Chinese string `json:"chinese"`
		Pinyin  string `json:"pinyin"`
		English string `json:"english"`
		Lapses  int    `json:"lapses"`
	}
	out := struct {
		Week       []day  `json:"week"`
		Ahead      []int  `json:"ahead"`
		Hard       []hard `json:"hard"`
		Streak     int    `json:"streak"`
		Characters int    `json:"characters"`
		Words      int    `json:"words"`
		Grammar    int    `json:"grammar"`
		Readings   int    `json:"readings"`
	}{Week: []day{}, Hard: []hard{}, Streak: c.store.Streak()}
	for _, d := range c.store.WeekStats() {
		out.Week = append(out.Week, day{d.Date, d.CardsReviewed, d.Correct})
	}
	ahead, err := c.store.DueForecast(now, 7)
	if err != nil {
		return "", err
	}
	out.Ahead = ahead
	hw, err := c.store.HardWords()
	if err != nil {
		return "", err
	}
	if len(hw) > 0 {
		cards, err := c.store.DrillableCards()
		if err != nil {
			return "", err
		}
		byID := map[string]int{}
		for i, cd := range cards {
			byID[cd.ID] = i
		}
		for _, h := range hw {
			if i, ok := byID[h.CardID]; ok && len(out.Hard) < 10 {
				out.Hard = append(out.Hard, hard{cards[i].Chinese, cards[i].Pinyin, cards[i].English, h.Lapses})
			}
		}
	}
	st, err := c.store.Stats()
	if err != nil {
		return "", err
	}
	out.Characters, out.Words, out.Grammar, out.Readings = st.CharacterPages, st.VocabPages, st.GrammarPages, st.ReadingPages
	data, err := json.Marshal(out)
	return string(data), err
}

// SpiderWeb is the spider's web (盘丝洞) around a word, as JSON: the word,
// the studied words tied to it by a shared character (slipping ones
// first), and the web's size. An empty or unknown centre opens on the
// weakest word, as the Mac does. Empty before any word is studied.
func (c *Core) SpiderWeb(centre string) (string, error) {
	studied, err := c.store.StudiedWords()
	if err != nil {
		return "", err
	}
	w := campnews.BuildWeb(studied, time.Now())
	type node struct {
		Word    string `json:"word"`
		Pinyin  string `json:"pinyin"`
		English string `json:"english"`
		Weak    bool   `json:"weak"`
		Shared  string `json:"shared,omitempty"`
		Links   int    `json:"links"`
	}
	out := struct {
		Centre      *node  `json:"centre,omitempty"`
		Links       []node `json:"links"`
		Words       int    `json:"words"`
		Strands     int    `json:"strands"`
		WeakStrands int    `json:"weakStrands"`
		Walk        int    `json:"walk"`
	}{Links: []node{}, Words: len(w.Words), Strands: w.Strands, WeakStrands: w.WeakStrands}
	i, ok := w.Index[centre]
	if !ok {
		i = w.Weakest()
	}
	if i >= 0 {
		cw := w.Words[i]
		out.Centre = &node{Word: cw.Card.Chinese, Pinyin: cw.Card.Pinyin, English: cw.Card.English, Weak: cw.Weak, Links: len(w.Links[i])}
		for _, j := range w.Links[i] {
			n := w.Words[j]
			out.Links = append(out.Links, node{n.Card.Chinese, n.Card.Pinyin, n.Card.English, n.Weak, campnews.Shared(cw.Card.Chinese, n.Card.Chinese), len(w.Links[j])})
		}
		out.Walk = len(w.Cluster(i))
	}
	data, err := json.Marshal(out)
	return string(data), err
}

// WalkWeb is the cards a walk of the web from a word revives, as the due
// cards' JSON: the word if it is slipping, then its slipping neighbours.
func (c *Core) WalkWeb(centre string) (string, error) {
	studied, err := c.store.StudiedWords()
	if err != nil {
		return "", err
	}
	w := campnews.BuildWeb(studied, time.Now())
	out := []card{}
	if i, ok := w.Index[centre]; ok {
		for _, cd := range w.Cluster(i) {
			out = append(out, toCard(cd))
		}
	}
	data, err := json.Marshal(out)
	return string(data), err
}
