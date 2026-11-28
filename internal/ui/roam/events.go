package roam

import (
	"fmt"
	"math/rand"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/therealsiege/tuilaoshi/internal/journey"
	"github.com/therealsiege/tuilaoshi/internal/road"
	"github.com/therealsiege/tuilaoshi/internal/ui"
)

// A trail event takes over the map until met: its Chinese text and
// choices; once chosen, the result, the English and what it changed (the
// road package's events.go has the rules).

// eventRoll draws 0-99 against journey.EventChance; tests fix it.
var eventRoll = func() int { return rand.Intn(100) }

// trailEvent is an event being met.
type trailEvent struct {
	journey.Event
	chose  int    // -1 until chosen
	effect string // what the choice changed, in short
	cross  bool   // the region's obstacle: crossing it once chosen
}

// updateEvent takes 1-3 for a choice, then enter to walk on.
func (m Model) updateEvent(k tea.KeyMsg) (Model, tea.Cmd) {
	e := m.event
	if e.chose >= 0 {
		if key.Matches(k, m.gkeys.Enter) || k.Type == tea.KeySpace || key.Matches(k, m.gkeys.Back) {
			m.event = nil
			if e.cross {
				m = m.crossed()
			}
		}
		return m, nil
	}
	s := k.String()
	if len(s) != 1 || s[0] < '1' || int(s[0]-'1') >= len(e.Choices) {
		return m, nil
	}
	e.chose = int(s[0] - '1')
	w := m.walk()
	effect, err := w.Choose(e.Event, e.chose) // see the road package for the rules
	m.trail, e.effect = w.Trail, effect
	if err != nil {
		m.loadErr = err
	}
	return m.refreshPurse(), nil
}

// refreshPurse rereads the purse after an event paid or charged.
func (m Model) refreshPurse() Model {
	if w, err := m.store.Wallet(); err == nil {
		m.wallet = w
	}
	return m
}

func (m Model) eventView() string {
	e := m.event
	var b strings.Builder
	b.WriteString(ui.SubtitleStyle.Render(fmt.Sprintf("  Day %d on the road", m.trail.Day)) + "\n\n")
	b.WriteString(m.wrapped(ui.ChineseStyle, e.Text, "  ") + "\n")
	if e.chose >= 0 {
		b.WriteString(m.wrapped(ui.MutedStyle, e.English, "  ") + "\n")
	}
	for i, c := range e.Choices {
		line := fmt.Sprintf("%d  %s", i+1, c.Text)
		switch {
		case i == e.chose:
			b.WriteString(m.wrapped(ui.WarningStyle, line+"  ◀", "  "))
		case e.chose >= 0:
			b.WriteString(m.wrapped(ui.MutedStyle, line, "  "))
		default:
			b.WriteString(m.wrapped(ui.UnselectedStyle, line, "  "))
		}
		if e.chose >= 0 {
			b.WriteString(m.wrapped(ui.MutedStyle, c.English, "     "))
		}
	}
	if e.chose < 0 {
		if m.trail.Forewarned {
			b.WriteString("\n  " + ui.SuccessStyle.Render("Forewarned: whatever harm comes is halved."))
		}
		b.WriteString("\n" + ui.HelpStyle.Render("  1-3 choose"))
		return b.String()
	}
	c := e.Choices[e.chose]
	b.WriteString("\n" + m.wrapped(ui.ChineseStyle, c.Result, "  "))
	b.WriteString(m.wrapped(ui.MutedStyle, c.ResultEnglish, "  "))
	b.WriteString(m.wrapped(ui.WarningStyle, e.effect, "  "))
	b.WriteString("\n" + ui.HelpStyle.Render("  enter walk on"))
	return b.String()
}

// crossed puts the party on the far side of the obstacle, for good.
func (m Model) crossed() Model {
	w := m.walk()
	if err := w.Cross(); err != nil {
		m.loadErr = err
	}
	m.pos, m.trail = w.Pos, w.Trail
	return m
}

// hasCrossed reports whether this region's obstacle is behind the party.
func (m Model) hasCrossed() bool { return m.walk().HasCrossed() }

// Hear is the Chinese of the event or crossing open, to read aloud: its
// text, and once chosen what came of it; or the overheard demons' lines.
func (m Model) Hear() string {
	e := m.event
	if e == nil {
		if m.plot != nil {
			return road.PlotText(m.plot.Plot)
		}
		return ""
	}
	if e.chose >= 0 {
		return e.Text + "\n" + e.Choices[e.chose].Result
	}
	return e.Text
}

// wrapped renders text in style wrapped to the window (80 columns at most
// is plenty), each line indented by pad. Beside the sidebar it wraps to
// the space left of it.
func (m Model) wrapped(style lipgloss.Style, text, pad string) string {
	width := m.room() - lipgloss.Width(pad)
	var b strings.Builder
	for _, l := range ui.WrapDisplay(text, width) {
		b.WriteString(pad + style.Render(l) + "\n")
	}
	return b.String()
}
