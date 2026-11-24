// Package collection is the card collection (图鉴): every enemy beaten and
// every character mastered is a card, grouped into sets by region and by
// radical family. A completed set pays its reward once, when claimed.
package collection

import (
	"errors"
	"fmt"
	"slices"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/therealsiege/tuilaoshi/internal/audio"
	"github.com/therealsiege/tuilaoshi/internal/journey"
	"github.com/therealsiege/tuilaoshi/internal/ui"
	"github.com/therealsiege/tuilaoshi/internal/wiki"
)

type tab int

const (
	tabEnemies tab = iota
	tabCharacters
)

// Model is the collection view model.
type Model struct {
	store   *wiki.Store
	gkeys   ui.GlobalKeys
	tab     tab
	sets    [2][]journey.CardSet
	got     [2]map[string]bool // collected members, per tab
	claimed []string
	cursor  [2]int
	offset  [2]int
	note    string
	noteErr bool
	width   int
	height  int
	loadErr error
}

// New creates a collection model.
func New(store *wiki.Store) Model {
	return Model{store: store, gkeys: ui.DefaultGlobalKeys()}
}

func (m Model) Init() tea.Cmd { return nil }

// SetSize updates the terminal dimensions.
func (m Model) SetSize(w, h int) Model {
	m.width, m.height = w, h
	return m
}

// Refresh reloads the enemies beaten and the characters mastered.
func (m Model) Refresh() Model {
	pages, pagesErr := m.store.ListPages()
	c, colErr := m.store.Collection()
	m.loadErr = errors.Join(pagesErr, colErr)
	m.sets[tabEnemies] = journey.EnemySets()
	m.sets[tabCharacters] = journey.RadicalSets(pages)
	m.got[tabEnemies] = map[string]bool{}
	for _, d := range c.Defeated {
		m.got[tabEnemies][d] = true
	}
	m.got[tabCharacters] = journey.MasteredCharacters(pages)
	m.claimed = c.Claimed
	for t := range m.cursor {
		m.cursor[t] = min(m.cursor[t], max(len(m.sets[t])-1, 0))
	}
	m.note = ""
	return m
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	k, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	switch {
	case key.Matches(k, m.gkeys.Back):
		return m, ui.SwitchTo(ui.ViewDashboard)
	case key.Matches(k, m.gkeys.Tab) || key.Matches(k, m.gkeys.Left) || key.Matches(k, m.gkeys.Right):
		m.tab = 1 - m.tab
		m.note = ""
	case key.Matches(k, m.gkeys.Up):
		m.move(-1)
	case key.Matches(k, m.gkeys.Down):
		m.move(1)
	case k.Type == tea.KeyPgUp:
		m.move(-m.listRows())
	case k.Type == tea.KeyPgDown:
		m.move(m.listRows())
	case key.Matches(k, m.gkeys.Enter) || k.Type == tea.KeySpace:
		return m.claim(), nil
	}
	return m, nil
}

func (m *Model) move(by int) {
	t := m.tab
	if len(m.sets[t]) == 0 {
		return
	}
	m.cursor[t] = min(max(m.cursor[t]+by, 0), len(m.sets[t])-1)
	rows := m.listRows()
	if m.cursor[t] < m.offset[t] {
		m.offset[t] = m.cursor[t]
	} else if m.cursor[t] >= m.offset[t]+rows {
		m.offset[t] = m.cursor[t] - rows + 1
	}
}

// claim pays the selected set's reward once it is complete.
func (m Model) claim() Model {
	if len(m.sets[m.tab]) == 0 {
		return m
	}
	s := m.sets[m.tab][m.cursor[m.tab]]
	switch have := s.Have(m.got[m.tab]); {
	case slices.Contains(m.claimed, s.ID):
		m.note, m.noteErr = "Already claimed.", false
	case have < len(s.Members):
		m.note, m.noteErr = fmt.Sprintf("%d more to complete this set.", len(s.Members)-have), true
	default:
		if _, err := m.store.Claim(s.ID, wiki.Wallet(s.Reward)); err != nil {
			m.note, m.noteErr = err.Error(), true
			return m
		}
		ui.Cue(audio.Coins)
		m = m.Refresh()
		m.note, m.noteErr = "Set complete! The Earth God pays "+priceText(s.Reward)+".", false
	}
	return m
}

// status counts what set s has, and whether it is complete and claimed.
func (m Model) status(t tab, s journey.CardSet) (have int, complete, claimed bool) {
	have = s.Have(m.got[t])
	return have, have == len(s.Members), slices.Contains(m.claimed, s.ID)
}
