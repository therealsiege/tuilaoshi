package camp

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/therealsiege/tuilaoshi/internal/claude"
	"github.com/therealsiege/tuilaoshi/internal/journey"
	"github.com/therealsiege/tuilaoshi/internal/model"
	"github.com/therealsiege/tuilaoshi/internal/wiki"
)

func TestStorySentences(t *testing.T) {
	got := StorySentences("他说：“走吧！”我们走了。真的？\n好。")
	want := []string{"他说：“走吧！”", "我们走了。", "真的？", "好。"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("sentences %q, want %q", got, want)
	}
}

func TestKnownWordsPreferLearnedAndEasyWords(t *testing.T) {
	pages := []model.Page{{Path: "vocabulary/easy.md", Level: "HSK2"}, {Path: "vocabulary/hard.md", Level: "HSK6"}}
	cards := []model.Card{
		{ID: "vocabulary/easy.md#吃饭", Chinese: "吃饭"},
		{ID: "vocabulary/hard.md#踌躇", Chinese: "踌躇"},
		{ID: "vocabulary/hard.md#斟酌", Chinese: "斟酌", ReviewInterval: 365},
		{ID: "grammar/ba.md", Type: model.CardGrammar, Chinese: "把", ReviewInterval: 365},
	}
	got := KnownWords(cards, pages)
	if !reflect.DeepEqual(got, []string{"斟酌", "吃饭"}) {
		t.Errorf("known words %q; want the learned word, then the HSK2 one", got)
	}
}

func TestTheStorytellerTellsOneChapterADay(t *testing.T) {
	store := wiki.NewStore(t.TempDir())
	var asked []int
	tell := func(_ context.Context, n int, _ []string) (claude.Chapter, error) {
		asked = append(asked, n)
		return claude.Chapter{Title: "石猴出世", Text: "猴子出来了。", Questions: []claude.Question{{Ask: "谁？", Choices: []string{"猴子", "猪"}}}}, nil
	}
	day := time.Date(2026, 11, 24, 20, 0, 0, 0, time.Local)
	ch, err := TellNextStory(t.Context(), store, tell, day)
	if err != nil || ch.Number != 1 || ch.Date != "2026-11-24" || len(ch.Questions) != 1 {
		t.Fatalf("told %+v %v", ch, err)
	}
	if _, err := TellNextStory(t.Context(), store, tell, day); !errors.Is(err, ErrToldToday) {
		t.Errorf("a second chapter the same day: %v", err)
	}
	ch, err = TellNextStory(t.Context(), store, tell, day.AddDate(0, 0, 1))
	if err != nil || ch.Number != 2 {
		t.Fatalf("the next day %+v %v", ch, err)
	}
	told, err := store.StoryChaptersTold()
	if err != nil || len(told) != 2 || !reflect.DeepEqual(asked, []int{1, 2}) {
		t.Errorf("kept %d, asked %v, %v", len(told), asked, err)
	}
}

func TestADiaryEntryIsGradedKeptAndLiftsSpirits(t *testing.T) {
	store := wiki.NewStore(t.TempDir())
	day := time.Date(2026, 11, 24, 20, 0, 0, 0, time.Local)
	tr, _, err := journey.LoadTrail(store, nil)
	if err != nil {
		t.Fatal(err)
	}
	tr.Morale[journey.Master] = 50
	if err := store.SaveTrail(tr); err != nil {
		t.Fatal(err)
	}
	score := 0
	grade := func(_ context.Context, task, criteria, answer string) (int, string, error) {
		if task != DiaryTask(1) || criteria != DiaryCriteria || answer != "今天我们走了很远。我很累。" {
			t.Errorf("graded %q for %q", answer, task)
		}
		return score, "很好。", nil
	}
	e, err := WriteDiary(t.Context(), store, grade, "  今天我们走了很远。我很累。 ", day)
	if err != nil || e.Score != 0 {
		t.Fatalf("a failing entry %+v %v", e, err)
	}
	if entries, err := store.Diary(); err != nil || len(entries) != 0 {
		t.Errorf("a failing entry was kept: %v", err)
	}
	score = 2
	e, err = WriteDiary(t.Context(), store, grade, "今天我们走了很远。我很累。", day)
	if err != nil || e.Score != 2 || e.Day != 1 || e.Feedback != "很好。" {
		t.Fatalf("a passing entry %+v %v", e, err)
	}
	if entries, err := store.Diary(); err != nil || len(entries) != 1 {
		t.Errorf("kept %d entries, %v", len(entries), err)
	}
	tr, _, err = journey.LoadTrail(store, nil)
	if err != nil || tr.Morale[journey.Master] != 50+DiaryMorale[2] {
		t.Errorf("morale %d, %v", tr.Morale[journey.Master], err)
	}
	if _, err := WriteDiary(t.Context(), store, grade, "又写了。", day); !errors.Is(err, ErrDiaryWritten) {
		t.Errorf("a second entry the same day: %v", err)
	}
}
