// Package world is a town on the road west, entered by walking into it on
// the region map: its notice board of townsfolk's requests, its inn and its
// stone. A reading request is a short notice or letter in Chinese with a
// question about it; answering right pays copper once, and a wrong answer
// closes it until tomorrow, so the reward follows reading rather than
// guessing.
package world

import (
	"errors"
	"math/rand"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/therealsiege/tuilaoshi/internal/claude"
	"github.com/therealsiege/tuilaoshi/internal/journey"
	"github.com/therealsiege/tuilaoshi/internal/ui"
	"github.com/therealsiege/tuilaoshi/internal/ui/input"
	"github.com/therealsiege/tuilaoshi/internal/wiki"
)

type screen int

const (
	screenTown screen = iota
	screenQuest
	screenStone    // a town's secret stone (see stone.go)
	screenAlms     // asking a townsperson for alms (see alms.go)
	screenRiddle   // a lantern riddle (see riddle.go)
	screenGate     // a kingdom's gate questions (see passport.go)
	screenPassport // the stamps collected
)

// Model is the world map view model.
type Model struct {
	store  *wiki.Store
	gkeys  ui.GlobalKeys
	towns  []journey.Town
	states map[string]wiki.QuestState
	screen screen
	town   int // cursor on the map, and the town open
	quest  int // cursor on the notice board, and the quest open

	// The open quest's choices, shuffled, and the answer once given.
	choices  []string
	right    int
	answered int // -1 until answered
	now      func() time.Time

	// Writing requests: the answer box, and Claude's grade (see write.go).
	grader   Grader
	input    input.Model
	grading  bool
	feedback string

	stones []string       // towns whose stone is open
	hunts  map[string]int // hunts accepted: id -> soldiers beaten
	hinted bool           // the open stone's pinyin has been shown

	alms      journey.Alms // the townsperson asked for alms
	almsOrder []int        // the alms choices' order on screen

	riddle journey.Riddle // the lantern riddle open
	tries  int            // wrong guesses at it

	stamps []string     // kingdoms that stamped the rescript, by town
	gate   journey.Gate // the gate being passed
	gateQ  int          // its question being asked
	misses int          // wrong answers to it

	note          string
	width, height int
	loadErr       error
}

// New creates a world map model; with no Claude client, writing requests
// can be read but not graded.
func New(store *wiki.Store, client *claude.Client, dict *input.Dict) Model {
	in := input.NewWithDict(dict)
	in.SetPlaceholder("write in Chinese: type pinyin, then pick characters")
	in.SetCharLimit(200)
	m := Model{store: store, gkeys: ui.DefaultGlobalKeys(), towns: journey.Towns(), now: time.Now, input: in}
	if client != nil {
		m.grader = client.GradeWriting
	}
	return m
}

// Typing reports whether a writing request or a stone has the keyboard.
func (m Model) Typing() bool {
	return m.screen == screenStone || m.screen == screenRiddle || m.screen == screenGate || (m.screen == screenQuest && m.towns[m.town].Quests[m.quest].Writing())
}

func (m Model) Init() tea.Cmd { return nil }

// SetSize updates the terminal dimensions.
func (m Model) SetSize(w, h int) Model {
	m.width, m.height = w, h
	return m
}

// OpenTown shows a town walked into from the region map.
func (m Model) OpenTown(id string) Model {
	for i, t := range m.towns {
		if t.ID == id {
			m.town, m.quest, m.screen, m.note = i, 0, screenTown, ""
		}
	}
	return m
}

// Refresh reloads the quests' states, the stones opened and the hunts.
func (m Model) Refresh() Model {
	states, questErr := m.store.Quests(m.now())
	stones, stoneErr := m.store.StonesOpened()
	hunts, huntErr := m.store.Hunts()
	stamps, stampErr := m.store.Stamps()
	m.loadErr = errors.Join(questErr, stoneErr, huntErr, stampErr)
	m.stamps = stamps
	m.hunts, m.stones, m.states = hunts, stones, states
	return m
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if g, ok := msg.(gradeMsg); ok {
		return m.onGrade(g), nil
	}
	k, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	back := key.Matches(k, m.gkeys.Back)
	switch m.screen {
	case screenPassport:
		if back {
			m.screen = screenTown
		}
		return m, nil
	case screenGate:
		if back && !m.input.Composing() {
			m.screen, m.note = screenTown, ""
			return m.Refresh(), nil
		}
		return m.updateGate(k)
	case screenRiddle:
		if back && !m.input.Composing() {
			m.screen, m.note = screenTown, ""
			return m.Refresh(), nil
		}
		return m.updateRiddle(k)
	case screenAlms:
		if back && m.answered < 0 {
			m.screen, m.note = screenTown, ""
			return m, nil
		}
		return m.updateAlms(k), nil
	case screenStone:
		if back && !m.input.Composing() {
			m.screen, m.note = screenTown, ""
			return m.Refresh(), nil
		}
		return m.updateStone(k)
	case screenQuest:
		if back && !m.input.Composing() {
			m.screen, m.note = screenTown, ""
			return m.Refresh(), nil
		}
		if m.Typing() {
			return m.updateWriting(k)
		}
		if m.towns[m.town].Quests[m.quest].Hunting() {
			return m.updateHunt(k)
		}
		return m.answer(k), nil
	case screenTown:
		switch {
		case back:
			return m, ui.SwitchTo(ui.ViewMap)
		case key.Matches(k, m.gkeys.Up):
			m.quest = max(m.quest-1, 0)
		case key.Matches(k, m.gkeys.Down):
			m.quest = min(m.quest+1, len(m.towns[m.town].Quests)-1)
		case key.Matches(k, m.gkeys.Enter):
			m = m.openQuest()
			if m.Typing() {
				return m, m.input.Focus()
			}
			return m, nil
		case k.String() == "g":
			return m, ui.SwitchTo(ui.ViewShrine)
		case k.String() == "s":
			return m.openStone()
		case k.String() == "r":
			return m.rest(), nil
		case k.String() == "b":
			return m.buy(), nil
		case k.String() == "a":
			return m.openAlms(), nil
		case k.String() == "t":
			return m.trade(), nil
		case k.String() == "x":
			return m.buySandals(), nil
		case k.String() == "o":
			return m.offer(), nil
		case k.String() == "l":
			return m.openRiddle()
		case k.String() == "p":
			return m.openPassport()
		}
		return m, nil
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
	m.screen, m.answered, m.note, m.feedback = screenQuest, -1, "", ""
	if q.Hunting() {
		return m
	}
	if q.Writing() {
		m.input.Reset()
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
		err = m.store.FinishQuest(q.ID, wiki.Wallet{Tongqian: q.Pay(1)})
	} else {
		err = m.store.FailQuest(q.ID, m.now())
	}
	if err != nil {
		m.loadErr = err
	}
	return m
}
