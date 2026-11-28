package camp

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/therealsiege/tuilaoshi/internal/journey"
	"github.com/therealsiege/tuilaoshi/internal/wiki"
)

// The travel diary (日记), shared by the Mac's camp and the phone's: once
// a day, a few sentences in Chinese about the day on the road, graded by
// Claude. A passing entry is kept and lifts the party's spirits on the
// trail; one that doesn't pass can be revised and sent again.

// DiaryMorale is the morale an entry lifts, by its grade (1 or 2).
var DiaryMorale = map[int]int{1: 5, 2: 10}

// DiaryCriteria is what an entry must do, for the grader.
const DiaryCriteria = "Two to four sentences of Chinese, in the first person, about a day on the journey west. Score 2 for natural, mostly correct writing; 1 for understandable writing with clear errors; 0 for writing that is not Chinese, off the task, or under two sentences."

// DiaryTask is the task for day n on the trail, for the grader.
func DiaryTask(day int) string {
	return fmt.Sprintf("写一篇简短的旅行日记：今天是取经路上的第%d天。写两三句话，说说今天走到了哪里、遇到了什么、心情怎么样。 (Write a short travel diary entry for day %d on the road west: two or three sentences on where you went, what happened and how you felt.)", day, day)
}

// Grader grades Chinese written for a task against criteria, 0-2, with
// feedback (claude.Client.GradeWriting).
type Grader func(ctx context.Context, task, criteria, answer string) (int, string, error)

// ErrDiaryWritten is returned when today's entry is already in the diary.
var ErrDiaryWritten = errors.New("today's diary entry is written; write the next one tomorrow")

// DiaryDay is the day on the trail, and whether today's entry is written.
func DiaryDay(store *wiki.Store, now time.Time) (day int, written bool, err error) {
	entries, err := store.Diary()
	if err != nil {
		return 0, false, err
	}
	cleared, err := store.ClearedStages()
	if err != nil {
		return 0, false, err
	}
	t, _, err := journey.LoadTrail(store, cleared)
	written = len(entries) > 0 && entries[len(entries)-1].Date == now.Format("2006-01-02")
	return t.Day, written, err
}

// KeepDiaryEntry keeps a passing entry and lifts the party's morale on the
// trail by its grade.
func KeepDiaryEntry(store *wiki.Store, e wiki.DiaryEntry) error {
	if err := store.WriteDiary(e); err != nil {
		return err
	}
	cleared, err := store.ClearedStages()
	if err != nil {
		return err
	}
	t, members, err := journey.LoadTrail(store, cleared)
	if err != nil {
		return err
	}
	t.Apply(journey.Outcome{Morale: DiaryMorale[e.Score]}, members)
	return store.SaveTrail(t)
}

// WriteDiary grades today's entry and keeps it if it passes. The entry
// comes back with its grade and feedback; a grade of 0 isn't kept, so it
// can be revised and sent again.
func WriteDiary(ctx context.Context, store *wiki.Store, grade Grader, text string, now time.Time) (wiki.DiaryEntry, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return wiki.DiaryEntry{}, errors.New("the entry is empty")
	}
	day, written, err := DiaryDay(store, now)
	if err != nil {
		return wiki.DiaryEntry{}, err
	}
	if written {
		return wiki.DiaryEntry{}, ErrDiaryWritten
	}
	score, feedback, err := grade(ctx, DiaryTask(day), DiaryCriteria, text)
	if err != nil {
		return wiki.DiaryEntry{}, err
	}
	e := wiki.DiaryEntry{Date: now.Format("2006-01-02"), Day: day, Text: text, Score: score, Feedback: feedback}
	if score == 0 {
		return e, nil
	}
	return e, KeepDiaryEntry(store, e)
}
