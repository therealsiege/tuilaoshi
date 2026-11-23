package rpg

import (
	"fmt"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/therealsiege/tuilaoshi/internal/journey"
	"github.com/therealsiege/tuilaoshi/internal/model"
	rpgengine "github.com/therealsiege/tuilaoshi/internal/rpg"
	"github.com/therealsiege/tuilaoshi/internal/ui"
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
	m := skipScene(New(stageWiki(t, 0, 0), nil, nil).RoadFight())
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

	m = skipScene(m.Spar()) // stage 2 needs characters this learner lacks
	if m.stage != nil || m.next == nil || m.next.Number != 2 {
		t.Fatalf("after clearing 1 with nothing mastered: stage %v, next %v; want practice before stage 2", m.stage, m.next)
	}
	if v := m.View(); !strings.Contains(v, "practice") || !strings.Contains(v, "to open tribulation 2") {
		t.Errorf("practice intro doesn't explain what opens the next stage:\n%s", v)
	}
}

func TestBossStageIsABossBattle(t *testing.T) {
	need := journey.TribulationAt(8).Unlock
	m := skipScene(New(stageWiki(t, need, 7), nil, nil).RoadFight())
	if m.stage == nil || m.stage.Number != 8 || m.stage.Boss == nil {
		t.Fatalf("stage %+v; want the open boss stage 8", m.stage)
	}
	normal := skipScene(New(stageWiki(t, need, 0), nil, nil).RoadFight()) // stage 1: same demon, not a boss
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

func TestReplayingAnEasyPracticeWinPaysNothing(t *testing.T) {
	win := func(m Model) Model {
		m.stage = nil // practice
		q := m.battle.Question
		m, _ = enter(m)
		m.battle.Question = q // always the same word
		m, _ = right(m)
		enemyDown(&m)
		m, _ = animationEnds(m)
		return m
	}
	m := win(fixedBattle(t))
	if m.earned.Tongqian != 1+m.battle.WinCoins() {
		t.Fatalf("first win paid %+v, want a coin for the word and %d for the win", m.earned, m.battle.WinCoins())
	}
	again := fixedBattle(t)
	again.store = m.store
	if again = win(again); again.earned != (wiki.Wallet{}) {
		t.Errorf("the same word won again paid %+v, want nothing", again.earned)
	}
}

func TestShakyRecentAnswersMakeBattlesGentler(t *testing.T) {
	store := stageWiki(t, 0, 0)
	sess := model.Session{StartedAt: time.Now(), EndedAt: time.Now(), Type: "drill"}
	for i := 0; i < 30; i++ {
		sess.Items = append(sess.Items, model.SessionItem{CardID: fmt.Sprint(i), Correct: i < 15})
	}
	sess.CardsTotal, sess.CardsCorrect = 30, 15
	if err := store.WriteSession(sess); err != nil {
		t.Fatal(err)
	}
	m := skipScene(New(store, nil, nil).RoadFight())
	if m.difficulty != rpgengine.Easier || m.battle.Difficulty != rpgengine.Easier || m.battle.Effects.Reveals == 0 {
		t.Fatalf("after 50%% right: difficulty %d, battle %d, reveals %d", m.difficulty, m.battle.Difficulty, m.battle.Effects.Reveals)
	}
	if !strings.Contains(m.View(), "50% right") {
		t.Errorf("intro doesn't explain the gentler battle:\n%s", m.View())
	}
}

func TestAnAssistThatHelpsRaisesTheBond(t *testing.T) {
	m := fixedBattle(t)
	m.stage, m.cleared = nil, map[int]string{8: "2026-11-01"} // 孙悟空 has joined
	m.applyAbilities(m.cleared)
	q := m.battle.Question
	m, _ = enter(m)
	m.battle.Question = q
	m = m.reveal()
	m, _ = right(m)
	enemyDown(&m)
	m, _ = animationEnds(m)
	if len(m.bonded) != 1 || m.bonded[0] != "孙悟空" {
		t.Fatalf("bonded %v, want 孙悟空 for the reveal", m.bonded)
	}
	if bonds, err := m.store.BondPoints(); err != nil || bonds["孙悟空"] != 1 {
		t.Errorf("bond points %v %v", bonds, err)
	}
	if !strings.Contains(m.View(), "Bond +1: 孙悟空") {
		t.Errorf("victory screen doesn't show the bond:\n%s", m.View())
	}
}

func TestMasteredFamilyWakesItsSpirit(t *testing.T) {
	pages := map[string]string{}
	for i := 0; i < journey.SpiritWakes+2; i++ {
		ch := string(rune(0x6C5F + i))
		pages[fmt.Sprintf("characters/w%d.md", i)] = fmt.Sprintf(
			"---\ntype: character\nreview_interval: 7\n---\n\n# %s\n\n- **Character**: %s\n- **Pinyin**: p%d\n- **Meaning**: water word %d\n- **Radical**: 氵(water)\n", ch, ch, i, i)
	}
	pages["journey.yaml"] = "cleared: {1: \"2026-11-01\", 2: \"2026-11-01\"}\n" // stage 3 needs 16: practice
	defer func(roll func() int) { spiritRoll = roll }(spiritRoll)
	spiritRoll = func() int { return 0 }
	m := skipScene(New(uitest.Wiki(t, pages), nil, nil).RoadFight())
	if m.spirit == nil || m.battle.Enemy.Chinese != "河伯" {
		t.Fatalf("no spirit: enemy %s", m.battle.Enemy.Chinese)
	}
	if v := m.View(); !strings.Contains(v, "hidden spirit") || !strings.Contains(v, "氵 characters") {
		t.Errorf("intro doesn't announce the spirit:\n%s", v)
	}
	m, _ = enter(m)
	m, _ = right(m)
	enemyDown(&m)
	m, _ = animationEnds(m)
	if m.earned.Yuanbao != 1 {
		t.Errorf("beating the spirit paid %+v, want an ingot", m.earned)
	}
	again := skipScene(New(m.store, nil, nil).RoadFight())
	if again.spirit != nil {
		t.Error("a beaten spirit appeared again")
	}
}

func TestNightmareRematchOfABeatenBoss(t *testing.T) {
	m := skipScene(New(stageWiki(t, 0, 8), nil, nil).RoadFight()) // boss of stage 8 beaten; practice now
	if m.stage != nil || !strings.Contains(m.View(), "n nightmare rematch") {
		t.Fatalf("practice intro doesn't offer a nightmare:\n%s", m.View())
	}
	normal := m.battle.Enemy.MaxHP
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("n")})
	if m.nightmare == nil || m.battle.Enemy.Chinese != m.nightmare.Name || m.battle.Enemy.MaxHP <= normal {
		t.Fatalf("nightmare: %v, enemy %s HP %d", m.nightmare, m.battle.Enemy.Chinese, m.battle.Enemy.MaxHP)
	}
	if !strings.Contains(m.View(), "梦魇") {
		t.Errorf("intro doesn't say nightmare:\n%s", m.View())
	}
	for i := 0; i < 2; i++ {
		m, _ = enter(m)
		m, _ = right(m)
		enemyDown(&m)
		m, _ = animationEnds(m)
		if want := 1 - i; m.earned.Yuanbao != want {
			t.Errorf("nightmare win %d paid %d ingots, want %d", i+1, m.earned.Yuanbao, want)
		}
		m = m.rematch()
		m.state = stateIntro
	}
}

func TestRoadBattlesCarryWoundsAndAFallWakesYouInTown(t *testing.T) {
	store := stageWiki(t, 0, 4)
	if err := store.SetPlace(wiki.Place{Region: "changan", X: 10, Y: 6}); err != nil {
		t.Fatal(err)
	}
	if err := store.SetWounds(40); err != nil {
		t.Fatal(err)
	}
	m := New(store, nil, nil).Encounter(journey.SoldiersAt(5)[0])
	if m.stage != nil || m.state != stateIntro || m.battle.Enemy.Chinese != journey.SoldiersAt(5)[0].Name {
		t.Fatalf("encounter: stage %v, state %s, enemy %s", m.stage, stateNames[m.state], m.battle.Enemy.Chinese)
	}
	if m.battle.Player.HP != m.battle.Player.MaxHP-40 {
		t.Fatalf("HP %d of %d, want 40 wounds carried in", m.battle.Player.HP, m.battle.Player.MaxHP)
	}
	m, _ = enter(m)
	m, _ = wrong(m)
	playerDown(&m)
	m, _ = enter(m)
	if m.state != stateDefeat {
		t.Fatalf("state %s, want defeat", stateNames[m.state])
	}
	m, cmd := enter(m)
	if msg, ok := cmd().(ui.SwitchViewMsg); !ok || msg.Target != ui.ViewMap {
		t.Fatalf("a fall should return to the map, got %#v", cmd())
	}
	place, err := store.Place()
	if err != nil || place.Wounds != m.battle.Player.MaxHP/2 {
		t.Errorf("woke with wounds %d (%v), want half HP", place.Wounds, err)
	}
	rm := journey.BuildMap(journey.RegionByID("changan"))
	if _, town := rm.Towns[journey.Point{X: place.X, Y: place.Y}]; !town && (journey.Point{X: place.X, Y: place.Y}) != rm.Start {
		t.Errorf("woke at %v, want a town or the region's start", place)
	}
}

func TestSparringNeverClearsATribulation(t *testing.T) {
	m := skipScene(New(stageWiki(t, 0, 0), nil, nil).Spar())
	if m.stage != nil {
		t.Fatalf("sparring picked up tribulation %d", m.stage.Number)
	}
}

func TestRoadWinsCountTowardHunts(t *testing.T) {
	store := stageWiki(t, 0, 4)
	if err := store.AcceptHunt("changan-hunt"); err != nil {
		t.Fatal(err)
	}
	var bull journey.Soldier
	for _, s := range journey.Soldiers() {
		if s.Name == "特处士" {
			bull = s
		}
	}
	win := func(m Model) {
		t.Helper()
		m, _ = enter(m)
		m, _ = right(m)
		enemyDown(&m)
		if m, _ = animationEnds(m); m.state != stateVictory {
			t.Fatalf("state %s, want victory", stateNames[m.state])
		}
	}
	win(New(store, nil, nil).Encounter(bull))
	sparred := New(store, nil, nil).Spar()
	sparred.battle = rpgengine.NewSoldierBattle(sparred.cards, bull, 0)
	win(sparred) // sparring at camp doesn't count
	hunts, err := store.Hunts()
	if err != nil || hunts["changan-hunt"] != 1 {
		t.Errorf("hunt progress %v %v, want 1 from the road win only", hunts, err)
	}
}

func TestAWildEncounterIsNotPractice(t *testing.T) {
	m := New(stageWiki(t, 0, 0), nil, nil).Encounter(journey.SoldiersAt(1)[0])
	v := m.View()
	if strings.Contains(v, "practice") || strings.Contains(v, "0 more") || !strings.Contains(v, "leaps out on the road") {
		t.Errorf("encounter intro:\n%s", v)
	}
}

func TestLootOnlyFromRoadWinsThatRecallHardWords(t *testing.T) {
	win := func(m Model, interval int) Model {
		t.Helper()
		m.battle.Question = &rpgengine.CombatQuestion{Choices: []string{"a", "b"}, Correct: 0, Damage: 20}
		m.battle.Question.Card.ReviewInterval = interval
		m.battle.Question.Card.ID = fmt.Sprintf("w%d", interval)
		m.state = stateFighting
		m, _ = right(m)
		enemyDown(&m)
		m, _ = animationEnds(m)
		return m
	}
	store := stageWiki(t, 0, 4)
	if err := store.SetPlace(wiki.Place{Region: "changan"}); err != nil {
		t.Fatal(err)
	}
	minion := journey.SoldiersAt(5)[0]
	if m := win(New(store, nil, nil).Encounter(minion), 30); len(m.loot) != 0 {
		t.Errorf("an easy word earned loot %v", m.loot)
	}
	if m := win(New(store, nil, nil).Spar(), 1); len(m.loot) != 0 {
		t.Errorf("sparring earned loot %v", m.loot)
	}
	m := win(New(store, nil, nil).Encounter(minion), 1)
	if len(m.loot) != 1 || !strings.Contains(m.View(), "Loot:") {
		t.Fatalf("a hard word on the road earned %v:\n%s", m.loot, m.View())
	}
	if mats, err := store.Materials(); err != nil || mats[m.loot[0]] != 1 {
		t.Errorf("loot not kept: %v %v", mats, err)
	}
}
