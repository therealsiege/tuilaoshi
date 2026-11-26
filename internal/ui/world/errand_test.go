package world

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/therealsiege/tuilaoshi/internal/claude"
	"github.com/therealsiege/tuilaoshi/internal/journey"
	"github.com/therealsiege/tuilaoshi/internal/ui/uitest"
	"github.com/therealsiege/tuilaoshi/internal/wiki"
)

// innErrand is the inn errand, whose points are room, five and water.
func innErrand(t *testing.T) journey.Errand {
	t.Helper()
	for _, e := range journey.Errands() {
		if e.ID == "inn-room" {
			return e
		}
	}
	t.Fatal("no inn-room errand")
	return journey.Errand{}
}

func TestParseErrand(t *testing.T) {
	e := innErrand(t)
	zh, en, done := parseErrand("好的，一间房。\n几位？\nEN: Fine, one room. How many of you?\nDONE: room, Five ,hat", e)
	if zh != "好的，一间房。 几位？" || en != "Fine, one room. How many of you?" || strings.Join(done, " ") != "room five" {
		t.Errorf("%q %q %v", zh, en, done)
	}
	if _, _, done := parseErrand("客官请说。\nEN: Go on.\nDONE: none", e); len(done) != 0 {
		t.Errorf("none done: %v", done)
	}
	if _, _, done := parseErrand("好。\nDONE: room，water，room", e); strings.Join(done, " ") != "room water" {
		t.Errorf("Chinese commas and repeats: %v", done)
	}
}

// errandWorld is a town with the inn errand open-able and a person who
// replies with the next of replies (or "听不懂" once they run out).
func errandWorld(t *testing.T, replies ...string) Model {
	t.Helper()
	m := worldAt(t, 11)
	for i := 0; journey.ErrandFor("changan", m.now()).ID != "inn-room"; i++ {
		d := day.AddDate(0, 0, i)
		m.now = func() time.Time { return d }
	}
	m.talker = func(_ context.Context, system string, _ []claude.ChatMessage, _ string) (string, error) {
		if !strings.Contains(system, "- five: ") || !strings.Contains(system, "DONE:") {
			t.Errorf("the innkeeper's prompt: %s", system)
		}
		if len(replies) == 0 {
			return "听不懂。\nEN: I don't follow.\nDONE: none", nil
		}
		r := replies[0]
		replies = replies[1:]
		return r, nil
	}
	return m
}

// say types a line and lets the person answer.
func say(t *testing.T, m Model, line string) Model {
	t.Helper()
	m.input.SetValue(line)
	m, cmd := m.Update(enter)
	if cmd == nil {
		t.Fatalf("nothing said for %q", line)
	}
	m, _ = m.Update(cmd())
	return m
}

func TestAnErrandChecksOffPointsAndPaysOnce(t *testing.T) {
	m := errandWorld(t,
		"好，住店。几位？\nEN: Right, a room. How many?\nDONE: room",
		"五位，好的。\nEN: Five, fine.\nDONE: room, five",
		"热水马上来！\nEN: Hot water coming!\nDONE: room, five, water",
		"慢走！\nEN: Take care!\nDONE: room, five, water")
	m.width, m.height = 0, 0
	m, _ = m.Update(press("e"))
	if m.screen != screenErrand || !m.Typing() || !strings.Contains(m.View(), "王掌柜") {
		t.Fatalf("the errand didn't open:\n%s", m.View())
	}
	m = say(t, m, "掌柜的，我们要住店。")
	if !m.errand.done["room"] || m.errand.done["five"] || m.errand.finished {
		t.Fatalf("after one line: %v", m.errand.done)
	}
	m = say(t, m, "我们一共五个人。")
	m = m.SetSize(80, 24)
	uitest.Golden(t, "errand-80x24", m.View())
	m = m.SetSize(120, 45)
	uitest.Golden(t, "errand-120x45", m.View())
	if !strings.Contains(m.Hear(), "五位") {
		t.Errorf("ctrl+r reads the innkeeper's last words: %q", m.Hear())
	}
	m.withTrail(func(tr *journey.Trail, _ []string) { tr.Morale[journey.Master] = 40 })
	m = say(t, m, "请给我们一些热水。")
	w, err := m.store.Wallet()
	if err != nil || !m.errand.finished || w.Tongqian != 30 || m.Typing() || !strings.Contains(m.View(), "任务完成") {
		t.Fatalf("all done: purse %+v %v\n%s", w, err, m.View())
	}
	tr, _, err := journey.LoadTrail(m.store, nil)
	if err != nil || tr.Morale[journey.Master] != 40+journey.ErrandMorale {
		t.Errorf("morale %v %v", tr.Morale, err)
	}
	m = m.onErrand(errandMsg{reply: "慢走！\nDONE: room, five, water"}) // a stray reply pays nothing more
	m.input.SetValue("谢谢！")
	if m, cmd := m.Update(enter); cmd != nil || m.screen != screenTown {
		t.Errorf("after the errand enter goes back to town: screen %d", m.screen)
	}
	if w, err := m.store.Wallet(); err != nil || w.Tongqian != 30 {
		t.Errorf("paid twice: %+v", w)
	}
	deeds, err := m.store.Deeds()
	if err != nil || len(deeds) != 1 || deeds[0].Kind != wiki.DeedErrand {
		t.Errorf("deeds %+v %v", deeds, err)
	}
}

func TestAnErrandRunsOutOfTurns(t *testing.T) {
	m := errandWorld(t, "什么？\nEN: What?\nDONE: room")
	m, _ = m.Update(press("e"))
	for i := range journey.ErrandTurns {
		if !m.Typing() {
			t.Fatalf("the box closed after %d lines", i)
		}
		m = say(t, m, "嗯……")
	}
	if !m.errand.over() || m.errand.finished || m.Typing() || !strings.Contains(m.View(), "someone else to see to") {
		t.Fatalf("after %d lines:\n%s", journey.ErrandTurns, m.View())
	}
	if w, err := m.store.Wallet(); err != nil || w.Tongqian != 0 {
		t.Errorf("an unfinished errand paid %+v", w)
	}
}

func TestAFailedCallCostsNoTurn(t *testing.T) {
	m := errandWorld(t)
	m.talker = func(context.Context, string, []claude.ChatMessage, string) (string, error) {
		return "", errors.New("offline")
	}
	m, _ = m.Update(press("e"))
	m = say(t, m, "你好")
	if m.errand.turns != 0 || len(m.errand.lines) != 1 || !strings.Contains(m.note, "offline") {
		t.Errorf("turns %d, lines %d, note %q", m.errand.turns, len(m.errand.lines), m.note)
	}
}

func TestAnErrandADayPerTownAndNoKey(t *testing.T) {
	m := errandWorld(t)
	m.talker = nil
	if m, _ = m.Update(press("e")); m.screen != screenTown || !strings.Contains(m.note, "ANTHROPIC_API_KEY") {
		t.Fatalf("no Claude: screen %d %q", m.screen, m.note)
	}
	if taken, err := m.store.ErrandTaken("changan", m.now()); err != nil || taken {
		t.Error("asking without a key spent the day's errand")
	}
	m = errandWorld(t)
	m, _ = m.Update(press("e"))
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if m, _ = m.Update(press("e")); m.screen != screenTown || !strings.Contains(m.note, "tomorrow") {
		t.Fatalf("a second errand the same day: screen %d %q", m.screen, m.note)
	}
	if m, _ = m.OpenTown("gao-village").Update(press("e")); m.screen != screenErrand {
		t.Errorf("another town's errand is closed: %q", m.note)
	}
}

func TestErrandScreensFitEightyColumns(t *testing.T) {
	m := errandWorld(t).SetSize(80, 24)
	long := "我们是从东土大唐来的和尚，要去西天取经，今天晚上想在这里住一晚，请问还有没有房间？"
	for _, er := range journey.Errands() {
		e := &errandTalk{errand: er, done: map[string]bool{er.Points[0].ID: true}, lines: []haggleLine{{true, er.Opening, er.OpeningEnglish}}}
		for range journey.ErrandTurns {
			e.lines = append(e.lines, haggleLine{false, long, ""}, haggleLine{true, long, "We are monks from the Tang lands in the east, on the way west for the scriptures."})
		}
		m.errand, m.screen = e, screenErrand
		for _, state := range []string{"talking", "waiting", "failed", "finished"} {
			e.waiting, e.finished, e.turns = state == "waiting", state == "finished", 3
			if state == "failed" {
				e.turns = journey.ErrandTurns
			}
			out := m.View()
			for _, l := range strings.Split(out, "\n") {
				if lipgloss.Width(l) > 80 {
					t.Errorf("%s %s: a line %d wide: %s", er.ID, state, lipgloss.Width(l), l)
				}
			}
			if n := strings.Count(out, "\n") + 1; n > 24 {
				t.Errorf("%s %s takes %d lines:\n%s", er.ID, state, n, out)
			}
		}
	}
}
