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
	if m.stage == nil || m.stage.Number != 1 || len(m.gauntlet) == 0 || m.battle.Enemy.Chinese != m.gauntlet[0].Name {
		t.Fatalf("new learner: stage %v, enemy %s; want stage 1's first soldier", m.stage, m.battle.Enemy.Chinese)
	}
	// Each soldier, then the demon; only the demon's defeat clears the stage.
	for i := 0; i <= len(m.gauntlet); i++ {
		m, _ = enter(m)
		m, _ = right(m)
		enemyDown(&m)
		m, _ = animationEnds(m)
		if m.state != stateVictory || m.saveErr != nil {
			t.Fatalf("fight %d: state %s, save error %v", i+1, stateNames[m.state], m.saveErr)
		}
		if m.earned.Tongqian <= 0 {
			t.Errorf("fight %d paid no copper", i+1)
		}
		if i == len(m.gauntlet) {
			break
		}
		if cleared, err := m.store.ClearedStages(); err != nil || len(cleared) > 0 {
			t.Fatalf("stage cleared after soldier %d of %d", i+1, len(m.gauntlet))
		}
		hurt := m.battle.Player.MaxHP / 2
		m.battle.Player.HP = hurt
		m, _ = enter(m)
		if m.state != stateIntro || m.battle.Player.HP != hurt+m.battle.Player.MaxHP*restShare/100 {
			t.Fatalf("next foe: state %s, HP %d; want the intro with HP carried from %d plus a rest", stateNames[m.state], m.battle.Player.HP, hurt)
		}
	}
	if m.battle.Enemy.Chinese != m.stage.Opponent.Name {
		t.Fatalf("last fight is against %s, want %s", m.battle.Enemy.Chinese, m.stage.Opponent.Name)
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
	m.foe, normal.foe = len(m.gauntlet), len(normal.gauntlet)
	m.battle, normal.battle = m.stageBattle(), normal.stageBattle()
	if m.battle.Enemy.MaxHP <= normal.battle.Enemy.MaxHP {
		t.Errorf("boss HP %d not above the same demon's %d on an ordinary stage", m.battle.Enemy.MaxHP, normal.battle.Enemy.MaxHP)
	}
	if !strings.Contains(m.View(), "BOSS!") {
		t.Errorf("boss intro missing:\n%s", m.View())
	}
	m, _ = enter(m)
	m, _ = right(m)
	enemyDown(&m)
	m, _ = animationEnds(m)
	if m.earned.Yuanbao != 1 || m.wallet.Yuanbao != 1 {
		t.Errorf("beating the boss paid %+v, purse %+v; want a silver ingot", m.earned, m.wallet)
	}
}
