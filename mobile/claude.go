package mobile

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	campnews "github.com/therealsiege/tuilaoshi/internal/camp"
	"github.com/therealsiege/tuilaoshi/internal/claude"
	"github.com/therealsiege/tuilaoshi/internal/journey"
	"github.com/therealsiege/tuilaoshi/internal/model"
	"github.com/therealsiege/tuilaoshi/internal/wiki"
)

// Claude on the phone goes through the game's server, signed in with the
// device's session token: no API key is ever on the phone. Each call takes
// the server and the token; the prompts and rules are the Mac's own.
//
// When the server says today's calls are used up, the error says so in
// the words of claude.ErrServerCap ("Claude limit"), and when it refuses
// the token, in those of claude.ErrSignedOut ("sign in again"), so the app
// can tell the learner what to do.

// claudeVia is a Claude client through the game's server.
func claudeVia(server, token string) (*claude.Client, error) {
	if strings.TrimSpace(server) == "" || token == "" {
		return nil, errors.New("sign in (Settings) first: Claude comes through the game's server")
	}
	return claude.NewServer(server, token, nil), nil
}

// WriteDiary grades today's diary entry with the Mac's task and criteria,
// keeps it if it passes, and lifts the party's morale as the Mac does. It
// returns the entry as JSON (date, day, text, score, feedback, kept); an
// entry scored 0 isn't kept and can be revised and sent again.
func (c *Core) WriteDiary(server, token, text string) (string, error) {
	cl, err := claudeVia(server, token)
	if err != nil {
		return "", err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	e, err := campnews.WriteDiary(ctx, c.store, cl.GradeWriting, text, time.Now())
	if err != nil {
		return "", err
	}
	data, err := json.Marshal(map[string]any{"date": e.Date, "day": e.Day, "text": e.Text, "score": e.Score, "feedback": e.Feedback, "kept": e.Score > 0})
	return string(data), err
}

// ClaudeGazette is this week's court gazette as JSON, as Gazette is, but
// written by Claude when it hasn't come out yet. When Claude can't write
// it the clerk sets it from the template, as on the Mac, and "missed"
// says why.
func (c *Core) ClaudeGazette(server, token string) (string, error) {
	cl, err := claudeVia(server, token)
	if err != nil {
		return "", err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	g, missed, err := campnews.PrintGazette(ctx, c.store, cl.WriteGazette, time.Now())
	if err != nil {
		return "", err
	}
	out := map[string]any{"week": g.Week, "text": g.Text, "english": g.English, "byHand": g.ByHand}
	if missed != nil {
		out["missed"] = missed.Error()
	}
	data, err := json.Marshal(out)
	return string(data), err
}

// storyID is a told chapter's id in Readings and Reading.
func storyID(n int) string { return fmt.Sprintf("story:%d", n) }

type toldChapter struct {
	ID     string `json:"id"`
	Number int    `json:"number"`
	Title  string `json:"title"`
	Date   string `json:"date"`
}

// Storyteller is the teahouse storyteller's shelf as JSON: the next
// chapter, whether it may be told today, and the chapters told so far,
// newest first, each to open with Reading.
func (c *Core) Storyteller() (string, error) {
	told, err := c.store.StoryChaptersTold()
	if err != nil {
		return "", err
	}
	next, ok := wiki.NextStoryChapter(told, time.Now())
	out := struct {
		Next     int           `json:"next"`
		CanTell  bool          `json:"canTell"`
		Done     bool          `json:"done"`
		Chapters []toldChapter `json:"chapters"`
	}{Next: next, CanTell: ok, Done: next > wiki.StoryChapters, Chapters: []toldChapter{}}
	for i := len(told) - 1; i >= 0; i-- {
		ch := told[i]
		out.Chapters = append(out.Chapters, toldChapter{storyID(ch.Number), ch.Number, ch.Title, ch.Date})
	}
	data, err := json.Marshal(out)
	return string(data), err
}

// TellChapter has the storyteller tell today's chapter in the words the
// learner knows, as on the Mac, and keeps it; one new chapter a day. It
// returns the chapter as JSON (id, number, title, date), to open with
// Reading.
func (c *Core) TellChapter(server, token string) (string, error) {
	cl, err := claudeVia(server, token)
	if err != nil {
		return "", err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Minute)
	defer cancel()
	ch, err := campnews.TellNextStory(ctx, c.store, cl.TellChapter, time.Now())
	if err != nil {
		return "", err
	}
	data, err := json.Marshal(toldChapter{storyID(ch.Number), ch.Number, ch.Title, ch.Date})
	return string(data), err
}

// storyReading is a told chapter's title and lines, a sentence a line,
// each carrying the chapter's English summary.
func (c *Core) storyReading(key string) (string, []readingLine, []model.Card, error) {
	n, err := strconv.Atoi(key)
	if err != nil {
		return "", nil, nil, fmt.Errorf("no chapter %q", key)
	}
	told, err := c.store.StoryChaptersTold()
	if err != nil {
		return "", nil, nil, err
	}
	for _, ch := range told {
		if ch.Number != n {
			continue
		}
		var lines []readingLine
		for _, s := range campnews.StorySentences(ch.Text) {
			l := readingLine{Tokens: []token{{Text: s}}}
			if ch.English != "" {
				l.English = "This chapter: " + ch.English
			}
			lines = append(lines, l)
		}
		return fmt.Sprintf("说书人 第%d回 %s", ch.Number, ch.Title), lines, journey.StoryGlossary(), nil
	}
	return "", nil, nil, fmt.Errorf("第%d回 hasn't been told", n)
}

// GradeAnswer asks Claude about an answer the app couldn't match exactly
// (mode recognition, production, listening or sentence), as the Mac's
// drill does. It returns JSON (correct, score 0-5, feedback, rating): the
// rating (0 again, 1 hard, 2 good, 3 easy) the score suggests, for the
// learner to accept or change. Nothing is recorded.
func (c *Core) GradeAnswer(server, token, cardID, answer, mode string) (string, error) {
	cl, err := claudeVia(server, token)
	if err != nil {
		return "", err
	}
	m := model.DrillMode(mode)
	switch m {
	case model.ModeRecognition, model.ModeProduction, model.ModeListening, model.ModeSentence:
	default:
		return "", fmt.Errorf("drill mode %q is not recognition, production, listening or sentence", mode)
	}
	cards, err := c.store.DrillableCards()
	if err != nil {
		return "", err
	}
	for _, cd := range cards {
		if cd.ID != cardID {
			continue
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()
		res, err := cl.EvaluateDrill(ctx, cd, answer, m)
		if err != nil {
			return "", err
		}
		data, err := json.Marshal(map[string]any{"correct": res.Correct, "score": res.Score, "feedback": res.Feedback, "rating": ratingFor(res)})
		return string(data), err
	}
	return "", fmt.Errorf("no card %q", cardID)
}

// ratingFor is the rating a graded answer suggests: wrong is again, then
// the score's 3 (with issues), 4 (good) and 5 (perfect) are hard, good
// and easy.
func ratingFor(r model.DrillResult) int {
	switch {
	case !r.Correct || r.Score <= 2:
		return 0
	case r.Score == 3:
		return 1
	case r.Score >= 5:
		return 3
	}
	return 2
}
