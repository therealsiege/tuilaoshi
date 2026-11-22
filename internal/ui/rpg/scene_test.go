package rpg

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/therealsiege/tuilaoshi/internal/journey"
	"github.com/therealsiege/tuilaoshi/internal/ui/uitest"
)

func TestStageSceneShowsOnceBeforeTheFight(t *testing.T) {
	m := New(stageWiki(t, 0, 0), nil, nil).SetSize(80, 30).RoadFight()
	if m.state != stateScene || m.sceneOf == nil || m.sceneOf.Stage != 1 {
		t.Fatalf("state %s; want stage 1's scene first", stateNames[m.state])
	}
	uitest.Golden(t, "scene", m.View())

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("e")})
	if !strings.Contains(m.View(), "Long, long ago") {
		t.Error("e did not show the translation")
	}
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRight})
	if m.state != stateScene {
		t.Fatal("moving through words left the scene")
	}
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	if m.state != stateIntro {
		t.Fatalf("tab: state %s, want the battle intro", stateNames[m.state])
	}
	if seen, err := m.store.SceneSeen(1); err != nil || !seen {
		t.Fatalf("scene not recorded as seen: %v %v", seen, err)
	}
	if m = m.RoadFight(); m.state != stateIntro {
		t.Errorf("second visit: state %s; the scene shows only once", stateNames[m.state])
	}
}

func TestStageWithoutSceneDescribesItsTribulation(t *testing.T) {
	t9 := journey.TribulationAt(9)
	m := New(stageWiki(t, t9.Unlock, 8), nil, nil).SetSize(80, 30).RoadFight()
	if m.state != stateIntro || m.stage == nil || m.stage.Number != 9 {
		t.Fatalf("state %s, stage %v; want stage 9's intro (it has no scene)", stateNames[m.state], m.stage)
	}
	if v := uitest.Plain(m.View()); !strings.Contains(strings.Join(strings.Fields(v), " "), strings.Join(strings.Fields(t9.Desc)[:6], " ")) {
		t.Errorf("intro doesn't describe tribulation 9:\n%s", v)
	}
}
