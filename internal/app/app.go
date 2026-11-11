package app

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/cj/mandarin-wiki/internal/audio"
	"github.com/cj/mandarin-wiki/internal/claude"
	"github.com/cj/mandarin-wiki/internal/journey"
	"github.com/cj/mandarin-wiki/internal/model"
	"github.com/cj/mandarin-wiki/internal/ui"
	"github.com/cj/mandarin-wiki/internal/ui/browser"
	"github.com/cj/mandarin-wiki/internal/ui/celebrate"
	"github.com/cj/mandarin-wiki/internal/ui/chat"
	"github.com/cj/mandarin-wiki/internal/ui/dashboard"
	"github.com/cj/mandarin-wiki/internal/ui/drill"
	"github.com/cj/mandarin-wiki/internal/ui/input"
	journeyview "github.com/cj/mandarin-wiki/internal/ui/journey"
	"github.com/cj/mandarin-wiki/internal/ui/radicals"
	"github.com/cj/mandarin-wiki/internal/ui/review"
	"github.com/cj/mandarin-wiki/internal/ui/stats"
	"github.com/cj/mandarin-wiki/internal/wiki"
)

// Model is the root Bubble Tea model.
type Model struct {
	store  *wiki.Store
	claude *claude.Client
	tts    *audio.Client
	keys   ui.GlobalKeys

	currentView ui.View
	showHelp    bool
	width       int
	height      int

	dashboard dashboard.Model
	drill     drill.Model
	browser   browser.Model
	chat      chat.Model
	review    review.Model
	stats     stats.Model
	journey   journeyview.Model
	radicals  radicals.Model

	// Celebration state
	celebration    *celebrate.Celebration // active celebration to display
	celebrationTTL int                    // remaining key presses before auto-dismiss
	prevRankName   string                 // to detect rank changes
	prevBossCount  int                    // to detect boss defeats

	// Voice picker state
	voiceCursor   int
	voiceCategory int // 0 = Chinese, 1 = English
}

// New creates a new app model.
func New(wikiDir string) Model {
	store := wiki.NewStore(wikiDir)
	claudeClient := claude.New()
	ttsClient := audio.New()

	// Build pinyin/character dictionary from wiki pages
	dict := buildDict(store)

	// Snapshot initial journey state to detect future milestones.
	initialState := journey.Calculate(store)
	initialBossCount := 0
	for _, b := range initialState.Bosses {
		if b.Defeated {
			initialBossCount++
		}
	}

	return Model{
		store:         store,
		claude:        claudeClient,
		tts:           ttsClient,
		keys:          ui.DefaultGlobalKeys(),
		currentView:   ui.ViewDashboard,
		dashboard:     dashboard.New(store, ttsClient),
		drill:         drill.New(store, claudeClient, ttsClient, dict),
		browser:       browser.New(store),
		chat:          chat.New(claudeClient, ttsClient, dict),
		review:        review.New(store, claudeClient, ttsClient, dict),
		stats:         stats.New(store),
		journey:       journeyview.New(store),
		radicals:      radicals.New(),
		prevRankName:  initialState.Rank.Name,
		prevBossCount: initialBossCount,
	}
}

// buildDict constructs a pinyin-to-character dictionary from all wiki pages.
func buildDict(store *wiki.Store) *input.Dict {
	pages, err := store.ListPages()
	if err != nil {
		return input.NewDict()
	}
	var drillable []model.Page
	for _, p := range pages {
		switch p.Type {
		case "character", "vocabulary", "grammar":
			drillable = append(drillable, p)
		}
	}
	cards := wiki.CardsFromPages(drillable)
	return input.NewFromCards(cards)
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.dashboard.Init(),
		tea.EnterAltScreen,
	)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.dashboard = m.dashboard.SetSize(msg.Width, msg.Height)
		m.drill = m.drill.SetSize(msg.Width, msg.Height)
		m.browser = m.browser.SetSize(msg.Width, msg.Height)
		m.chat = m.chat.SetSize(msg.Width, msg.Height)
		m.review = m.review.SetSize(msg.Width, msg.Height)
		m.stats = m.stats.SetSize(msg.Width, msg.Height)
		m.journey = m.journey.SetSize(msg.Width, msg.Height)
		m.radicals = m.radicals.SetSize(msg.Width, msg.Height)
		return m, nil

	case ui.SwitchViewMsg:
		m.currentView = msg.Target
		m.showHelp = false
		switch msg.Target {
		case ui.ViewDashboard:
			m.dashboard = m.dashboard.Refresh()
			m.checkCelebrations()
			return m, m.dashboard.Init()
		case ui.ViewDrill:
			m.drill = m.drill.Refresh()
			return m, m.drill.Init()
		case ui.ViewBrowser:
			m.browser = m.browser.Refresh()
			return m, m.browser.Init()
		case ui.ViewChat:
			return m, m.chat.Init()
		case ui.ViewReview:
			m.review = m.review.Refresh()
			return m, m.review.Init()
		case ui.ViewVoicePicker:
			m.voiceCursor = 0
			m.voiceCategory = 0
			return m, nil
		case ui.ViewStats:
			m.stats = m.stats.Refresh()
			return m, m.stats.Init()
		case ui.ViewJourney:
			m.journey = m.journey.Refresh()
			return m, m.journey.Init()
		case ui.ViewRadicals:
			m.radicals = m.radicals.Refresh()
			return m, m.radicals.Init()
		}
		return m, nil

	case tea.KeyMsg:
		// Dismiss celebration on any keypress
		if m.celebration != nil {
			m.celebrationTTL--
			if m.celebrationTTL <= 0 || msg.String() == "esc" || msg.String() == "enter" {
				m.celebration = nil
				m.celebrationTTL = 0
			}
			return m, nil
		}

		// Help toggle from anywhere
		if key.Matches(msg, m.keys.Help) && m.currentView != ui.ViewChat {
			m.showHelp = !m.showHelp
			return m, nil
		}

		// Dismiss help on any key if showing
		if m.showHelp {
			m.showHelp = false
			return m, nil
		}

		// Voice picker input handling
		if m.currentView == ui.ViewVoicePicker {
			return m.updateVoicePicker(msg)
		}

		// Global quit from dashboard
		if key.Matches(msg, m.keys.Quit) && m.currentView == ui.ViewDashboard {
			return m, tea.Quit
		}
		// Global back to dashboard (unless input is composing in Chinese mode)
		if key.Matches(msg, m.keys.Back) && m.currentView != ui.ViewDashboard {
			if !m.isInputComposing() {
				m.currentView = ui.ViewDashboard
				m.dashboard = m.dashboard.Refresh()
				return m, m.dashboard.Init()
			}
		}
	}

	// Delegate to current view
	var cmd tea.Cmd
	switch m.currentView {
	case ui.ViewDashboard:
		m.dashboard, cmd = m.dashboard.Update(msg)
	case ui.ViewDrill:
		m.drill, cmd = m.drill.Update(msg)
	case ui.ViewBrowser:
		m.browser, cmd = m.browser.Update(msg)
	case ui.ViewChat:
		m.chat, cmd = m.chat.Update(msg)
	case ui.ViewReview:
		m.review, cmd = m.review.Update(msg)
	case ui.ViewStats:
		m.stats, cmd = m.stats.Update(msg)
	case ui.ViewJourney:
		m.journey, cmd = m.journey.Update(msg)
	case ui.ViewRadicals:
		m.radicals, cmd = m.radicals.Update(msg)
	}
	return m, cmd
}

func (m *Model) updateVoicePicker(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	zhVoices := audio.ChineseVoices()
	enVoices := audio.EnglishVoices()

	var maxItems int
	if m.voiceCategory == 0 {
		maxItems = len(zhVoices)
	} else {
		maxItems = len(enVoices)
	}

	switch {
	case key.Matches(msg, m.keys.Up):
		if m.voiceCursor > 0 {
			m.voiceCursor--
		}
	case key.Matches(msg, m.keys.Down):
		if m.voiceCursor < maxItems-1 {
			m.voiceCursor++
		}
	case key.Matches(msg, m.keys.Tab):
		if m.voiceCategory == 0 {
			m.voiceCategory = 1
		} else {
			m.voiceCategory = 0
		}
		m.voiceCursor = 0
	case key.Matches(msg, m.keys.Enter):
		if m.voiceCategory == 0 && m.voiceCursor < len(zhVoices) {
			m.tts.SetChineseVoice(zhVoices[m.voiceCursor])
		} else if m.voiceCategory == 1 && m.voiceCursor < len(enVoices) {
			m.tts.SetEnglishVoice(enVoices[m.voiceCursor])
		}
		m.currentView = ui.ViewDashboard
		m.dashboard = m.dashboard.Refresh()
		return m, m.dashboard.Init()
	case key.Matches(msg, m.keys.Back):
		m.currentView = ui.ViewDashboard
		m.dashboard = m.dashboard.Refresh()
		return m, m.dashboard.Init()
	}
	return m, nil
}

func (m Model) View() string {
	var content string

	switch m.currentView {
	case ui.ViewDrill:
		content = m.drill.View()
	case ui.ViewBrowser:
		content = m.browser.View()
	case ui.ViewChat:
		content = m.chat.View()
	case ui.ViewReview:
		content = m.review.View()
	case ui.ViewVoicePicker:
		content = m.voicePickerView()
	case ui.ViewStats:
		content = m.stats.View()
	case ui.ViewJourney:
		content = m.journey.View()
	case ui.ViewRadicals:
		content = m.radicals.View()
	default:
		content = m.dashboard.View()
	}

	if m.celebration != nil {
		content = m.renderCelebrationOverlay(content)
	}

	if m.showHelp {
		content = m.renderHelpOverlay(content)
	}

	return content
}

// isInputComposing checks if the current view's input is in Chinese composition mode.
func (m Model) isInputComposing() bool {
	switch m.currentView {
	case ui.ViewDrill:
		return m.drill.InputComposing()
	case ui.ViewChat:
		return m.chat.InputComposing()
	case ui.ViewReview:
		return m.review.InputComposing()
	}
	return false
}

// checkCelebrations compares current journey state against previous snapshots
// and triggers a celebration when a rank-up or boss defeat is detected.
func (m *Model) checkCelebrations() {
	state := journey.Calculate(m.store)

	// Count currently defeated bosses.
	bossCount := 0
	for _, b := range state.Bosses {
		if b.Defeated {
			bossCount++
		}
	}

	// Boss defeat takes priority (more dramatic).
	if bossCount > m.prevBossCount {
		// Find the most recently defeated boss (highest-index defeated one).
		for i := len(state.Bosses) - 1; i >= 0; i-- {
			if state.Bosses[i].Defeated {
				c := celebrate.BossDefeatCelebration(state.Bosses[i].Name, state.Bosses[i].English)
				m.celebration = &c
				m.celebrationTTL = 3
				break
			}
		}
		m.prevBossCount = bossCount
		m.prevRankName = state.Rank.Name
		return
	}

	// Rank change.
	if state.Rank.Name != m.prevRankName {
		c := celebrate.RankUpCelebration(state.Rank.Name, state.Rank.English)
		m.celebration = &c
		m.celebrationTTL = 3
		m.prevRankName = state.Rank.Name
		m.prevBossCount = bossCount
		return
	}

	// Keep snapshots current even when no celebration fires.
	m.prevRankName = state.Rank.Name
	m.prevBossCount = bossCount
}

func (m Model) voicePickerView() string {
	var b strings.Builder

	title := ui.TitleStyle.Render("Voice Picker 语音选择")
	b.WriteString(title)
	b.WriteString("\n\n")

	zhVoices := audio.ChineseVoices()
	enVoices := audio.EnglishVoices()

	// Category tabs
	if m.voiceCategory == 0 {
		b.WriteString(ui.ActiveTabStyle.Render("Chinese 中文"))
	} else {
		b.WriteString(ui.TabStyle.Render("Chinese 中文"))
	}
	b.WriteString("  ")
	if m.voiceCategory == 1 {
		b.WriteString(ui.ActiveTabStyle.Render("English"))
	} else {
		b.WriteString(ui.TabStyle.Render("English"))
	}
	b.WriteString("\n\n")

	if m.voiceCategory == 0 {
		current := m.tts.ChineseVoice()
		for i, v := range zhVoices {
			label := fmt.Sprintf("  %s", v.Name)
			if v.ID == current.ID {
				label += " " + ui.SuccessStyle.Render("(current)")
			}
			if i == m.voiceCursor {
				b.WriteString(ui.SelectedStyle.Render("▸ " + label))
			} else {
				b.WriteString(ui.UnselectedStyle.Render("  " + label))
			}
			b.WriteString("\n")
		}
	} else {
		current := m.tts.EnglishVoice()
		for i, v := range enVoices {
			label := fmt.Sprintf("  %s", v.Name)
			if v.ID == current.ID {
				label += " " + ui.SuccessStyle.Render("(current)")
			}
			if i == m.voiceCursor {
				b.WriteString(ui.SelectedStyle.Render("▸ " + label))
			} else {
				b.WriteString(ui.UnselectedStyle.Render("  " + label))
			}
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(ui.HelpStyle.Render("  ↑↓ select · tab switch category · enter confirm · esc back"))

	return b.String()
}

func (m Model) renderCelebrationOverlay(base string) string {
	celebContent := m.celebration.Render()
	hint := ui.HelpStyle.Render("  Press any key to continue")
	celebContent += hint

	overlay := lipgloss.NewStyle().
		Padding(1, 2).
		Render(celebContent)

	overlayLines := strings.Split(overlay, "\n")
	overlayHeight := len(overlayLines)
	// Estimate overlay width from the longest line.
	overlayWidth := 0
	for _, line := range overlayLines {
		if w := lipgloss.Width(line); w > overlayWidth {
			overlayWidth = w
		}
	}

	baseLines := strings.Split(base, "\n")
	startY := (m.height - overlayHeight) / 2
	if startY < 0 {
		startY = 0
	}
	startX := (m.width - overlayWidth) / 2
	if startX < 0 {
		startX = 0
	}

	for i, oLine := range overlayLines {
		targetLine := startY + i
		if targetLine < len(baseLines) {
			padded := strings.Repeat(" ", startX) + oLine
			baseLines[targetLine] = padded
		}
	}

	return strings.Join(baseLines, "\n")
}

func (m Model) renderHelpOverlay(base string) string {
	helpContent := ui.HelpText(m.currentView)

	overlay := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ui.ColorAccent).
		Padding(1, 3).
		Width(40).
		Render(helpContent)

	// Center the overlay
	overlayWidth := 46 // 40 + padding + borders
	overlayHeight := strings.Count(overlay, "\n") + 1

	baseLines := strings.Split(base, "\n")
	startY := (m.height - overlayHeight) / 2
	if startY < 0 {
		startY = 0
	}
	startX := (m.width - overlayWidth) / 2
	if startX < 0 {
		startX = 0
	}

	// Simple overlay: replace lines in the base view
	overlayLines := strings.Split(overlay, "\n")
	for i, oLine := range overlayLines {
		targetLine := startY + i
		if targetLine < len(baseLines) {
			padded := strings.Repeat(" ", startX) + oLine
			baseLines[targetLine] = padded
		}
	}

	return strings.Join(baseLines, "\n")
}
