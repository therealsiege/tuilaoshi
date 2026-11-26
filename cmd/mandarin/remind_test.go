package main

import (
	"strings"
	"testing"
	"time"

	"github.com/therealsiege/tuilaoshi/internal/wiki"
)

func TestTheReminderSaysWhatTodayNeeds(t *testing.T) {
	store := wiki.NewStore(t.TempDir())
	defer store.Close() //nolint:errcheck // test
	text, err := reminder(store, time.Now())
	if err != nil || !strings.Contains(text, "start a streak") {
		t.Errorf("a new learner: %q %v", text, err)
	}
}
