package roam

import (
	"fmt"
	"math/rand"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/therealsiege/tuilaoshi/internal/audio"
	"github.com/therealsiege/tuilaoshi/internal/journey"
	"github.com/therealsiege/tuilaoshi/internal/ui"
)

// Overheard demons (妖怪密谋): some nights two demons are heard plotting an
// ambush, each in their own voice. The learner answers where or when they
// will strike; answered rightly, the party is forewarned and the next
// danger on the road does half its harm. With audio the Chinese stays
// hidden until answered (listening); without, it is shown (reading).

// plotRoll draws 0-99 against journey.PlotChance; tests fix it.
var plotRoll = func() int { return rand.Intn(100) }

// overheard is a plot being listened to.
type overheard struct {
	journey.Plot
	order []int // the choices as shown: indices into Choices, 0 the right one
	chose int   // the shown choice picked, -1 until answered
	heard bool  // spoken aloud, so the Chinese is hidden until answered
}

// right reports whether the answer given was the right one.
func (p *overheard) right() bool { return p.chose >= 0 && p.order[p.chose] == 0 }

// maybePlot may bring demons plotting at the end of a day without an
// event, unless the party is already forewarned.
func (m Model) maybePlot() Model {
	if m.event != nil || m.trail.Forewarned || len(journey.Plots()) == 0 || plotRoll() >= journey.PlotChance {
		return m
	}
	rng := rand.New(rand.NewSource(rand.Int63()))
	return m.overhear(journey.PickPlot(rng), rng)
}

// overhear opens a plot, its choices shuffled by rng.
func (m Model) overhear(p journey.Plot, rng *rand.Rand) Model {
	m.plot = &overheard{Plot: p, order: rng.Perm(len(p.Choices)), chose: -1, heard: ui.Audible(m.tts)}
	return m
}

// speakPlot speaks the plot's lines, the first demon in the learner's
// Chinese voice and the second in another; nil without audio.
func (m Model) speakPlot() tea.Cmd {
	if m.plot == nil || !ui.Audible(m.tts) {
		return nil
	}
	first := m.plot.Lines[0].Who
	var lines []audio.Line
	for _, l := range m.plot.Lines {
		v := m.tts.ChineseVoice()
		if l.Who != first {
			v = m.tts.OtherChineseVoice()
		}
		lines = append(lines, audio.Line{Text: l.Text, Voice: v})
	}
	return ui.SpeakLinesCmd(m.tts, lines)
}

// Replay speaks the overheard demons again, in their two voices; nil when
// none are being heard (ctrl+r then reads Hear).
func (m Model) Replay() tea.Cmd { return m.speakPlot() }

// updatePlot takes 1-3 for an answer, then enter to walk on.
func (m Model) updatePlot(k tea.KeyMsg) (Model, tea.Cmd) {
	p := m.plot
	if p.chose >= 0 {
		if key.Matches(k, m.gkeys.Enter) || k.Type == tea.KeySpace || key.Matches(k, m.gkeys.Back) {
			m.plot = nil
		}
		return m, nil
	}
	s := k.String()
	if len(s) != 1 || s[0] < '1' || int(s[0]-'1') >= len(p.order) {
		return m, nil
	}
	p.chose = int(s[0] - '1')
	if p.right() {
		m.trail.Forewarned = true
		m.saveTrail()
	}
	return m, nil
}

// plotView shows the demons' lines (hidden while listening), the question
// and its choices; once answered, the English and what came of it.
func (m Model) plotView() string {
	p := m.plot
	answered := p.chose >= 0
	var b strings.Builder
	b.WriteString(ui.SubtitleStyle.Render(fmt.Sprintf("  Night %d on the road · 妖怪密谋", m.trail.Day)) + "\n")
	if !answered {
		b.WriteString(ui.MutedStyle.Render("  The fire burns low. In the dark, two demons are whispering...") + "\n\n")
	}
	for _, l := range p.Lines {
		if p.heard && !answered {
			b.WriteString("  " + ui.WarningStyle.Render(l.Who) + "  " + ui.MutedStyle.Render("……") + "\n")
			continue
		}
		said := strings.TrimPrefix(m.wrapped(ui.ChineseStyle, l.Text, "      "), "    ")
		b.WriteString(strings.Replace(said, "  ", "  "+ui.WarningStyle.Render(l.Who)+" ", 1))
		if answered {
			b.WriteString(m.wrapped(ui.MutedStyle, l.English, "      "))
		}
	}
	if !answered {
		b.WriteString("\n")
	}
	b.WriteString(m.wrapped(ui.ChineseStyle, p.Question, "  "))
	if answered {
		b.WriteString(m.wrapped(ui.MutedStyle, p.QuestionEnglish, "  "))
	}
	for i, c := range p.order {
		line := fmt.Sprintf("%d  %s", i+1, p.Choices[c])
		switch {
		case answered && c == 0:
			b.WriteString(m.wrapped(ui.SuccessStyle, line+"  ✓", "  "))
		case i == p.chose:
			b.WriteString(m.wrapped(ui.ErrorStyle, line+"  ◀", "  "))
		case answered:
			b.WriteString(m.wrapped(ui.MutedStyle, line, "  "))
		default:
			b.WriteString(m.wrapped(ui.UnselectedStyle, line, "  "))
		}
		if answered && (c == 0 || i == p.chose) {
			b.WriteString(m.wrapped(ui.MutedStyle, p.EnglishChoices[c], "     "))
		}
	}
	if !answered {
		help := "  1-3 answer"
		if p.heard {
			help += " · ctrl+r hear them again"
		}
		b.WriteString("\n" + ui.HelpStyle.Render(help))
		return b.String()
	}
	if p.right() {
		b.WriteString("\n" + m.wrapped(ui.ChineseStyle, p.Warning, "  "))
		b.WriteString(m.wrapped(ui.MutedStyle, p.WarningEnglish, "  "))
		b.WriteString("  " + ui.SuccessStyle.Render("Forewarned: the next danger on the road does half its harm.") + "\n")
	} else {
		b.WriteString("\n  " + ui.WarningStyle.Render("The demons slip away into the night. The road ahead is as it was.") + "\n")
	}
	b.WriteString("\n" + ui.HelpStyle.Render("  enter walk on"))
	return b.String()
}

// plotText is the demons' Chinese, line by line, for ctrl+r.
func (p *overheard) plotText() string {
	var lines []string
	for _, l := range p.Lines {
		lines = append(lines, l.Text)
	}
	return strings.Join(lines, "\n")
}
