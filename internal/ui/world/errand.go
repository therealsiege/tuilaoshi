package world

import (
	"context"
	"fmt"
	"slices"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/therealsiege/tuilaoshi/internal/claude"
	"github.com/therealsiege/tuilaoshi/internal/journey"
	"github.com/therealsiege/tuilaoshi/internal/wiki"
)

// Errands (任务对话): a town offers one a day, a goal and a person to see,
// such as booking a room for five at the inn and asking for hot water.
// Claude plays the person and says, after each reply, which of the goal's
// points the learner has got across; when every point is done the errand
// pays its reward and lifts the party's spirits. After journey.ErrandTurns
// things said without finishing, the person has to see to someone else.

// errandTalk is an errand under way.
type errandTalk struct {
	errand   journey.Errand
	done     map[string]bool // points got across, by id
	turns    int             // things said
	history  []claude.ChatMessage
	lines    []haggleLine
	waiting  bool
	finished bool // every point done and the reward paid
}

// over reports whether the errand has ended, done or out of turns.
func (e *errandTalk) over() bool {
	return e.finished || (e.turns >= journey.ErrandTurns && !e.waiting)
}

type errandMsg struct {
	reply string
	err   error
}

// errandPrompt sets the person's character, the points to listen for and
// the reply format.
const errandPrompt = `You are %s (%s), a %s in %s, a town on the road west in Journey to the West. %s
A travelling monk who is learning Chinese (around HSK3-4) has come to you on an errand. Without telling you outright, they need to get these points across, each with an id:
%s
Stay in character and react naturally to what they say. Don't guess or offer what they want before they say it; if something is unclear, ask about it simply. A point counts once its meaning has been said in understandable Chinese, even with small mistakes.
Answer in one to three short sentences of simple spoken Chinese (HSK3-4), in character.
Then on separate lines, exactly:
EN: <the English of what you said>
DONE: <the ids of every point the traveller has got across so far in the whole conversation, separated by commas, or none>`

// errandSystem is the person's system prompt for an errand in a town.
func errandSystem(e journey.Errand, town string) string {
	var points strings.Builder
	for _, p := range e.Points {
		fmt.Fprintf(&points, "- %s: %s\n", p.ID, p.Listen)
	}
	return fmt.Sprintf(errandPrompt, e.Name, e.Role, e.RoleEnglish, town, e.Persona, strings.TrimRight(points.String(), "\n"))
}

// parseErrand reads the person's reply: the Chinese, its English, and the
// ids of the points done, keeping only the errand's own.
func parseErrand(reply string, e journey.Errand) (zh, en string, done []string) {
	var said []string
	for _, line := range strings.Split(reply, "\n") {
		line = strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(line, "EN:"):
			en = strings.TrimSpace(strings.TrimPrefix(line, "EN:"))
		case strings.HasPrefix(line, "DONE:"):
			for _, id := range strings.FieldsFunc(strings.TrimPrefix(line, "DONE:"), func(r rune) bool { return r == ',' || r == '，' || r == ' ' }) {
				id = strings.ToLower(strings.TrimSpace(id))
				if slices.ContainsFunc(e.Points, func(p journey.ErrandPoint) bool { return p.ID == id }) && !slices.Contains(done, id) {
					done = append(done, id)
				}
			}
		case line != "":
			said = append(said, line)
		}
	}
	return strings.Join(said, " "), en, done
}

// openErrand starts the town's errand for today, unless it was taken.
func (m Model) openErrand() (Model, tea.Cmd) {
	if m.talker == nil {
		m.note = "Errands " + claude.NeedsKey + "."
		return m, nil
	}
	t := m.towns[m.town]
	taken, err := m.store.ErrandTaken(t.ID, m.now())
	if err != nil {
		m.loadErr = err
		return m, nil
	}
	if taken {
		m.note = "You've done this town's errand today: there's another tomorrow."
		return m, nil
	}
	if err := m.store.TakeErrand(t.ID, m.now()); err != nil {
		m.loadErr = err
		return m, nil
	}
	e := journey.ErrandFor(t.ID, m.now())
	m.errand = &errandTalk{
		errand:  e,
		done:    map[string]bool{},
		lines:   []haggleLine{{true, e.Opening, e.OpeningEnglish}},
		history: []claude.ChatMessage{{Role: "user", Content: "（客人走过来）"}, {Role: "assistant", Content: e.Opening}},
	}
	m.screen, m.note = screenErrand, ""
	m.input.Reset()
	return m, m.input.Focus()
}

// updateErrand types what to say; enter says it. Once the errand is over,
// enter goes back to town; esc leaves at any time (the day's errand is
// spent).
func (m Model) updateErrand(k tea.KeyMsg, back bool) (Model, tea.Cmd) {
	e := m.errand
	switch {
	case back && !m.input.Composing():
		m.screen, m.note = screenTown, ""
		m.input.Blur()
		return m.Refresh(), nil
	case e.over():
		if k.Type == tea.KeyEnter || k.Type == tea.KeySpace {
			m.screen, m.note = screenTown, ""
			return m.Refresh(), nil
		}
		return m, nil
	case e.waiting:
		return m, nil
	case k.Type == tea.KeyEnter && !m.input.Composing():
		said := strings.TrimSpace(m.input.Value())
		if said == "" {
			return m, nil
		}
		e.turns++
		e.lines = append(e.lines, haggleLine{false, said, ""})
		e.waiting = true
		m.input.Reset()
		talker, history := m.talker, append([]claude.ChatMessage{}, e.history...)
		system := errandSystem(e.errand, m.towns[m.town].Name)
		e.history = append(e.history, claude.ChatMessage{Role: "user", Content: said})
		return m, func() tea.Msg {
			reply, err := talker(context.Background(), system, history, said)
			return errandMsg{reply, err}
		}
	}
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(k)
	return m, cmd
}

// onErrand takes the person's reply and checks off the points done; when
// all are, it pays the reward once.
func (m Model) onErrand(msg errandMsg) Model {
	e := m.errand
	if e == nil || !e.waiting {
		return m
	}
	e.waiting = false
	if msg.err != nil {
		e.turns-- // a failed call costs nothing
		e.lines = e.lines[:len(e.lines)-1]
		e.history = e.history[:len(e.history)-1]
		m.note = "Couldn't reach Claude: " + msg.err.Error()
		return m
	}
	zh, en, done := parseErrand(msg.reply, e.errand)
	e.lines = append(e.lines, haggleLine{true, zh, en})
	e.history = append(e.history, claude.ChatMessage{Role: "assistant", Content: zh})
	for _, id := range done {
		e.done[id] = true
	}
	switch {
	case e.finished:
	case len(e.done) == len(e.errand.Points):
		return m.finishErrand()
	case e.turns >= journey.ErrandTurns:
		m.input.Blur()
	}
	return m
}

// finishErrand pays the reward and lifts the party's spirits.
func (m Model) finishErrand() Model {
	e := m.errand
	if err := m.store.FinishErrand(m.towns[m.town].ID, e.errand.Goal, m.now(), wiki.Wallet{Tongqian: e.errand.Reward}); err != nil {
		m.loadErr = err
		return m
	}
	e.finished = true
	m.withTrail(func(t *journey.Trail, members []string) {
		t.Apply(journey.Outcome{Morale: journey.ErrandMorale}, members)
	})
	m.input.Blur()
	return m
}

// errandHear is the person's last words, to read aloud.
func (m Model) errandHear() string {
	for i := len(m.errand.lines) - 1; i >= 0; i-- {
		if l := m.errand.lines[i]; l.merchant {
			return l.zh
		}
	}
	return ""
}
