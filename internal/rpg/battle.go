package rpg

import (
	"math/rand"
	"strconv"
	"strings"

	"github.com/therealsiege/tuilaoshi/internal/journey"
	"github.com/therealsiege/tuilaoshi/internal/model"
)

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
	BattlePlayerTurn               // showing question
	BattlePlayerAttack             // correct answer animation
	BattleEnemyAttack              // wrong answer, enemy hits back
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
	DamageDealt int  // last damage dealt
	DamageTaken int  // last damage taken
	TotalDealt  int  // damage dealt across the whole battle
	Answered    int  // questions answered
	CorrectHits int  // questions answered correctly
	Listening   bool // audio works, so some questions are heard rather than read
	Writing     bool // the player can type Chinese, so some questions ask for it

	Effects      Effects // abilities in play; reveals and shields are used up
	startEffects Effects // as the battle began, for rematches
	struck       bool    // the player has landed a hit (for FirstStrike)
	Event        string  // what an ability did on the last answer, for the log

	// Radicals maps characters to their radical ("火 (fire)") so answers can
	// show how a word's characters carry its meaning.
	Radicals map[string]string

	// Rank is the enemy's: a soldier rank, RankDemon or RankBoss. It sets
	// the copper cash a win pays.
	Rank string

	// Difficulty tunes the question kinds to recent accuracy.
	Difficulty Difficulty
}

// Accuracy returns the percentage of questions answered correctly.
func (b *Battle) Accuracy() int {
	if b.Answered == 0 {
		return 0
	}
	return b.CorrectHits * 100 / b.Answered
}

// Rematch returns a fresh battle against the same enemy.
func (b *Battle) Rematch(cards []model.Card) *Battle {
	r := &Battle{Player: b.Player, Enemy: b.Enemy, State: BattleIntro, Listening: b.Listening, Writing: b.Writing}
	r.SetEffects(b.startEffects)
	r.Radicals = b.Radicals
	r.Rank = b.Rank
	r.Difficulty = b.Difficulty
	r.Player.HP = r.Player.MaxHP
	r.Enemy.HP = r.Enemy.MaxHP
	r.Question = r.GenerateQuestion(cards)
	return r
}

// NewBattle creates a battle against a specific demon.
func NewBattle(cards []model.Card, demon journey.BestiaryEntry, playerRank string, masteredCount int) *Battle {
	playerHP := PlayerHP(masteredCount)
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
		Rank:  RankDemon,
	}

	if len(cards) > 0 {
		b.Question = b.GenerateQuestion(cards)
	}

	return b
}

// Answer processes the player's choice and returns whether it was correct
// along with the damage dealt or taken.
func (b *Battle) Answer(choiceIndex int) (correct bool, dmg int) {
	if b.Question == nil {
		return false, 0
	}
	return b.resolve(choiceIndex == b.Question.Correct)
}

// AnswerText processes a typed answer to a "write" question.
func (b *Battle) AnswerText(text string) (correct bool, dmg int) {
	if b.Question == nil {
		return false, 0
	}
	return b.resolve(b.Question.Accepts(text))
}

// resolve applies a right or wrong answer: the player strikes, or the
// enemy does.
func (b *Battle) resolve(right bool) (correct bool, dmg int) {
	b.Answered++
	b.Event = ""

	if right {
		// Correct answer: player attacks enemy.
		b.CorrectHits++
		b.Combo++
		if b.Combo > b.MaxCombo {
			b.MaxCombo = b.Combo
		}

		totalDmg := b.Question.Damage
		actualDmg := totalDmg - b.Enemy.Defense/2
		if actualDmg < 5 {
			actualDmg = 5
		}
		actualDmg += b.Effects.BonusDamage
		if b.Effects.FirstStrike && !b.struck {
			actualDmg *= 2
			b.Event = "白龙马 carries you in: a double-strength first strike!"
		}
		b.struck = true

		b.Enemy.HP -= actualDmg
		if b.Enemy.HP < 0 {
			b.Enemy.HP = 0
		}

		b.DamageDealt = actualDmg
		b.DamageTaken = 0
		b.TotalDealt += actualDmg
		b.XPEarned += actualDmg
		b.Turn++
		b.State = BattlePlayerAttack

		return true, actualDmg
	}

	// Wrong answer: enemy attacks player.
	if b.Effects.HalfCombo {
		b.Combo /= 2
	} else {
		b.Combo = 0
	}

	enemyDmg := b.Enemy.Attack - b.Player.Defense/2 + rand.Intn(5) + 1
	if enemyDmg < 5 {
		enemyDmg = 5
	}
	enemyDmg += enemyDmg * b.Effects.MissDamage / 100
	if b.Effects.Shields > 0 {
		b.Effects.Shields--
		enemyDmg = 0
		b.Event = "沙悟净 steps in and takes the blow!"
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

// PlayerHP is the player's max HP before gear, from the level that
// mastered characters reach.
func PlayerHP(masteredCount int) int { return journey.LevelHP(journey.LevelFor(masteredCount).N) }
