// Package roam is the journey's hub: the pilgrim walks the region's tile
// map with the arrow keys. Towns are entered by walking into them, each
// tribulation bars the road until it is won, and the last is the demon's
// lair; the doors at either edge lead to the neighbouring regions. Camp
// (studying, party, shrine, collection) is one key away. Where the pilgrim
// stands is saved with the journey.
package roam

import (
	"errors"
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/therealsiege/tuilaoshi/internal/journey"
	"github.com/therealsiege/tuilaoshi/internal/ui"
	"github.com/therealsiege/tuilaoshi/internal/wiki"
)

// EnterTownMsg asks the app to open a town's screen.
type EnterTownMsg struct{ Town string }

// FightMsg asks the app to start the battle for the tribulation stood on.
type FightMsg struct{}

// Model is the walking map.
type Model struct {
	store   *wiki.Store
	gkeys   ui.GlobalKeys
	region  journey.RegionMap
	pos     journey.Point
	cleared map[int]string
	counts  journey.Counts
	stage   *journey.Stage // the stage being worked on; nil once all are cleared
	wallet  wiki.Wallet
	note    string

	width, height int
	loadErr       error
}

// New creates the walking map.
func New(store *wiki.Store) Model {
	return Model{store: store, gkeys: ui.DefaultGlobalKeys()}
}

func (m Model) Init() tea.Cmd { return nil }

// SetSize updates the terminal dimensions.
func (m Model) SetSize(w, h int) Model {
	m.width, m.height = w, h
	return m
}

// Refresh reloads the journey and puts the pilgrim where they last stood,
// or at the start of the region being worked on.
func (m Model) Refresh() Model {
	pages, pagesErr := m.store.ListPages()
	cleared, clearedErr := m.store.ClearedStages()
	place, placeErr := m.store.Place()
	wallet, walletErr := m.store.Wallet()
	m.loadErr = errors.Join(pagesErr, clearedErr, placeErr, walletErr)
	m.cleared, m.wallet = cleared, wallet
	m.counts = journey.CountProgress(pages)
	m.stage = journey.CurrentStage(journey.Stages(m.counts, cleared))
	if i := journey.RegionByID(place.Region); i >= 0 {
		m.region, m.pos = journey.BuildMap(i), journey.Point{X: place.X, Y: place.Y}
	} else {
		n := 81 // the pilgrimage is complete: stand at its end
		if m.stage != nil {
			n = m.stage.Number
		}
		m.region = journey.BuildMap(journey.RegionOf(n))
		m.pos = m.region.Start
		m.save()
	}
	m.note = ""
	return m
}

// save records where the pilgrim stands.
func (m *Model) save() {
	if err := m.store.SetPlace(wiki.Place{Region: m.region.Region.ID, X: m.pos.X, Y: m.pos.Y}); err != nil {
		m.loadErr = err
	}
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	k, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	switch {
	case key.Matches(k, m.gkeys.Up):
		return m.step(0, -1)
	case key.Matches(k, m.gkeys.Down):
		return m.step(0, 1)
	case key.Matches(k, m.gkeys.Left):
		return m.step(-1, 0)
	case key.Matches(k, m.gkeys.Right):
		return m.step(1, 0)
	case key.Matches(k, m.gkeys.Back) || k.String() == "c":
		return m, ui.SwitchTo(ui.ViewDashboard)
	case key.Matches(k, m.gkeys.Enter):
		return m.act()
	}
	return m, nil
}

// step moves the pilgrim one cell, if the ground allows.
func (m Model) step(dx, dy int) (Model, tea.Cmd) {
	to := journey.Point{X: m.pos.X + dx, Y: m.pos.Y + dy}
	m.note = ""
	switch m.region.At(to) {
	case journey.TileBlock:
		return m, nil
	case journey.TileEast:
		return m.travel(m.region.Index-1, false), nil
	case journey.TileWest:
		return m.travel(m.region.Index+1, true), nil
	case journey.TileStage:
		if !m.stageCleared(to) {
			return m.face(to) // the tribulation bars the road: face it, don't pass
		}
	}
	m.pos = to
	m.save()
	if t, ok := m.region.Towns[to]; ok {
		id := t.ID
		return m, func() tea.Msg { return EnterTownMsg{Town: id} }
	}
	return m, nil
}

// travel walks through a door into the neighbouring region.
func (m Model) travel(i int, west bool) Model {
	m.region = journey.BuildMap(i)
	m.pos = m.region.End
	if west {
		m.pos = m.region.Start
	}
	m.note = "You come to " + m.region.Region.Name + " " + m.region.Region.English + "."
	m.save()
	return m
}

// act enters the town the pilgrim stands in, or faces an uncleared
// tribulation next to them.
func (m Model) act() (Model, tea.Cmd) {
	if t, ok := m.region.Towns[m.pos]; ok {
		id := t.ID
		return m, func() tea.Msg { return EnterTownMsg{Town: id} }
	}
	for _, d := range []journey.Point{{X: -1}, {X: 1}, {Y: -1}, {Y: 1}} {
		p := journey.Point{X: m.pos.X + d.X, Y: m.pos.Y + d.Y}
		if _, ok := m.region.Stages[p]; ok && !m.stageCleared(p) {
			return m.face(p)
		}
	}
	return m, nil
}

// face meets the tribulation at p: its battle when it is the open stage,
// otherwise what it still needs.
func (m Model) face(p journey.Point) (Model, tea.Cmd) {
	n := m.region.Stages[p]
	s := m.stage
	switch {
	case s == nil || s.Number != n:
		m.note = "An earlier tribulation still waits on the road behind you."
	case s.Status != journey.StageOpen:
		m.note = fmt.Sprintf("Tribulation %d is sealed: %s. Study at camp (c) to open it.", n, journey.NeedText(*s, m.counts))
	default:
		return m, func() tea.Msg { return FightMsg{} }
	}
	return m, nil
}

func (m Model) stageCleared(p journey.Point) bool {
	_, ok := m.cleared[m.region.Stages[p]]
	return ok
}
