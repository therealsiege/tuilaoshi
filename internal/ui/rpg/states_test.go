package rpg

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/therealsiege/tuilaoshi/internal/journey"
	"github.com/therealsiege/tuilaoshi/internal/ui"
)

// The battle state chart in docs/battle-states.md is generated from this
// table, and each row is played against the real model, so the chart can't
// drift from the code. Run make golden to regenerate the chart.

var stateNames = map[viewState]string{
	stateIntro:        "Intro",
	stateFighting:     "Fighting",
	statePlayerAttack: "PlayerAttack",
	stateEnemyAttack:  "EnemyAttack",
	stateVictory:      "Victory",
	stateDefeat:       "Defeat",
	stateScene:        "Scene",
}

const outside viewState = -1 // [*]: opening the battle, or leaving to the dashboard

type transition struct {
	from, to viewState
	event    string
	setup    func(*Model) // adjust HP etc. before the event
	act      func(Model) (Model, tea.Cmd)
}

var (
	enter = func(m Model) (Model, tea.Cmd) { return m.Update(tea.KeyMsg{Type: tea.KeyEnter}) }
	esc   = func(m Model) (Model, tea.Cmd) { return m.Update(tea.KeyMsg{Type: tea.KeyEsc}) }
	tab   = func(m Model) (Model, tea.Cmd) { return m.Update(tea.KeyMsg{Type: tea.KeyTab}) }
	right = func(m Model) (Model, tea.Cmd) {
		m.battle.Selected = m.battle.Question.Correct
		return m.submitAnswer()
	}
	wrong = func(m Model) (Model, tea.Cmd) {
		m.battle.Selected = (m.battle.Question.Correct + 1) % len(m.battle.Question.Choices)
		return m.submitAnswer()
	}
	// animationEnds sends ticks until the attack animation finishes.
	animationEnds = func(m Model) (Model, tea.Cmd) {
		var cmd tea.Cmd
		for i := 0; i <= playerAttackTicks && m.state == statePlayerAttack; i++ {
			m, cmd = m.Update(animTickMsg{})
		}
		return m, cmd
	}
	// The hit has already landed by the attack states, so these set the result.
	enemyDown  = func(m *Model) { m.battle.Enemy.HP = 0 }
	playerDown = func(m *Model) { m.battle.Player.HP = 0 }
	// moreFoes puts the beaten enemy first in a gauntlet before the stage's demon.
	onTheRoad = func(m *Model) { m.road = true }
	moreFoes  = func(m *Model) {
		st := journey.Stages(journey.Counts{}, nil)[0]
		m.stage, m.gauntlet, m.foe = &st, journey.SoldiersAt(1)[:1], 0
	}
)

var transitions = []transition{
	{outside, stateScene, "open battle on a stage whose story scene is unseen", nil, nil},
	{outside, stateIntro, "open battle (the open stage's demon or boss, else practice)", nil, nil},
	{stateScene, stateIntro, "tab (scene marked seen)", nil, tab},
	{stateScene, outside, "esc (flee)", nil, esc},
	{stateIntro, stateFighting, "enter / space", nil, enter},
	{stateFighting, statePlayerAttack, "right answer", nil, right},
	{stateFighting, stateEnemyAttack, "wrong answer", nil, wrong},
	{stateFighting, stateEnemyAttack, "time runs out (burning enemy)", nil, timeRunsOut},
	{statePlayerAttack, stateFighting, "animation ends, enemy alive", nil, animationEnds},
	{statePlayerAttack, stateVictory, "animation ends, enemy HP 0 (saves; clears the stage)", enemyDown, animationEnds},
	{stateEnemyAttack, stateFighting, "enter / space, player alive", nil, enter},
	{stateEnemyAttack, stateDefeat, "enter / space, player HP 0 (saves)", playerDown, enter},
	{stateDefeat, stateFighting, "enter / space, sparring at camp (rematch)", nil, enter},
	{stateDefeat, outside, "enter / space, on the road (wake at the last town, half HP)", onTheRoad, enter},
	{stateVictory, stateIntro, "enter / space, more of the stage's foes to fight (HP carries)", moreFoes, enter},
	{stateVictory, outside, "enter / space, last foe beaten", nil, enter},
	{stateFighting, outside, "esc (flee, any state)", nil, esc},
}

// reach drives a fresh battle into state s through the table's own rows.
func reach(t *testing.T, s viewState, setup func(*Model)) Model {
	t.Helper()
	m := fixedBattle(t)
	path := map[viewState][]func(Model) (Model, tea.Cmd){
		stateIntro:        nil,
		stateScene:        {func(m Model) (Model, tea.Cmd) { return m.showScene(journey.SceneFor(1)), nil }},
		stateFighting:     {enter},
		statePlayerAttack: {enter, right},
		stateEnemyAttack:  {enter, wrong},
		stateVictory:      {enter, right, func(m Model) (Model, tea.Cmd) { enemyDown(&m); return animationEnds(m) }},
		stateDefeat:       {enter, wrong, func(m Model) (Model, tea.Cmd) { playerDown(&m); return enter(m) }},
	}[s]
	for _, step := range path {
		m, _ = step(m)
	}
	if m.state != s {
		t.Fatalf("could not reach %s: at %s", stateNames[s], stateNames[m.state])
	}
	if setup != nil {
		setup(&m)
	}
	return m
}

func TestBattleTransitions(t *testing.T) {
	for _, tr := range transitions {
		if tr.act == nil {
			continue // entry from outside: covered by fixedBattle starting at Intro
		}
		t.Run(chartName(tr.from)+"->"+chartName(tr.to), func(t *testing.T) {
			m := reach(t, tr.from, tr.setup)
			m, cmd := tr.act(m)
			if tr.to == outside {
				if cmd == nil {
					t.Fatal("no command; want switch to the map")
				}
				if msg, ok := cmd().(ui.SwitchViewMsg); !ok || msg.Target != ui.ViewMap {
					t.Fatalf("got %#v, want switch to the map", cmd())
				}
				return
			}
			if m.state != tr.to {
				t.Fatalf("%s on %q: got %s, want %s", stateNames[tr.from], tr.event, stateNames[m.state], stateNames[tr.to])
			}
		})
	}
}

func chartName(s viewState) string {
	if s == outside {
		return "[*]"
	}
	return stateNames[s]
}

func TestBattleStateChartIsCurrent(t *testing.T) {
	var b strings.Builder
	b.WriteString("# Battle state chart\n\n")
	b.WriteString("Generated from `internal/ui/rpg/states_test.go`, which also plays every\n")
	b.WriteString("transition against the battle model. Regenerate with `make golden`.\n\n")
	b.WriteString("```mermaid\nstateDiagram-v2\n")
	for _, tr := range transitions {
		fmt.Fprintf(&b, "    %s --> %s : %s\n", chartName(tr.from), chartName(tr.to), tr.event)
	}
	b.WriteString("```\n")

	path := filepath.Join("..", "..", "..", "docs", "battle-states.md")
	if os.Getenv("UPDATE_GOLDEN") != "" {
		if err := os.WriteFile(path, []byte(b.String()), 0644); err != nil {
			t.Fatal(err)
		}
		return
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("%v (run make golden)", err)
	}
	if string(want) != b.String() {
		t.Errorf("%s is stale; run make golden", path)
	}
}
