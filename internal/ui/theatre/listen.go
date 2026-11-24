package theatre

import (
	"math/rand"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/therealsiege/tuilaoshi/internal/audio"
	"github.com/therealsiege/tuilaoshi/internal/episode"
	"github.com/therealsiege/tuilaoshi/internal/model"
	"github.com/therealsiege/tuilaoshi/internal/ui"
	"github.com/therealsiege/tuilaoshi/internal/ui/input"
)

// The listening drill plays a line of the episode's own audio: a short line
// is typed back in Chinese, a longer one matched to its meaning. Answers
// are saved as a session when the drill ends, so each line pays a copper
// coin once a day and heals like any other fresh recall.

// clipDoneMsg reports that a clip finished, or why it couldn't play.
type clipDoneMsg struct{ err error }

// playClip plays part of an episode's audio; tests replace it.
var playClip = audio.PlayClip

type listenState struct {
	drill    episode.Drill
	typing   bool // a typed answer has the keyboard
	answered bool
	right    bool
	picked   int
	input    input.Model
	session  model.Session
}

// startListen begins a drill on the open episode.
func (m Model) startListen() (Model, tea.Cmd) {
	m.listen.session = model.Session{StartedAt: m.now(), Type: "listening"}
	m.screen = screenListen
	return m.nextLine()
}

// nextLine picks and plays the next line.
func (m Model) nextLine() (Model, tea.Cmd) {
	e := m.eps[m.ep]
	d, ok := e.NextDrill(rand.New(rand.NewSource(m.now().UnixNano())))
	if !ok {
		m.screen, m.note = screenEpisode, "This episode has no audio to drill (add its mp3 beside the subtitles)."
		return m, nil
	}
	m.listen.drill, m.listen.answered, m.listen.right, m.listen.typing = d, false, false, d.Type
	var cmds []tea.Cmd
	if d.Type {
		m.listen.input.Reset()
		m.listen.input.SetMode(ui.AnswerMode())
		cmds = append(cmds, m.listen.input.Focus())
	}
	return m, tea.Batch(append(cmds, m.play())...)
}

// play plays the current line.
func (m Model) play() tea.Cmd {
	e := m.eps[m.ep]
	l := e.Lines[m.listen.drill.Line]
	path := e.Audio
	return func() tea.Msg { return clipDoneMsg{playClip(path, l.Start, l.End)} }
}

func (m Model) updateListen(k tea.KeyMsg) (Model, tea.Cmd) {
	l := &m.listen
	switch {
	case k.Type == tea.KeyEsc && (!l.typing || !l.input.Composing()):
		return m.endListen(), nil
	case k.Type == tea.KeyTab:
		return m, m.play() // hear it again
	case l.answered:
		if key.Matches(k, m.gkeys.Enter) || k.Type == tea.KeySpace {
			return m.nextLine()
		}
		return m, nil
	case l.typing:
		if k.Type == tea.KeyEnter && !l.input.Composing() {
			typed := strings.TrimSpace(l.input.Value())
			if typed == "" {
				return m, nil
			}
			return m.record(episode.Heard(m.eps[m.ep].Lines[l.drill.Line].Chinese, typed), typed), nil
		}
		var cmd tea.Cmd
		l.input, cmd = l.input.Update(k)
		return m, cmd
	}
	s := k.String()
	if len(s) == 1 && s[0] >= '1' && int(s[0]-'1') < len(l.drill.Choices) {
		l.picked = int(s[0] - '1')
		return m.record(l.picked == l.drill.Right, l.drill.Choices[l.picked]), nil
	}
	return m, nil
}

// record keeps an answer for the session saved when the drill ends.
func (m Model) record(right bool, answer string) Model {
	l := &m.listen
	l.answered, l.right, l.typing = true, right, false
	l.input.Blur()
	l.session.Items = append(l.session.Items, model.SessionItem{CardID: m.eps[m.ep].CardID(l.drill.Line), UserAnswer: answer, Correct: right})
	l.session.CardsTotal++
	if right {
		l.session.CardsCorrect++
	}
	return m
}

// endListen saves the drill's answers and returns to the episode.
func (m Model) endListen() Model {
	m.screen = screenEpisode
	s := m.listen.session
	if len(s.Items) > 0 {
		s.EndedAt = m.now()
		if s.EndedAt.Equal(s.StartedAt) {
			s.EndedAt = s.EndedAt.Add(time.Second)
		}
		if err := m.store.WriteSession(s); err != nil {
			m.loadErr = err
		} else {
			m.note = "Listening saved: each line you caught pays a coin once today."
		}
	}
	m.listen.session = model.Session{}
	return m.Refresh()
}
