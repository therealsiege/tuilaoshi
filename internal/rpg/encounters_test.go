package rpg

import (
	"fmt"
	"strings"
	"testing"

	"github.com/therealsiege/tuilaoshi/internal/journey"
	"github.com/therealsiege/tuilaoshi/internal/model"
)

func TestReachedDemonsGrowWithProgress(t *testing.T) {
	start, end := ReachedDemons(0), ReachedDemons(100)
	if len(start) == 0 {
		t.Fatal("a new learner meets no demons")
	}
	if len(end) != len(journey.Bestiary()) {
		t.Errorf("full progress reaches %d demons, want all %d", len(end), len(journey.Bestiary()))
	}
	for _, d := range start {
		if n := parseChapterNumber(d.Chapter); n > openChapters {
			t.Errorf("%s (chapter %d) met before reaching it", d.English, n)
		}
	}
	prev := 0
	for p := 0.0; p <= 100; p += 10 {
		if n := len(ReachedDemons(p)); n < prev {
			t.Errorf("progress %.0f%% reaches %d demons, fewer than before (%d)", p, n, prev)
		} else {
			prev = n
		}
	}
}

func TestBossBattleIsTougher(t *testing.T) {
	demon := journey.BestiaryEntry{Name: "白骨精", English: "White Bone Demon", Chapter: "Chapter 27"}
	normal, boss := NewBattle(testDeck(), demon, "", 0), NewBossBattle(testDeck(), demon, 0)
	if boss.Enemy.MaxHP != normal.Enemy.MaxHP*bossHPMultiplier || boss.Enemy.HP != boss.Enemy.MaxHP {
		t.Errorf("boss HP %d/%d, want %d full", boss.Enemy.HP, boss.Enemy.MaxHP, normal.Enemy.MaxHP*bossHPMultiplier)
	}
	if boss.Enemy.Attack <= normal.Enemy.Attack {
		t.Errorf("boss attack %d not above %d", boss.Enemy.Attack, normal.Enemy.Attack)
	}
}

func TestWeakCardsPreferShakyWords(t *testing.T) {
	var deck []model.Card
	for i := 0; i < minBossDeck+2; i++ {
		deck = append(deck, model.Card{ID: fmt.Sprint("weak", i), Chinese: fmt.Sprint("弱", i), English: "weak", ReviewInterval: 1})
		deck = append(deck, model.Card{ID: fmt.Sprint("solid", i), Chinese: fmt.Sprint("强", i), English: "strong", ReviewInterval: 30})
	}
	for _, c := range WeakCards(deck) {
		if c.ReviewInterval >= weakInterval {
			t.Errorf("boss deck includes settled word %s", c.ID)
		}
	}
	if got := WeakCards(deck[:4]); len(got) != 4 {
		t.Errorf("small deck: got %d cards, want all 4 when too few are weak", len(got))
	}
}

func TestBossEntryFindsBestiaryArt(t *testing.T) {
	for _, b := range journey.Bosses(0, nil) {
		if e := journey.BossEntry(b); e.Art == "" {
			t.Errorf("boss %s %s has no bestiary entry", b.Name, b.English)
		}
	}
}

func TestListeningQuestionsOnlyWhenAudioWorks(t *testing.T) {
	count := func(listening bool) int {
		b := testBattle(testDeck())
		b.Listening = listening
		n := 0
		for i := 0; i < 300; i++ {
			if q := b.GenerateQuestion(testDeck()); q != nil && q.Category == "listening" {
				n++
				if strings.Contains(q.Prompt, q.Chinese) || strings.Contains(q.Prompt, q.Pinyin) {
					t.Fatalf("listening prompt gives the answer away: %q", q.Prompt)
				}
			}
		}
		return n
	}
	if n := count(false); n != 0 {
		t.Errorf("%d listening questions without audio", n)
	}
	if n := count(true); n == 0 {
		t.Error("no listening questions with audio available")
	}
}

func TestWriteQuestionsGradeTypedChinese(t *testing.T) {
	deck := append(testDeck(), card("6", "書", "shū", "book (traditional)"))
	b := testBattle(deck)
	b.Writing = true
	var q *CombatQuestion
	for i := 0; i < 300 && (q == nil || !q.written() || q.Chinese != "书"); i++ {
		q = b.GenerateQuestion(deck)
	}
	if q == nil || !q.written() || q.Chinese != "书" {
		t.Fatal("no write question for 书")
	}
	if strings.Contains(q.Prompt, q.Chinese) {
		t.Fatalf("prompt gives the answer away: %q", q.Prompt)
	}
	for text, want := range map[string]bool{"书": true, " 书 ": true, "書": true, "水": false, "": false} {
		if got := q.Accepts(text); got != want {
			t.Errorf("Accepts(%q) = %v, want %v", text, got, want)
		}
	}
	b.Question = q
	if ok, _ := b.AnswerText("水"); ok || b.Answered != 1 || b.Player.HP == b.Player.MaxHP {
		t.Errorf("wrong typed answer: ok=%v answered=%d hp=%d", ok, b.Answered, b.Player.HP)
	}
	if ok, dmg := b.AnswerText("书"); !ok || dmg <= 0 || b.CorrectHits != 1 {
		t.Errorf("right typed answer: ok=%v dmg=%d hits=%d", ok, dmg, b.CorrectHits)
	}
}
