package stats

import (
	"errors"
	"fmt"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/therealsiege/tuilaoshi/internal/ui"
	"github.com/therealsiege/tuilaoshi/internal/wiki"
)

// What the review history shows: the week ahead, how much is remembered,
// and the hard words that keep slipping.

// shownHard is how many hard words the screen lists.
const shownHard = 8

// hardWord is a hard word as shown: the word and how often it was missed.
type hardWord struct {
	Chinese, Pinyin, English string
	Lapses                   int
}

// history is what Refresh reads from the review log.
type history struct {
	forecast     []int // studied words due on each of the next 7 days; [0] includes overdue
	right, total int   // reviews remembered, of those given in the last 30 days
	hard         []hardWord
	hardCount    int
}

func loadHistory(store *wiki.Store, now time.Time) (history, error) {
	var h history
	var forecastErr, retentionErr, hardErr, cardsErr error
	h.forecast, forecastErr = store.DueForecast(now, 7)
	h.right, h.total, retentionErr = store.Retention(now, 30)
	hard, hardErr := store.HardWords()
	h.hardCount = len(hard)
	if len(hard) > 0 {
		cards, err := store.DrillableCards()
		cardsErr = err
		byID := map[string]int{}
		for i, c := range cards {
			byID[c.ID] = i
		}
		for _, w := range hard[:min(len(hard), shownHard)] {
			if i, ok := byID[w.CardID]; ok {
				c := cards[i]
				h.hard = append(h.hard, hardWord{c.Chinese, c.Pinyin, c.English, w.Lapses})
			}
		}
	}
	return h, errors.Join(forecastErr, retentionErr, hardErr, cardsErr)
}

// renderAhead is the next seven days of studied words due, and retention.
func (m Model) renderAhead() string {
	lines := []string{ui.SubtitleStyle.Render("The Week Ahead"), ""}
	peak := 1
	for _, n := range m.hist.forecast {
		peak = max(peak, n)
	}
	now := m.now()
	for d, n := range m.hist.forecast {
		day := "Today"
		if d > 0 {
			day = now.AddDate(0, 0, d).Format("Mon")
		}
		lines = append(lines, fmt.Sprintf("  %-5s %s %4d", day, miniBar(n*16/peak, 16), n))
	}
	lines = append(lines, "")
	if m.hist.total == 0 {
		lines = append(lines, ui.MutedStyle.Render("  Retention: -- (no reviews in 30 days)"))
	} else {
		pct := m.hist.right * 100 / m.hist.total
		text := fmt.Sprintf("  Retention: %d%% of %d reviews (30 days)", pct, m.hist.total)
		switch {
		case pct >= 85:
			lines = append(lines, ui.SuccessStyle.Render(text))
		case pct >= 70:
			lines = append(lines, ui.WarningStyle.Render(text))
		default:
			lines = append(lines, ui.ErrorStyle.Render(text))
		}
	}
	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

// renderHardWords lists the words missed most, the leeches worth extra work.
func (m Model) renderHardWords() string {
	lines := []string{ui.SubtitleStyle.Render("Hard Words 顽词"), ""}
	if m.hist.hardCount == 0 {
		lines = append(lines, ui.MutedStyle.Render(fmt.Sprintf("  None yet: a word missed %d times", wiki.HardWordLapses)),
			ui.MutedStyle.Render("  shows up here."))
		return lipgloss.JoinVertical(lipgloss.Left, lines...)
	}
	for _, w := range m.hist.hard {
		lines = append(lines, fmt.Sprintf("  %s %s %s", ui.ChineseStyle.Render(w.Chinese), ui.PinyinStyle.Render(w.Pinyin),
			ui.MutedStyle.Render(fmt.Sprintf("×%d", w.Lapses))))
	}
	if more := m.hist.hardCount - len(m.hist.hard); more > 0 {
		lines = append(lines, ui.MutedStyle.Render(fmt.Sprintf("  and %d more", more)))
	}
	lines = append(lines, "", ui.MutedStyle.Render("  Try them in the forge (z)."))
	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}
