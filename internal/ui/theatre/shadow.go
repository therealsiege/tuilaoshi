package theatre

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/therealsiege/tuilaoshi/internal/audio"
	"github.com/therealsiege/tuilaoshi/internal/episode"
	"github.com/therealsiege/tuilaoshi/internal/ui"
)

// Shadowing (跟读): hear a line of the episode, say it back, and see which
// characters whisper.cpp heard, all on this Mac. s in an episode starts it;
// space records, tab hears the line again, n moves to another line.

// shadowListen records the learner and returns what was heard; tests
// replace it.
var shadowListen = audio.Shadow

// shadowReady says what shadowing still needs; tests replace it.
var shadowReady = audio.ShadowReady

type shadowPhase int

const (
	shadowReadyToSay shadowPhase = iota
	shadowRecording
	shadowHeard
)

type shadowState struct {
	line   int
	phase  shadowPhase
	heard  string
	result episode.Shadowing
	tries  int // lines shadowed this visit
	best   int // the best match this visit, in percent
}

type shadowMsg struct {
	heard string
	err   error
}

// recordFor is how long to record for a line: its own length and a
// little more, or by its characters when the episode has no audio.
func recordFor(start, end time.Duration, chars int) time.Duration {
	d := time.Duration(chars)*350*time.Millisecond + 1500*time.Millisecond
	if end > start {
		d = (end-start)*3/2 + 1500*time.Millisecond
	}
	return min(d, 12*time.Second)
}

// startShadow begins shadowing the open episode, if this Mac can.
func (m Model) startShadow() (Model, tea.Cmd) {
	if why := shadowReady(); why != "" {
		m.note = why
		return m, nil
	}
	if len(m.eps[m.ep].ShadowLines()) == 0 {
		m.note = "This episode has no short lines to shadow."
		return m, nil
	}
	m.shadow = shadowState{}
	m.screen = screenShadow
	return m.nextShadow()
}

// nextShadow picks another line and plays it.
func (m Model) nextShadow() (Model, tea.Cmd) {
	lines := m.eps[m.ep].ShadowLines()
	m.shadow.line = lines[rand.New(rand.NewSource(m.now().UnixNano())).Intn(len(lines))]
	m.shadow.phase, m.shadow.heard, m.note = shadowReadyToSay, "", ""
	return m, m.playShadow()
}

// playShadow plays the line being shadowed, when the episode has audio.
func (m Model) playShadow() tea.Cmd {
	e := m.eps[m.ep]
	if e.Audio == "" {
		return nil
	}
	l := e.Lines[m.shadow.line]
	path := e.Audio
	return func() tea.Msg { return clipDoneMsg{playClip(path, l.Start, l.End)} }
}

func (m Model) updateShadow(k tea.KeyMsg) (Model, tea.Cmd) {
	s := &m.shadow
	switch {
	case k.Type == tea.KeyEsc:
		m.screen, m.note = screenEpisode, ""
		if s.tries > 0 {
			m.note = fmt.Sprintf("Shadowed %d %s; best %d%%.", s.tries, plural(s.tries, "line", "lines"), s.best)
		}
		return m, nil
	case s.phase == shadowRecording:
		return m, nil
	case k.Type == tea.KeyTab:
		return m, m.playShadow()
	case k.String() == "n":
		return m.nextShadow()
	case k.Type == tea.KeySpace || k.Type == tea.KeyEnter:
		l := m.eps[m.ep].Lines[s.line]
		d := recordFor(l.Start, l.End, len([]rune(l.Chinese)))
		s.phase = shadowRecording
		listen := shadowListen
		return m, func() tea.Msg {
			heard, err := listen(d)
			return shadowMsg{heard, err}
		}
	}
	return m, nil
}

// onShadow scores what was heard against the line.
func (m Model) onShadow(msg shadowMsg) Model {
	s := &m.shadow
	if msg.err != nil {
		s.phase, m.note = shadowReadyToSay, msg.err.Error()
		return m
	}
	s.phase, s.heard = shadowHeard, msg.heard
	s.result = episode.CompareShadow(m.eps[m.ep].Lines[s.line].Chinese, msg.heard)
	s.tries++
	s.best = max(s.best, s.result.Percent)
	return m
}

func (m Model) shadowView() string {
	e, s := m.eps[m.ep], m.shadow
	l := e.Lines[s.line]
	var b strings.Builder
	b.WriteString(ui.TitleStyle.MarginBottom(0).Render("跟读 Shadowing · "+e.Title) + "\n\n")
	if s.phase != shadowHeard {
		if l.Speaker != "" {
			b.WriteString(ui.MutedStyle.Render("  "+l.Speaker+"：") + "\n")
		}
		b.WriteString("  " + ui.ChineseStyle.Render(l.Chinese) + "\n\n")
	} else {
		var line strings.Builder
		for _, c := range s.result.Chars {
			if c.Heard {
				line.WriteString(ui.SuccessStyle.Render(string(c.Char)))
			} else {
				line.WriteString(ui.ErrorStyle.Underline(true).Render(string(c.Char)))
			}
		}
		b.WriteString("  " + line.String() + "  " + ui.WarningStyle.Render(fmt.Sprintf("%d%%", s.result.Percent)) + "\n")
		heard := s.heard
		if heard == "" {
			heard = "（没听到）"
		}
		b.WriteString(ui.MutedStyle.Render("  heard: ") + heard + "\n")
		if l.English != "" {
			b.WriteString(ui.MutedStyle.Render("  "+l.English) + "\n")
		}
		b.WriteString("\n")
	}
	switch s.phase {
	case shadowRecording:
		b.WriteString(ui.WarningStyle.Render("  ● Listening… say the line now.") + "\n")
	case shadowHeard:
		b.WriteString(ui.MutedStyle.Render("  Characters underlined in red weren't heard.") + "\n")
		b.WriteString("\n" + ui.HelpStyle.Render("  space try again · tab hear it · n another line · esc back"))
	default:
		b.WriteString("\n" + ui.HelpStyle.Render("  space say it · tab hear it again · n another line · esc back"))
	}
	if m.note != "" {
		b.WriteString("\n" + ui.WarningStyle.Render("  "+m.note))
	}
	return b.String()
}

func plural(n int, one, many string) string {
	if n == 1 {
		return one
	}
	return many
}
