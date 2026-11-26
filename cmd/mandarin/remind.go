package main

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/therealsiege/tuilaoshi/internal/wiki"
)

// The daily reminder (-remind, scheduled by make reminder): a macOS
// notification saying how many studied words are due and whether today's
// study is still needed to keep the streak. Nothing is shown once today is
// done and nothing is due.

// reminder is the notification's text, or "" when there is nothing to say.
func reminder(store *wiki.Store, now time.Time) (string, error) {
	due, err := store.DueForecast(now, 1)
	if err != nil {
		return "", err
	}
	studied := store.TodayStats().CardsReviewed > 0
	streak := store.Streak()
	var parts []string
	if due[0] > 0 {
		parts = append(parts, fmt.Sprintf("%d %s due for review", due[0], map[bool]string{true: "word", false: "words"}[due[0] == 1]))
	}
	switch {
	case !studied && streak > 0:
		parts = append(parts, fmt.Sprintf("your %d-day streak ends tonight unless you study", streak))
	case !studied:
		parts = append(parts, "a few answers today start a streak")
	}
	if len(parts) == 0 {
		return "", nil
	}
	text := strings.Join(parts, "; ") + "."
	return strings.ToUpper(text[:1]) + text[1:], nil
}

// remind shows the day's reminder as a notification.
func remind(dir string) error {
	store := wiki.NewStore(dir)
	defer store.Close() //nolint:errcheck // exiting
	text, err := reminder(store, time.Now())
	if err != nil || text == "" {
		return err
	}
	script := fmt.Sprintf("display notification %q with title %q", text, "TuiLaoShi 推老师")
	if out, err := exec.Command("osascript", "-e", script).CombinedOutput(); err != nil {
		return fmt.Errorf("notification: %v %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}
