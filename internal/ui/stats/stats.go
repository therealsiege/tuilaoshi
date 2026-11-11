package stats

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/cj/mandarin-wiki/internal/model"
	"github.com/cj/mandarin-wiki/internal/ui"
	"github.com/cj/mandarin-wiki/internal/wiki"
)

// Model is the statistics view model.
type Model struct {
	store  *wiki.Store
	stats  wiki.ContentStats
	daily  wiki.DailyStats
	week   [7]wiki.DailyStats
	streak int
	pages  []model.Page
	width  int
	height int
}

// New creates a new stats model.
func New(store *wiki.Store) Model {
	return Model{
		store: store,
	}
}

type refreshMsg struct{}

func (m Model) Init() tea.Cmd {
	return func() tea.Msg { return refreshMsg{} }
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg.(type) {
	case refreshMsg:
		m.stats = m.store.Stats()
		m.daily = m.store.TodayStats()
		m.streak = m.store.Streak()
		m.week = m.store.WeekStats()
		m.pages, _ = m.store.ListPages()
		return m, nil
	}
	return m, nil
}

func (m Model) View() string {
	var b strings.Builder

	title := ui.TitleStyle.Render("Study Statistics 学习统计")
	b.WriteString(title)
	b.WriteString("\n\n")

	// Section 1: Study Overview
	overviewBox := m.renderOverview()

	// Section 2: 7-Day Activity
	activityBox := m.renderWeekActivity()

	leftCol := ui.BorderStyle.Width(34).Render(overviewBox)
	rightCol := ui.BorderStyle.Width(42).Render(activityBox)
	b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, leftCol, "  ", rightCol))
	b.WriteString("\n\n")

	// Section 3: SRS Distribution
	srsBox := m.renderSRSDistribution()

	// Section 4: Mastery Levels
	masteryBox := m.renderMasteryLevels()

	leftCol2 := ui.BorderStyle.Width(34).Render(srsBox)
	rightCol2 := ui.BorderStyle.Width(42).Render(masteryBox)
	b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, leftCol2, "  ", rightCol2))
	b.WriteString("\n\n")

	// Section 5: Streak
	b.WriteString(m.renderStreak())
	b.WriteString("\n\n")

	b.WriteString(ui.HelpStyle.Render("  esc back to dashboard • ? help"))

	return b.String()
}

func (m Model) renderOverview() string {
	var lines []string
	lines = append(lines, ui.SubtitleStyle.Render("Study Overview"))
	lines = append(lines, "")
	lines = append(lines, fmt.Sprintf("  Characters: %d", m.stats.CharacterPages))
	lines = append(lines, fmt.Sprintf("  Vocabulary: %d", m.stats.VocabPages))
	lines = append(lines, fmt.Sprintf("  Grammar:    %d", m.stats.GrammarPages))
	lines = append(lines, fmt.Sprintf("  Readings:   %d", m.stats.ReadingPages))
	lines = append(lines, fmt.Sprintf("  Total:      %d pages", m.stats.TotalPages))
	lines = append(lines, "")
	lines = append(lines, fmt.Sprintf("  Reviewed today: %d cards", m.daily.CardsReviewed))
	if m.daily.CardsReviewed > 0 {
		pct := float64(m.daily.Correct) / float64(m.daily.CardsReviewed) * 100
		accText := fmt.Sprintf("  Accuracy:       %.0f%%", pct)
		if pct >= 80 {
			lines = append(lines, ui.SuccessStyle.Render(accText))
		} else if pct >= 50 {
			lines = append(lines, ui.WarningStyle.Render(accText))
		} else {
			lines = append(lines, ui.ErrorStyle.Render(accText))
		}
	} else {
		lines = append(lines, ui.MutedStyle.Render("  Accuracy:       --"))
	}
	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func (m Model) renderWeekActivity() string {
	var lines []string
	lines = append(lines, ui.SubtitleStyle.Render("7-Day Activity"))
	lines = append(lines, "")

	// Find max for scaling
	maxCards := 0
	for _, d := range m.week {
		if d.CardsReviewed > maxCards {
			maxCards = d.CardsReviewed
		}
	}

	bars := [8]string{"▁", "▂", "▃", "▄", "▅", "▆", "▇", "█"}
	maxHeight := 8

	if maxCards == 0 {
		lines = append(lines, ui.MutedStyle.Render("  No activity this week"))
	} else {
		// Build bar chart rows from top to bottom
		for row := maxHeight; row >= 1; row-- {
			line := "  "
			for _, d := range m.week {
				level := 0
				if maxCards > 0 {
					level = d.CardsReviewed * maxHeight / maxCards
				}
				if level >= row {
					line += ui.SuccessStyle.Render(bars[7]) + "  "
				} else if level == row-1 && d.CardsReviewed > 0 {
					frac := d.CardsReviewed * maxHeight % maxCards
					barIdx := frac * 7 / maxCards
					if barIdx > 7 {
						barIdx = 7
					}
					line += ui.SuccessStyle.Render(bars[barIdx]) + "  "
				} else {
					line += "   "
				}
			}
			lines = append(lines, line)
		}

		// Day labels
		dayLine := "  "
		now := time.Now()
		for i := 0; i < 7; i++ {
			day := now.AddDate(0, 0, i-6)
			dayLine += day.Format("Mo") + " "
		}
		lines = append(lines, ui.MutedStyle.Render(dayLine))

		// Counts
		countLine := "  "
		for _, d := range m.week {
			countLine += fmt.Sprintf("%-3d", d.CardsReviewed)
		}
		lines = append(lines, ui.MutedStyle.Render(countLine))
	}

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func (m Model) renderSRSDistribution() string {
	var lines []string
	lines = append(lines, ui.SubtitleStyle.Render("SRS Distribution"))
	lines = append(lines, "")

	// Bucket cards by review interval
	buckets := map[string]int{
		"1d":    0,
		"2-3d":  0,
		"4-7d":  0,
		"8-14d": 0,
		"15-30d": 0,
		"30d+":  0,
	}
	bucketOrder := []string{"1d", "2-3d", "4-7d", "8-14d", "15-30d", "30d+"}

	for _, p := range m.pages {
		switch p.Type {
		case "character", "vocabulary", "grammar":
		default:
			continue
		}
		iv := p.ReviewInterval
		if iv <= 0 {
			continue
		}
		switch {
		case iv <= 1:
			buckets["1d"]++
		case iv <= 3:
			buckets["2-3d"]++
		case iv <= 7:
			buckets["4-7d"]++
		case iv <= 14:
			buckets["8-14d"]++
		case iv <= 30:
			buckets["15-30d"]++
		default:
			buckets["30d+"]++
		}
	}

	for _, name := range bucketOrder {
		count := buckets[name]
		lines = append(lines, fmt.Sprintf("  %-7s %s %d", name, miniBar(count, 15), count))
	}

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func (m Model) renderMasteryLevels() string {
	var lines []string
	lines = append(lines, ui.SubtitleStyle.Render("Mastery Levels"))
	lines = append(lines, "")

	levels := map[string]int{
		"New":       0,
		"Learning":  0,
		"Familiar":  0,
		"Confident": 0,
		"Mastered":  0,
	}
	levelOrder := []string{"New", "Learning", "Familiar", "Confident", "Mastered"}

	for _, p := range m.pages {
		switch p.Type {
		case "character", "vocabulary", "grammar":
		default:
			continue
		}
		iv := p.ReviewInterval
		switch {
		case iv <= 0:
			levels["New"]++
		case iv <= 3:
			levels["Learning"]++
		case iv <= 14:
			levels["Familiar"]++
		case iv <= 30:
			levels["Confident"]++
		default:
			levels["Mastered"]++
		}
	}

	styles := map[string]lipgloss.Style{
		"New":       ui.MutedStyle,
		"Learning":  ui.ErrorStyle,
		"Familiar":  ui.WarningStyle,
		"Confident": ui.SuccessStyle,
		"Mastered":  lipgloss.NewStyle().Foreground(lipgloss.Color("#61AFEF")),
	}

	total := 0
	for _, c := range levels {
		total += c
	}

	for _, name := range levelOrder {
		count := levels[name]
		pct := ""
		if total > 0 {
			pct = fmt.Sprintf(" (%d%%)", count*100/total)
		}
		label := styles[name].Render(fmt.Sprintf("  %-10s", name))
		lines = append(lines, fmt.Sprintf("%s %s %d%s", label, miniBar(count, 15), count, ui.MutedStyle.Render(pct)))
	}

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func (m Model) renderStreak() string {
	var parts []string

	if m.streak == 0 {
		parts = append(parts, ui.MutedStyle.Render("  Streak: No active streak. Start studying today!"))
	} else {
		streakText := fmt.Sprintf("  Streak: %d day", m.streak)
		if m.streak != 1 {
			streakText += "s"
		}
		if m.streak >= 7 {
			parts = append(parts, ui.SuccessStyle.Render(streakText+" 🔥"))
		} else {
			parts = append(parts, streakText)
		}
	}

	// Daily goal progress
	goal := 10 // target cards per day
	reviewed := m.daily.CardsReviewed
	if reviewed >= goal {
		parts = append(parts, ui.SuccessStyle.Render(fmt.Sprintf("  Daily goal: %d/%d cards", reviewed, goal)))
	} else {
		parts = append(parts, fmt.Sprintf("  Daily goal: %d/%d cards  %s",
			reviewed, goal,
			ui.ProgressBar(reviewed, goal, 15)))
	}

	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}

// miniBar renders a small horizontal bar for distribution displays.
func miniBar(count, maxWidth int) string {
	if count == 0 {
		return ui.MutedStyle.Render(strings.Repeat("░", 1))
	}
	// Scale: use log-ish scaling for visual clarity
	width := count
	if width > maxWidth {
		width = maxWidth
	}
	return ui.SuccessStyle.Render(strings.Repeat("█", width)) +
		ui.MutedStyle.Render(strings.Repeat("░", maxWidth-width))
}

// SetSize updates terminal dimensions.
func (m Model) SetSize(w, h int) Model {
	m.width = w
	m.height = h
	return m
}

// Refresh reloads data from the wiki store.
func (m Model) Refresh() Model {
	m.stats = m.store.Stats()
	m.daily = m.store.TodayStats()
	m.streak = m.store.Streak()
	m.week = m.store.WeekStats()
	m.pages, _ = m.store.ListPages()
	return m
}
