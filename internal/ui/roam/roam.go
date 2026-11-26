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
	"math/rand"
	"sort"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/therealsiege/tuilaoshi/internal/audio"
	"github.com/therealsiege/tuilaoshi/internal/journey"
	"github.com/therealsiege/tuilaoshi/internal/ui"
	"github.com/therealsiege/tuilaoshi/internal/ui/input"
	rpgview "github.com/therealsiege/tuilaoshi/internal/ui/rpg"
	"github.com/therealsiege/tuilaoshi/internal/wiki"
)

// EnterTownMsg asks the app to open a town's screen.
type EnterTownMsg struct{ Town string }

// FightMsg asks the app to start the battle for the tribulation faced.
type FightMsg struct{}

// EncounterMsg asks the app to start a fight with a soldier met on wild ground.
type EncounterMsg struct{ Soldier journey.Soldier }

// encounterChance is the chance, in percent, that a step onto wild ground
// meets one of the region's clans; the road itself is safe.
const encounterChance = 12

// encounterRate is the chance in effect, by the learner's setting.
func encounterRate() int {
	switch ui.Prefs.Encounters {
	case "fewer":
		return encounterChance / 2
	case "off":
		return 0
	}
	return encounterChance
}

// encounterRoll draws 0-99 against encounterRate; tests fix it.
var encounterRoll = func() int { return rand.Intn(100) }

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
	hp, max int                        // the pilgrim's HP on the road, and its most
	quests  []string                   // the quest log: hunts under way
	trail   journey.Trail              // days, supplies and the party (see trail.go)
	members []string                   // the party walking
	event   *trailEvent                // an event being met (see events.go)
	plot    *overheard                 // demons overheard plotting (see plots.go)
	tts     *audio.Client              // speaks the overheard demons; nil: read them
	states  map[string]wiki.QuestState // side quests' states, for the road strip
	gates   []string                   // regions whose grammar gate is passed
	gate    *gateRun                   // the grammar gate being taken (see gate.go)
	grader  Grader                     // Claude judging the gate's typed round; nil: tiles only
	gateIn  input.Model                // the gate's typed round
	note    string

	width, height int
	loadErr       error
}

// New creates the walking map; tts (nil for none) voices overheard demons.
func New(store *wiki.Store, tts *audio.Client) Model {
	return Model{store: store, gkeys: ui.DefaultGlobalKeys(), tts: tts, gateIn: gateInput(nil)}
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
	gates, gatesErr := m.store.GrammarGates()
	m.loadErr = errors.Join(pagesErr, clearedErr, placeErr, walletErr, gatesErr)
	m.cleared, m.wallet, m.gates, m.gate = cleared, wallet, gates, nil
	states, questErr := m.store.Quests(time.Now())
	m.states, m.loadErr = states, errors.Join(m.loadErr, questErr)
	m.quests = m.questLog()
	m.loadTrail()
	hp, most, hpErr := rpgview.PilgrimHP(m.store)
	m.hp, m.max = hp, most
	m.loadErr = errors.Join(m.loadErr, hpErr)
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

// questLog lists the hunts under way: target, progress and where to report.
func (m Model) questLog() []string {
	hunts, err := m.store.Hunts()
	if err != nil {
		return nil
	}
	states := m.states
	var log []string
	for id, beaten := range hunts {
		q, town := journey.QuestByID(id)
		if q == nil || states[id] == wiki.QuestDone {
			continue
		}
		line := fmt.Sprintf("%s %d/%d", q.Target, beaten, q.Count)
		if beaten >= q.Count {
			line = fmt.Sprintf("%s done: back to %s", q.Target, town.Name)
		}
		log = append(log, line)
	}
	sort.Strings(log)
	return log
}

// save records where the pilgrim stands.
func (m *Model) save() {
	if err := m.store.SetPlace(wiki.Place{Region: m.region.Region.ID, X: m.pos.X, Y: m.pos.Y}); err != nil {
		m.loadErr = err
	}
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if g, ok := msg.(gateGradeMsg); ok {
		return m.onGateGrade(g), nil
	}
	k, ok := msg.(tea.KeyMsg)
	if !ok {
		if m.Typing() {
			var cmd tea.Cmd
			m.gateIn, cmd = m.gateIn.Update(msg) // the cursor's blink
			return m, cmd
		}
		return m, nil
	}
	if m.gate != nil {
		return m.updateGate(k)
	}
	if m.event != nil {
		return m.updateEvent(k)
	}
	if m.plot != nil {
		return m.updatePlot(k)
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
	if m, ok := m.trailKey(k); ok {
		return m, nil
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
	case journey.TileCrossing:
		if !m.hasCrossed() {
			return m.faceCrossing(), nil // the obstacle bars the road until crossed
		}
	}
	m.pos = to
	m.save()
	if m = m.walked(); m.event != nil {
		return m, nil // the day's event comes before anything else on this step
	}
	if m.plot != nil {
		return m, m.speakPlot() // and so do the night's plotting demons
	}
	if t, ok := m.region.Towns[to]; ok {
		id := t.ID
		return m, func() tea.Msg { return EnterTownMsg{Town: id} }
	}
	if m.region.At(to) == journey.TileWild && encounterRoll() < encounterRate() {
		if s, ok := m.region.Region.Encounter(rand.New(rand.NewSource(rand.Int63()))); ok {
			m.note = s.Name + " " + s.English + " leaps out of the undergrowth!"
			ui.Cue(audio.Encounter)
			return m, func() tea.Msg { return EncounterMsg{Soldier: s} }
		}
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

// face meets the tribulation at p: its battle when it is the open stage
// (for the demon's lair, once its grammar gate is passed), otherwise what
// it still needs.
func (m Model) face(p journey.Point) (Model, tea.Cmd) {
	n := m.region.Stages[p]
	s := m.stage
	switch {
	case s == nil || s.Number != n:
		m.note = "An earlier tribulation still waits on the road behind you."
	case s.Status != journey.StageOpen:
		m.note = fmt.Sprintf("Tribulation %d is sealed: %s. %s", n, journey.NeedText(*s, m.counts), journey.HowToOpen(*s))
	case n == m.region.Region.To && !m.gatePassed():
		return m.openGate(), nil // the gate stands before the lair
	default:
		return m, func() tea.Msg { return FightMsg{} }
	}
	return m, nil
}

func (m Model) stageCleared(p journey.Point) bool {
	_, ok := m.cleared[m.region.Stages[p]]
	return ok
}
