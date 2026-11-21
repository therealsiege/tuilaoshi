package wiki

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/cj/mandarin-wiki/internal/model"
)

// Session files in wiki/sessions/ record each drill, review, and battle.
// WriteSession is the only writer and readSessions the only reader, so the
// file format lives in this file alone.

const (
	sessionsDir = "sessions"
	// xpLinePrefix starts the session-file line that records experience earned.
	xpLinePrefix = "- **XP**: "
	resultHeader = "| Card | Answer | Result | Difficulty |"
)

// DailyStats holds aggregated stats for a single day.
type DailyStats struct {
	Date          string
	CardsReviewed int
	Correct       int
	NewCards      int
}

// sessionRecord is what the stats read back from one session file.
type sessionRecord struct {
	Date    string // YYYY-MM-DD, from the file name
	Answers int
	Correct int
	XP      int
}

// WriteSession saves a drill, review, or battle session as a markdown file in
// wiki/sessions/.
func (s *Store) WriteSession(sess model.Session) error {
	sessDir := filepath.Join(s.Dir, sessionsDir)
	if err := os.MkdirAll(sessDir, 0755); err != nil {
		return fmt.Errorf("create sessions dir: %w", err)
	}

	filename := sess.StartedAt.Format("2006-01-02-150405") + ".md"
	path := filepath.Join(sessDir, filename)

	var b strings.Builder
	b.WriteString("---\n")
	b.WriteString("type: session\n")
	b.WriteString("status: active\n")
	fmt.Fprintf(&b, "last_reviewed: %s\n", FormatDate(time.Now()))
	fmt.Fprintf(&b, "tags: [%s]\n", sess.Type)
	b.WriteString("---\n\n")

	fmt.Fprintf(&b, "# Session — %s\n\n", sess.StartedAt.Format("2006-01-02 15:04"))
	fmt.Fprintf(&b, "- **Type**: %s\n", sess.Type)
	fmt.Fprintf(&b, "- **Duration**: %s\n", sess.EndedAt.Sub(sess.StartedAt).Truncate(time.Second))
	fmt.Fprintf(&b, "- **Cards**: %d\n", sess.CardsTotal)
	fmt.Fprintf(&b, "- **Correct**: %d\n", sess.CardsCorrect)
	if sess.CardsTotal > 0 {
		pct := float64(sess.CardsCorrect) / float64(sess.CardsTotal) * 100
		fmt.Fprintf(&b, "- **Accuracy**: %.0f%%\n", pct)
	}
	if sess.XP > 0 {
		fmt.Fprintf(&b, "%s%d\n", xpLinePrefix, sess.XP)
	}
	b.WriteString("\n## Results\n\n")
	b.WriteString(resultHeader + "\n")
	b.WriteString("|------|--------|--------|------------|\n")
	for _, item := range sess.Items {
		result := "incorrect"
		if item.Correct {
			result = "correct"
		}
		diff := difficultyLabel(item.Difficulty)
		fmt.Fprintf(&b, "| %s | %s | %s | %s |\n",
			item.CardID, item.UserAnswer, result, diff)
	}

	return os.WriteFile(path, []byte(b.String()), 0644)
}

func difficultyLabel(d model.DrillDifficulty) string {
	switch d {
	case model.DiffAgain:
		return "Again"
	case model.DiffHard:
		return "Hard"
	case model.DiffGood:
		return "Good"
	case model.DiffEasy:
		return "Easy"
	default:
		return "Good"
	}
}

// readSessions parses every session file. Unreadable files are skipped.
func (s *Store) readSessions() []sessionRecord {
	entries, err := os.ReadDir(filepath.Join(s.Dir, sessionsDir))
	if err != nil {
		return nil
	}

	var records []sessionRecord
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".md") || len(name) < len(dateFmt) {
			continue
		}
		data, err := os.ReadFile(filepath.Join(s.Dir, sessionsDir, name))
		if err != nil {
			continue
		}
		records = append(records, parseSession(name[:len(dateFmt)], string(data)))
	}
	return records
}

// parseSession reads the XP line and the results table of one session file.
func parseSession(date, content string) sessionRecord {
	rec := sessionRecord{Date: date}
	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if xp, ok := strings.CutPrefix(trimmed, xpLinePrefix); ok {
			// A hand-edited, malformed XP line counts as no XP rather than failing the whole history.
			if n, err := strconv.Atoi(strings.TrimSpace(xp)); err == nil {
				rec.XP += n
			}
			continue
		}
		correct, ok := parseResultRow(trimmed)
		if !ok {
			continue
		}
		rec.Answers++
		if correct {
			rec.Correct++
		}
	}
	return rec
}

// parseResultRow reads one row of the results table. Columns are counted from
// the right because an answer may itself contain "|".
func parseResultRow(line string) (correct, ok bool) {
	if !strings.HasPrefix(line, "|") || line == resultHeader || strings.HasPrefix(line, "|-") {
		return false, false
	}
	cells := strings.Split(line, "|")
	if len(cells) < 6 {
		return false, false
	}
	switch strings.TrimSpace(cells[len(cells)-3]) {
	case "correct":
		return true, true
	case "incorrect":
		return false, true
	}
	return false, false
}

// TodayStats computes stats from session files for today.
func (s *Store) TodayStats() DailyStats {
	today := time.Now().Format(dateFmt)
	stats := DailyStats{Date: today}
	for _, rec := range s.readSessions() {
		if rec.Date == today {
			stats.CardsReviewed += rec.Answers
			stats.Correct += rec.Correct
		}
	}
	return stats
}

// Streak counts consecutive days (ending today or yesterday) with at least one session.
func (s *Store) Streak() int {
	dates := make(map[string]bool)
	for _, rec := range s.readSessions() {
		dates[rec.Date] = true
	}

	day := time.Now()
	// Allow streak to start from today or yesterday
	if !dates[day.Format(dateFmt)] {
		day = day.AddDate(0, 0, -1)
	}
	streak := 0
	for dates[day.Format(dateFmt)] {
		streak++
		day = day.AddDate(0, 0, -1)
	}
	return streak
}

// WeekStats returns DailyStats for each of the last 7 days (index 0 = 6 days ago, 6 = today).
func (s *Store) WeekStats() [7]DailyStats {
	var week [7]DailyStats
	now := time.Now()
	dateIndex := make(map[string]int)
	for i := 0; i < 7; i++ {
		week[i].Date = now.AddDate(0, 0, i-6).Format(dateFmt)
		dateIndex[week[i].Date] = i
	}

	for _, rec := range s.readSessions() {
		if idx, ok := dateIndex[rec.Date]; ok {
			week[idx].CardsReviewed += rec.Answers
			week[idx].Correct += rec.Correct
		}
	}
	return week
}

// TotalXP sums the experience recorded across all session files.
func (s *Store) TotalXP() int {
	total := 0
	for _, rec := range s.readSessions() {
		total += rec.XP
	}
	return total
}
