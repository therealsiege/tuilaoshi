package rpg

import (
	"context"
	"errors"
	"math/rand"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/therealsiege/tuilaoshi/internal/journey"
	"github.com/therealsiege/tuilaoshi/internal/ui"
)

// baiguAt is a battle facing the White Bone Demon's first disguise, each
// disguise told in its first round and the choices shuffled the same way
// every run.
func baiguAt(t *testing.T, w, h int) Model {
	t.Helper()
	m := storeModel(t).SetSize(w, h)
	st := journey.Stages(journey.Counts{}, nil)[19]
	if st.Boss == nil || st.Boss.Name != journey.BaiguDemon {
		t.Fatalf("stage 20 is %+v, want the White Bone Demon's", st.Boss)
	}
	m.road, m.stage = true, &st
	m.startBaigu()
	m.bg.variant = []int{0, 0, 0}
	m.bg.rng = rand.New(rand.NewSource(1))
	m, _ = m.askBaigu()
	return m
}

// pressKey presses one key: enter, esc, or the runes of s.
func pressKey(m Model, s string) (Model, tea.Cmd) {
	switch s {
	case "enter":
		return m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	case "esc":
		return m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	}
	return m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)})
}

// choose picks the right choice, or a wrong one, by its number on screen.
func choose(m Model, right bool) Model {
	for i, j := range m.bg.order {
		if (j == 0) == right {
			m, _ = pressKey(m, string(rune('1'+i)))
			return m
		}
	}
	return m
}

// gradeWith stands in for Claude, giving every written answer score.
func gradeWith(score int, err error) Grader {
	return func(context.Context, string, string, string) (int, string, error) {
		return score, "Claude's note on the answer.", err
	}
}

// writeAndSend types an answer for the old man and plays Claude's grade.
func writeAndSend(t *testing.T, m Model, text string) Model {
	t.Helper()
	m.bgIn.SetValue(text)
	m, cmd := pressKey(m, "enter")
	if m.bg.phase != baiguGrading || cmd == nil {
		t.Fatalf("sending a written answer: phase %d, cmd %v", m.bg.phase, cmd)
	}
	m, _ = m.Update(cmd())
	return m
}

func masterMorale(t *testing.T, m Model) int {
	t.Helper()
	tr, _, err := journey.LoadTrail(m.store, nil)
	if err != nil {
		t.Fatal(err)
	}
	tr.Join([]string{journey.Master})
	return tr.Morale[journey.Master]
}

func TestBaiguIsTheWhiteBoneDemonsTribulation(t *testing.T) {
	m := skipScene(New(stageWiki(t, journey.TribulationAt(20).Unlock, 19), nil, nil).RoadFight())
	if m.stage == nil || m.stage.Number != 20 || !m.baiguFight() || len(m.gauntlet) != 0 {
		t.Fatalf("stage %v, gauntlet %d: want the White Bone Demon's tribulation as her disguises", m.stage, len(m.gauntlet))
	}
	if v := m.View(); !strings.Contains(v, "三打白骨精") || strings.Contains(v, "BOSS!") {
		t.Errorf("intro doesn't say she is fought in three disguises:\n%s", v)
	}
	m, _ = pressKey(m, "enter")
	if m.state != stateBaigu || m.bg.disguise != 0 || m.bg.phase != baiguAsk {
		t.Fatalf("enter on the intro: state %d, disguise %d, phase %d; want the village girl", m.state, m.bg.disguise, m.bg.phase)
	}
	if New(stageWiki(t, 0, 0), nil, nil).Spar().baiguFight() {
		t.Error("sparring at camp is fought as the disguises")
	}
}

func TestBaiguReadingRound(t *testing.T) {
	m := choose(baiguAt(t, 80, 24), true)
	if m.bg.phase != baiguRight || !strings.Contains(m.View(), "Wukong sees through 村姑") {
		t.Fatalf("right answer: phase %d\n%s", m.bg.phase, m.View())
	}
	m, _ = pressKey(m, "enter")
	if m.bg.disguise != 1 || m.bg.phase != baiguAsk {
		t.Errorf("after the girl: disguise %d, phase %d; want the old woman", m.bg.disguise, m.bg.phase)
	}
}

func TestBaiguWrongUnmaskingAngersTheMasterAndComesAgain(t *testing.T) {
	m := baiguAt(t, 80, 24)
	was := masterMorale(t, m)
	m = choose(m, false)
	if m.bg.phase != baiguWrong || m.bg.misses != 1 {
		t.Fatalf("wrong answer: phase %d, misses %d", m.bg.phase, m.bg.misses)
	}
	if got := masterMorale(t, m); got != was-journey.BaiguMorale {
		t.Errorf("master's morale %d -> %d, want -%d", was, got, journey.BaiguMorale)
	}
	if v := m.View(); !strings.Contains(v, "悟空") || !strings.Contains(v, "morale") {
		t.Errorf("wrong unmasking doesn't show the master's anger:\n%s", v)
	}
	m, _ = pressKey(m, "enter")
	if m.bg.disguise != 0 || m.bg.variant[0] != 1 || m.bg.phase != baiguAsk {
		t.Errorf("after a wrong answer: disguise %d, round %d, phase %d; want the girl again in another telling",
			m.bg.disguise, m.bg.variant[0], m.bg.phase)
	}
}

func TestBaiguListeningHidesTheStoryUntilAnswered(t *testing.T) {
	m := baiguAt(t, 80, 24)
	m.bg.disguise = 1
	m, _ = m.askBaigu()
	_, r := m.baiguRound()
	story := r.Text[:9] // its first three characters
	if v := m.View(); !strings.Contains(v, story) || !strings.Contains(v, "No voice") {
		t.Errorf("with no audio the old woman's story isn't shown to read:\n%s", v)
	}
	m.bg.listen = true // as when a voice speaks it
	if v := m.View(); strings.Contains(v, story) || !strings.Contains(v, "r hears it again") {
		t.Errorf("while her story is heard its words show:\n%s", v)
	}
	m.ttsErr = errors.New("no speaker")
	if v := m.View(); !strings.Contains(v, story) {
		t.Errorf("when the audio fails her story isn't shown:\n%s", v)
	}
	m.ttsErr = nil
	m = choose(m, true)
	if v := m.View(); m.bg.phase != baiguRight || !strings.Contains(v, story) || !strings.Contains(v, r.English[:20]) {
		t.Errorf("answered: her story and its English don't show:\n%s", v)
	}
}

func TestBaiguWritingGradedByClaude(t *testing.T) {
	setup := func(g Grader) Model {
		m := baiguAt(t, 80, 24)
		m.grader, m.bg.disguise = g, 2
		m, _ = m.askBaigu()
		return m
	}
	m := writeAndSend(t, setup(gradeWith(2, nil)), "他没有影子")
	if m.bg.phase != baiguRight || !strings.Contains(m.View(), "他没有影子") {
		t.Errorf("passing answer: phase %d\n%s", m.bg.phase, m.View())
	}

	m = setup(gradeWith(0, nil))
	was := masterMorale(t, m)
	m = writeAndSend(t, m, "他很老")
	if m.bg.phase != baiguWrong || masterMorale(t, m) != was-journey.BaiguMorale {
		t.Errorf("failing answer: phase %d, morale %d -> %d", m.bg.phase, was, masterMorale(t, m))
	}
	if v := m.View(); !strings.Contains(v, "Claude's note") || !strings.Contains(v, "It was: ") {
		t.Errorf("failing answer doesn't show Claude's note and a right answer:\n%s", v)
	}
	m, _ = pressKey(m, "enter")
	if m.bg.disguise != 2 || m.bg.variant[2] != 1 || m.bg.phase != baiguAsk || m.bgIn.Value() != "" {
		t.Errorf("after failing: disguise %d, round %d, phase %d, box %q", m.bg.disguise, m.bg.variant[2], m.bg.phase, m.bgIn.Value())
	}

	m = writeAndSend(t, setup(gradeWith(0, errors.New("offline"))), "他没有影子")
	if m.bg.phase != baiguAsk || !strings.Contains(m.View(), "Couldn't reach Claude") || m.bg.misses != 0 {
		t.Errorf("Claude unreachable: phase %d, misses %d\n%s", m.bg.phase, m.bg.misses, m.View())
	}

	m = setup(gradeWith(2, nil))
	m, _ = pressKey(m, "enter")
	if m.bg.phase != baiguAsk {
		t.Errorf("an empty answer was sent")
	}
}

func TestBaiguWritingWithoutAKeyIsChosen(t *testing.T) {
	m := baiguAt(t, 80, 24)
	m.bg.disguise = 2
	m, _ = m.askBaigu()
	if m.baiguWrites() || !strings.Contains(m.View(), "1-3 answer") {
		t.Fatalf("without Claude the old man's round isn't chosen from three:\n%s", m.View())
	}
	if m = choose(m, true); m.bg.phase != baiguRight {
		t.Errorf("right sentence: phase %d", m.bg.phase)
	}
}

func clearedNow(t *testing.T, m Model) map[int]string {
	t.Helper()
	cleared, err := m.store.ClearedStages()
	if err != nil {
		t.Fatal(err)
	}
	return cleared
}

func TestBaiguThreeInOrderClearTheStage(t *testing.T) {
	m := baiguAt(t, 80, 24)
	before, err := m.store.Wallet()
	if err != nil {
		t.Fatal(err)
	}
	for i, skill := range []string{"reading", "listening", "writing"} {
		if d, _ := m.baiguRound(); m.bg.disguise != i || d.Skill != skill {
			t.Fatalf("disguise %d is %s, want %s", m.bg.disguise, d.Skill, skill)
		}
		if clearedNow(t, m)[20] != "" {
			t.Fatalf("stage cleared before the %s disguise", skill)
		}
		m = choose(m, true)
		m, _ = pressKey(m, "enter")
	}
	if m.bg.phase != baiguWon || !strings.Contains(m.View(), "Tribulation 20 cleared") {
		t.Fatalf("after all three: phase %d\n%s", m.bg.phase, m.View())
	}
	if clearedNow(t, m)[20] == "" {
		t.Error("stage 20 not recorded as cleared")
	}
	after, err := m.store.Wallet()
	if err != nil {
		t.Fatal(err)
	}
	if after.Yuanbao != before.Yuanbao+1 || after.Tongqian != before.Tongqian+baiguPay {
		t.Errorf("purse %+v -> %+v, want a boss's pay", before, after)
	}
	m, cmd := pressKey(m, "enter")
	if cmd == nil {
		t.Fatal("enter after winning does nothing")
	}
	if msg, ok := cmd().(ui.SwitchViewMsg); !ok || msg.Target != ui.ViewMap {
		t.Errorf("enter after winning: %#v, want the map", cmd())
	}
	if m = m.Leave(); m.saveErr != nil {
		t.Errorf("leaving after the win: %v", m.saveErr)
	}
}
