package rpg

import (
	"fmt"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"

	"github.com/therealsiege/tuilaoshi/internal/ui/uitest"
)

// baiguScreens are the disguises' screens worth a snapshot, each set up
// from a fresh battle at the given size.
var baiguScreens = []struct {
	name  string
	setup func(t *testing.T, m Model) Model
}{
	{"reading", func(_ *testing.T, m Model) Model { return m }},
	{"wrong", func(_ *testing.T, m Model) Model { return choose(m, false) }},
	{"listening", func(_ *testing.T, m Model) Model {
		m.bg.disguise = 1
		m, _ = m.askBaigu()
		m.bg.listen = true
		return m
	}},
	{"writing", func(_ *testing.T, m Model) Model {
		m.grader, m.bg.disguise = gradeWith(2, nil), 2
		m, _ = m.askBaigu()
		return m
	}},
	{"won", func(t *testing.T, m Model) Model {
		for range m.bg.trial.Disguises {
			m = choose(m, true)
			m, _ = pressKey(m, "enter")
		}
		if m.bg.phase != baiguWon {
			t.Fatalf("not won after three right answers: phase %d", m.bg.phase)
		}
		return m
	}},
}

func TestBaiguScreens(t *testing.T) {
	for _, size := range [][2]int{{80, 24}, {120, 45}} {
		for _, s := range baiguScreens {
			name := fmt.Sprintf("baigu-%s-%dx%d", s.name, size[0], size[1])
			t.Run(name, func(t *testing.T) {
				m := s.setup(t, baiguAt(t, size[0], size[1]))
				uitest.Golden(t, name, m.View())
			})
		}
	}
}

// TestBaiguScreensFitEightyColumns plays every round of every disguise
// through each of its phases at 80x24 and checks each screen fits.
func TestBaiguScreensFitEightyColumns(t *testing.T) {
	check := func(what string, m Model) {
		t.Helper()
		out := m.View()
		for _, l := range strings.Split(out, "\n") {
			if lipgloss.Width(l) > 80 {
				t.Errorf("%s: a line %d wide: %s", what, lipgloss.Width(l), l)
			}
		}
		if n := strings.Count(out, "\n") + 1; n > 24 {
			t.Errorf("%s takes %d lines:\n%s", what, n, out)
		}
	}
	base := baiguAt(t, 80, 24)
	long := strings.Repeat("Close, but say what is wrong with him, not what he looks like. ", 3)
	for d, dis := range base.bg.trial.Disguises {
		for v, r := range dis.Rounds {
			at := func(keyed bool) Model {
				m := base
				m.bg.variant = []int{0, 0, 0}
				m.bg.variant[d], m.bg.disguise = v, d
				if keyed {
					m.grader = gradeWith(0, nil)
				}
				m, _ = m.askBaigu()
				return m
			}
			what := fmt.Sprintf("%s round %s", dis.Name, r.ID)
			check(what+" asked", at(false))
			listen := at(false)
			listen.bg.listen = true
			check(what+" heard", listen)
			check(what+" right", choose(at(false), true))
			wrong := choose(at(false), false)
			check(what+" wrong", wrong)
			wrong.bg.english = true
			check(what+" wrong, English shown", wrong)
			if dis.Skill != "writing" {
				continue
			}
			typed := at(true)
			typed.bg.note = long
			check(what+" written, Claude unreachable", typed)
			typed.bg.typed, typed.bg.phase = "他看起来很奇怪，因为他的鞋子太干净了，而且没有影子", baiguGrading
			check(what+" grading", typed)
			typed.bg.feedback = long
			check(what+" written wrong", typed.judgeBaigu(false))
			check(what+" written right", typed.judgeBaigu(true))
		}
	}
	won := base
	for range won.bg.trial.Disguises {
		won = choose(won, true)
		won, _ = pressKey(won, "enter")
	}
	check("won", won)
}
