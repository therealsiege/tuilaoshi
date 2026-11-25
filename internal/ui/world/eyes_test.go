package world

import (
	"context"
	"errors"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/therealsiege/tuilaoshi/internal/claude"
	"github.com/therealsiege/tuilaoshi/internal/journey"
	"github.com/therealsiege/tuilaoshi/internal/ui/roam"
	"github.com/therealsiege/tuilaoshi/internal/ui/uitest"
)

var esc = tea.KeyMsg{Type: tea.KeyEsc}

// villagersAt is a town whose villagers answer with a fixed line, keeping
// the system prompts they were given.
func villagersAt(t *testing.T, prompts *[]string) Model {
	t.Helper()
	m := worldAt(t, 11)
	m.eyes.talker = func(_ context.Context, system string, history []claude.ChatMessage, q string) (string, error) {
		*prompts = append(*prompts, system)
		return "我昨天晚上在家，哪儿也没去。\nEN: I was home last night and went nowhere.", nil
	}
	m, _ = m.Update(press("f"))
	if m.screen != screenEyes {
		t.Fatalf("the rumour didn't open: %q", m.note)
	}
	return m
}

// askEyes puts a question to villager n (1-3) and delivers the answer.
func askEyes(t *testing.T, m Model, n int, q string) Model {
	t.Helper()
	if !m.eyes.asking || m.eyes.who != n-1 {
		m, _ = m.Update(press(string(rune('0' + n))))
	}
	if !m.Typing() {
		t.Fatalf("the question box didn't open: %q", m.note)
	}
	m.input.SetValue(q)
	m, cmd := m.Update(enter)
	if cmd == nil {
		t.Fatal("the question wasn't sent")
	}
	m, _ = m.Update(cmd())
	return m
}

func TestFieryEyesNeedsAKey(t *testing.T) {
	m := worldAt(t, 11)
	m, _ = m.Update(press("f"))
	if m.screen != screenTown || !strings.Contains(m.note, "ANTHROPIC_API_KEY") {
		t.Errorf("without Claude the rumour should say it needs a key: screen %d, %q", m.screen, m.note)
	}
}

func TestQuestioningAVillager(t *testing.T) {
	var prompts []string
	m := villagersAt(t, &prompts)
	m = askEyes(t, m, 2, "你昨天晚上在哪儿？")
	v := m.eyes.mystery.Villagers[1]
	if len(prompts) != 1 || !strings.Contains(prompts[0], v.Name) {
		t.Fatalf("the question went to the wrong villager: %v", prompts)
	}
	if want := m.eyes.demon == 1; strings.Contains(prompts[0], v.Lie) != want {
		t.Errorf("villager 2 tells the lie only as the demon (demon %d):\n%s", m.eyes.demon, prompts[0])
	}
	if !strings.Contains(m.View(), "哪儿也没去") || strings.Contains(m.View(), "went nowhere") {
		t.Fatalf("the answer should show, its English hidden:\n%s", m.View())
	}
	if m.Hear() != "我昨天晚上在家，哪儿也没去。" {
		t.Errorf("ctrl+r should read the last answer: %q", m.Hear())
	}
	m, _ = m.Update(esc)
	m, _ = m.Update(press("e"))
	if !strings.Contains(m.View(), "went nowhere") {
		t.Errorf("e should show the English:\n%s", m.View())
	}
	m, _ = m.Update(esc)
	if m.screen != screenTown {
		t.Fatalf("esc should leave for town, screen %d", m.screen)
	}
	m, _ = m.Update(press("f"))
	if m.eyes.asked != 1 || len(m.eyes.talks[1]) != 1 {
		t.Errorf("the mystery should resume where it was: asked %d", m.eyes.asked)
	}
}

func TestAFailedQuestionIsGivenBack(t *testing.T) {
	var prompts []string
	m := villagersAt(t, &prompts)
	m.eyes.talker = func(context.Context, string, []claude.ChatMessage, string) (string, error) {
		return "", errors.New("offline")
	}
	m = askEyes(t, m, 1, "你好")
	if m.eyes.asked != 0 || len(m.eyes.talks[0]) != 0 || !strings.Contains(m.note, "offline") {
		t.Errorf("a failed call should cost nothing: asked %d, %q", m.eyes.asked, m.note)
	}
}

func TestTheQuestionsRunOut(t *testing.T) {
	var prompts []string
	m := villagersAt(t, &prompts)
	for i := range journey.MysteryQuestions {
		m = askEyes(t, m, i%3+1, "你看见什么了？")
	}
	if m.eyes.asking || !strings.Contains(m.note, "last question") {
		t.Fatalf("the box should close after the last question: %q", m.note)
	}
	m, _ = m.Update(press("1"))
	if m.Typing() || !strings.Contains(m.note, "No questions left") {
		t.Errorf("a ninth question was allowed: %q", m.note)
	}
	if len(prompts) != journey.MysteryQuestions {
		t.Errorf("%d questions sent", len(prompts))
	}
}

func TestAccusingTheDemonUnmasksIt(t *testing.T) {
	var prompts []string
	m := villagersAt(t, &prompts)
	m, _ = m.Update(press("a"))
	m, _ = m.Update(press(string(rune('1' + m.eyes.demon))))
	w, err := m.store.Wallet()
	if err != nil || w.Tongqian != journey.MysteryBounty || !strings.Contains(m.View(), "Fight!") {
		t.Fatalf("unmasked: purse %+v %v\n%s", w, err, m.View())
	}
	m, cmd := m.Update(enter)
	if cmd == nil {
		t.Fatal("unmasking should start a fight")
	}
	if msg, ok := cmd().(roam.EncounterMsg); !ok || msg.Soldier.Name == "" {
		t.Fatalf("want a fight with a soldier, got %#v", cmd())
	}
	m, _ = m.Update(press("f"))
	if m.screen != screenTown || !strings.Contains(m.note, "tomorrow") {
		t.Errorf("a town's rumour is settled once a day: screen %d, %q", m.screen, m.note)
	}
}

func TestAccusingAnInnocentLetsTheDemonGo(t *testing.T) {
	var prompts []string
	m := villagersAt(t, &prompts)
	cleared, err := m.store.ClearedStages()
	if err != nil {
		t.Fatal(err)
	}
	before, _, err := journey.LoadTrail(m.store, cleared)
	if err != nil {
		t.Fatal(err)
	}
	m, _ = m.Update(press("a"))
	m, _ = m.Update(press(string(rune('1' + (m.eyes.demon+1)%3))))
	after, _, err := journey.LoadTrail(m.store, cleared)
	if err != nil || after.Morale[journey.Master] != before.Morale[journey.Master]-journey.MysteryMorale {
		t.Fatalf("morale %d -> %d (%v)", before.Morale[journey.Master], after.Morale[journey.Master], err)
	}
	if w, err := m.store.Wallet(); err != nil || w.Tongqian != 0 || !strings.Contains(m.View(), "slips away") {
		t.Fatalf("a wrong accusation paid %d (%v):\n%s", w.Tongqian, err, m.View())
	}
	m, cmd := m.Update(enter)
	if cmd != nil || m.screen != screenTown {
		t.Errorf("a wrong accusation should go back to town with no fight")
	}
}

func TestFieryEyesScreen(t *testing.T) {
	var prompts []string
	m := villagersAt(t, &prompts)
	m = askEyes(t, m, 1, "你昨天晚上在哪儿？")
	m, _ = m.Update(esc)
	m, _ = m.Update(press("e"))
	uitest.Golden(t, "eyes", m.View())
}
