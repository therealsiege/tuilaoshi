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
	ViewTheatre    // the learner's episodes: scenes, listening, talks
	ViewForge      // the character forge: build characters from their parts
	ViewFestival   // Chinese festivals on the real calendar
	ViewSettings   // the learner's preferences
	ViewDiary      // the travel diary kept at camp
)

// InnerDemonMsg asks the app to start the fight with the week's Inner Demon.
type InnerDemonMsg struct{}

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
	Episodes key.Binding
	Forge    key.Binding
	Festival key.Binding
	Settings key.Binding
	Inner    key.Binding
	Diary    key.Binding
	Letter   key.Binding
	Gazette  key.Binding
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
		Episodes: key.NewBinding(
			key.WithKeys("e"),
			key.WithHelp("e", "episodes"),
		),
		Forge: key.NewBinding(
			key.WithKeys("z"),
			key.WithHelp("z", "forge"),
		),
		Festival: key.NewBinding(
			key.WithKeys("y"),
			key.WithHelp("y", "festivals"),
		),
		Settings: key.NewBinding(
			key.WithKeys(","),
			key.WithHelp(",", "settings"),
		),
		Inner: key.NewBinding(
			key.WithKeys("m"),
			key.WithHelp("m", "inner demon"),
		),
		Diary: key.NewBinding(
			key.WithKeys("w"),
			key.WithHelp("w", "diary"),
		),
		Letter: key.NewBinding(
			key.WithKeys("n"),
			key.WithHelp("n", "letter"),
		),
		Gazette: key.NewBinding(
			key.WithKeys("u"),
			key.WithHelp("u", "gazette"),
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
			"  ↑↓     Scroll the numbers (pgup/pgdn a page)\n" +
			"  w      Spider's web 盘丝洞 / back to the numbers\n" +
			"  ↑↓     Web: pick a strand\n" +
			"  →/←    Web: follow it / go back\n" +
			"  c      Web: centre on the weakest word\n" +
			"  enter  Web: revive the slipping words around the centre\n" +
			"  1-4    Reviving: rate (space shows the answer)\n" +
			"  esc    Back to dashboard (while reviving: to the web)\n" +
			"  ?      Toggle help"
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
	case ViewSettings:
		return "Settings Keys\n\n" +
			"  ↑↓     Choose a setting\n" +
			"  ←→     Change it (saved at once)\n" +
			"  esc    Back"
	case ViewDiary:
		return "Diary Keys\n\n" +
			"  enter       Send today's entry\n" +
			"  ctrl+space  Pinyin / English input\n" +
			"  ctrl+r      Hear your entry\n" +
			"  esc         Back to camp"
	case ViewFestival:
		return "Festival Keys\n\n" +
			"  1-2    Answer a question\n" +
			"  e      English translation\n" +
			"  ctrl+r Hear it in Chinese\n" +
			"  esc    Back"
	case ViewForge:
		return "Forge Keys\n\n" +
			"  1-8        Pick a part\n" +
			"  backspace  Undo a part\n" +
			"  enter      Forge / next character\n" +
			"  esc        Back (saves what you forged)"
	case ViewTheatre:
		return "Episodes Keys\n\n" +
			"  ↑↓     Choose an episode / scene\n" +
			"  enter  Open / answer a scene\n" +
			"  r      Read & listen line by line\n" +
			"  l      Listening drill (tab replays)\n" +
			"  s      Shadow: hear a line, say it back (space records)\n" +
			"  t      Talk it over with 孙悟空\n" +
			"  esc    Back"
	case ViewMap:
		return "Road Keys\n\n" +
			"  arrows/hjkl  Walk\n" +
			"  enter        Enter a town / face a tribulation\n" +
			"  p            Pace · f rations\n" +
			"  r            Rest a day\n" +
			"  ctrl+r       Hear an event in Chinese\n" +
			"  1-3          Answer where or when the\n" +
			"               overheard demons strike\n" +
			"  c / esc      Camp: study, party, shrine\n\n" +
			"  @ you   · road (safe)   , wild (clans)\n" +
			"  H town  ! tribulation   & demon's lair\n" +
			"  + won   < > the next regions\n\n" +
			"  Studying at camp opens sealed\n" +
			"  tribulations; winning one opens\n" +
			"  the road past it."
	case ViewWorld:
		return "Town Keys\n\n" +
			"  ↑↓     Choose a request\n" +
			"  enter  Read / accept / collect\n" +
			"  1-4    Answer a request\n" +
			"  s      The town's stone\n" +
			"  r      Rest at the inn\n" +
			"  b      Buy 斋饭 at the store\n" +
			"  x      Buy straw sandals (草鞋) for the party\n" +
			"  o      Make a temple offering, once a day\n" +
			"  l      Guess a lantern riddle (灯谜)\n" +
			"  p      Show your travel papers (通关文牒)\n" +
			"  h      Haggle with the merchant (讨价还价)\n" +
			"  t      Trade a material for 斋饭\n" +
			"  a      Ask for alms (化缘), once a day\n" +
			"  f      Fiery Eyes (火眼金睛): question three\n" +
			"         villagers in Chinese, accuse the demon\n" +
			"  e      An errand (任务): get its points across\n" +
			"         to a townsperson in Chinese, once a day\n" +
			"  ctrl+r Hear the Chinese on screen\n" +
			"  g      Shrine\n" +
			"  esc    Back on the road"
	case ViewParty:
		return "Party Keys\n\n" +
			"  ↑↓     Choose a companion\n" +
			"  t      Talk in Chinese (raises the bond)\n" +
			"  i      Idiom-chain duel (成语接龙)\n" +
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
			"  esc    Flee battle\n\n" +
			"三打白骨精 (the White Bone Demon's disguises)\n\n" +
			"  1-3    Unmask: pick the answer\n" +
			"  r      Hear the old woman again\n" +
			"  enter  Say what gives the old man away\n" +
			"  e      English of the passage (short windows)"
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
	case ViewReading:
		return "Reading Keys\n\n" +
			"  ↑↓     Choose a text / line\n" +
			"  enter  Read / didn't know this word\n" +
			"  s      说书人 The Storyteller: a chapter a day\n" +
			"  e      English · p hear the line\n" +
			"  ctrl+r Hear the storyteller's chapter\n" +
			"  esc    Back"
	default:
		return "  ?  Toggle help\n  esc  Back"
	}
}
