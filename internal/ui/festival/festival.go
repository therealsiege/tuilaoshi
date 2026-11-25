// Package festival is the festivals screen (节日): the Chinese festival on
// now, or the next one with a countdown. While one is on, its story can be
// read (e for English), its two reading questions answered (pay once a
// year; a miss locks until tomorrow; both answered pays a silver ingot),
// and its words join the reviews.
package festival

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

// Rewards: copper per question, and the festival's reward for both.
const (
	QuestionPay    = 20
	FestivalPay    = 30
	FestivalIngots = 1
)

// Model is the festivals screen.
type Model struct {
	store *wiki.Store
	gkeys ui.GlobalKeys
	now   func() time.Time

	occ     journey.Occasion
	on, ok  bool
	states  map[string]wiki.QuestState
	claimed bool

	english  bool
	asking   int // the question open, or -1
	choices  []string
	right    int
	answered int // -1 until answered
	added    int // festival words added to the reviews on this visit

	note          string
	width, height int
	loadErr       error
}

// New creates the festivals screen.
func New(store *wiki.Store) Model {
	return Model{store: store, gkeys: ui.DefaultGlobalKeys(), now: time.Now, asking: -1}
}

func (m Model) Init() tea.Cmd { return nil }

// SetSize updates the terminal dimensions.
func (m Model) SetSize(w, h int) Model {
	m.width, m.height = w, h
	return m
}

// Busy reports whether a question is open, so esc closes it first.
func (m Model) Busy() bool { return m.asking >= 0 }

// Refresh finds the festival on now (or next), and while one is on adds
// its words to the reviews.
func (m Model) Refresh() Model {
	m.occ, m.on, m.ok = journey.FestivalNow(m.now())
	states, questErr := m.store.Quests(m.now())
	col, colErr := m.store.Collection()
	m.loadErr = errors.Join(questErr, colErr)
	m.states, m.asking, m.note = states, -1, ""
	m.claimed = false
	for _, id := range col.Claimed {
		if id == "festival:"+m.occ.Year() {
			m.claimed = true
		}
	}
	if m.on {
		var words []wiki.Word
		for _, w := range m.occ.Words {
			words = append(words, wiki.Word{Chinese: w.Chinese, Pinyin: w.Pinyin, Meaning: w.Meaning})
		}
		n, err := m.store.AddFestivalWords(words, m.now())
		m.added = n
		m.loadErr = errors.Join(m.loadErr, err)
	}
	return m
}

// questID names a question of this year's festival.
func (m Model) questID(i int) string { return fmt.Sprintf("festival:%s:q%d", m.occ.Year(), i) }

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	k, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	if m.asking >= 0 {
		return m.updateQuestion(k), nil
	}
	switch {
	case key.Matches(k, m.gkeys.Back):
		return m, ui.SwitchTo(ui.ViewDashboard)
	case k.String() == "e":
		m.english = !m.english
	case m.on && len(k.String()) == 1 && k.String()[0] >= '1' && int(k.String()[0]-'1') < len(m.occ.Questions):
		return m.open(int(k.String()[0] - '1')), nil
	}
	return m, nil
}

// open shows a question unless answered or closed for today.
func (m Model) open(i int) Model {
	switch m.states[m.questID(i)] {
	case wiki.QuestDone:
		m.note = "Answered already."
		return m
	case wiki.QuestLocked:
		m.note = "Not today: read the story again and try tomorrow."
		return m
	}
	q := m.occ.Questions[i]
	rng := rand.New(rand.NewSource(m.now().UnixNano()))
	order := rng.Perm(len(q.Choices))
	m.choices = make([]string, len(q.Choices))
	for j, o := range order {
		m.choices[j] = q.Choices[o]
		if o == 0 {
			m.right = j
		}
	}
	m.asking, m.answered, m.note = i, -1, ""
	return m
}

func (m Model) updateQuestion(k tea.KeyMsg) Model {
	leave := key.Matches(k, m.gkeys.Enter) || k.Type == tea.KeySpace || key.Matches(k, m.gkeys.Back)
	if m.answered >= 0 || key.Matches(k, m.gkeys.Back) {
		if !leave {
			return m
		}
		note := m.note
		m = m.Refresh()
		m.note = note
		return m
	}
	s := k.String()
	if len(s) != 1 || s[0] < '1' || int(s[0]-'1') >= len(m.choices) {
		return m
	}
	m.answered = int(s[0] - '1')
	id := m.questID(m.asking)
	if m.answered != m.right {
		if err := m.store.FailQuest(id, m.now()); err != nil {
			m.loadErr = err
		}
		return m
	}
	if err := m.store.FinishQuest(id, wiki.Wallet{Tongqian: QuestionPay}); err != nil {
		m.loadErr = err
		return m
	}
	m.states[id] = wiki.QuestDone
	for i := range m.occ.Questions {
		if m.states[m.questID(i)] != wiki.QuestDone {
			return m
		}
	}
	if paid, err := m.store.Claim("festival:"+m.occ.Year(), wiki.Wallet{Tongqian: FestivalPay, Yuanbao: FestivalIngots}); err != nil {
		m.loadErr = err
	} else if paid {
		m.note = fmt.Sprintf("%s快乐! The festival's reward: +%d 铜钱 +%d 元宝", m.occ.Name, FestivalPay, FestivalIngots)
	}
	return m
}

// Hear is the festival's Chinese, to read aloud: the question open, or its
// story.
func (m Model) Hear() string {
	if !m.ok {
		return ""
	}
	if m.asking >= 0 {
		return m.occ.Questions[m.asking].Question
	}
	return m.occ.Text
}
