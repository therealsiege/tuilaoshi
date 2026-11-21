package rpg

import (
	"testing"

	"github.com/therealsiege/tuilaoshi/internal/journey"
)

func TestBeatingBossRecordsVictory(t *testing.T) {
	m := fixedBattle(t)
	unlockBoss(&m)
	boss := *m.boss

	m, _ = pressB(m)
	if !m.bossFight || m.battle.Enemy.Chinese != boss.Name {
		t.Fatalf("after b: bossFight=%v enemy=%s, want a fight with %s", m.bossFight, m.battle.Enemy.Chinese, boss.Name)
	}
	m, _ = enter(m)
	m, _ = right(m)
	enemyDown(&m)
	m, _ = animationEnds(m)
	if m.state != stateVictory || m.saveErr != nil {
		t.Fatalf("state %s, save error %v; want a saved victory", stateNames[m.state], m.saveErr)
	}

	victories, err := m.store.BossVictories()
	if err != nil {
		t.Fatal(err)
	}
	if victories[boss.Name] == "" {
		t.Fatalf("victory over %s not recorded: %v", boss.Name, victories)
	}
	if next := journey.NextBoss(journey.Bosses(99999, victories)); next == nil || next.Name == boss.Name {
		t.Errorf("next boss is still %v after beating %s", next, boss.Name)
	}
}

func TestBossChallengeNeedsUnlockedBoss(t *testing.T) {
	m := fixedBattle(t) // fixture wiki masters no characters
	m, _ = pressB(m)
	if m.bossFight {
		t.Error("b started a boss fight with no boss unlocked")
	}
}
