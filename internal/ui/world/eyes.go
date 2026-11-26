package world

import (
	"context"
	"fmt"
	"math/rand"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/therealsiege/tuilaoshi/internal/claude"
	"github.com/therealsiege/tuilaoshi/internal/journey"
	"github.com/therealsiege/tuilaoshi/internal/ui/roam"
	"github.com/therealsiege/tuilaoshi/internal/wiki"
)

// 火眼金睛 Fiery Eyes: a strange rumour in town. Three villagers are
// questioned in Chinese, each played by Claude from a hidden story, and
// one is a demon in disguise whose story can't be true. Accusing the demon
// unmasks it: a bounty, then a fight with one of the region's strongest
// soldiers (or, where none serve, a demon core it drops as it flees).
// Accusing an innocent lets it slip away and dampens the party's spirits.
// Either way the town's rumour is settled for the day.

// Talker answers as a character: a system prompt, the talk so far, and the
// new message (claude.Client.ChatAs).
type Talker func(ctx context.Context, system string, history []claude.ChatMessage, message string) (string, error)

// eyesTurns is how many past messages per villager are sent with a question.
const eyesTurns = 8

// exchange is one question put to a villager and the answer.
type exchange struct {
	question, answer, english string
}

// eyes is the state of a town's mystery.
type eyes struct {
	talker Talker      // nil without Claude
	keyed  func() bool // whether Claude has a key; nil counts as yes

	town      string // the town the mystery is in, empty when none is open
	mystery   journey.Mystery
	demon     int
	talks     [][]exchange // per villager
	who       int          // the villager being questioned
	asked     int          // questions used
	asking    bool         // the question box is open
	waiting   bool         // a villager is answering
	accusing  bool         // choosing whom to accuse
	accused   int          // -1 until someone is accused
	english   bool         // show the English of facts and answers
	foe       *journey.Soldier
	materials bool // the unmasked demon fled and dropped a core
}

// eyesReplyMsg is a villager's answer.
type eyesReplyMsg struct {
	who   int
	reply string
	err   error
}

// openEyes follows the town's rumour: resumes the mystery under way here,
// or starts one unless today's is already settled.
func (m Model) openEyes() (Model, tea.Cmd) {
	t := m.towns[m.town]
	if m.eyes.talker == nil || (m.eyes.keyed != nil && !m.eyes.keyed()) {
		m.note = "Questioning villagers " + claude.NeedsKey + "."
		return m, nil
	}
	settled, err := m.store.MysterySettled(t.ID, m.now())
	if err != nil {
		m.loadErr = err
		return m, nil
	}
	if settled {
		m.note = "The rumour here is settled for today: come back tomorrow."
		return m, nil
	}
	m.screen, m.note = screenEyes, ""
	if m.eyes.town == t.ID && m.eyes.accused < 0 {
		return m, nil
	}
	rng := rand.New(rand.NewSource(m.now().UnixNano()))
	e := eyes{talker: m.eyes.talker, keyed: m.eyes.keyed, town: t.ID, accused: -1}
	e.mystery, e.demon = journey.PickMystery(rng)
	e.talks = make([][]exchange, len(e.mystery.Villagers))
	m.eyes = e
	return m, nil
}

// updateEyes handles the mystery's keys: 1-3 question a villager, a
// accuse, e the English, esc back out a step.
func (m Model) updateEyes(k tea.KeyMsg, back bool) (Model, tea.Cmd) {
	e := &m.eyes
	switch {
	case e.accused >= 0:
		if k.Type == tea.KeyEnter || k.Type == tea.KeySpace {
			return m.leaveEyes()
		}
		return m, nil
	case e.asking:
		return m.updateAsking(k, back)
	case e.accusing:
		if back {
			e.accusing = false
			return m, nil
		}
		if n, ok := villagerKey(k, len(e.mystery.Villagers)); ok {
			return m.accuse(n), nil
		}
		return m, nil
	case back:
		m.screen, m.note = screenTown, ""
		return m, nil
	}
	switch k.String() {
	case "e":
		e.english = !e.english
	case "a":
		e.accusing, m.note = true, ""
	default:
		if n, ok := villagerKey(k, len(e.mystery.Villagers)); ok {
			e.who, m.note = n, ""
			if e.asked >= journey.MysteryQuestions {
				m.note = "No questions left: time to accuse someone (a)."
				return m, nil
			}
			e.asking = true
			m.input.Reset()
			return m, m.input.Focus()
		}
	}
	return m, nil
}

// villagerKey reads 1-n as a villager's index.
func villagerKey(k tea.KeyMsg, n int) (int, bool) {
	s := k.String()
	if len(s) != 1 || s[0] < '1' || int(s[0]-'1') >= n {
		return 0, false
	}
	return int(s[0] - '1'), true
}

// updateAsking types a question; enter asks it, esc closes the box.
func (m Model) updateAsking(k tea.KeyMsg, back bool) (Model, tea.Cmd) {
	e := &m.eyes
	switch {
	case back && !m.input.Composing():
		e.asking = false
		m.input.Blur()
		return m, nil
	case k.Type == tea.KeyEnter && !m.input.Composing():
		return m.ask()
	}
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(k)
	return m, cmd
}

// ask sends the typed question to the villager being questioned.
func (m Model) ask() (Model, tea.Cmd) {
	e := &m.eyes
	q := strings.TrimSpace(m.input.Value())
	if q == "" || e.waiting {
		return m, nil
	}
	e.asked++
	e.waiting, m.note = true, ""
	e.talks[e.who] = append(e.talks[e.who], exchange{question: q})
	var history []claude.ChatMessage
	for _, x := range e.talks[e.who] {
		if x.answer != "" {
			history = append(history, claude.ChatMessage{Role: "user", Content: x.question}, claude.ChatMessage{Role: "assistant", Content: x.answer + "\nEN: " + x.english})
		}
	}
	history = history[max(len(history)-eyesTurns, 0):]
	who, system, talk := e.who, e.mystery.VillagerPrompt(e.who, e.demon), e.talker
	m.input.Reset()
	return m, func() tea.Msg {
		reply, err := talk(context.Background(), system, history, q)
		return eyesReplyMsg{who, reply, err}
	}
}

// onEyesReply records a villager's answer; a failed call gives the
// question back.
func (m Model) onEyesReply(r eyesReplyMsg) Model {
	e := &m.eyes
	e.waiting = false
	talks := e.talks[r.who]
	if len(talks) == 0 {
		return m
	}
	last := &talks[len(talks)-1]
	if r.err != nil {
		e.talks[r.who] = talks[:len(talks)-1]
		e.asked--
		m.note = "Couldn't reach Claude: " + r.err.Error()
		return m
	}
	last.answer, last.english = journey.SplitReply(r.reply)
	if e.asked >= journey.MysteryQuestions {
		e.asking = false
		m.input.Blur()
		m.note = "That was your last question: accuse someone (a)."
	}
	return m
}

// accuse names villager n the demon and settles the town's rumour today.
func (m Model) accuse(n int) Model {
	e := &m.eyes
	e.accused, e.accusing = n, false
	pay := wiki.Wallet{}
	if n == e.demon {
		pay.Tongqian = journey.MysteryBounty
		t := m.towns[m.town]
		if at := journey.SoldiersAt(t.Stage); len(at) > 0 {
			foe := at[len(at)-1] // the strongest serving here
			e.foe = &foe
		} else if err := m.store.AddMaterials([]string{"core"}); err != nil {
			m.loadErr = err
		} else {
			e.materials = true
		}
	} else {
		m.withTrail(func(t *journey.Trail, members []string) {
			for _, who := range members {
				t.Morale[who] = max(t.Morale[who]-journey.MysteryMorale, 0)
			}
		})
	}
	if err := m.store.SettleMystery(e.town, m.now(), pay); err != nil {
		m.loadErr = err
	}
	return m
}

// leaveEyes closes a settled mystery: into the fight with an unmasked
// demon, otherwise back to town.
func (m Model) leaveEyes() (Model, tea.Cmd) {
	foe := m.eyes.foe
	m.eyes = eyes{talker: m.eyes.talker, keyed: m.eyes.keyed}
	m.screen = screenTown
	m = m.Refresh()
	if foe != nil {
		s := *foe
		return m, func() tea.Msg { return roam.EncounterMsg{Soldier: s} }
	}
	return m, nil
}

// eyesHear is the Chinese to read aloud: the last answer, or the rumour.
func (m Model) eyesHear() string {
	e := m.eyes
	if talks := e.talks[e.who]; len(talks) > 0 && talks[len(talks)-1].answer != "" {
		return talks[len(talks)-1].answer
	}
	return e.mystery.Situation
}

// eyesLine is the rumour's line in the town view.
func eyesLine() string {
	return fmt.Sprintf("火眼金睛 A strange rumour: one of three villagers is a demon. Question them and unmask it for %d 铜钱 (f)", journey.MysteryBounty)
}
