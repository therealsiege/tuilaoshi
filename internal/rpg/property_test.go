package rpg

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"testing"

	"github.com/therealsiege/tuilaoshi/internal/model"
	"github.com/therealsiege/tuilaoshi/internal/wiki"
)

// rightAnswers lists every choice in q that some card in pool makes a correct
// answer. It is the test's own oracle: a choice is right when a card for the
// asked word (or a card for the chosen word) matches what the prompt asks for.
func rightAnswers(q *CombatQuestion, pool []model.Card) []string {
	var right []string
	for _, choice := range q.Choices {
		for _, k := range pool {
			var ok bool
			switch q.Category {
			case "meaning": // prompt shows q.Chinese; choices are English
				ok = k.Chinese == q.Chinese && k.English != "" && sharesGloss(k.English, choice)
			case "pinyin": // prompt shows q.Chinese; choices are pinyin
				ok = k.Chinese == q.Chinese && k.Pinyin != "" && normalizePinyin(k.Pinyin) == normalizePinyin(choice)
			case "reading": // prompt shows q.English; choices are Chinese
				ok = k.Chinese == choice && k.English != "" && sharesGloss(k.English, q.English)
			case "character": // prompt shows q.Pinyin; choices are Chinese
				ok = k.Chinese == choice && k.Pinyin != "" && normalizePinyin(k.Pinyin) == normalizePinyin(q.Pinyin)
			}
			if ok {
				right = append(right, choice)
				break
			}
		}
	}
	return right
}

// checkQuestion fails the test unless q has distinct choices and exactly one
// right answer, at q.Correct.
func checkQuestion(t *testing.T, q *CombatQuestion, pool []model.Card) {
	t.Helper()
	if len(q.Choices) < 2 || len(q.Choices) > distractorCount+1 {
		t.Fatalf("%q: %d choices %v", q.Prompt, len(q.Choices), q.Choices)
	}
	seen := map[string]bool{}
	for _, c := range q.Choices {
		if seen[c] {
			t.Fatalf("%q: duplicate choice %q in %v", q.Prompt, c, q.Choices)
		}
		seen[c] = true
	}
	right := rightAnswers(q, pool)
	if len(right) != 1 || right[0] != q.CorrectAnswer() {
		t.Fatalf("%q (%s): right answers %v, want only %q; choices %v",
			q.Prompt, q.Category, right, q.CorrectAnswer(), q.Choices)
	}
}

// collidingPool builds a small deck from tiny vocabularies so homophones,
// shared glosses, and repeated characters are common.
func collidingPool(r *rand.Rand) []model.Card {
	chinese := []string{"书", "看", "见", "人", "大", "天", "口", "日"}
	pinyin := []string{"shu1", "kan4", "Kan4", "kan 4", "jian4", "ren2", "da4", "tian1", "kou3", ""}
	english := []string{"book", "(to) look", "look; see", "to see", "person", "big", "sky; day", "day", "mouth", "sun", ""}
	pool := make([]model.Card, 3+r.Intn(10))
	for i := range pool {
		pool[i] = model.Card{
			ID:      fmt.Sprintf("c%d", i),
			Chinese: chinese[r.Intn(len(chinese))],
			Pinyin:  pinyin[r.Intn(len(pinyin))],
			English: english[r.Intn(len(english))],
		}
	}
	return pool
}

func TestQuestionsHaveExactlyOneRightAnswer(t *testing.T) {
	r := rand.New(rand.NewSource(1))
	asked := 0
	for i := 0; i < 5000; i++ {
		pool := collidingPool(r)
		q := testBattle(pool).GenerateQuestion(pool)
		if q == nil {
			continue // no unambiguous question exists for this pool
		}
		asked++
		checkQuestion(t, q, pool)
	}
	if asked < 1000 {
		t.Fatalf("only %d of 5000 pools produced a question; the generator is too strict to test", asked)
	}
}

func TestRealWikiQuestionsHaveExactlyOneRightAnswer(t *testing.T) {
	dir := filepath.Join("..", "..", "wiki")
	if _, err := os.Stat(dir); err != nil {
		t.Skip("wiki directory not present")
	}
	if testing.Short() {
		t.Skip("loads the whole wiki")
	}
	cards, err := wiki.NewStore(dir).DrillableCards()
	if err != nil {
		t.Fatal(err)
	}
	pool := filterUsableCards(cards)
	b := testBattle(pool)
	for i := 0; i < 300; i++ {
		q := b.GenerateQuestion(pool)
		if q == nil {
			t.Fatal("no question from the real wiki")
		}
		checkQuestion(t, q, pool)
	}
}
