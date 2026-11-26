package app

import (
	"time"

	"github.com/therealsiege/tuilaoshi/internal/ui"
	"github.com/therealsiege/tuilaoshi/internal/wiki"
)

// meter counts Claude's use in the progress file and holds each day to the
// cap set in settings (ui.Prefs.ClaudeCap; 0 means no limit).
type meter struct{ store *wiki.Store }

// Allow says whether another call fits today's cap.
func (m meter) Allow() bool {
	if ui.Prefs.ClaudeCap <= 0 {
		return true
	}
	u, err := m.store.UsageOn(time.Now())
	return err != nil || u.Calls < ui.Prefs.ClaudeCap // a counting error never blocks play
}

// Used records a call.
func (m meter) Used(model string, input, output int64) {
	_ = m.store.RecordUsage(model, input, output, time.Now()) //nolint:errcheck // counting is best effort; play goes on
}
