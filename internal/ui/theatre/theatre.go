// Package theatre is the episodes screen (剧场): the learner's own episodes,
// each opening at its point on the road. An opened episode can be read and
// heard line by line, its scenes answered for comprehension, its lines
// drilled by ear, and talked over with a companion; answering every scene
// pays a finishing reward. Episodes' text stays in wiki/private/episodes.
package theatre

import (
	"errors"
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/therealsiege/tuilaoshi/internal/episode"
	"github.com/therealsiege/tuilaoshi/internal/journey"
	"github.com/therealsiege/tuilaoshi/internal/model"
	"github.com/therealsiege/tuilaoshi/internal/ui"
	"github.com/therealsiege/tuilaoshi/internal/ui/input"
	"github.com/therealsiege/tuilaoshi/internal/wiki"
)

// ReadMsg asks the app to open an episode in the reader.
type ReadMsg struct{ Episode model.Episode }

// TalkMsg asks the app to talk the episode over with a companion.
type TalkMsg struct {
	Title string
	Lines []model.DialogueLine
	Words []model.Card
}

type screen int

const (
	screenList    screen = iota
	screenEpisode        // one episode's menu
	screenScene          // a scene's question
	screenListen         // a listening drill (see listen.go)
	screenShadow         // shadowing a line (see shadow.go)
)

// Model is the episodes screen.
type Model struct {
	store   *wiki.Store
	gkeys   ui.GlobalKeys
	eps     []episode.Episode
	current int // the stage being worked on; 0 once all are cleared
	cards   []model.Card
	states  map[string]wiki.QuestState
	claimed map[string]bool

	screen   screen
	cursor   int // on the list
	ep       int // the episode open
	scenes   []episode.Scene
	scene    int // the scene open
	answered int // -1 until the open scene is answered
	now      func() time.Time

	listen listenState // the listening drill (see listen.go)
	shadow shadowState // shadowing (see shadow.go)

	note          string
	width, height int
	loadErr       error
}

// New creates the episodes screen; dict powers typed listening answers.
func New(store *wiki.Store, dict *input.Dict) Model {
	in := input.NewWithDict(dict)
	in.SetPlaceholder("type what you heard: pinyin, then pick characters")
	in.SetCharLimit(60)
	return Model{store: store, gkeys: ui.DefaultGlobalKeys(), now: time.Now, listen: listenState{input: in}}
}

func (m Model) Init() tea.Cmd { return nil }

// SetSize updates the terminal dimensions.
func (m Model) SetSize(w, h int) Model {
	m.width, m.height = w, h
	return m
}

// Busy reports whether an episode is open, so esc steps back within the
// screen rather than leaving it.
func (m Model) Busy() bool { return m.screen != screenList }

// Typing reports whether a listening answer has the keyboard.
func (m Model) Typing() bool { return m.screen == screenListen && m.listen.typing }

// Refresh reloads the episodes, how far the road has come and what is done.
func (m Model) Refresh() Model {
	eps, epErr := episode.Load(m.store)
	pages, pagesErr := m.store.ListPages()
	cleared, clearedErr := m.store.ClearedStages()
	cards, cardsErr := m.store.DrillableCards()
	states, questErr := m.store.Quests(m.now())
	col, colErr := m.store.Collection()
	m.loadErr = errors.Join(epErr, pagesErr, clearedErr, cardsErr, questErr, colErr)
	m.eps, m.cards, m.states = eps, cards, states
	m.claimed = map[string]bool{}
	for _, id := range col.Claimed {
		m.claimed[id] = true
	}
	m.current = 0
	if s := journey.CurrentStage(journey.Stages(journey.CountProgress(pages), cleared)); s != nil {
		m.current = s.Number
	}
	m.cursor = min(m.cursor, max(len(m.eps)-1, 0))
	return m
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if s, ok := msg.(shadowMsg); ok {
		return m.onShadow(s), nil
	}
	if done, ok := msg.(clipDoneMsg); ok {
		if done.err != nil {
			m.note = done.err.Error()
		}
		return m, nil
	}
	k, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	back := key.Matches(k, m.gkeys.Back)
	switch m.screen {
	case screenShadow:
		return m.updateShadow(k)
	case screenListen:
		return m.updateListen(k)
	case screenScene:
		if back {
			m.screen, m.note = screenEpisode, ""
			return m.Refresh(), nil
		}
		return m.answer(k), nil
	case screenEpisode:
		return m.updateEpisode(k)
	}
	switch {
	case back:
		return m, ui.SwitchTo(ui.ViewDashboard)
	case key.Matches(k, m.gkeys.Up):
		m.cursor, m.note = max(m.cursor-1, 0), ""
	case key.Matches(k, m.gkeys.Down):
		m.cursor, m.note = min(m.cursor+1, len(m.eps)-1), ""
	case key.Matches(k, m.gkeys.Enter):
		if len(m.eps) == 0 {
			return m, nil
		}
		e := m.eps[m.cursor]
		if !e.Open(m.current) {
			m.note = fmt.Sprintf("Episode %d opens at tribulation %d.", e.Number, e.Stage)
			return m, nil
		}
		m.ep, m.scenes, m.scene, m.screen, m.note = m.cursor, e.Scenes(), 0, screenEpisode, ""
	}
	return m, nil
}

// updateEpisode moves through an episode's scenes and starts its activities.
func (m Model) updateEpisode(k tea.KeyMsg) (Model, tea.Cmd) {
	e := m.eps[m.ep]
	switch {
	case key.Matches(k, m.gkeys.Back):
		m.screen, m.note = screenList, ""
	case key.Matches(k, m.gkeys.Up):
		m.scene, m.note = max(m.scene-1, 0), ""
	case key.Matches(k, m.gkeys.Down):
		m.scene, m.note = min(m.scene+1, len(m.scenes)-1), ""
	case key.Matches(k, m.gkeys.Enter):
		return m.openScene(), nil
	case k.String() == "r":
		ep := e.Episode
		return m, func() tea.Msg { return ReadMsg{Episode: ep} }
	case k.String() == "l":
		return m.startListen()
	case k.String() == "s":
		return m.startShadow()
	case k.String() == "t":
		title, lines, words := e.Title, e.Lines, e.Words(m.cards)
		return m, func() tea.Msg { return TalkMsg{Title: title, Lines: lines, Words: words} }
	}
	return m, nil
}

// openScene shows a scene's question, unless done or closed for today.
func (m Model) openScene() Model {
	if len(m.scenes) == 0 {
		return m
	}
	id := m.eps[m.ep].QuestID(m.scenes[m.scene].Index)
	switch m.states[id] {
	case wiki.QuestDone:
		m.note = "Answered. Read it again with r."
		return m
	case wiki.QuestLocked:
		m.note = "Not today: watch or read the scene again, and try tomorrow."
		return m
	}
	m.screen, m.answered, m.note = screenScene, -1, ""
	return m
}

// answer takes 1-4; a right answer pays once, a wrong one locks the scene
// until tomorrow. Finishing every scene pays the episode's reward.
func (m Model) answer(k tea.KeyMsg) Model {
	s := m.scenes[m.scene]
	if m.answered >= 0 {
		if key.Matches(k, m.gkeys.Enter) || k.Type == tea.KeySpace {
			m.screen = screenEpisode
			return m.Refresh()
		}
		return m
	}
	pressed := k.String()
	if len(pressed) != 1 || pressed[0] < '1' || int(pressed[0]-'1') >= len(s.Choices) {
		return m
	}
	m.answered = int(pressed[0] - '1')
	e := m.eps[m.ep]
	id := e.QuestID(s.Index)
	if m.answered != s.Right {
		if err := m.store.FailQuest(id, m.now()); err != nil {
			m.loadErr = err
		}
		return m
	}
	if err := m.store.FinishQuest(id, wiki.Wallet{Tongqian: episode.ScenePay}); err != nil {
		m.loadErr = err
		return m
	}
	m.states[id] = wiki.QuestDone
	if m.done(e) == len(m.scenes) {
		paid, err := m.store.Claim(e.FinishID(), wiki.Wallet{Tongqian: episode.FinishCopper, Yuanbao: episode.FinishIngots})
		if err != nil {
			m.loadErr = err
		} else if paid {
			m.note = fmt.Sprintf("Episode finished! +%d 铜钱 +%d 元宝", episode.FinishCopper, episode.FinishIngots)
		}
	}
	return m
}

// done counts an episode's scenes answered.
func (m Model) done(e episode.Episode) int {
	n := 0
	for _, s := range e.Scenes() {
		if m.states[e.QuestID(s.Index)] == wiki.QuestDone {
			n++
		}
	}
	return n
}
