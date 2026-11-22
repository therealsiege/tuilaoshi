package party

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/therealsiege/tuilaoshi/internal/claude"
	"github.com/therealsiege/tuilaoshi/internal/journey"
	"github.com/therealsiege/tuilaoshi/internal/model"
	"github.com/therealsiege/tuilaoshi/internal/ui"
	"github.com/therealsiege/tuilaoshi/internal/wiki"
)

// Talking with a companion: they speak in their own voice around three
// words the learner is studying, and grade each reply; real Chinese that
// uses the words raises the bond, up to wiki.TalkCap points a day.

// Talker sends a conversation to Claude with a system prompt.
type Talker func(ctx context.Context, system string, history []claude.ChatMessage, msg string) (string, error)

// talkWords is how many study words a conversation is built around.
const talkWords = 3

type talkReplyMsg struct {
	text string
	err  error
}

// talk is an open conversation.
type talk struct {
	with    member
	system  string
	history []claude.ChatMessage // what Claude has seen, including the hidden opener
	words   []model.Card
	waiting bool
	gained  int // bond points earned in this conversation
}

// startTalk opens a conversation with the selected companion.
func (m Model) startTalk() (Model, tea.Cmd) {
	if len(m.members) == 0 {
		return m, nil
	}
	b := m.members[m.cursor]
	switch {
	case !b.Unlocked:
		m.note = b.Name + " hasn't joined the pilgrimage yet."
		return m, nil
	case m.talker == nil:
		m.note = "Talking needs ANTHROPIC_API_KEY (in ~/.config/tuilaoshi/env)."
		return m, nil
	}
	cards, err := m.store.DrillableCards()
	if err != nil {
		m.loadErr = err
	}
	words := studyWords(cards, rand.New(rand.NewSource(time.Now().UnixNano())))
	m.talk = &talk{with: b, system: journey.TalkPrompt(b.Companion, b.bond, words), words: words}
	m.input.Reset()
	m.note = ""
	return m, tea.Batch(m.send(journey.TalkOpener), m.input.Focus())
}

// studyWords picks words still being learned (review interval under two
// weeks), or any words when there are too few.
func studyWords(cards []model.Card, rng *rand.Rand) []model.Card {
	var learning, all []model.Card
	for _, c := range cards {
		if c.Chinese == "" || c.English == "" {
			continue
		}
		all = append(all, c)
		if c.ReviewInterval >= 1 && c.ReviewInterval < 14 {
			learning = append(learning, c)
		}
	}
	pool := learning
	if len(pool) < talkWords {
		pool = all
	}
	rng.Shuffle(len(pool), func(i, j int) { pool[i], pool[j] = pool[j], pool[i] })
	return pool[:min(talkWords, len(pool))]
}

// send asks the companion to answer msg.
func (m *Model) send(msg string) tea.Cmd {
	t := m.talk
	t.waiting = true
	history := append([]claude.ChatMessage(nil), t.history...)
	t.history = append(t.history, claude.ChatMessage{Role: "user", Content: msg})
	talker, system := m.talker, t.system
	return func() tea.Msg {
		reply, err := talker(context.Background(), system, history, msg)
		return talkReplyMsg{reply, err}
	}
}

// onReply shows the companion's answer and credits the learner's last message.
func (m Model) onReply(msg talkReplyMsg) Model {
	t := m.talk
	if t == nil {
		return m
	}
	t.waiting = false
	if msg.err != nil {
		t.history = t.history[:len(t.history)-1] // let the learner try again
		m.note = "Couldn't reach Claude: " + msg.err.Error()
		return m
	}
	text, grade := journey.ParseTalkReply(msg.text)
	t.history = append(t.history, claude.ChatMessage{Role: "assistant", Content: text})
	if len(t.history) > 2 { // not the opener
		added, err := m.store.AddTalkBond(t.with.Name, grade, time.Now())
		switch {
		case err != nil:
			m.note = err.Error()
		case added > 0:
			t.gained += added
			m.note = fmt.Sprintf("Bond +%d with %s", added, t.with.Name)
		case grade > 0:
			m.note = fmt.Sprintf("You've talked enough with %s today to raise the bond; keep chatting for practice.", t.with.Name)
		default:
			m.note = "Answer in Chinese, using one of the words, to raise the bond."
		}
	}
	return m
}

// updateTalk types and sends; esc ends the conversation.
func (m Model) updateTalk(k tea.KeyMsg) (Model, tea.Cmd) {
	switch {
	case k.Type == tea.KeyEsc && !m.input.Composing():
		m.talk = nil
		return m.Refresh(), nil
	case k.Type == tea.KeyEnter && !m.input.Composing():
		text := strings.TrimSpace(m.input.Value())
		if text == "" || m.talk.waiting {
			return m, nil
		}
		m.input.Reset()
		return m, m.send(text)
	}
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(k)
	return m, cmd
}

func (m Model) talkView() string {
	t := m.talk
	var b strings.Builder
	b.WriteString(ui.TitleStyle.MarginBottom(0).Render("谈心 Talking with "+t.with.Name+" "+t.with.English) + "\n")
	var words []string
	for _, w := range t.words {
		words = append(words, ui.ChineseStyle.Render(w.Chinese)+" "+w.Pinyin+" "+ui.MutedStyle.Render(w.English))
	}
	b.WriteString("  Use these: " + strings.Join(words, " · ") + "\n\n")

	width := 76
	if m.width > 0 {
		width = max(m.width-8, 30)
	}
	var lines []string
	for _, h := range t.history[1:] { // the opener stays hidden
		who := ui.ChineseStyle.Render(t.with.Name + "：")
		if h.Role == "user" {
			who = ui.SubtitleStyle.Render("你：")
		}
		lines = append(lines, who)
		for _, l := range strings.Split(ui.WrapText(h.Content, width), "\n") {
			lines = append(lines, "  "+l)
		}
	}
	if t.waiting {
		lines = append(lines, ui.MutedStyle.Render(t.with.Name+" is thinking…"))
	}
	if keep := max(m.height-12, 6); len(lines) > keep {
		lines = lines[len(lines)-keep:]
	}
	for _, l := range lines {
		b.WriteString("  " + l + "\n")
	}
	b.WriteString("\n  " + m.input.View() + "\n")
	if m.note != "" {
		b.WriteString(ui.WarningStyle.Render("  "+m.note) + "\n")
	}
	b.WriteString(ui.HelpStyle.Render(fmt.Sprintf("  enter send · ctrl+space pinyin/English · esc end · bond from talk today, up to %d", wiki.TalkCap)))
	return b.String()
}
