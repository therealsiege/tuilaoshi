package rpg

import (
	"strings"
	"testing"
	"time"

	"github.com/therealsiege/tuilaoshi/internal/journey"
	"github.com/therealsiege/tuilaoshi/internal/model"
)

func TestEffectsForEnemyAndJoinedParty(t *testing.T) {
	enemy := journey.BestiaryEntry{Name: "黄风怪", Abilities: []string{"blind"}}
	party := []journey.Companion{{Name: "孙悟空", Effect: "reveal", Unlocked: true}, {Name: "沙悟净", Effect: "shield"}}
	e, lines := EffectsFor(enemy, party)
	if !e.HidePinyin || e.Reveals != 1 || e.Shields != 0 {
		t.Errorf("effects %+v: want blind and one reveal; Sha Wujing hasn't joined", e)
	}
	if len(lines) != 2 || !strings.HasPrefix(lines[0], "黄风怪 blinds") {
		t.Errorf("intro lines %q", lines)
	}
}

func TestBlindHidesPinyin(t *testing.T) {
	b := testBattle(testDeck())
	b.Writing = true
	b.SetEffects(Effects{HidePinyin: true})
	for i := 0; i < 300; i++ {
		q := b.GenerateQuestion(testDeck())
		if q.Category == "pinyin" || q.Category == "character" || (q.Category == "write" && strings.Contains(q.Prompt, q.Pinyin)) {
			t.Fatalf("blind battle shows pinyin: %s %q", q.Category, q.Prompt)
		}
	}
}

func TestDisguisePrefersLookAlikes(t *testing.T) {
	deck := []model.Card{card("1", "大", "dà", "big"), card("2", "太", "tài", "too"), card("3", "犬", "quǎn", "dog"),
		card("4", "书", "shū", "book"), card("5", "水", "shuǐ", "water"), card("6", "山", "shān", "mountain")}
	for i := 0; i < 50; i++ {
		got := generateDistractors(deck[0], newDeck(deck), "reading", 2, true)
		if len(got) != 2 || !strings.Contains("太犬", got[0]) || !strings.Contains("太犬", got[1]) {
			t.Fatalf("disguised distractors for 大 = %v, want 太 and 犬", got)
		}
	}
}

func TestMissEffects(t *testing.T) {
	b := testBattle(testDeck())
	b.Enemy.Attack, b.Player.Defense = 40, 0
	b.SetEffects(Effects{MissDamage: 50, Shields: 1, HalfCombo: true})
	b.Combo = 4
	if _, dmg := b.resolve(false); dmg != 0 || b.Event == "" || b.Combo != 2 {
		t.Errorf("first miss: damage %d, event %q, combo %d; want blocked, combo halved to 2", dmg, b.Event, b.Combo)
	}
	if _, dmg := b.resolve(false); dmg < 61 {
		t.Errorf("venom miss hit for %d, want at least 61 (41+50%%)", dmg)
	}
	if r := b.Rematch(testDeck()); r.Effects.Shields != 1 {
		t.Error("a rematch should restore the shield")
	}
}

func TestFirstStrikeDoublesOnlyTheFirstHit(t *testing.T) {
	b := testBattle(testDeck())
	b.SetEffects(Effects{FirstStrike: true})
	b.Question = &CombatQuestion{Damage: 20}
	_, first := b.resolve(true)
	_, second := b.resolve(true)
	if first != 2*second {
		t.Errorf("first hit %d, second %d; want the first doubled", first, second)
	}
}

func TestRevealLeavesTheRightAnswerAndOneWrong(t *testing.T) {
	b := testBattle(testDeck())
	b.SetEffects(Effects{Reveals: 1})
	b.Question = &CombatQuestion{Choices: []string{"a", "b", "c", "d"}, Correct: 2}
	if !b.Reveal() {
		t.Fatal("reveal refused")
	}
	q := b.Question
	if len(q.Choices) != 2 || q.CorrectAnswer() != "c" {
		t.Errorf("after reveal: %v correct %d", q.Choices, q.Correct)
	}
	if b.Reveal() {
		t.Error("a second reveal worked with none left")
	}
}

func TestPickTargetFavoursWordsThatNeedPractice(t *testing.T) {
	now := time.Date(2026, 11, 21, 12, 0, 0, 0, time.UTC)
	reviewed := now.AddDate(0, 0, -3)
	cards := []model.Card{
		{ID: "due", LastReviewed: reviewed, ReviewInterval: 2, NextReview: now.AddDate(0, 0, -1)},
		{ID: "shaky", LastReviewed: reviewed, ReviewInterval: 4, NextReview: now.AddDate(0, 0, 1)},
		{ID: "new"},
		{ID: "settled", LastReviewed: reviewed, ReviewInterval: 30, NextReview: now.AddDate(0, 0, 27)},
	}
	counts := map[string]int{}
	for i := 0; i < 16000; i++ {
		counts[pickTarget(cards, now).ID]++
	}
	if counts["due"] <= counts["shaky"] || counts["shaky"] <= counts["new"] || counts["new"] <= counts["settled"] {
		t.Errorf("picks %v; want due > shaky > new > settled", counts)
	}
	if counts["settled"] == 0 {
		t.Error("settled words never come up")
	}
}

func TestMeaningHintsExplainCharacters(t *testing.T) {
	b := testBattle(testDeck())
	b.Radicals = map[string]string{"灯": "火 (fire)", "笼": "⺮ (bamboo)", "火": "火 (fire)"}
	if got := b.meaningHints("灯笼"); len(got) != 2 || got[0] != "灯 ← 火 (fire)" {
		t.Errorf("hints for 灯笼 = %v", got)
	}
	if got := b.meaningHints("火"); len(got) != 0 {
		t.Errorf("火 is its own radical, got hints %v", got)
	}
}
