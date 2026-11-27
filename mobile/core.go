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

// camp is the numbers camp shows, as the app reads them.
type camp struct {
	Due      int        `json:"due"`
	Streak   int        `json:"streak"`
	Tongqian int        `json:"tongqian"`
	Yuanbao  int        `json:"yuanbao"`
	Rank     string     `json:"rank"`
	RankEN   string     `json:"rankEnglish"`
	Level    int        `json:"level"`
	Learned  int        `json:"learned"`
	Mastered int        `json:"mastered"`
	Stage    *campStage `json:"stage,omitempty"`
	Party    []string   `json:"party"`
	Term     string     `json:"term"`
	TermEN   string     `json:"termEnglish"`
}

type campStage struct {
	Number   int    `json:"number"`
	Name     string `json:"name"`
	English  string `json:"english"`
	Location string `json:"location"`
	Region   string `json:"region"`
	Demon    string `json:"demon"`
	DemonEN  string `json:"demonEnglish"`
}

// Camp gathers camp's numbers in one call, as JSON: cards due, the streak,
// the purse, rank and level, the stage being worked on, who has joined the
// party, and today's solar term.
func (c *Core) Camp() (string, error) {
	due, err := wiki.DueItems(c.store, time.Now())
	if err != nil {
		return "", err
	}
	state, err := journey.Calculate(c.store)
	if err != nil {
		return "", err
	}
	purse, err := c.store.Wallet()
	if err != nil {
		return "", err
	}
	cleared, err := c.store.ClearedStages()
	if err != nil {
		return "", err
	}
	term, _ := journey.TermOn(time.Now())
	out := camp{
		Due: len(due), Streak: c.store.Streak(), Tongqian: purse.Tongqian, Yuanbao: purse.Yuanbao,
		Rank: state.Rank.Name, RankEN: state.Rank.English, Level: state.Level.N,
		Learned: state.Counts.Learned, Mastered: state.Counts.Mastered,
		Party: []string{}, Term: term.Name, TermEN: term.English,
	}
	for _, p := range journey.Party(cleared) {
		if p.Unlocked {
			out.Party = append(out.Party, p.Name)
		}
	}
	if s := state.Stage; s != nil {
		r := journey.Regions()[journey.RegionOf(s.Number)]
		out.Stage = &campStage{Number: s.Number, Name: s.Name, English: s.English, Location: s.Location,
			Region: r.Name, Demon: s.Opponent.Name, DemonEN: s.Opponent.English}
	}
	data, err := json.Marshal(out)
	return string(data), err
}
