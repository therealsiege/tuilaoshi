// Package mobile is the game's Go engine as Swift sees it on iOS, built
// into an xcframework with gomobile bind. It keeps to types gomobile can
// carry (strings, numbers, errors and these structs); lists cross as JSON.
package mobile

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/therealsiege/tuilaoshi/internal/journey"
	"github.com/therealsiege/tuilaoshi/internal/model"
	"github.com/therealsiege/tuilaoshi/internal/wiki"
)

// Core is an open wiki and its progress file.
type Core struct {
	store *wiki.Store
}

// Camp is the read model used by the iOS camp.  Keeping this payload in the
// bridge means the Mac and iOS clients read the same progress store instead
// of maintaining a second copy of the game rules in Swift.
type Camp struct {
	Today    json.RawMessage     `json:"today"`
	Bounties []mobileBounty `json:"bounties"`
	Coins    wiki.Wallet   `json:"coins"`
	Letters  []wiki.Letter `json:"letters"`
	Gazettes []wiki.Gazette `json:"gazettes"`
	Diary    []wiki.DiaryEntry `json:"diary"`
	Due      int `json:"due"`
	Streak   int `json:"streak"`
	XP       int `json:"xp"`
}

type mobileBounty struct {
	Text string `json:"text"`
	Target int `json:"target"`
	Progress int `json:"progress"`
	Done bool `json:"done"`
}

// Camp returns the shared camp/dashboard data and its linked screens' data.
func (c *Core) Camp() (string, error) {
	today, err := Today()
	if err != nil { return "", err }
	var b Camp
	b.Today = json.RawMessage(today)
	bounties, _, err := c.store.Bounties(time.Now()); if err != nil { return "", err }
	for _, bounty := range bounties { b.Bounties = append(b.Bounties, mobileBounty{Text: bounty.Text, Target: bounty.Target, Progress: bounty.Have, Done: bounty.Done}) }
	b.Coins, err = c.store.Wallet(); if err != nil { return "", err }
	b.Letters, err = c.store.Letters(); if err != nil { return "", err }
	b.Gazettes, err = c.store.Gazettes(); if err != nil { return "", err }
	b.Diary, err = c.store.Diary(); if err != nil { return "", err }
	b.Due, err = c.DueCount(); if err != nil { return "", err }
	b.Streak = c.store.Streak()
	b.XP = c.store.TotalXP()
	data, err := json.Marshal(b)
	return string(data), err
}

// Settings returns the preferences shared with the Mac progress store.
func (c *Core) Settings() (string, error) {
	set, err := c.store.Settings(); if err != nil { return "", err }
	b, err := json.Marshal(set); return string(b), err
}

// Open opens the wiki in wikiDir (the app bundle's copy, read-only) with
// its progress kept in progressDir (the app's own container).
func Open(wikiDir, progressDir string) (*Core, error) {
	s := wiki.NewStore(wikiDir)
	s.Progress = progressDir
	if _, err := s.Settings(); err != nil {
		return nil, fmt.Errorf("open progress: %w", err)
	}
	return &Core{store: s}, nil
}

// Close closes the progress file.
func (c *Core) Close() error { return c.store.Close() }

// Card is a flashcard as the app shows it.
type card struct {
	ID       string   `json:"id"`
	Chinese  string   `json:"chinese"`
	Pinyin   string   `json:"pinyin"`
	English  string   `json:"english"`
	Examples []string `json:"examples,omitempty"`
	Interval int      `json:"interval"`
	Next     string   `json:"next,omitempty"`
}

func toCard(c model.Card) card {
	out := card{ID: c.ID, Chinese: c.Chinese, Pinyin: c.Pinyin, English: c.English, Examples: c.Examples, Interval: c.ReviewInterval}
	if !c.NextReview.IsZero() {
		out.Next = wiki.FormatDate(c.NextReview)
	}
	return out
}

// DueCount is how many cards are due today.
func (c *Core) DueCount() (int, error) {
	due, err := wiki.DueItems(c.store, time.Now())
	return len(due), err
}

// DueCards returns up to limit due cards as a JSON list.
func (c *Core) DueCards(limit int) (string, error) {
	due, err := wiki.DueItems(c.store, time.Now())
	if err != nil {
		return "", err
	}
	out := make([]card, 0, min(limit, len(due)))
	for _, d := range due[:min(limit, len(due))] {
		out = append(out, toCard(d))
	}
	data, err := json.Marshal(out)
	return string(data), err
}

// Review records an answer to a card (rating 0 again, 1 hard, 2 good,
// 3 easy) and returns the card's new schedule as JSON.
func (c *Core) Review(cardID string, rating int) (string, error) {
	if rating < 0 || rating > 3 {
		return "", fmt.Errorf("rating %d is not 0-3", rating)
	}
	cards, err := c.store.DrillableCards()
	if err != nil {
		return "", err
	}
	for i := range cards {
		if cards[i].ID == cardID {
			if err := c.store.RecordReview(&cards[i], model.DrillDifficulty(rating), time.Now()); err != nil {
				return "", err
			}
			data, err := json.Marshal(toCard(cards[i]))
			return string(data), err
		}
	}
	return "", fmt.Errorf("no card %q", cardID)
}

// Today is a few things the journey's embedded content says about today,
// as JSON: the solar term and how many tribulations there are, so the app
// can check the content loaded.
func Today() (string, error) {
	term, starts := journey.TermOn(time.Now())
	data, err := json.Marshal(map[string]any{
		"term": term.Name, "termEnglish": term.English, "termStarts": starts,
		"tribulations": len(journey.AllTribulations()),
	})
	return string(data), err
}
