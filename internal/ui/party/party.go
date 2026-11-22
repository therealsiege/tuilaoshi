// Package party shows the bonds with the four disciples (Persona-style
// social links): each bond's rank, what raises it, the upgraded assist at
// 信任, and the scenes each rank opens, read word by word in the reader.
package party

import (
	"errors"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/therealsiege/tuilaoshi/internal/journey"
	"github.com/therealsiege/tuilaoshi/internal/ui"
	"github.com/therealsiege/tuilaoshi/internal/ui/reader"
	"github.com/therealsiege/tuilaoshi/internal/wiki"
)

// member is a disciple with their bond.
type member struct {
	journey.Companion
	bond   journey.Bond
	points int
}

// rank is the member's bond rank index.
func (b member) rank() int { return journey.BondRankFor(b.points) }

// Model is the party view model.
type Model struct {
	store   *wiki.Store
	gkeys   ui.GlobalKeys
	members []member
	cursor  int

	// The scene being read, if any.
	reading     bool
	scene       reader.Model
	sceneOf     journey.BondScene
	showEnglish bool

	note          string
	width, height int
	loadErr       error
}

// New creates a party model.
func New(store *wiki.Store) Model {
	return Model{store: store, gkeys: ui.DefaultGlobalKeys()}
}

func (m Model) Init() tea.Cmd { return nil }

// SetSize updates the terminal dimensions.
func (m Model) SetSize(w, h int) Model {
	m.width, m.height = w, h
	return m
}

// Reading reports whether a scene is open, so esc closes it rather than
// leaving the screen.
func (m Model) Reading() bool { return m.reading }

// Refresh reloads the party and bond points.
func (m Model) Refresh() Model {
	cleared, clearedErr := m.store.ClearedStages()
	points, bondErr := m.store.BondPoints()
	m.loadErr = errors.Join(clearedErr, bondErr)
	m.members = nil
	for _, c := range journey.Party(cleared) {
		if b := journey.BondFor(c.Name); b != nil {
			m.members = append(m.members, member{Companion: c, bond: *b, points: points[c.Name]})
		}
	}
	m.cursor = min(m.cursor, max(len(m.members)-1, 0))
	m.reading, m.note = false, ""
	return m
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	k, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	if m.reading {
		return m.updateScene(k)
	}
	switch {
	case key.Matches(k, m.gkeys.Back):
		return m, ui.SwitchTo(ui.ViewDashboard)
	case key.Matches(k, m.gkeys.Up):
		m.cursor = max(m.cursor-1, 0)
		m.note = ""
	case key.Matches(k, m.gkeys.Down):
		m.cursor = min(m.cursor+1, len(m.members)-1)
		m.note = ""
	default:
		if s := k.String(); len(s) == 1 && s[0] >= '1' && s[0] <= '4' {
			return m.openScene(int(s[0] - '1')), nil
		}
	}
	return m, nil
}

// openScene opens the selected member's scene n (0-based) if its rank is reached.
func (m Model) openScene(n int) Model {
	if len(m.members) == 0 {
		return m
	}
	b := m.members[m.cursor]
	switch {
	case !b.Unlocked:
		m.note = b.Name + " hasn't joined the pilgrimage yet."
	case n >= len(b.bond.Scenes):
	case b.rank() < n+1:
		m.note = "This scene opens at bond rank " + journey.BondRanks[n+1].Name + "."
	default:
		sc := b.bond.Scenes[n]
		m.sceneOf, m.showEnglish, m.reading = sc, false, true
		m.scene = reader.New(m.store).Refresh().OpenText(sc.Title+"  "+sc.EnglishTitle,
			strings.Split(strings.TrimSpace(sc.Text), "\n"), journey.StoryGlossary())
	}
	return m
}

// updateScene reads a scene word by word; e shows the translation, esc or
// tab closes it.
func (m Model) updateScene(k tea.KeyMsg) (Model, tea.Cmd) {
	switch {
	case key.Matches(k, m.gkeys.Back) || k.String() == "tab":
		m.reading = false
		return m, nil
	case k.String() == "e":
		m.showEnglish = !m.showEnglish
		return m, nil
	}
	var cmd tea.Cmd
	m.scene, cmd = m.scene.Update(k)
	return m, cmd
}
