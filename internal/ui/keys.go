package ui

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// GlobalKeys defines keybindings available in all views.
type GlobalKeys struct {
	Quit   key.Binding
	Help   key.Binding
	Back   key.Binding
	Enter  key.Binding
	Tab    key.Binding
	Up     key.Binding
	Down   key.Binding
	Left   key.Binding
	Right  key.Binding
	Search key.Binding
}

// DefaultGlobalKeys returns the default global keybindings.
func DefaultGlobalKeys() GlobalKeys {
	return GlobalKeys{
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "help"),
		),
		Back: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "back"),
		),
		Enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "select"),
		),
		Tab: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "next"),
		),
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
		),
		Left: key.NewBinding(
			key.WithKeys("left", "h"),
			key.WithHelp("←/h", "left"),
		),
		Right: key.NewBinding(
			key.WithKeys("right", "l"),
			key.WithHelp("→/l", "right"),
		),
		Search: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "search"),
		),
	}
}

// View identifies which screen is active (shared across packages).
type View int

const (
	ViewDashboard View = iota
	ViewDrill
	ViewBrowser
	ViewChat
	ViewReview
	ViewVoicePicker
	ViewStats
	ViewJourney
	ViewRadicals
	ViewBattle
	ViewReading
	ViewAssess
	ViewSearch     // the browser, opened with its search box active
	ViewShrine     // the Earth God's shrine: gear shop and character sheet
	ViewCollection // enemy and character cards, by set
	ViewParty      // bonds with the four disciples
	ViewWorld      // a town walked into: notice board, inn, stone
	ViewTitle      // the opening screen: continue, new or load a journey
	ViewMap        // walking the region's map: the journey's hub
)

// SwitchViewMsg tells the app to change views.
type SwitchViewMsg struct {
	Target View
}

// SwitchTo returns a tea.Cmd that sends a SwitchViewMsg.
func SwitchTo(v View) tea.Cmd {
	return func() tea.Msg { return SwitchViewMsg{Target: v} }
}

// DashboardKeys defines keybindings for the dashboard view.
type DashboardKeys struct {
	Drill    key.Binding
	Browse   key.Binding
	Chat     key.Binding
	Review   key.Binding
	Voice    key.Binding
	Stats    key.Binding
	Search   key.Binding
	Journey  key.Binding
	Radicals key.Binding
	Reading  key.Binding
	Assess   key.Binding
	Battle   key.Binding
	Shrine   key.Binding
	Cards    key.Binding
	Party    key.Binding
}

// DefaultDashboardKeys returns dashboard-specific keybindings.
func DefaultDashboardKeys() DashboardKeys {
	return DashboardKeys{
		Drill: key.NewBinding(
			key.WithKeys("d"),
			key.WithHelp("d", "drill"),
		),
		Browse: key.NewBinding(
			key.WithKeys("b"),
			key.WithHelp("b", "browse"),
		),
		Chat: key.NewBinding(
			key.WithKeys("c"),
			key.WithHelp("c", "chat"),
		),
		Review: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "review"),
		),
		Voice: key.NewBinding(
			key.WithKeys("v"),
			key.WithHelp("v", "voice"),
		),
		Stats: key.NewBinding(
			key.WithKeys("s"),
			key.WithHelp("s", "stats"),
		),
		Search: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "search"),
		),
		Journey: key.NewBinding(
			key.WithKeys("j"),
			key.WithHelp("j", "journey"),
		),
		Radicals: key.NewBinding(
			key.WithKeys("x"),
			key.WithHelp("x", "radicals"),
		),
		Battle: key.NewBinding(
			key.WithKeys("f"),
			key.WithHelp("f", "fight"),
		),
		Shrine: key.NewBinding(
			key.WithKeys("g"),
			key.WithHelp("g", "gear"),
		),
		Cards: key.NewBinding(
			key.WithKeys("o"),
			key.WithHelp("o", "collection"),
		),
		Party: key.NewBinding(
			key.WithKeys("p"),
			key.WithHelp("p", "party"),
		),
		Reading: key.NewBinding(
			key.WithKeys("t"),
			key.WithHelp("t", "read"),
		),
		Assess: key.NewBinding(
			key.WithKeys("a"),
			key.WithHelp("a", "assess"),
		),
	}
}

// HelpText returns a formatted help string for a given view.
func HelpText(v View) string {
	switch v {
	case ViewDashboard:
		return menuHelp()
	case ViewStats:
		return "Stats Keys\n\n" +
			"  esc  Back to dashboard\n" +
			"  ?    Toggle help"
	case ViewJourney:
		return "Journey Keys\n\n" +
			"  tab    Switch tab\n" +
			"  ←/→    Switch tab\n" +
			"  esc    Back to dashboard\n" +
			"  ?      Toggle help"
	case ViewRadicals:
		return "Radicals Keys\n\n" +
			"  ↑↓     Select radical\n" +
			"  esc    Back to dashboard\n" +
			"  ?      Toggle help"
	case ViewWorld:
		return "Town Keys\n\n" +
			"  ↑↓     Choose a request\n" +
			"  enter  Read / accept / collect\n" +
			"  1-4    Answer a request\n" +
			"  s      The town's stone\n" +
			"  r      Rest at the inn\n" +
			"  g      Shrine\n" +
			"  esc    Back on the road"
	case ViewParty:
		return "Party Keys\n\n" +
			"  ↑↓     Choose a companion\n" +
			"  t      Talk in Chinese (raises the bond)\n" +
			"  1-4    Read a bond scene\n" +
			"  e      Translation (in a scene)\n" +
			"  esc    Close scene / back"
	case ViewCollection:
		return "Collection Keys\n\n" +
			"  tab    Enemies / characters\n" +
			"  ↑↓     Choose a set\n" +
			"  enter  Claim a completed set\n" +
			"  esc    Back to dashboard"
	case ViewShrine:
		return "Shrine Keys\n\n" +
			"  ↑↓     Choose an item\n" +
			"  enter  Buy, put on or take off\n" +
			"  esc    Back to dashboard"
	case ViewBattle:
		return "Battle Keys\n\n" +
			"  ↑↓     Navigate choices\n" +
			"  enter  Select answer\n" +
			"  1-5    Quick select\n" +
			"  space  Continue\n" +
			"  esc    Flee battle"
	case ViewDrill:
		return "Drill Keys\n\n" +
			"  ↑↓     Select mode/navigate\n" +
			"  enter  Submit answer / select\n" +
			"  space  Next card (after feedback)\n" +
			"  1-4    Rate difficulty\n" +
			"  ctrl+space  Toggle input mode\n" +
			"  esc    Back to dashboard"
	case ViewBrowser:
		return "Browser Keys\n\n" +
			"  ↑↓     Navigate pages\n" +
			"  tab    Switch pane\n" +
			"  enter  Open page\n" +
			"  type   Filter pages\n" +
			"  esc    Back to dashboard"
	case ViewChat:
		return "Chat Keys\n\n" +
			"  enter       Send message\n" +
			"  ctrl+space  Toggle input mode\n" +
			"  pgup/dn     Scroll history\n" +
			"  esc         Back to dashboard"
	case ViewReview:
		return "Review Keys\n\n" +
			"  ↑↓     Navigate items\n" +
			"  enter  Start drill / submit\n" +
			"  1-4    Rate difficulty\n" +
			"  ctrl+space  Toggle input mode\n" +
			"  esc    Back to dashboard"
	default:
		return "  ?  Toggle help\n  esc  Back"
	}
}
