package dashboard

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/cj/mandarin-wiki/internal/audio"
	"github.com/cj/mandarin-wiki/internal/ui"
	"github.com/cj/mandarin-wiki/internal/wiki"
)

// Model is the dashboard view model.
type Model struct {
	store  *wiki.Store
	tts    *audio.Client
	keys   ui.DashboardKeys
	gkeys  ui.GlobalKeys
	stats  wiki.ContentStats
	daily  wiki.DailyStats
	streak int
	due    int
	width  int
	height int
}

// New creates a new dashboard model.
func New(store *wiki.Store, tts *audio.Client) Model {
	return Model{
		store: store,
		tts:   tts,
		keys:  ui.DefaultDashboardKeys(),
		gkeys: ui.DefaultGlobalKeys(),
	}
}

func (m Model) Init() tea.Cmd {
	return func() tea.Msg {
		return refreshMsg{}
	}
}

type refreshMsg struct{}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case refreshMsg:
		m.stats = m.store.Stats()
		due, _ := wiki.DueItems(m.store, time.Now())
		m.due = len(due)
		m.daily = m.store.TodayStats()
		m.streak = m.store.Streak()
		return m, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Drill):
			return m, ui.SwitchTo(ui.ViewDrill)
		case key.Matches(msg, m.keys.Browse):
			return m, ui.SwitchTo(ui.ViewBrowser)
		case key.Matches(msg, m.keys.Chat):
			return m, ui.SwitchTo(ui.ViewChat)
		case key.Matches(msg, m.keys.Review):
			return m, ui.SwitchTo(ui.ViewReview)
		case key.Matches(msg, m.keys.Voice):
			return m, ui.SwitchTo(ui.ViewVoicePicker)
		}
	}
	return m, nil
}

func (m Model) View() string {
	var b strings.Builder

	// Title
	title := ui.TitleStyle.Render("中文 Mandarin Tutor")
	b.WriteString(title)
	b.WriteString("\n\n")

	// Two-column layout: stats + today's progress
	statsBox := lipgloss.JoinVertical(lipgloss.Left,
		ui.SubtitleStyle.Render("Wiki Content"),
		"",
		fmt.Sprintf("  Characters: %d", m.stats.CharacterPages),
		fmt.Sprintf("  Vocabulary: %d", m.stats.VocabPages),
		fmt.Sprintf("  Grammar:    %d", m.stats.GrammarPages),
		fmt.Sprintf("  Readings:   %d", m.stats.ReadingPages),
		fmt.Sprintf("  Total:      %d pages", m.stats.TotalPages),
	)

	todayBox := lipgloss.JoinVertical(lipgloss.Left,
		ui.SubtitleStyle.Render("Today's Progress"),
		"",
		fmt.Sprintf("  Reviewed: %d cards", m.daily.CardsReviewed),
		m.accuracyLine(),
		m.streakLine(),
	)

	leftCol := ui.BorderStyle.Width(30).Render(statsBox)
	rightCol := ui.BorderStyle.Width(30).Render(todayBox)

	b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, leftCol, "  ", rightCol))
	b.WriteString("\n\n")

	// Review status with progress bar
	if m.due > 0 {
		dueText := ui.WarningStyle.Render(fmt.Sprintf("  %d items due for review", m.due))
		b.WriteString(dueText)
		b.WriteString("\n")
		reviewed := m.daily.CardsReviewed
		total := reviewed + m.due
		b.WriteString("  " + ui.ProgressBar(reviewed, total, 25))
		b.WriteString(ui.MutedStyle.Render(fmt.Sprintf(" %d/%d", reviewed, total)))
	} else {
		dueText := ui.SuccessStyle.Render("  All reviews complete!")
		b.WriteString(dueText)
	}
	b.WriteString("\n\n")

	// Navigation
	nav := lipgloss.JoinVertical(lipgloss.Left,
		ui.SubtitleStyle.Render("Actions"),
		"",
		fmt.Sprintf("  %s  Start a flashcard drill", ui.ChineseStyle.Render("[d]")),
		fmt.Sprintf("  %s  Browse wiki pages", ui.ChineseStyle.Render("[b]")),
		fmt.Sprintf("  %s  Chat with tutor", ui.ChineseStyle.Render("[c]")),
		fmt.Sprintf("  %s  Review due items", ui.ChineseStyle.Render("[r]")),
		fmt.Sprintf("  %s  Change voice", ui.ChineseStyle.Render("[v]")),
		fmt.Sprintf("  %s  Quit", ui.ChineseStyle.Render("[q]")),
	)
	b.WriteString(nav)
	b.WriteString("\n\n")

	// Footer
	footer := "  mandarin-wiki • esc back • ? help"
	if m.tts != nil && m.tts.Available() {
		footer += " • " + ui.SuccessStyle.Render("audio on") +
			" (" + m.tts.ChineseVoice().Name + "/" + m.tts.EnglishVoice().Name + ")"
	}
	b.WriteString(ui.MutedStyle.Render(footer))

	return b.String()
}

func (m Model) accuracyLine() string {
	if m.daily.CardsReviewed == 0 {
		return ui.MutedStyle.Render("  Accuracy: --")
	}
	pct := float64(m.daily.Correct) / float64(m.daily.CardsReviewed) * 100
	text := fmt.Sprintf("  Accuracy: %.0f%%", pct)
	if pct >= 80 {
		return ui.SuccessStyle.Render(text)
	} else if pct >= 50 {
		return ui.WarningStyle.Render(text)
	}
	return ui.ErrorStyle.Render(text)
}

func (m Model) streakLine() string {
	if m.streak == 0 {
		return ui.MutedStyle.Render("  Streak:   --")
	}
	text := fmt.Sprintf("  Streak:   %d days", m.streak)
	if m.streak >= 7 {
		return ui.SuccessStyle.Render(text + " 🔥")
	}
	return text
}

// SetSize updates the terminal dimensions.
func (m Model) SetSize(w, h int) Model {
	m.width = w
	m.height = h
	return m
}

// Refresh reloads data from the wiki store.
func (m Model) Refresh() Model {
	m.stats = m.store.Stats()
	due, _ := wiki.DueItems(m.store, time.Now())
	m.due = len(due)
	m.daily = m.store.TodayStats()
	m.streak = m.store.Streak()
	return m
}
