// Package world is the road west as a map: the towns between the stages,
// each with a notice board of townsfolk's requests. A request is a short
// notice or letter in Chinese with a question about it; answering right
// pays copper once, and a wrong answer closes it until tomorrow, so the
// reward follows reading rather than guessing.
package world

import (
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/therealsiege/tuilaoshi/internal/journey"
	"github.com/therealsiege/tuilaoshi/internal/ui"
	"github.com/therealsiege/tuilaoshi/internal/wiki"
)

type screen int

const (
	screenMap screen = iota
	screenTown
	screenQuest
)

// Model is the world map view model.
type Model struct {
	store   *wiki.Store
	gkeys   ui.GlobalKeys
	towns   []journey.Town
	current int // the stage being worked on; 0 once all are cleared
	states  map[string]wiki.QuestState
	screen  screen
	town    int // cursor on the map, and the town open
	quest   int // cursor on the notice board, and the quest open

	// The open quest's choices, shuffled, and the answer once given.
	choices  []string
	right    int
	answered int // -1 until answered
	now      func() time.Time

	note          string
	width, height int
	loadErr       error
}

// New creates a world map model.
func New(store *wiki.Store) Model {
	return Model{store: store, gkeys: ui.DefaultGlobalKeys(), towns: journey.Towns(), now: time.Now}
}

func (m Model) Init() tea.Cmd { return nil }

// SetSize updates the terminal dimensions.
func (m Model) SetSize(w, h int) Model {
	m.width, m.height = w, h
	return m
}

// Busy reports whether a town or quest is open, so esc steps back to the
// map rather than leaving the screen.
func (m Model) Busy() bool { return m.screen != screenMap }

// Refresh reloads how far the pilgrimage has come and the quests' states.
func (m Model) Refresh() Model {
	pages, pagesErr := m.store.ListPages()
	cleared, clearedErr := m.store.ClearedStages()
	states, questErr := m.store.Quests(m.now())
	m.loadErr = errors.Join(pagesErr, clearedErr, questErr)
	m.states = states
	m.current = 0
	if s := journey.CurrentStage(journey.Stages(journey.CountProgress(pages), cleared)); s != nil {
		m.current = s.Number
	}
	if m.screen == screenMap {
		// Start at the furthest town reached.
		for i, t := range m.towns {
			if t.Reached(m.current) {
				m.town = i
			}
		}
	}
	return m
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	k, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	back := key.Matches(k, m.gkeys.Back)
	switch m.screen {
	case screenQuest:
		if back {
			m.screen, m.note = screenTown, ""
			return m.Refresh(), nil
		}
		return m.answer(k), nil
	case screenTown:
		switch {
		case back:
			m.screen, m.note = screenMap, ""
		case key.Matches(k, m.gkeys.Up):
			m.quest = max(m.quest-1, 0)
		case key.Matches(k, m.gkeys.Down):
			m.quest = min(m.quest+1, len(m.towns[m.town].Quests)-1)
		case key.Matches(k, m.gkeys.Enter):
			return m.openQuest(), nil
		case k.String() == "g":
			return m, ui.SwitchTo(ui.ViewShrine)
		}
		return m, nil
	}
	switch {
	case back:
		return m, ui.SwitchTo(ui.ViewDashboard)
	case key.Matches(k, m.gkeys.Up):
		m.town, m.note = max(m.town-1, 0), ""
	case key.Matches(k, m.gkeys.Down):
		m.town, m.note = min(m.town+1, len(m.towns)-1), ""
	case key.Matches(k, m.gkeys.Enter):
		t := m.towns[m.town]
		if !t.Reached(m.current) {
			m.note = fmt.Sprintf("The road reaches %s at tribulation %d.", t.Name, t.Stage)
			return m, nil
		}
		m.screen, m.quest, m.note = screenTown, 0, ""
	}
	return m, nil
}

// openQuest opens the selected request, unless done or closed for today.
func (m Model) openQuest() Model {
	q := m.towns[m.town].Quests[m.quest]
	switch m.states[q.ID] {
	case wiki.QuestDone:
		m.note = "Done. " + q.Giver + " thanks you again."
		return m
	case wiki.QuestLocked:
		m.note = "Not today: read it again tomorrow."
		return m
	}
	rng := rand.New(rand.NewSource(m.now().UnixNano()))
	order := rng.Perm(len(q.Choices))
	m.choices = make([]string, len(q.Choices))
	for i, j := range order {
		m.choices[i] = q.Choices[j]
		if j == 0 {
			m.right = i
		}
	}
	m.screen, m.answered, m.note = screenQuest, -1, ""
	return m
}

// answer takes 1-4 for a choice; after answering, any key returns.
func (m Model) answer(k tea.KeyMsg) Model {
	if m.answered >= 0 {
		if key.Matches(k, m.gkeys.Enter) || k.Type == tea.KeySpace {
			m.screen = screenTown
			return m.Refresh()
		}
		return m
	}
	s := k.String()
	if len(s) != 1 || s[0] < '1' || int(s[0]-'1') >= len(m.choices) {
		return m
	}
	m.answered = int(s[0] - '1')
	q := m.towns[m.town].Quests[m.quest]
	var err error
	if m.answered == m.right {
		err = m.store.FinishQuest(q.ID, wiki.Wallet{Tongqian: journey.QuestReward})
	} else {
		err = m.store.FailQuest(q.ID, m.now())
	}
	if err != nil {
		m.loadErr = err
	}
	return m
}
