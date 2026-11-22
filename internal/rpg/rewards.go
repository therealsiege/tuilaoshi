package rpg

import (
	"math/rand"

	"github.com/therealsiege/tuilaoshi/internal/journey"
	"github.com/therealsiege/tuilaoshi/internal/model"
)

// Enemy ranks beyond the soldiers': a stage's own demon, and a boss.
const (
	RankDemon = "demon"
	RankBoss  = "boss"
)

// rankStrength scales a soldier's HP and attack, in percent of a demon of
// the same chapter.
var rankStrength = map[string]struct{ hp, atk int }{
	journey.RankMinion:  {45, 70},
	journey.RankElite:   {65, 85},
	journey.RankCaptain: {85, 95},
}

// winCoins is the copper cash (铜钱) paid for beating each rank, on top of
// one coin per right answer.
var winCoins = map[string]int{
	journey.RankMinion:  5,
	journey.RankElite:   10,
	journey.RankCaptain: 20,
	RankDemon:           30,
	RankBoss:            50,
}

// WinCoins is the copper cash paid for winning this battle.
func (b *Battle) WinCoins() int {
	if !b.PlayerWon() {
		return 0
	}
	return winCoins[b.Rank]
}

// NewSoldierBattle creates a fight against a foot soldier, weaker than a
// demon of the same chapter by its rank.
func NewSoldierBattle(cards []model.Card, s journey.Soldier, masteredCount int) *Battle {
	b := NewBattle(cards, s.BestiaryEntry, "", masteredCount)
	if st, ok := rankStrength[s.Rank]; ok {
		b.Enemy.HP = max(b.Enemy.HP*st.hp/100, 20)
		b.Enemy.MaxHP = b.Enemy.HP
		b.Enemy.Attack = max(b.Enemy.Attack*st.atk/100, 5)
	}
	b.Rank = s.Rank
	return b
}

// ReachedSoldiers returns the soldiers serving on stages up to current, for
// practice battles.
func ReachedSoldiers(current int) []journey.Soldier {
	var reached []journey.Soldier
	for _, s := range journey.Soldiers() {
		if s.From <= current {
			reached = append(reached, s)
		}
	}
	return reached
}

// soldierShare is the chance, in percent, that a practice battle meets a
// soldier rather than a demon.
const soldierShare = 75

// NewPracticeBattle creates a practice battle: usually against one of
// soldiers, sometimes one of demons.
func NewPracticeBattle(cards []model.Card, masteredCount int, demons []journey.BestiaryEntry, soldiers []journey.Soldier) *Battle {
	if len(soldiers) > 0 && (len(demons) == 0 || rand.Intn(100) < soldierShare) {
		return NewSoldierBattle(cards, soldiers[rand.Intn(len(soldiers))], masteredCount)
	}
	return NewRandomBattle(cards, masteredCount, demons)
}
