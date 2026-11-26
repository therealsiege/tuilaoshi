package reader

import (
	"fmt"
	"math/rand"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/therealsiege/tuilaoshi/internal/ui"
	"github.com/therealsiege/tuilaoshi/internal/wiki"
)

// Hearing a chapter first (听书): on the shelf, l on a told chapter hides
// its text; ctrl+r reads it aloud, and three questions about it are
// answered from what was heard. Then the text opens for checking.

// listening is a chapter being heard.
type listening struct {
	chapter wiki.StoryChapter
	q       int     // the question being asked
	order   [][]int // each question's choices as shown: indices, 0 the right one
	picked  int     // the choice picked for the question, -1 until answered
	right   int     // questions answered right
}

// startListening hides a told chapter's text and asks its questions.
func (m Model) startListening(c wiki.StoryChapter) Model {
	if len(c.Questions) == 0 {
		m.status = "This chapter was told before questions came with it; ctrl+r still reads it aloud."
		return m
	}
	rng := rand.New(rand.NewSource(m.now().UnixNano()))
	l := &listening{chapter: c, picked: -1}
	for _, q := range c.Questions {
		l.order = append(l.order, rng.Perm(len(q.Choices)))
	}
	m.listening, m.status = l, ""
	return m
}

// updateListening takes 1-4 for an answer, enter to go on; esc leaves.
func (m Model) updateListening(k tea.KeyMsg) (Model, tea.Cmd) {
	l := m.listening
	switch {
	case key.Matches(k, m.gkeys.Back):
		m.listening = nil
	case l.q >= len(l.chapter.Questions):
		if key.Matches(k, m.gkeys.Enter) {
			c := l.chapter
			m.listening = nil
			return m.openChapter(c), nil
		}
	case l.picked >= 0:
		if key.Matches(k, m.gkeys.Enter) || k.Type == tea.KeySpace {
			l.q, l.picked = l.q+1, -1
		}
	default:
		s := k.String()
		if len(s) == 1 && s[0] >= '1' && int(s[0]-'1') < len(l.order[l.q]) {
			l.picked = int(s[0] - '1')
			if l.order[l.q][l.picked] == 0 {
				l.right++
			}
		}
	}
	return m, nil
}

func (m Model) listeningView() string {
	l := m.listening
	var b strings.Builder
	b.WriteString(ui.SubtitleStyle.Render(fmt.Sprintf("  听书 第%d回 %s", l.chapter.Number, l.chapter.Title)) + "\n")
	b.WriteString(ui.MutedStyle.Render(ui.WrapIndent("Hear it first: ctrl+r reads the chapter aloud (again any time). The text opens after the questions.", m.width)) + "\n\n")
	if l.q >= len(l.chapter.Questions) {
		b.WriteString(ui.SuccessStyle.Render(fmt.Sprintf("  %d of %d right from listening.", l.right, len(l.chapter.Questions))) + "\n\n")
		b.WriteString(ui.HelpStyle.Render("  enter read the text · esc back"))
		return b.String()
	}
	q := l.chapter.Questions[l.q]
	b.WriteString(ui.ChineseStyle.Render(ui.WrapIndent(fmt.Sprintf("%d. %s", l.q+1, q.Ask), m.width)) + "\n")
	for i, c := range l.order[l.q] {
		line := fmt.Sprintf("  %d  %s", i+1, q.Choices[c])
		switch {
		case l.picked >= 0 && c == 0:
			line = ui.SuccessStyle.Render(line + "  ✓")
		case i == l.picked:
			line = ui.ErrorStyle.Render(line + "  ✗")
		}
		b.WriteString(line + "\n")
	}
	help := "  1-4 answer · ctrl+r hear it · esc back"
	if l.picked >= 0 {
		help = "  enter next question · ctrl+r hear it · esc back"
	}
	b.WriteString("\n" + ui.HelpStyle.Render(help))
	return b.String()
}
