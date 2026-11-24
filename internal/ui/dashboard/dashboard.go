package dashboard

import (
	"errors"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/therealsiege/tuilaoshi/internal/audio"
	"github.com/therealsiege/tuilaoshi/internal/journey"
	"github.com/therealsiege/tuilaoshi/internal/ui"
	"github.com/therealsiege/tuilaoshi/internal/wiki"
)

// Model is the dashboard view model.
type Model struct {
	store        *wiki.Store
	tts          *audio.Client
	keys         ui.DashboardKeys
	gkeys        ui.GlobalKeys
	daily        wiki.DailyStats
	journeyState journey.State
	streak       int
	wallet       wiki.Wallet
	bounties     []wiki.BountyStatus
	shields      int              // streak shields held
	now          func() time.Time // the clock, fixed in tests
	due          int
	xp           int   // lifetime battle XP
	loadErr      error // why the wiki could not be read, shown under the header
	width        int
	height       int
}

// New creates a new dashboard model.
func New(store *wiki.Store, tts *audio.Client) Model {
	return Model{
		store: store,
		tts:   tts,
		keys:  ui.DefaultDashboardKeys(),
		gkeys: ui.DefaultGlobalKeys(),
		now:   time.Now,
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
		return m.Refresh(), nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.gkeys.Back):
			return m, ui.SwitchTo(ui.ViewMap) // back on the road
		case msg.String() == "T":
			return m, ui.SwitchTo(ui.ViewTitle) // switch journeys
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
		case key.Matches(msg, m.keys.Stats):
			return m, ui.SwitchTo(ui.ViewStats)
		case key.Matches(msg, m.keys.Journey):
			return m, ui.SwitchTo(ui.ViewJourney)
		case key.Matches(msg, m.keys.Radicals):
			return m, ui.SwitchTo(ui.ViewRadicals)
		case key.Matches(msg, m.keys.Battle):
			return m, ui.SwitchTo(ui.ViewBattle)
		case key.Matches(msg, m.keys.Shrine):
			return m, ui.SwitchTo(ui.ViewShrine)
		case key.Matches(msg, m.keys.Cards):
			return m, ui.SwitchTo(ui.ViewCollection)
		case key.Matches(msg, m.keys.Party):
			return m, ui.SwitchTo(ui.ViewParty)
		case key.Matches(msg, m.keys.Episodes):
			return m, ui.SwitchTo(ui.ViewTheatre)
		case key.Matches(msg, m.keys.Forge):
			return m, ui.SwitchTo(ui.ViewForge)
		case key.Matches(msg, m.keys.Festival):
			return m, ui.SwitchTo(ui.ViewFestival)
		case key.Matches(msg, m.keys.Settings):
			return m, ui.SwitchTo(ui.ViewSettings)
		case key.Matches(msg, m.keys.Reading):
			return m, ui.SwitchTo(ui.ViewReading)
		case key.Matches(msg, m.keys.Assess):
			return m, ui.SwitchTo(ui.ViewAssess)
		case key.Matches(msg, m.keys.Search):
			return m, ui.SwitchTo(ui.ViewSearch)
		}
	}
	return m, nil
}

// SetSize updates the terminal dimensions.
func (m Model) SetSize(w, h int) Model {
	m.width = w
	m.height = h
	return m
}

// Refresh reloads data from the wiki store.
func (m Model) Refresh() Model {
	var dueErr, journeyErr, walletErr, bountyErr error
	due, dueErr := wiki.DueItems(m.store, time.Now())
	m.due = len(due)
	m.daily = m.store.TodayStats()
	m.streak = m.store.Streak()
	m.xp = m.store.TotalXP()
	m.wallet, walletErr = m.store.Wallet()
	m.bounties, m.shields, bountyErr = m.store.Bounties(m.now())
	m.journeyState, journeyErr = journey.Calculate(m.store)
	m.loadErr = errors.Join(dueErr, journeyErr, walletErr, bountyErr)
	return m
}
