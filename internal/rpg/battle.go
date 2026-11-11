package rpg

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"

	"github.com/cj/mandarin-wiki/internal/journey"
	"github.com/cj/mandarin-wiki/internal/model"
)

// CombatQuestion is a multiple-choice question generated from wiki cards.
type CombatQuestion struct {
	Prompt   string   // the question text (e.g. "What does 书 mean?")
	Choices  []string // 4-5 answer options
	Correct  int      // index of correct answer
	Category string   // "reading", "meaning", "pinyin", "character"
	Damage   int      // damage dealt on correct answer
	Chinese  string   // the Chinese text being tested
	Pinyin   string   // pinyin of the tested item
}

// Fighter represents a combatant (player or enemy).
type Fighter struct {
	Name      string
	Chinese   string // Chinese name
	HP        int
	MaxHP     int
	Attack    int
	Defense   int
	Art       string // ASCII art
	ImagePath string
}

// BattleState tracks the current combat state.
type BattleState int

const (
	BattleIntro        BattleState = iota
	BattlePlayerTurn                      // showing question
	BattlePlayerAttack                    // correct answer animation
	BattleEnemyAttack                     // wrong answer, enemy hits back
	BattleVictory
	BattleDefeat
)

// Battle is the core battle model.
type Battle struct {
	Player      Fighter
	Enemy       Fighter
	State       BattleState
	Question    *CombatQuestion
	Selected    int // cursor position on choices
	Message     string
	Turn        int
	Combo       int // consecutive correct answers
	MaxCombo    int
	XPEarned    int
	DamageDealt int // last damage dealt
	DamageTaken int // last damage taken
}

// NewBattle creates a battle against a specific demon.
func NewBattle(cards []model.Card, demon journey.BestiaryEntry, playerRank string, masteredCount int) *Battle {
	playerHP := 100 + masteredCount/10
	playerAtk := 15 + masteredCount/50
	playerDef := 5

	playerName := playerRank
	if playerName == "" {
		playerName = "修行者"
	}

	enemyHP, enemyAtk, enemyDef := demonStats(demon)

	b := &Battle{
		Player: Fighter{
			Name:    playerName,
			Chinese: "修行者",
			HP:      playerHP,
			MaxHP:   playerHP,
			Attack:  playerAtk,
			Defense: playerDef,
		},
		Enemy: Fighter{
			Name:      demon.English,
			Chinese:   demon.Name,
			HP:        enemyHP,
			MaxHP:     enemyHP,
			Attack:    enemyAtk,
			Defense:   enemyDef,
			Art:       demon.Art,
			ImagePath: demon.ImagePath,
		},
		State: BattleIntro,
	}

	if len(cards) > 0 {
		b.Question = b.GenerateQuestion(cards)
	}

	return b
}

// NewRandomBattle creates a battle against a randomly selected demon.
func NewRandomBattle(cards []model.Card, masteredCount int) *Battle {
	demons := journey.Bestiary()
	if len(demons) == 0 {
		return nil
	}
	demon := demons[rand.Intn(len(demons))]
	return NewBattle(cards, demon, "", masteredCount)
}

// GenerateQuestion creates a new combat question from the available cards.
func (b *Battle) GenerateQuestion(cards []model.Card) *CombatQuestion {
	usable := filterUsableCards(cards)
	if len(usable) == 0 {
		return nil
	}

	card := usable[rand.Intn(len(usable))]

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

	distractors := generateDistractors(card, usable, category, 3)

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
	}
}

// Answer processes the player's choice and returns whether it was correct
// along with the damage dealt or taken.
func (b *Battle) Answer(choiceIndex int) (correct bool, dmg int) {
	if b.Question == nil {
		return false, 0
	}

	if choiceIndex == b.Question.Correct {
		// Correct answer: player attacks enemy.
		b.Combo++
		if b.Combo > b.MaxCombo {
			b.MaxCombo = b.Combo
		}

		totalDmg := b.Question.Damage
		actualDmg := totalDmg - b.Enemy.Defense/2
		if actualDmg < 5 {
			actualDmg = 5
		}

		b.Enemy.HP -= actualDmg
		if b.Enemy.HP < 0 {
			b.Enemy.HP = 0
		}

		b.DamageDealt = actualDmg
		b.DamageTaken = 0
		b.XPEarned += actualDmg
		b.Turn++
		b.State = BattlePlayerAttack

		return true, actualDmg
	}

	// Wrong answer: enemy attacks player.
	b.Combo = 0

	enemyDmg := b.Enemy.Attack - b.Player.Defense/2 + rand.Intn(5) + 1
	if enemyDmg < 5 {
		enemyDmg = 5
	}

	b.Player.HP -= enemyDmg
	if b.Player.HP < 0 {
		b.Player.HP = 0
	}

	b.DamageDealt = 0
	b.DamageTaken = enemyDmg
	b.Turn++
	b.State = BattleEnemyAttack

	return false, enemyDmg
}

// IsOver returns true if the battle has concluded.
func (b *Battle) IsOver() bool {
	return b.Player.HP <= 0 || b.Enemy.HP <= 0
}

// PlayerWon returns true if the player won.
func (b *Battle) PlayerWon() bool {
	return b.Enemy.HP <= 0 && b.Player.HP > 0
}

// EnemyWon returns true if the enemy won.
func (b *Battle) EnemyWon() bool {
	return b.Player.HP <= 0
}

// demonStats derives HP, Attack, and Defense from a BestiaryEntry's chapter
// position. Early chapters produce weaker demons; later chapters are tougher.
func demonStats(demon journey.BestiaryEntry) (hp, atk, def int) {
	chapterNum := parseChapterNumber(demon.Chapter)

	switch {
	case chapterNum <= 20:
		hp = 80
		atk = 10
		def = 3
	case chapterNum <= 50:
		hp = 120
		atk = 16
		def = 6
	case chapterNum <= 75:
		hp = 150
		atk = 20
		def = 8
	default:
		hp = 180
		atk = 25
		def = 10
	}

	return hp, atk, def
}

// parseChapterNumber extracts the first numeric chapter from a string like
// "Chapters 27-28" or "Chapter 55-56".
func parseChapterNumber(chapter string) int {
	s := strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' || r == ' ' {
			return r
		}
		return ' '
	}, chapter)

	for _, part := range strings.Fields(s) {
		n, err := strconv.Atoi(part)
		if err == nil && n > 0 {
			return n
		}
	}

	return 50 // default to mid-range
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

// generateDistractors picks wrong answers from other cards for a given
// question category. It returns up to count unique distractors.
func generateDistractors(card model.Card, allCards []model.Card, category string, count int) []string {
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

	for _, c := range allCards {
		if c.ID == card.ID {
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
