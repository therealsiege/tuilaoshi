package camp

import (
	"context"
	"errors"
	"fmt"
	"math/rand/v2"
	"strings"
	"time"

	"github.com/therealsiege/tuilaoshi/internal/claude"
	"github.com/therealsiege/tuilaoshi/internal/journey"
	"github.com/therealsiege/tuilaoshi/internal/model"
	"github.com/therealsiege/tuilaoshi/internal/wiki"
)

// The teahouse storyteller (说书人), shared by the Mac's reader and the
// phone's: one new chapter of the novel a day, told by Claude in the words
// the learner knows and kept in the progress file.

// Teller writes chapter n of the novel using mostly the known words
// (claude.Client.TellChapter).
type Teller func(ctx context.Context, n int, known []string) (claude.Chapter, error)

// knownSample is how many known words the storyteller is given.
const knownSample = 300

// KnownWords picks the words to write the story in: words learned in
// reviews first, then words from HSK1-3 pages, a sample of each shuffled
// so the stories don't always lean on the same ones.
func KnownWords(cards []model.Card, pages []model.Page) []string {
	easy := map[string]bool{}
	for _, p := range pages {
		switch p.Level {
		case "HSK1", "HSK2", "HSK3":
			easy[p.Path] = true
		}
	}
	seen := map[string]bool{}
	var learned, basic []string
	for _, c := range cards {
		if c.Type == model.CardGrammar || c.Chinese == "" || seen[c.Chinese] {
			continue
		}
		page, _, _ := strings.Cut(c.ID, "#")
		switch {
		case c.ReviewInterval >= journey.LearnedInterval():
			learned = append(learned, c.Chinese)
		case easy[page]:
			basic = append(basic, c.Chinese)
		default:
			continue
		}
		seen[c.Chinese] = true
	}
	rand.Shuffle(len(learned), func(i, j int) { learned[i], learned[j] = learned[j], learned[i] })
	rand.Shuffle(len(basic), func(i, j int) { basic[i], basic[j] = basic[j], basic[i] })
	words := append(learned, basic...)
	return words[:min(len(words), knownSample)]
}

// StorySentences splits a chapter into one line per sentence, so a reader
// moves through it a sentence at a time.
func StorySentences(text string) []string {
	var lines []string
	var cur strings.Builder
	flush := func() {
		if s := strings.TrimSpace(cur.String()); s != "" {
			lines = append(lines, s)
		}
		cur.Reset()
	}
	runes := []rune(text)
	for i, r := range runes {
		if r == '\n' {
			flush()
			continue
		}
		cur.WriteRune(r)
		// End after 。！？, taking a closing quote with it.
		next := rune(0)
		if i+1 < len(runes) {
			next = runes[i+1]
		}
		if strings.ContainsRune("。！？", r) && next != '”' && next != '」' {
			flush()
		} else if strings.ContainsRune("”」", r) && i > 0 && strings.ContainsRune("。！？", runes[i-1]) {
			flush()
		}
	}
	flush()
	return lines
}

// ErrToldToday is returned when today's chapter has been told already.
var ErrToldToday = errors.New("today's chapter is told; the storyteller tells the next one tomorrow")

// ErrTaleTold is returned once all the chapters are told.
var ErrTaleTold = fmt.Errorf("the tale is told: all %d chapters", wiki.StoryChapters)

// TellStory has the storyteller tell chapter n, dated today, and keeps it.
func TellStory(ctx context.Context, store *wiki.Store, tell Teller, n int, known []string, now time.Time) (wiki.StoryChapter, error) {
	ch, err := tell(ctx, n, known)
	if err != nil {
		return wiki.StoryChapter{}, err
	}
	told := wiki.StoryChapter{Number: n, Title: ch.Title, Text: ch.Text, English: ch.English, Date: now.Format("2006-01-02")}
	for _, q := range ch.Questions {
		told.Questions = append(told.Questions, wiki.StoryQuestion{Ask: q.Ask, Choices: q.Choices})
	}
	return told, store.SaveStoryChapter(told)
}

// TellNextStory tells the next chapter if one may be told today, in the
// words the learner knows.
func TellNextStory(ctx context.Context, store *wiki.Store, tell Teller, now time.Time) (wiki.StoryChapter, error) {
	told, err := store.StoryChaptersTold()
	if err != nil {
		return wiki.StoryChapter{}, err
	}
	n, ok := wiki.NextStoryChapter(told, now)
	switch {
	case n > wiki.StoryChapters:
		return wiki.StoryChapter{}, ErrTaleTold
	case !ok:
		return wiki.StoryChapter{}, ErrToldToday
	}
	cards, err := store.DrillableCards()
	if err != nil {
		return wiki.StoryChapter{}, err
	}
	pages, err := store.ListPages()
	if err != nil {
		return wiki.StoryChapter{}, err
	}
	return TellStory(ctx, store, tell, n, KnownWords(cards, pages), now)
}
