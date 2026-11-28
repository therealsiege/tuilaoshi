// Package mobile is the game's Go engine as Swift sees it on iOS, built
// into an xcframework with gomobile bind. It keeps to types gomobile can
// carry (strings, numbers, errors and these structs); lists cross as JSON.
package mobile

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"sync"
	"time"

	"github.com/therealsiege/tuilaoshi/internal/episode"
	"github.com/therealsiege/tuilaoshi/internal/journey"
	"github.com/therealsiege/tuilaoshi/internal/model"
	"github.com/therealsiege/tuilaoshi/internal/remote"
	"github.com/therealsiege/tuilaoshi/internal/road"
	"github.com/therealsiege/tuilaoshi/internal/wiki"
)

// Core is an open wiki and its progress file.
type Core struct {
	store *wiki.Store

	mu    sync.Mutex
	walk  *road.Walk // the road as last loaded (see road.go); nil: load it
	event *openEvent // an event or crossing being met (see trail.go)
	plot  *openPlot  // demons overheard plotting
	gate  *road.Gate // the grammar gate being taken (see gate.go)
}

// Open opens the wiki in wikiDir (the app bundle's copy, read-only) with
// its progress, and the pages the game writes, kept in progressDir (the
// app's own container).
func Open(wikiDir, progressDir string) (*Core, error) {
	s := wiki.NewStore(wikiDir)
	s.Progress = progressDir
	s.Pages = filepath.Join(progressDir, "pages") // the bundled wiki is read-only
	set, err := s.Settings()
	if err != nil {
		return nil, fmt.Errorf("open progress: %w", err)
	}
	s.Scheduler = set.Scheduler
	return &Core{store: s}, nil
}

// Sync merges in every other device's folder under root (the shared
// iCloud Drive folder), then writes this device's progress to root/name,
// the same format the Mac's TUILAOSHI_SYNC_DIR uses.
func (c *Core) Sync(root, name string) error {
	defer c.forgetRoad() // the merge may move the pilgrim
	return c.store.Sync(root, name)
}

// Export writes the progress as readable files into dir.
func (c *Core) Export(dir string) error { return c.store.Export(dir) }

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
	Cloze    string   `json:"cloze,omitempty"` // an example with the word blanked, for sentence mode
	Sentence string   `json:"sentence,omitempty"`
}

func toCard(c model.Card) card {
	out := card{ID: c.ID, Chinese: c.Chinese, Pinyin: c.Pinyin, English: c.English, Examples: c.Examples, Interval: c.ReviewInterval}
	if !c.NextReview.IsZero() {
		out.Next = wiki.FormatDate(c.NextReview)
	}
	for _, ex := range c.Examples {
		if cloze := wiki.Cloze(ex, c.Chinese); cloze != "" {
			out.Cloze, out.Sentence = cloze, ex
			break
		}
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

// answer is one answer in a session, as the app reports it.
type answer struct {
	ID      string `json:"id"`
	Correct bool   `json:"correct"`
	Rating  int    `json:"rating"`
}

// FinishSession records a study session (kind "review" or "drill") from
// its start time (Unix seconds) and its answers as JSON, so the streak,
// the day's stats and copper count it. Each answer's review is already
// saved by Review; this only writes the session. No answers, no session.
func (c *Core) FinishSession(kind string, started int64, answersJSON string) error {
	var answers []answer
	if err := json.Unmarshal([]byte(answersJSON), &answers); err != nil {
		return fmt.Errorf("answers: %w", err)
	}
	if len(answers) == 0 {
		return nil
	}
	if kind != "review" && kind != "drill" {
		return fmt.Errorf("session kind %q is not review or drill", kind)
	}
	sess := model.Session{Type: kind, StartedAt: time.Unix(started, 0), EndedAt: time.Now(), CardsTotal: len(answers)}
	for _, a := range answers {
		if a.Correct {
			sess.CardsCorrect++
		}
		sess.Items = append(sess.Items, model.SessionItem{CardID: a.ID, Correct: a.Correct, Difficulty: model.DrillDifficulty(a.Rating)})
	}
	return c.store.WriteSession(sess)
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

// Shadow scores a line said back against the line, character by character,
// as the Mac's shadowing does, and returns the result as JSON.
func Shadow(line, heard string) (string, error) {
	type char struct {
		Char  string `json:"char"`
		Heard bool   `json:"heard"`
	}
	sh := episode.CompareShadow(line, heard)
	out := struct {
		Chars   []char `json:"chars"`
		Percent int    `json:"percent"`
	}{Chars: []char{}, Percent: sh.Percent}
	for _, c := range sh.Chars {
		out.Chars = append(out.Chars, char{string(c.Char), c.Heard})
	}
	data, err := json.Marshal(out)
	return string(data), err
}

// SignInWithApple trades the identity token from Sign in with Apple for
// this device's session with the game's server, as JSON (token, device,
// expiresAt).
func SignInWithApple(server, identityToken, device string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	s, err := remote.SignInWithApple(ctx, server, identityToken, device)
	if err != nil {
		return "", err
	}
	data, err := json.Marshal(s)
	return string(data), err
}

// SyncWithServer merges the other devices' progress from the game's server
// and uploads this device's.
func (c *Core) SyncWithServer(server, token, device string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	defer c.forgetRoad() // the merge may move the pilgrim
	return remote.Client{Base: server, Token: token}.Sync(ctx, c.store, device)
}

// PairingCode asks the server for a code the Mac claims with -pair.
func PairingCode(server, token string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	code, _, err := remote.Client{Base: server, Token: token}.PairingCode(ctx)
	return code, err
}
