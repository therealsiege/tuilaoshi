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
	ViewImagen
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
	Drill  key.Binding
	Browse key.Binding
	Chat   key.Binding
	Review key.Binding
	Voice  key.Binding
	Imagen key.Binding
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
		Imagen: key.NewBinding(
			key.WithKeys("i"),
			key.WithHelp("i", "imagen"),
		),
	}
}

// HelpText returns a formatted help string for a given view.
func HelpText(v View) string {
	switch v {
	case ViewDashboard:
		return "Dashboard Keys\n\n" +
			"  d  Start drill\n" +
			"  b  Browse wiki\n" +
			"  c  Chat with tutor\n" +
			"  r  Review due items\n" +
			"  i  Generate images\n" +
			"  v  Change voice\n" +
			"  ?  Toggle help\n" +
			"  q  Quit"
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
	case ViewImagen:
		return "Image Studio Keys\n\n" +
			"  enter       Generate image\n" +
			"  o           Open last image\n" +
			"  ctrl+space  Toggle input mode\n" +
			"  pgup/dn     Scroll history\n" +
			"  esc         Back to dashboard"
	default:
		return "  ?  Toggle help\n  esc  Back"
	}
}
