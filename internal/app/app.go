package app

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/therealsiege/tuilaoshi/internal/audio"
	"github.com/therealsiege/tuilaoshi/internal/claude"
	"github.com/therealsiege/tuilaoshi/internal/journey"
	"github.com/therealsiege/tuilaoshi/internal/ui"
	"github.com/therealsiege/tuilaoshi/internal/ui/assessment"
	"github.com/therealsiege/tuilaoshi/internal/ui/browser"
	"github.com/therealsiege/tuilaoshi/internal/ui/celebrate"
	"github.com/therealsiege/tuilaoshi/internal/ui/chat"
	"github.com/therealsiege/tuilaoshi/internal/ui/collection"
	"github.com/therealsiege/tuilaoshi/internal/ui/dashboard"
	"github.com/therealsiege/tuilaoshi/internal/ui/drill"
	journeyview "github.com/therealsiege/tuilaoshi/internal/ui/journey"
	"github.com/therealsiege/tuilaoshi/internal/ui/party"
	"github.com/therealsiege/tuilaoshi/internal/ui/radicals"
	"github.com/therealsiege/tuilaoshi/internal/ui/reader"
	"github.com/therealsiege/tuilaoshi/internal/ui/review"
	rpgview "github.com/therealsiege/tuilaoshi/internal/ui/rpg"
	"github.com/therealsiege/tuilaoshi/internal/ui/shrine"
	"github.com/therealsiege/tuilaoshi/internal/ui/stats"
	"github.com/therealsiege/tuilaoshi/internal/ui/world"
	"github.com/therealsiege/tuilaoshi/internal/wiki"
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
	shrine    shrine.Model
	cards     collection.Model
	party     party.Model
	world     world.Model
	battle    rpgview.Model
	reader    reader.Model
	assess    assessment.Model

	// Celebration state
	celebration    *celebrate.Celebration // active celebration to display
	celebrationTTL int                    // remaining key presses before auto-dismiss
	prevRankName   string                 // to detect rank changes
	prevBossCount  int                    // to detect boss defeats
	prevLevel      int                    // to detect level-ups
	prevOpenStage  int                    // the last stage seen open, to say a level-up opened one

	// Voice picker state
	voiceCursor   int
	voiceCategory int // 0 = Chinese, 1 = English
}

// New creates a new app model. It fails when the wiki cannot be read.
func New(wikiDir string) (Model, error) {
	store := wiki.NewStore(wikiDir)
	claudeClient := claude.New()
	ttsClient := audio.New()

	pages, err := store.ListPages()
	if err != nil {
		return Model{}, fmt.Errorf("read wiki: %w", err)
	}
	// Build pinyin/character dictionary from wiki pages
	dict := buildDict(pages)

	// Snapshot initial journey state to detect future milestones.
	initialState, err := journey.Calculate(store)
	if err != nil {
		return Model{}, fmt.Errorf("read wiki: %w", err)
	}
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
		shrine:        shrine.New(store),
		cards:         collection.New(store),
		party:         party.New(store, claudeClient, dict),
		world:         world.New(store),
		battle:        rpgview.New(store, ttsClient, dict),
		reader:        reader.New(store),
		assess:        assessment.New(store),
		prevRankName:  initialState.Rank.Name,
		prevBossCount: initialBossCount,
		prevLevel:     initialState.Level.N,
		prevOpenStage: openStage(initialState),
	}, nil
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
		m.shrine = m.shrine.SetSize(msg.Width, msg.Height)
		m.cards = m.cards.SetSize(msg.Width, msg.Height)
		m.party = m.party.SetSize(msg.Width, msg.Height)
		m.world = m.world.SetSize(msg.Width, msg.Height)
		m.battle = m.battle.SetSize(msg.Width, msg.Height)
		m.reader = m.reader.SetSize(msg.Width, msg.Height)
		m.assess = m.assess.SetSize(msg.Width, msg.Height)
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
		case ui.ViewSearch:
			m.currentView = ui.ViewBrowser
			m.browser = m.browser.Refresh().StartSearch()
			return m, nil
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
		case ui.ViewShrine:
			m.shrine = m.shrine.Refresh()
			return m, m.shrine.Init()
		case ui.ViewCollection:
			m.cards = m.cards.Refresh()
			return m, m.cards.Init()
		case ui.ViewParty:
			m.party = m.party.Refresh()
			return m, m.party.Init()
		case ui.ViewWorld:
			m.world = m.world.Refresh()
			return m, m.world.Init()
		case ui.ViewBattle:
			m.battle = m.battle.Refresh()
			return m, m.battle.Init()
		case ui.ViewReading:
			m.reader = m.reader.Refresh()
			return m, m.reader.Init()
		case ui.ViewAssess:
			m.assess = m.assess.Refresh()
			return m, m.assess.Init()
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
		typing := m.currentView == ui.ViewChat || (m.currentView == ui.ViewParty && m.party.Busy())
		if key.Matches(msg, m.keys.Help) && !typing {
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
		// Global back to dashboard (unless input is composing in Chinese mode,
		// or esc closes an open reading)
		readingOpen := (m.currentView == ui.ViewReading && m.reader.Open()) ||
			(m.currentView == ui.ViewBrowser && m.browser.Searching()) ||
			(m.currentView == ui.ViewParty && m.party.Busy()) ||
			(m.currentView == ui.ViewWorld && m.world.Busy())
		if key.Matches(msg, m.keys.Back) && m.currentView != ui.ViewDashboard && !readingOpen {
			if !m.isInputComposing() {
				if m.currentView == ui.ViewBattle {
					m.battle = m.battle.Leave()
				}
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
	case ui.ViewShrine:
		m.shrine, cmd = m.shrine.Update(msg)
	case ui.ViewCollection:
		m.cards, cmd = m.cards.Update(msg)
	case ui.ViewParty:
		m.party, cmd = m.party.Update(msg)
	case ui.ViewWorld:
		m.world, cmd = m.world.Update(msg)
	case ui.ViewBattle:
		m.battle, cmd = m.battle.Update(msg)
	case ui.ViewReading:
		m.reader, cmd = m.reader.Update(msg)
	case ui.ViewAssess:
		m.assess, cmd = m.assess.Update(msg)
	}
	return m, cmd
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
	case ui.ViewShrine:
		content = m.shrine.View()
	case ui.ViewCollection:
		content = m.cards.View()
	case ui.ViewParty:
		content = m.party.View()
	case ui.ViewWorld:
		content = m.world.View()
	case ui.ViewBattle:
		content = m.battle.View()
	case ui.ViewReading:
		content = m.reader.View()
	case ui.ViewAssess:
		content = m.assess.View()
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
	case ui.ViewBattle:
		return m.battle.InputComposing()
	}
	return false
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
