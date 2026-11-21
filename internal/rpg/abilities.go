package rpg

import (
	"math/rand"
	"time"

	"github.com/therealsiege/tuilaoshi/internal/journey"
)

// Effects are the abilities in play in a battle: the enemy's make it harder,
// the party's help. Demons and companions name their abilities in the
// journey content; the registry below says what each name does.
type Effects struct {
	HidePinyin  bool          // no pinyin in prompts or as answers
	LookAlike   bool          // wrong answers look like the right one
	TimeLimit   time.Duration // each question must be answered in time
	MissDamage  int           // extra damage on a miss, in percent
	Reveals     int           // questions that can have two wrong answers struck out
	HalfCombo   bool          // a miss halves the combo instead of resetting it
	Shields     int           // enemy hits blocked outright
	FirstStrike bool          // the first hit of the battle deals double damage
}

// ability is one registered ability: what it does and how the intro says it.
type ability struct {
	says  string
	apply func(*Effects)
}

var demonAbilities = map[string]ability{
	"blind":    {"blinds you: no pinyin is shown", func(e *Effects) { e.HidePinyin = true }},
	"disguise": {"disguises its lies: wrong answers look like the right one", func(e *Effects) { e.LookAlike = true }},
	"burn":     {"burns: answer each question within 12 seconds", func(e *Effects) { e.TimeLimit = 12 * time.Second }},
	"venom":    {"stings: misses hurt half as much again", func(e *Effects) { e.MissDamage += 50 }},
}

var companionAbilities = map[string]ability{
	"reveal":       {"once per battle, w strikes out two wrong answers", func(e *Effects) { e.Reveals++ }},
	"half_combo":   {"a miss only halves your combo", func(e *Effects) { e.HalfCombo = true }},
	"shield":       {"blocks the first enemy hit", func(e *Effects) { e.Shields++ }},
	"first_strike": {"your first hit deals double damage", func(e *Effects) { e.FirstStrike = true }},
}

// KnownAbility reports whether name is a registered demon (demon true) or
// companion ability, so content can be validated.
func KnownAbility(name string, demon bool) bool {
	if demon {
		_, ok := demonAbilities[name]
		return ok
	}
	_, ok := companionAbilities[name]
	return ok
}

// EffectsFor combines the enemy's abilities with those of the party members
// who have joined, and describes each for the battle intro.
func EffectsFor(enemy journey.BestiaryEntry, party []journey.Companion) (Effects, []string) {
	var e Effects
	var lines []string
	for _, name := range enemy.Abilities {
		if a, ok := demonAbilities[name]; ok {
			a.apply(&e)
			lines = append(lines, enemy.Name+" "+a.says)
		}
	}
	for _, c := range party {
		if a, ok := companionAbilities[c.Effect]; ok && c.Unlocked {
			a.apply(&e)
			lines = append(lines, c.Name+": "+a.says)
		}
	}
	return e, lines
}

// SetEffects puts effects in play for this battle and its rematches.
func (b *Battle) SetEffects(e Effects) {
	b.Effects, b.startEffects = e, e
}

// Reveal strikes out all but one wrong answer of a choice question, using a
// reveal. It reports whether it did.
func (b *Battle) Reveal() bool {
	q := b.Question
	if b.Effects.Reveals == 0 || q == nil || len(q.Choices) < 3 {
		return false
	}
	b.Effects.Reveals--
	wrong := (q.Correct + 1 + rand.Intn(len(q.Choices)-1)) % len(q.Choices)
	right := q.Choices[q.Correct]
	if wrong < q.Correct {
		q.Choices, q.Correct = []string{q.Choices[wrong], right}, 1
	} else {
		q.Choices, q.Correct = []string{right, q.Choices[wrong]}, 0
	}
	b.Selected = 0
	return true
}
