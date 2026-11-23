// Package forgeview is the character forge (字炉): a character still being
// learned is shown by its sound and meaning, and the learner picks its
// written parts out of a tray of look-alikes to forge it. Each character
// forged pays like other recall (once a day, through a study session) and,
// when fresh, yields a material for crafting, by its radical's element.
package forgeview

import (
	"context"
	"errors"
	"math/rand"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/therealsiege/tuilaoshi/internal/claude"
	"github.com/therealsiege/tuilaoshi/internal/forge"
	"github.com/therealsiege/tuilaoshi/internal/journey"
	"github.com/therealsiege/tuilaoshi/internal/model"
	"github.com/therealsiege/tuilaoshi/internal/ui"
	"github.com/therealsiege/tuilaoshi/internal/wiki"
)

// Decomposer asks for characters' written parts.
type Decomposer func(ctx context.Context, chars []string) (map[string][]string, error)

// batch is how many characters a round of the forge works on.
const batch = 12

type state int

const (
	stateLoading state = iota // asking for decompositions
	stateForging
	stateResult
	stateEmpty
)

type decomposedMsg struct {
	parts map[string][]string
	err   error
}

// Model is the forge screen.
type Model struct {
	store     *wiki.Store
	gkeys     ui.GlobalKeys
	decompose Decomposer
	now       func() time.Time
	rng       *rand.Rand

	comps    map[string][]string
	radicals map[string]string
	queue    []model.Card // characters to forge this round, with known parts
	cur      int
	tray     []string
	picked   []int // tray indices, in the order picked

	state   state
	right   bool
	paid    map[string]bool // cards already paying fresh this session
	session model.Session
	loot    []string // materials earned this session

	note          string
	width, height int
	loadErr       error
}

// New creates the forge; without a Claude client only characters already
// decomposed can be forged.
func New(store *wiki.Store, client *claude.Client) Model {
	m := Model{store: store, gkeys: ui.DefaultGlobalKeys(), now: time.Now, rng: rand.New(rand.NewSource(time.Now().UnixNano()))}
	if client != nil {
		m.decompose = client.DecomposeCharacters
	}
	return m
}

func (m Model) Init() tea.Cmd { return nil }

// SetSize updates the terminal dimensions.
func (m Model) SetSize(w, h int) Model {
	m.width, m.height = w, h
	return m
}

// Refresh starts a round: pick characters, and ask for any parts not cached.
func (m Model) Refresh() (Model, tea.Cmd) {
	cards, cardsErr := m.store.DrillableCards()
	pages, pagesErr := m.store.ListPages()
	comps, compErr := m.store.Components()
	m.loadErr = errors.Join(cardsErr, pagesErr, compErr)
	m.comps, m.radicals = comps, journey.RadicalIndex(pages)
	m.session = model.Session{StartedAt: m.now(), Type: "forge"}
	m.paid, m.loot, m.note = map[string]bool{}, nil, ""
	targets := forge.Targets(cards, batch, m.rng)
	var missing []string
	for _, c := range targets {
		if _, ok := comps[c.Chinese]; !ok {
			missing = append(missing, c.Chinese)
		}
	}
	m.queue = targets
	if len(missing) > 0 && m.decompose != nil {
		m.state = stateLoading
		decompose := m.decompose
		return m, func() tea.Msg {
			ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
			defer cancel()
			parts, err := decompose(ctx, missing)
			return decomposedMsg{parts, err}
		}
	}
	return m.start(), nil
}

// start keeps the characters with usable parts and shows the first.
func (m Model) start() Model {
	var ready []model.Card
	for _, c := range m.queue {
		if forge.Forgeable(c.Chinese, m.comps[c.Chinese]) {
			ready = append(ready, c)
		}
	}
	m.queue, m.cur = ready, 0
	if len(ready) == 0 {
		m.state = stateEmpty
		return m
	}
	return m.puzzle()
}

// puzzle lays out the tray for the current character.
func (m Model) puzzle() Model {
	c := m.queue[m.cur]
	m.tray = forge.Tray(m.comps[c.Chinese], forge.Pool(m.comps), m.rng)
	m.picked, m.state = nil, stateForging
	return m
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if d, ok := msg.(decomposedMsg); ok {
		if d.err != nil {
			m.note = "Couldn't reach Claude for new characters; forging the ones already known."
		} else if err := m.store.SaveComponents(d.parts); err != nil {
			m.loadErr = err
		} else {
			for ch, parts := range d.parts {
				m.comps[ch] = parts
			}
		}
		return m.start(), nil
	}
	k, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	if key.Matches(k, m.gkeys.Back) {
		return m.end(), ui.SwitchTo(ui.ViewDashboard)
	}
	switch m.state {
	case stateResult:
		if key.Matches(k, m.gkeys.Enter) || k.Type == tea.KeySpace {
			m.cur++
			if m.cur >= len(m.queue) {
				m = m.end()
				var cmd tea.Cmd
				m, cmd = m.Refresh()
				return m, cmd
			}
			return m.puzzle(), nil
		}
	case stateForging:
		return m.pick(k), nil
	}
	return m, nil
}

// pick toggles tray parts with 1-8; backspace undoes; enter forges.
func (m Model) pick(k tea.KeyMsg) Model {
	switch {
	case k.Type == tea.KeyBackspace && len(m.picked) > 0:
		m.picked = m.picked[:len(m.picked)-1]
	case key.Matches(k, m.gkeys.Enter) && len(m.picked) > 0:
		return m.forge()
	default:
		s := k.String()
		if len(s) == 1 && s[0] >= '1' && int(s[0]-'1') < len(m.tray) {
			m.picked = append(m.picked, int(s[0]-'1'))
		}
	}
	return m
}

// forge checks the picked parts and records the answer.
func (m Model) forge() Model {
	c := m.queue[m.cur]
	var picked []string
	for _, i := range m.picked {
		picked = append(picked, m.tray[i])
	}
	m.right = forge.Forged(m.comps[c.Chinese], picked)
	m.state = stateResult
	m.session.Items = append(m.session.Items, model.SessionItem{CardID: c.ID, UserAnswer: joinParts(picked), Correct: m.right})
	m.session.CardsTotal++
	if !m.right {
		return m
	}
	m.session.CardsCorrect++
	if paid, err := m.store.PaidToday(c.ID, m.now()); err == nil && !paid && !m.paid[c.ID] {
		m.paid[c.ID] = true
		m.loot = append(m.loot, forge.Material(m.radicals[c.Chinese]))
	}
	return m
}

// end saves the session (paying for fresh recall) and the materials earned.
func (m Model) end() Model {
	if len(m.session.Items) == 0 {
		return m
	}
	m.session.EndedAt = m.now().Add(time.Second)
	if err := m.store.WriteSession(m.session); err != nil {
		m.loadErr = err
		return m
	}
	if err := m.store.AddMaterials(m.loot); err != nil {
		m.loadErr = err
	}
	m.session.Items = nil
	return m
}

func joinParts(parts []string) string {
	out := ""
	for i, p := range parts {
		if i > 0 {
			out += "+"
		}
		out += p
	}
	return out
}
