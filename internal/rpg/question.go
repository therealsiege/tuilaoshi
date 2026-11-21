package rpg

import (
	"fmt"
	"math/rand"
	"regexp"
	"strings"

	"github.com/therealsiege/tuilaoshi/internal/model"
)

// CombatQuestion is a multiple-choice question generated from wiki cards.
type CombatQuestion struct {
	Prompt   string     // the question text (e.g. "What does 书 mean?")
	Choices  []string   // 4-5 answer options
	Correct  int        // index of correct answer
	Category string     // "reading", "meaning", "pinyin", "character"
	Damage   int        // damage dealt on correct answer
	Chinese  string     // the Chinese text being tested
	Pinyin   string     // pinyin of the tested item
	English  string     // meaning of the tested item
	Card     model.Card // the card being tested, for saving the result
}

// CorrectAnswer returns the text of the correct choice.
func (q *CombatQuestion) CorrectAnswer() string {
	return q.Choices[q.Correct]
}

const (
	// distractorCount is how many wrong choices a full question offers.
	distractorCount = 3
	// questionAttempts limits how many cards are tried to build a full question.
	questionAttempts = 12
)

// parenthetical matches glosses like "(to) look" so they compare as "look".
var parenthetical = regexp.MustCompile(`\([^)]*\)`)

// GenerateQuestion creates a new combat question from the available cards. It
// prefers a question with a full set of choices, and returns nil when no card
// has an unambiguous wrong answer.
func (b *Battle) GenerateQuestion(cards []model.Card) *CombatQuestion {
	d := newDeck(filterUsableCards(cards))
	if len(d.cards) == 0 {
		return nil
	}

	var best *CombatQuestion
	for i := 0; i < questionAttempts; i++ {
		q := b.buildQuestion(d.cards[rand.Intn(len(d.cards))], d)
		if q == nil {
			continue
		}
		if best == nil || len(q.Choices) > len(best.Choices) {
			best = q
		}
		if len(best.Choices) == distractorCount+1 {
			break
		}
	}
	return best
}

// buildQuestion creates a question testing card, or nil if no valid wrong
// answers exist for the category chosen.
func (b *Battle) buildQuestion(card model.Card, d deck) *CombatQuestion {
	categories := availableCategories(card)
	if len(categories) == 0 {
		return nil
	}
	category := categories[rand.Intn(len(categories))]

	var prompt, correctAnswer string
	switch category {
	case "meaning":
		prompt = fmt.Sprintf("What does %s mean?", card.Chinese)
		correctAnswer = card.English
	case "pinyin":
		prompt = fmt.Sprintf("What is the pinyin for %s?", card.Chinese)
		correctAnswer = card.Pinyin
	case "reading":
		prompt = fmt.Sprintf("Which Chinese word means \"%s\"?", card.English)
		correctAnswer = card.Chinese
	case "character":
		prompt = fmt.Sprintf("Which character is pronounced %s?", card.Pinyin)
		correctAnswer = card.Chinese
	}

	distractors := generateDistractors(card, d, category, distractorCount)
	if len(distractors) == 0 {
		return nil
	}

	choices := make([]string, 0, len(distractors)+1)
	choices = append(choices, correctAnswer)
	choices = append(choices, distractors...)

	// Shuffle choices.
	rand.Shuffle(len(choices), func(i, j int) {
		choices[i], choices[j] = choices[j], choices[i]
	})

	correctIdx := 0
	for i, c := range choices {
		if c == correctAnswer {
			correctIdx = i
			break
		}
	}

	damage := 15 + rand.Intn(10) + 1 + b.Combo*3

	return &CombatQuestion{
		Prompt:   prompt,
		Choices:  choices,
		Correct:  correctIdx,
		Category: category,
		Damage:   damage,
		Chinese:  card.Chinese,
		Pinyin:   card.Pinyin,
		English:  card.English,
		Card:     card,
	}
}

// filterUsableCards returns cards that have enough data for question generation.
func filterUsableCards(cards []model.Card) []model.Card {
	var usable []model.Card
	for _, c := range cards {
		if c.Chinese != "" && (c.English != "" || c.Pinyin != "") {
			usable = append(usable, c)
		}
	}
	return usable
}

// availableCategories returns the question categories a card can support.
func availableCategories(card model.Card) []string {
	var cats []string
	if card.Chinese != "" && card.English != "" {
		cats = append(cats, "meaning")
		cats = append(cats, "reading")
	}
	if card.Chinese != "" && card.Pinyin != "" {
		cats = append(cats, "pinyin")
		cats = append(cats, "character")
	}
	return cats
}

// deck holds the usable cards and every card teaching each word, since one
// word can appear on several pages or have several readings (调 tiáo/diào).
type deck struct {
	cards     []model.Card
	byChinese map[string][]model.Card
}

func newDeck(cards []model.Card) deck {
	d := deck{cards: cards, byChinese: map[string][]model.Card{}}
	for _, c := range cards {
		d.byChinese[c.Chinese] = append(d.byChinese[c.Chinese], c)
	}
	return d
}

// generateDistractors picks wrong answers from other cards for a given
// question category. It returns up to count unique distractors.
func generateDistractors(card model.Card, d deck, category string, count int) []string {
	correctAnswer := ""
	switch category {
	case "meaning":
		correctAnswer = card.English
	case "pinyin":
		correctAnswer = card.Pinyin
	case "reading", "character":
		correctAnswer = card.Chinese
	}

	// Collect all candidate distractor values.
	seen := map[string]bool{correctAnswer: true}
	var candidates []string

	for _, c := range d.cards {
		if c.Chinese == card.Chinese {
			continue
		}
		var val string
		switch category {
		case "meaning":
			val = c.English
		case "pinyin":
			val = c.Pinyin
		case "reading", "character":
			val = c.Chinese
		}
		if val == "" || seen[val] {
			continue
		}
		seen[val] = true
		if couldBeRight(card, val, category, d) {
			continue
		}
		candidates = append(candidates, val)
	}

	// Shuffle candidates and take up to count.
	rand.Shuffle(len(candidates), func(i, j int) {
		candidates[i], candidates[j] = candidates[j], candidates[i]
	})

	if len(candidates) > count {
		candidates = candidates[:count]
	}

	return candidates
}

// couldBeRight reports whether val, offered as a wrong answer to a question
// about card, is also right according to any card teaching the asked word
// (meaning, pinyin) or the offered word (reading, character). A card missing
// the compared field counts as right, since its answer is unknown.
func couldBeRight(card model.Card, val, category string, d deck) bool {
	switch category {
	case "meaning":
		for _, s := range d.byChinese[card.Chinese] {
			if sharesGloss(s.English, val) {
				return true
			}
		}
	case "pinyin":
		for _, s := range d.byChinese[card.Chinese] {
			if normalizePinyin(s.Pinyin) == normalizePinyin(val) {
				return true
			}
		}
	case "reading":
		for _, k := range d.byChinese[val] {
			if k.English == "" || sharesGloss(k.English, card.English) {
				return true
			}
		}
	case "character":
		for _, k := range d.byChinese[val] {
			if k.Pinyin == "" || normalizePinyin(k.Pinyin) == normalizePinyin(card.Pinyin) {
				return true
			}
		}
	}
	return false
}

// sharesGloss reports whether two English definitions have a meaning in common.
func sharesGloss(a, b string) bool {
	glosses := map[string]bool{}
	for _, g := range splitGlosses(a) {
		glosses[g] = true
	}
	for _, g := range splitGlosses(b) {
		if glosses[g] {
			return true
		}
	}
	return false
}

// splitGlosses breaks a definition like "(to) look; a view" into comparable
// meanings: "look" and "view".
func splitGlosses(english string) []string {
	cleaned := parenthetical.ReplaceAllString(strings.ToLower(english), "")
	parts := strings.FieldsFunc(cleaned, func(r rune) bool {
		return r == ';' || r == ',' || r == '/'
	})

	var glosses []string
	for _, part := range parts {
		g := strings.TrimSpace(part)
		for _, article := range []string{"to ", "a ", "an ", "the "} {
			g = strings.TrimPrefix(g, article)
		}
		if g != "" {
			glosses = append(glosses, g)
		}
	}
	return glosses
}

// normalizePinyin makes pinyin comparable across spacing and capitalization.
func normalizePinyin(pinyin string) string {
	return strings.ToLower(strings.Join(strings.Fields(pinyin), ""))
}
