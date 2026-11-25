package wiki

import (
	"testing"
	"time"
)

func TestOneStoryChapterADay(t *testing.T) {
	day := time.Date(2026, 11, 25, 20, 0, 0, 0, time.Local)
	if n, ok := NextStoryChapter(nil, day); n != 1 || !ok {
		t.Errorf("a new learner: chapter %d, ok %v; want 1, true", n, ok)
	}
	told := []StoryChapter{{Number: 1, Date: "2026-11-24"}, {Number: 2, Date: "2026-11-25"}}
	if n, ok := NextStoryChapter(told, day); n != 3 || ok {
		t.Errorf("after today's chapter: %d, %v; want 3, false", n, ok)
	}
	if n, ok := NextStoryChapter(told, day.AddDate(0, 0, 1)); n != 3 || !ok {
		t.Errorf("the next day: %d, %v; want 3, true", n, ok)
	}
	if n, ok := NextStoryChapter([]StoryChapter{{Number: StoryChapters, Date: "2026-01-01"}}, day); ok || n != StoryChapters+1 {
		t.Errorf("after the last chapter: %d, %v; want nothing more to tell", n, ok)
	}
}

func TestStoryChaptersAreKept(t *testing.T) {
	store := NewStore(t.TempDir())
	if told, err := store.StoryChaptersTold(); err != nil || len(told) != 0 {
		t.Fatalf("fresh store: %v, %v", told, err)
	}
	for _, c := range []StoryChapter{
		{Number: 2, Title: "二", Text: "第二回。", Date: "2026-11-25"},
		{Number: 1, Title: "一", Text: "第一回。", Date: "2026-11-24"},
		{Number: 2, Title: "二 again", Text: "又是第二回。", Date: "2026-11-25"},
	} {
		if err := store.SaveStoryChapter(c); err != nil {
			t.Fatal(err)
		}
	}
	told, err := store.StoryChaptersTold()
	if err != nil {
		t.Fatal(err)
	}
	if len(told) != 2 || told[0].Number != 1 || told[1].Title != "二 again" {
		t.Errorf("told %+v; want chapters 1 and 2 in order, 2 replaced", told)
	}
}
