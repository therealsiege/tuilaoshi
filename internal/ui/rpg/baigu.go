package rpg

import (
	"context"
	"math/rand"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/therealsiege/tuilaoshi/internal/audio"
	"github.com/therealsiege/tuilaoshi/internal/claude"
	"github.com/therealsiege/tuilaoshi/internal/journey"
	"github.com/therealsiege/tuilaoshi/internal/ui"
	"github.com/therealsiege/tuilaoshi/internal/ui/input"
	"github.com/therealsiege/tuilaoshi/internal/wiki"
)

// 三打白骨精: on the road, the White Bone Demon's tribulation is not fought
// as a battle. She comes three times, as a village girl, an old woman and an
// old man (content/baigu.yaml), and Wukong must see through each in turn:
// by reading what the girl says of her basket, by listening to the old
// woman's story, and by saying in Chinese what gives the old man away
// (graded by Claude, or chosen from three sentences without a key). A wrong
// unmasking angers the master, the party's morale drops, and that disguise
// comes again in another telling. Seeing through all three clears the
// tribulation as a won boss battle does.

// stateBaigu is the three disguises, entered from the tribulation's intro.
const stateBaigu viewState = 100

type baiguPhase int

const (
	baiguAsk     baiguPhase = iota // reading, hearing or writing, not yet answered
	baiguGrading                   // a written answer is with Claude
	baiguRight                     // this disguise is seen through
	baiguWrong                     // the wrong thing was unmasked
	baiguWon                       // all three seen through
)

// baiguPay is the copper paid for seeing through all three, as for a boss.
const baiguPay = 50

// Grader grades Chinese written for a task against criteria, 0-2, with feedback.
type Grader func(ctx context.Context, task, criteria, answer string) (int, string, error)

// baiguGradeMsg carries Claude's grade of a written answer; seq ties it to
// the answer it grades.
type baiguGradeMsg struct {
	seq      int
	score    int
	feedback string
	err      error
}

// baiguState is how far the three disguises have gone.
type baiguState struct {
	trial    journey.Baigu
	rng      *rand.Rand
	disguise int        // the disguise faced now, 0-2
	variant  []int      // the round told for each disguise
	order    []int      // the round's choices as shown, by index into Choices
	phase    baiguPhase // where the current disguise stands
	picked   int        // the choice made, by index into Choices; -1 for a written answer
	typed    string     // the written answer
	feedback string     // Claude's note on the written answer
	note     string     // why grading failed
	listen   bool       // the old woman's story is spoken, so its words stay hidden
	english  bool       // e: show the passage's English where the screen is too short for both
	misses   int        // wrong unmaskings so far
	seq      int        // written answers sent, so a stale grade is ignored
}

// WithClaude lets the old man's round be answered in writing and graded by
// Claude; without a client it is answered by choosing a sentence.
func (m Model) WithClaude(c *claude.Client) Model {
	if c != nil {
		m.grader = c.GradeWriting
	}
	return m
}

// InputComposing reports whether a typed answer is mid-composition, so esc
// cancels it rather than fleeing.
func (m Model) InputComposing() bool {
	return m.input.Composing() || (m.state == stateBaigu && m.bgIn.Composing())
}

// newBaiguInput is the answer box for the old man's round.
func newBaiguInput(dict *input.Dict) input.Model {
	in := input.NewWithDict(dict)
	in.SetPlaceholder("write in Chinese: type pinyin, then pick characters")
	in.SetCharLimit(60)
	return in
}

// baiguFight reports whether this road battle is the White Bone Demon's
// tribulation, fought as her three disguises.
func (m Model) baiguFight() bool {
	return m.road && m.stage != nil && m.stage.Boss != nil && m.stage.Boss.Name == journey.BaiguDemon &&
		len(journey.BaiguTrial().Disguises) > 0
}

// startBaigu readies the three disguises in place of the stage's gauntlet;
// the boss battle stays loaded for the intro's portrait and the wounds.
func (m *Model) startBaigu() {
	m.gauntlet, m.foe = nil, 0
	m.battle = m.stageBattle()
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	trial := journey.BaiguTrial()
	variant := make([]int, len(trial.Disguises))
	for i, d := range trial.Disguises {
		variant[i] = rng.Intn(len(d.Rounds))
	}
	m.bg = baiguState{trial: trial, rng: rng, variant: variant}
}

// baiguRound is the disguise faced now and the round it is told in.
func (m Model) baiguRound() (journey.Disguise, journey.BaiguRound) {
	d := m.bg.trial.Disguises[m.bg.disguise]
	return d, d.Rounds[m.bg.variant[m.bg.disguise]%len(d.Rounds)]
}

// baiguWrites reports whether the current round is answered by writing.
func (m Model) baiguWrites() bool {
	d, _ := m.baiguRound()
	return d.Skill == "writing" && m.grader != nil
}

// askBaigu starts the current disguise's round: choices shuffled, the old
// woman's story spoken when there is a voice, the answer box readied for
// the old man.
func (m Model) askBaigu() (Model, tea.Cmd) {
	m.state = stateBaigu
	d, r := m.baiguRound()
	bg := &m.bg
	bg.order = bg.rng.Perm(len(r.Choices))
	bg.phase, bg.picked, bg.typed, bg.feedback, bg.note = baiguAsk, -1, "", "", ""
	bg.listen, bg.english = false, false
	switch {
	case d.Skill == "listening" && ui.Audible(m.tts):
		m.ttsErr = nil
		bg.listen = true
		return m, ui.SpeakCmd(m.tts, r.Text, "zh")
	case m.baiguWrites():
		m.bgIn.Reset()
		m.bgIn.SetMode(ui.AnswerMode())
		return m, m.bgIn.Focus()
	}
	return m, nil
}

// updateBaigu handles keys while the disguises are faced.
func (m Model) updateBaigu(msg tea.KeyMsg) (Model, tea.Cmd) {
	enter := key.Matches(msg, m.gkeys.Enter) || msg.Type == tea.KeySpace
	switch m.bg.phase {
	case baiguAsk:
		if m.baiguWrites() {
			if key.Matches(msg, m.gkeys.Enter) && !m.bgIn.Composing() {
				return m.submitBaigu()
			}
			var cmd tea.Cmd
			m.bgIn, cmd = m.bgIn.Update(msg)
			return m, cmd
		}
		if s := msg.String(); s == "r" && m.bg.listen {
			_, r := m.baiguRound()
			return m, ui.SpeakCmd(m.tts, r.Text, "zh")
		} else if len(s) == 1 && s[0] >= '1' && int(s[0]-'1') < len(m.bg.order) {
			m.bg.picked = m.bg.order[s[0]-'1']
			return m.judgeBaigu(m.bg.picked == 0), nil
		}
	case baiguRight:
		if enter {
			m.bg.disguise++
			if m.bg.disguise == len(m.bg.trial.Disguises) {
				m.bg.disguise--
				return m.winBaigu(), nil
			}
			return m.askBaigu()
		}
	case baiguWrong:
		if enter {
			d := m.bg.trial.Disguises[m.bg.disguise]
			m.bg.variant[m.bg.disguise] = (m.bg.variant[m.bg.disguise] + 1) % len(d.Rounds)
			return m.askBaigu()
		}
	case baiguWon:
		if enter {
			return m, ui.SwitchTo(ui.ViewMap)
		}
	}
	if msg.String() == "e" && (m.bg.phase == baiguRight || m.bg.phase == baiguWrong) {
		m.bg.english = !m.bg.english
	}
	return m, nil
}

// submitBaigu sends the written answer to Claude.
func (m Model) submitBaigu() (Model, tea.Cmd) {
	text := strings.TrimSpace(m.bgIn.Value())
	if text == "" {
		return m, nil
	}
	_, r := m.baiguRound()
	m.bgIn.Blur()
	m.bg.typed, m.bg.phase, m.bg.note = text, baiguGrading, ""
	m.bg.seq++
	seq, grader := m.bg.seq, m.grader
	task := r.Text + "\n" + r.Question + " (" + r.QuestionEnglish + ")"
	return m, func() tea.Msg {
		score, fb, err := grader(context.Background(), task, r.Criteria, text)
		return baiguGradeMsg{seq: seq, score: score, feedback: fb, err: err}
	}
}

// onBaiguGrade judges the written answer by Claude's grade; when Claude
// can't be reached the answer can be sent again.
func (m Model) onBaiguGrade(msg baiguGradeMsg) (Model, tea.Cmd) {
	if m.state != stateBaigu || m.bg.phase != baiguGrading || msg.seq != m.bg.seq {
		return m, nil
	}
	if msg.err != nil {
		m.bg.phase, m.bg.note = baiguAsk, "Couldn't reach Claude: "+msg.err.Error()
		return m, m.bgIn.Focus()
	}
	m.bg.feedback = msg.feedback
	return m.judgeBaigu(msg.score > 0), nil
}

// judgeBaigu settles the current disguise: seen through, or the wrong
// thing unmasked, which angers the master.
func (m Model) judgeBaigu(right bool) Model {
	m.bg.listen = false // her words show once answered
	if right {
		ui.Cue(audio.Right)
		m.bg.phase = baiguRight
		return m
	}
	ui.Cue(audio.Wrong)
	m.bg.phase = baiguWrong
	m.bg.misses++
	m.angerMaster()
	return m
}

// angerMaster lowers the party's morale for a wrong unmasking.
func (m *Model) angerMaster() {
	if m.store == nil {
		return
	}
	cleared, err := m.store.ClearedStages()
	if err != nil {
		m.saveErr = err
		return
	}
	t, members, err := journey.LoadTrail(m.store, cleared)
	if err != nil {
		m.saveErr = err
		return
	}
	t.Apply(journey.Outcome{Morale: -journey.BaiguMorale}, members)
	if err := m.store.SaveTrail(t); err != nil {
		m.saveErr = err
	}
}

// winBaigu clears the tribulation as a won boss battle does: the boss's
// copper and ingot, her card in the collection, and the stage recorded.
func (m Model) winBaigu() Model {
	ui.Cue(audio.Victory)
	m.bg.phase = baiguWon
	if m.store == nil || m.saved {
		return m
	}
	before, err := m.store.Wallet()
	if err == nil {
		err = m.store.Earn(wiki.Wallet{Tongqian: baiguPay, Yuanbao: 1})
	}
	if err == nil {
		err = m.store.Tally("battle_won", 1)
	}
	if err == nil {
		err = m.store.RecordDefeat(journey.BaiguDemon)
	}
	if err == nil {
		err = m.store.RecordStageCleared(m.stage.Number, time.Now())
	}
	if err != nil {
		m.saveErr = err
		return m
	}
	if after, err := m.store.Wallet(); err == nil {
		m.earned = wiki.Wallet{Tongqian: after.Tongqian - before.Tongqian, Yuanbao: after.Yuanbao - before.Yuanbao}
		m.wallet = after
	}
	m.saved = true
	return m
}

// Typing reports whether a text box has the keyboard (a typed answer, or
// the old man's round), so the app lets ? be typed.
func (m Model) Typing() bool {
	return m.input.Focused() || (m.state == stateBaigu && m.bgIn.Focused())
}
