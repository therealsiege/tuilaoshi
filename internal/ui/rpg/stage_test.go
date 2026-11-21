package rpg

import (
	"fmt"
	"strings"
	"testing"

	"github.com/therealsiege/tuilaoshi/internal/journey"
	"github.com/therealsiege/tuilaoshi/internal/ui/uitest"
	"github.com/therealsiege/tuilaoshi/internal/wiki"
)

// stageWiki is a test wiki with mastered characters and stages 1..cleared won.
func stageWiki(t *testing.T, mastered, cleared int) *wiki.Store {
	t.Helper()
	pages := map[string]string{}
	for i := 0; i < mastered+4; i++ {
		interval := 7
		if i >= mastered {
			interval = 1
		}
		ch := string(rune(0x4E00 + i))
		pages[fmt.Sprintf("characters/c%d.md", i)] = fmt.Sprintf(
			"---\ntype: character\nreview_interval: %d\n---\n\n# %s\n\n- **Character**: %s\n- **Pinyin**: p%d\n- **Meaning**: meaning %d\n", interval, ch, ch, i, i)
	}
	if cleared > 0 {
		var b strings.Builder
		b.WriteString("cleared:\n")
		for i := 1; i <= cleared; i++ {
			fmt.Fprintf(&b, "    %d: \"2026-11-21\"\n", i)
		}
		pages["journey.yaml"] = b.String()
	}
	return uitest.Wiki(t, pages)
}

func TestWinningOpenStageClearsIt(t *testing.T) {
	m := skipScene(New(stageWiki(t, 0, 0), nil, nil).Refresh())
	if m.stage == nil || m.stage.Number != 1 || m.battle.Enemy.Chinese != m.stage.Opponent.Name {
		t.Fatalf("new learner: stage %v, enemy %s; want the stage 1 fight", m.stage, m.battle.Enemy.Chinese)
	}
	m, _ = enter(m)
	m, _ = right(m)
	enemyDown(&m)
	m, _ = animationEnds(m)
	if m.state != stateVictory || m.saveErr != nil {
		t.Fatalf("state %s, save error %v", stateNames[m.state], m.saveErr)
	}
	if !strings.Contains(m.View(), "Tribulation 1 cleared") {
		t.Errorf("victory screen doesn't say the stage cleared:\n%s", m.View())
	}

	m = skipScene(m.Refresh()) // stage 2 needs characters this learner lacks
	if m.stage != nil || m.next == nil || m.next.Number != 2 {
		t.Fatalf("after clearing 1 with nothing mastered: stage %v, next %v; want practice before stage 2", m.stage, m.next)
	}
	if v := m.View(); !strings.Contains(v, "practice") || !strings.Contains(v, "to open tribulation 2") {
		t.Errorf("practice intro doesn't explain what opens the next stage:\n%s", v)
	}
}

func TestBossStageIsABossBattle(t *testing.T) {
	need := journey.TribulationAt(8).Unlock
	m := skipScene(New(stageWiki(t, need, 7), nil, nil).Refresh())
	if m.stage == nil || m.stage.Number != 8 || m.stage.Boss == nil {
		t.Fatalf("stage %+v; want the open boss stage 8", m.stage)
	}
	normal := skipScene(New(stageWiki(t, need, 0), nil, nil).Refresh()) // stage 1: same demon, not a boss
	if m.battle.Enemy.MaxHP <= normal.battle.Enemy.MaxHP {
		t.Errorf("boss HP %d not above the same demon's %d on an ordinary stage", m.battle.Enemy.MaxHP, normal.battle.Enemy.MaxHP)
	}
	if !strings.Contains(m.View(), "BOSS!") {
		t.Errorf("boss intro missing:\n%s", m.View())
	}
}
