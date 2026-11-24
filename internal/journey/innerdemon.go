package journey

import (
	"fmt"
	"time"
)

// The Inner Demon (心魔) is a weekly boss made of the learner's own
// mistakes: it gathers the words answered wrong since Monday and wakes once
// there are InnerDemonWakes of them, quizzing only those words.

// InnerDemonWakes is how many missed words in a week wake the Inner Demon.
const InnerDemonWakes = 8

// InnerDemon is the boss's bestiary entry.
var InnerDemon = BestiaryEntry{
	Name:      "心魔",
	English:   "The Inner Demon",
	TitleEng:  "the shape of your own mistakes",
	Origin:    "The novel's pilgrims fight demons outside and the restless mind within; this one is made of every word you stumbled over this week",
	Powers:    "Asks only what you got wrong",
	Chapter:   "Chapter 58",
	ImagePath: "assets/images/demons/心魔.png",
	Art:       "   .~~~.\n  ( o o )\n  /) ~ (\\\n  ~~~~~~~",
}

// WeekStart is the Monday that starts now's week.
func WeekStart(now time.Time) time.Time {
	d := dayOf(now)
	return d.AddDate(0, 0, -((int(d.Weekday()) + 6) % 7))
}

// WeekID names now's week, for the weekly reward ("2026-W48").
func WeekID(now time.Time) string {
	y, w := now.ISOWeek()
	return fmt.Sprintf("%d-W%02d", y, w)
}
