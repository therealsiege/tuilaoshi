package stats

import (
	"errors"
	"fmt"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/therealsiege/tuilaoshi/internal/journey"
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
	claude       wiki.Usage // today's calls to Claude
}

func loadHistory(store *wiki.Store, now time.Time) (history, error) {
	var h history
	var forecastErr, retentionErr, hardErr, cardsErr error
	h.forecast, forecastErr = store.DueForecast(now, 7)
	h.right, h.total, retentionErr = store.Retention(now, 30)
	var usageErr error
	h.claude, usageErr = store.UsageOn(now)
	retentionErr = errors.Join(retentionErr, usageErr)
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
	used := fmt.Sprintf("  Claude today: %d %s", m.hist.claude.Calls, plural(m.hist.claude.Calls, "call", "calls"))
	if m.hist.claude.Calls > 0 {
		used += fmt.Sprintf(", %s tokens", tokens(m.hist.claude.Input+m.hist.claude.Output))
	}
	if ui.Prefs.ClaudeCap > 0 {
		used += fmt.Sprintf(" (cap %d)", ui.Prefs.ClaudeCap)
	}
	lines = append(lines, ui.MutedStyle.Render(used))
	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

// tokens writes 12345 as "12.3k".
func tokens(n int64) string {
	if n < 1000 {
		return fmt.Sprint(n)
	}
	return fmt.Sprintf("%.1fk", float64(n)/1000)
}

func plural(n int, one, many string) string {
	if n == 1 {
		return one
	}
	return many
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

// hskLevels are the levels the known-words lines count.
var hskLevels = []string{"HSK1", "HSK2", "HSK3", "HSK4", "HSK5", "HSK6"}

// renderKnownByLevel counts, for each HSK level, the character and
// vocabulary pages learned (interval at the journey's learned mark) of all
// the pages at that level.
func (m Model) renderKnownByLevel() []string {
	known, all := map[string]int{}, map[string]int{}
	for _, p := range m.pages {
		if p.Type != "character" && p.Type != "vocabulary" {
			continue
		}
		all[p.Level]++
		if p.ReviewInterval >= journey.LearnedInterval() {
			known[p.Level]++
		}
	}
	lines := []string{"", ui.SubtitleStyle.Render("  Known by level")}
	for _, l := range hskLevels {
		if all[l] == 0 {
			continue
		}
		lines = append(lines, fmt.Sprintf("  %-5s %s %d/%d", l, miniBar(known[l]*15/all[l], 15), known[l], all[l]))
	}
	return lines
}
