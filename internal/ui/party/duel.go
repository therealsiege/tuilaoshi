package party

import (
	"context"
	"fmt"
	"strings"
	"time"
	"unicode"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/therealsiege/tuilaoshi/internal/claude"
	"github.com/therealsiege/tuilaoshi/internal/episode"
	"github.com/therealsiege/tuilaoshi/internal/model"
	"github.com/therealsiege/tuilaoshi/internal/ui"
	"github.com/therealsiege/tuilaoshi/internal/wiki"
)

// An idiom-chain duel (成语接龙) with a companion: they open with an idiom,
// and the learner answers with one starting where theirs ended (a homophone
// counts); Claude referees and plays the companion. Three misses end the
// duel; six links, or stumping the companion, win it. Each idiom recalled
// pays once a day like any word (through a study session) and raises the
// bond, and every idiom met joins the reviews.

// IdiomMover plays the companion's side of a duel.
type IdiomMover func(ctx context.Context, companion string, chain []string, attempt string, hint bool) (claude.IdiomTurn, error)

const (
	duelLives  = 3
	duelLinks  = 6  // links by the learner that win the duel
	duelWinPay = 20 // copper for winning
)

type duel struct {
	with    member
	chain   []string     // every idiom played, in order
	last    claude.Idiom // the idiom to chain from
	lives   int
	links   int
	hint    *claude.Idiom // shown hint, if asked; the next answer earns no bond
	waiting bool
	won     bool
	over    bool
	note    string
	session model.Session
	met     []wiki.Word
}

type duelMsg struct {
	turn    claude.IdiomTurn
	attempt string
	hint    bool
	err     error
}

// startDuel opens a duel with the selected companion. The Monkey King duels
// before joining, as he talks over his episodes.
func (m Model) startDuel() (Model, tea.Cmd) {
	if len(m.members) == 0 {
		return m, nil
	}
	b := m.members[m.cursor]
	switch {
	case m.mover == nil:
		m.note = "Idiom duels need ANTHROPIC_API_KEY (in ~/.config/tuilaoshi/env)."
		return m, nil
	case !b.Unlocked && b.Name != episode.Talker:
		m.note = b.Name + " hasn't joined the pilgrimage yet; 孙悟空 will duel any time."
		return m, nil
	}
	m.duel = &duel{with: b, lives: duelLives, session: model.Session{StartedAt: time.Now(), Type: "duel"}}
	m.input.Reset()
	m.note = ""
	return m, tea.Batch(m.move("", false), m.input.Focus())
}

// move asks the referee for the companion's next move.
func (m *Model) move(attempt string, hint bool) tea.Cmd {
	d := m.duel
	d.waiting = true
	mover, name, chain := m.mover, d.with.Name, append([]string(nil), d.chain...)
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()
		turn, err := mover(ctx, name, chain, attempt, hint)
		return duelMsg{turn, attempt, hint, err}
	}
}

// onDuel applies the referee's answer.
func (m Model) onDuel(msg duelMsg) Model {
	d := m.duel
	if d == nil {
		return m
	}
	d.waiting = false
	switch {
	case msg.err != nil:
		d.note = "Couldn't reach Claude: " + msg.err.Error()
	case msg.hint:
		if msg.turn.Reply.Text != "" {
			d.hint = &msg.turn.Reply
		}
	case len(d.chain) == 0: // the opening
		d.chain, d.last = []string{msg.turn.Reply.Text}, msg.turn.Reply
		d.meet(msg.turn.Reply)
	case !msg.turn.Valid:
		d.lives--
		d.note = msg.turn.Note
		d.record(msg.attempt, false)
		d.over = d.lives == 0
	default:
		played := msg.turn.Played
		if played.Text == "" {
			played.Text = msg.attempt
		}
		d.links++
		d.record(played.Text, true)
		d.meet(played)
		d.chain = append(d.chain, played.Text)
		if d.hint == nil {
			if _, err := m.store.AddTalkBond(d.with.Name, 1, time.Now()); err != nil {
				m.loadErr = err
			}
		}
		d.hint, d.note = nil, msg.turn.Note
		if msg.turn.Reply.Text == "" {
			d.won, d.over = true, true
			d.note = d.with.Name + " is stumped!"
		} else {
			d.chain, d.last = append(d.chain, msg.turn.Reply.Text), msg.turn.Reply
			d.meet(msg.turn.Reply)
			if d.links >= duelLinks {
				d.won, d.over = true, true
			}
		}
	}
	if d.over {
		m.finishDuel()
	}
	return m
}

func (d *duel) record(idiom string, right bool) {
	d.session.Items = append(d.session.Items, model.SessionItem{CardID: "idiom:" + idiom, UserAnswer: idiom, Correct: right})
	d.session.CardsTotal++
	if right {
		d.session.CardsCorrect++
	}
}

func (d *duel) meet(i claude.Idiom) {
	if i.Text != "" {
		d.met = append(d.met, wiki.Word{Chinese: i.Text, Pinyin: i.Pinyin, Meaning: i.Meaning})
	}
}

// finishDuel saves the duel: the answers as a study session (each idiom
// pays once a day), the idioms met as review words, and a win's reward.
func (m *Model) finishDuel() {
	d := m.duel
	if len(d.session.Items) > 0 {
		d.session.EndedAt = time.Now().Add(time.Second)
		if err := m.store.WriteSession(d.session); err != nil {
			m.loadErr = err
		}
		d.session.Items = nil
	}
	if _, err := m.store.AddIdioms(d.met, time.Now()); err != nil {
		m.loadErr = err
	}
	d.met = nil
	if d.won {
		if err := m.store.Earn(wiki.Wallet{Tongqian: duelWinPay}); err != nil {
			m.loadErr = err
		}
	}
}

// updateDuel types and sends idioms; h asks for a hint, esc ends the duel.
func (m Model) updateDuel(k tea.KeyMsg) (Model, tea.Cmd) {
	d := m.duel
	switch {
	case k.Type == tea.KeyEsc && !m.input.Composing():
		if !d.over {
			m.finishDuel()
		}
		m.duel = nil
		return m.Refresh(), nil
	case d.over || d.waiting:
		return m, nil
	case k.Type == tea.KeyCtrlH:
		return m, m.move("", true)
	case k.Type == tea.KeyEnter && !m.input.Composing():
		attempt := hanOnly(m.input.Value())
		if len([]rune(attempt)) != 4 {
			d.note = "A 成语 here has four characters."
			return m, nil
		}
		m.input.Reset()
		return m, m.move(attempt, false)
	}
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(k)
	return m, cmd
}

func hanOnly(s string) string {
	var b strings.Builder
	for _, r := range s {
		if unicode.Is(unicode.Han, r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func (m Model) duelView() string {
	d := m.duel
	var b strings.Builder
	b.WriteString(ui.TitleStyle.MarginBottom(0).Render("成语接龙 Idiom chain with "+d.with.Name+" "+d.with.English) + "\n")
	b.WriteString(ui.MutedStyle.Render(fmt.Sprintf("  lives %s · links %d/%d · answer with an idiom that starts where theirs ends",
		strings.Repeat("♥", d.lives)+strings.Repeat("·", duelLives-d.lives), d.links, duelLinks)) + "\n\n")
	if len(d.chain) > 1 {
		b.WriteString(ui.MutedStyle.Render("  "+strings.Join(d.chain[:len(d.chain)-1], " → ")+" →") + "\n")
	}
	if d.last.Text != "" {
		fmt.Fprintf(&b, "  %s  %s\n  %s\n\n", ui.ChineseStyle.Bold(true).Render(d.last.Text), ui.Pinyin(d.last.Pinyin), ui.MutedStyle.Render(d.last.Meaning))
	}
	if h := d.hint; h != nil {
		r := []rune(h.Text)
		fmt.Fprintf(&b, "  %s %s  %s\n\n", ui.WarningStyle.Render("Hint:"), ui.ChineseStyle.Render(string(r[:1])+strings.Repeat("□", max(len(r)-1, 0))), ui.MutedStyle.Render(h.Pinyin+" · "+h.Meaning))
	}
	switch {
	case d.waiting:
		b.WriteString(ui.MutedStyle.Render("  "+d.with.Name+" thinks…") + "\n")
	case d.over && d.won:
		b.WriteString(ui.SuccessStyle.Render(fmt.Sprintf("  You win the duel! +%d 铜钱. Every idiom met joins your reviews.", duelWinPay)) + "\n")
	case d.over:
		b.WriteString(ui.WarningStyle.Render("  Out of lives. Every idiom met joins your reviews; try again tomorrow or now.") + "\n")
	default:
		b.WriteString("  " + m.input.View() + "\n")
	}
	if d.note != "" {
		b.WriteString(ui.MutedStyle.Render("  "+ui.WrapText(d.note, 72)) + "\n")
	}
	b.WriteString("\n" + ui.HelpStyle.Render("  enter play · ctrl+h hint (no bond for that idiom) · esc end"))
	return b.String()
}
