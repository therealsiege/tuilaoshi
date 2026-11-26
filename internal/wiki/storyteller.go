package wiki

import (
	"sort"
	"time"
)

// storytellerFile keeps the teahouse storyteller's chapters: each is
// written by Claude once, the day it is first told, and kept so it can be
// read again. It is progress, not content, so it never touches the pages.
const storytellerFile = "storyteller.yaml"

// StoryChapters is how many chapters the novel has.
const StoryChapters = 100

// StoryChapter is one retold chapter of Journey to the West.
type StoryChapter struct {
	Number  int    `yaml:"number"`
	Title   string `yaml:"title"`
	Text    string `yaml:"text"`
	English string `yaml:"english,omitempty"` // a short summary
	Date    string `yaml:"date"`              // the day it was told, YYYY-MM-DD

	Questions []StoryQuestion `yaml:"questions,omitempty"` // for hearing it first (right choice first)
}

// StoryQuestion is a question about a chapter, its right choice first.
type StoryQuestion struct {
	Ask     string   `yaml:"ask"`
	Choices []string `yaml:"choices"`
}

type storyteller struct {
	Chapters []StoryChapter `yaml:"chapters"`
}

// StoryChaptersTold returns the chapters told so far, in order.
func (s *Store) StoryChaptersTold() ([]StoryChapter, error) {
	var rec storyteller
	if err := s.readYAML(storytellerFile, &rec); err != nil {
		return nil, err
	}
	sort.Slice(rec.Chapters, func(i, j int) bool { return rec.Chapters[i].Number < rec.Chapters[j].Number })
	return rec.Chapters, nil
}

// SaveStoryChapter keeps a told chapter, replacing one with its number.
func (s *Store) SaveStoryChapter(c StoryChapter) error {
	told, err := s.StoryChaptersTold()
	if err != nil {
		return err
	}
	kept := []StoryChapter{c}
	for _, t := range told {
		if t.Number != c.Number {
			kept = append(kept, t)
		}
	}
	sort.Slice(kept, func(i, j int) bool { return kept[i].Number < kept[j].Number })
	return s.writeYAML(storytellerFile, storyteller{Chapters: kept})
}

// NextStoryChapter is the chapter to tell next, and whether it may be told
// now: one new chapter a calendar day, until the hundredth.
func NextStoryChapter(told []StoryChapter, now time.Time) (next int, ok bool) {
	today := now.Format(dateFmt)
	next = 1
	toldToday := false
	for _, c := range told {
		if c.Number >= next {
			next = c.Number + 1
		}
		if c.Date == today {
			toldToday = true
		}
	}
	return next, next <= StoryChapters && !toldToday
}
