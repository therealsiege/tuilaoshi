// Package title is the opening screen: continue the last journey, start a
// new one, or load another. Each journey keeps its own story, purse, gear,
// bonds and quests; study progress is shared by all of them.
package title

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/therealsiege/tuilaoshi/internal/ui"
	"github.com/therealsiege/tuilaoshi/internal/ui/input"
	"github.com/therealsiege/tuilaoshi/internal/wiki"
)

// StartedMsg says a journey has been chosen and is now the one being played.
type StartedMsg struct{}

func started() tea.Msg { return StartedMsg{} }

type screen int

const (
	screenMenu screen = iota
	screenName        // naming a new journey
	screenLoad        // choosing a saved journey
)

type choice int

const (
	choiceContinue choice = iota
	choiceNew
	choiceLoad
	choiceQuit
)

// Model is the title screen model.
type Model struct {
	store   *wiki.Store
	gkeys   ui.GlobalKeys
	saves   []wiki.SaveInfo
	menu    []choice
	cursor  int
	screen  screen
	pick    int  // cursor on the load list
	confirm bool // d pressed once on the load list: d again deletes
	name    input.Model
	now     func() time.Time

	width, height int
	loadErr       error
}

// New creates the title screen.
func New(store *wiki.Store, dict *input.Dict) Model {
	in := input.NewWithDict(dict)
	in.SetCharLimit(30)
	return Model{store: store, gkeys: ui.DefaultGlobalKeys(), name: in, now: time.Now}
}

func (m Model) Init() tea.Cmd { return nil }

// SetSize updates the terminal dimensions.
func (m Model) SetSize(w, h int) Model {
	m.width, m.height = w, h
	return m
}

// Typing reports whether the name box has the keyboard.
func (m Model) Typing() bool { return m.screen == screenName }

// Refresh lists the saved journeys and builds the menu from them.
func (m Model) Refresh() Model {
	m.saves, m.loadErr = m.store.Saves()
	m.menu = nil
	if len(m.saves) > 0 {
		m.menu = append(m.menu, choiceContinue)
	}
	m.menu = append(m.menu, choiceNew)
	if len(m.saves) > 0 {
		m.menu = append(m.menu, choiceLoad)
	}
	m.menu = append(m.menu, choiceQuit)
	m.cursor = min(m.cursor, len(m.menu)-1)
	m.pick = min(m.pick, max(len(m.saves)-1, 0))
	m.screen, m.confirm = screenMenu, false
	return m
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	k, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	switch m.screen {
	case screenName:
		return m.updateName(k)
	case screenLoad:
		return m.updateLoad(k)
	}
	switch {
	case key.Matches(k, m.gkeys.Up):
		m.cursor = max(m.cursor-1, 0)
	case key.Matches(k, m.gkeys.Down):
		m.cursor = min(m.cursor+1, len(m.menu)-1)
	case key.Matches(k, m.gkeys.Enter) || k.Type == tea.KeySpace:
		return m.choose(m.menu[m.cursor])
	case key.Matches(k, m.gkeys.Quit):
		return m, tea.Quit
	}
	return m, nil
}

func (m Model) choose(c choice) (Model, tea.Cmd) {
	switch c {
	case choiceContinue:
		return m.use(m.saves[0].ID)
	case choiceNew:
		m.screen = screenName
		m.name.Reset()
		m.name.SetPlaceholder(fmt.Sprintf("第%d次西行  (enter for this name)", len(m.saves)+1))
		return m, m.name.Focus()
	case choiceLoad:
		m.screen, m.confirm = screenLoad, false
		return m, nil
	}
	return m, tea.Quit
}

// use starts playing a saved journey.
func (m Model) use(id string) (Model, tea.Cmd) {
	if err := m.store.UseSave(id, m.now()); err != nil {
		m.loadErr = err
		return m, nil
	}
	return m, started
}

func (m Model) updateName(k tea.KeyMsg) (Model, tea.Cmd) {
	switch {
	case k.Type == tea.KeyEsc && !m.name.Composing():
		m.screen = screenMenu
		return m, nil
	case k.Type == tea.KeyEnter && !m.name.Composing():
		name := strings.TrimSpace(m.name.Value())
		if name == "" {
			name = fmt.Sprintf("第%d次西行", len(m.saves)+1)
		}
		if _, err := m.store.NewSave(name, m.now()); err != nil {
			m.loadErr = err
			return m, nil
		}
		return m, started
	}
	var cmd tea.Cmd
	m.name, cmd = m.name.Update(k)
	return m, cmd
}

func (m Model) updateLoad(k tea.KeyMsg) (Model, tea.Cmd) {
	switch {
	case key.Matches(k, m.gkeys.Back):
		m.screen = screenMenu
	case key.Matches(k, m.gkeys.Up):
		m.pick, m.confirm = max(m.pick-1, 0), false
	case key.Matches(k, m.gkeys.Down):
		m.pick, m.confirm = min(m.pick+1, len(m.saves)-1), false
	case key.Matches(k, m.gkeys.Enter):
		return m.use(m.saves[m.pick].ID)
	case k.String() == "d":
		if !m.confirm {
			m.confirm = true
			return m, nil
		}
		if err := m.store.DeleteSave(m.saves[m.pick].ID); err != nil {
			m.loadErr = err
			return m, nil
		}
		m = m.Refresh()
		if len(m.saves) > 0 {
			m.screen = screenLoad
		}
	default:
		m.confirm = false
	}
	return m, nil
}
