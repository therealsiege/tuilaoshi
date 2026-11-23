package rpg

import (
	"fmt"
	"testing"

	"github.com/therealsiege/tuilaoshi/internal/journey"
	"github.com/therealsiege/tuilaoshi/internal/model"
)

func card(id, chinese, pinyin, english string) model.Card {
	return model.Card{ID: id, Chinese: chinese, Pinyin: pinyin, English: english}
}

func testDeck() []model.Card {
	return []model.Card{
		card("1", "书", "shū", "book"),
		card("2", "水", "shuǐ", "water"),
		card("3", "火", "huǒ", "fire"),
		card("4", "山", "shān", "mountain"),
		card("5", "人", "rén", "person"),
	}
}

func testBattle(cards []model.Card) *Battle {
	return NewBattle(cards, journey.BestiaryEntry{Name: "白骨精", English: "White Bone Demon", Chapter: "Chapter 27"}, "", 0)
}

func wrongChoice(q *CombatQuestion) int {
	return (q.Correct + 1) % len(q.Choices)
}

func TestGenerateDistractorsExcludesOtherRightAnswers(t *testing.T) {
	tests := []struct {
		name     string
		category string
		target   model.Card
		other    model.Card
		wantExcl bool
	}{
		{"homophone for character", "character", card("1", "是", "shì", "to be"), card("2", "事", "shì", "matter"), true},
		{"spacing-only pinyin difference", "pinyin", card("1", "中国", "zhōngguó", "China"), card("2", "钟国", "Zhōng guó", "bell country"), true},
		{"synonym for reading", "reading", card("1", "看", "kàn", "to look; to see"), card("2", "瞧", "qiáo", "(to) look"), true},
		{"shared gloss for meaning", "meaning", card("1", "美", "měi", "beautiful, pretty"), card("2", "靓", "liàng", "pretty"), true},
		{"second reading of same word", "pinyin", card("1", "行", "xíng", "to walk"), card("2", "行", "háng", "row"), true},
		{"missing meaning cannot be ruled out", "reading", card("1", "书", "shū", "book"), card("2", "水", "shuǐ", ""), true},
		{"unrelated word stays available", "meaning", card("1", "书", "shū", "book"), card("2", "水", "shuǐ", "water"), false},
		{"different tone stays available", "character", card("1", "买", "mǎi", "to buy"), card("2", "卖", "mài", "to sell"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := generateDistractors(tt.target, newDeck([]model.Card{tt.target, tt.other}), tt.category, 3, false)
			if excluded := len(got) == 0; excluded != tt.wantExcl {
				t.Errorf("distractors = %v, want excluded = %v", got, tt.wantExcl)
			}
		})
	}
}

func TestGenerateQuestionHasExactlyOneCorrectChoice(t *testing.T) {
	b := testBattle(testDeck())

	for i := 0; i < 200; i++ {
		q := b.GenerateQuestion(testDeck())
		if q == nil {
			t.Fatal("GenerateQuestion returned nil for a usable deck")
		}
		if len(q.Choices) != distractorCount+1 {
			t.Fatalf("got %d choices, want %d", len(q.Choices), distractorCount+1)
		}
		seen := map[string]bool{}
		for _, c := range q.Choices {
			if seen[c] {
				t.Fatalf("duplicate choice %q in %v", c, q.Choices)
			}
			seen[c] = true
		}
		if q.English == "" || q.Chinese == "" || q.Pinyin == "" {
			t.Fatalf("question is missing word details: %+v", q)
		}
	}
}

func TestGenerateQuestionReturnsNilWithoutValidDistractors(t *testing.T) {
	deck := []model.Card{card("1", "书", "shū", "book")}

	if q := testBattle(deck).GenerateQuestion(deck); q != nil {
		t.Errorf("got %+v, want nil for a one-card deck", q)
	}
}

func TestBattleTotalsCoverWholeFight(t *testing.T) {
	deck := testDeck()
	b := testBattle(deck)

	_, first := b.Answer(b.Question.Correct)
	b.Question = b.GenerateQuestion(deck)
	b.Answer(wrongChoice(b.Question))

	if b.TotalDealt != first {
		t.Errorf("TotalDealt = %d, want %d after a later miss", b.TotalDealt, first)
	}
	if b.DamageDealt != 0 {
		t.Errorf("DamageDealt = %d, want 0 for the last hit", b.DamageDealt)
	}
	if got := b.Accuracy(); got != 50 {
		t.Errorf("Accuracy() = %d, want 50", got)
	}
}

func TestAccuracyWithNoAnswers(t *testing.T) {
	if got := testBattle(testDeck()).Accuracy(); got != 0 {
		t.Errorf("Accuracy() = %d, want 0", got)
	}
}

func TestRematchRestoresFightAgainstSameEnemy(t *testing.T) {
	deck := testDeck()
	b := testBattle(deck)
	b.Answer(b.Question.Correct)
	b.Player.HP = 0

	r := b.Rematch(deck)

	if r.Enemy.Chinese != b.Enemy.Chinese {
		t.Errorf("enemy = %q, want %q", r.Enemy.Chinese, b.Enemy.Chinese)
	}
	if r.Player.HP != r.Player.MaxHP || r.Enemy.HP != r.Enemy.MaxHP {
		t.Errorf("HP not restored: player %d/%d, enemy %d/%d", r.Player.HP, r.Player.MaxHP, r.Enemy.HP, r.Enemy.MaxHP)
	}
	if r.Answered != 0 || r.TotalDealt != 0 || r.Combo != 0 {
		t.Errorf("stats not reset: %+v", r)
	}
	if r.Question == nil {
		t.Error("rematch has no question")
	}
}

func TestWrongAnswersMatchTheRightOnesLength(t *testing.T) {
	var cards []model.Card
	for i, w := range []string{"九", "明", "很", "上", "练习", "生病", "公共汽车", "护照"} {
		cards = append(cards, model.Card{ID: fmt.Sprint(i), Chinese: w, Pinyin: fmt.Sprintf("p%d", i), English: fmt.Sprintf("meaning %d", i)})
	}
	d := newDeck(cards)
	for i := 0; i < 50; i++ {
		for _, c := range generateDistractors(cards[0], d, "character", 3, false) {
			if len([]rune(c)) != 1 {
				t.Fatalf("a one-character answer got the wrong answer %q", c)
			}
		}
	}
}
