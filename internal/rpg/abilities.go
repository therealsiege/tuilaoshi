package rpg

import (
	"math/rand"
	"slices"
	"strings"
	"time"

	"github.com/therealsiege/tuilaoshi/internal/confusables"
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
	BonusDamage int           // added to every hit (gear)
	BonusHP     int           // added to the player's max HP (gear)
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

// gearAbilities are what worn gear does; they apply after the enemy's, so
// gear can soften them.
var gearAbilities = map[string]ability{
	"power":       {"+3 damage on every hit", func(e *Effects) { e.BonusDamage += 3 }},
	"power2":      {"+6 damage on every hit", func(e *Effects) { e.BonusDamage += 6 }},
	"power3":      {"+10 damage on every hit", func(e *Effects) { e.BonusDamage += 10 }},
	"hp":          {"+15 max HP", func(e *Effects) { e.BonusHP += 15 }},
	"hp2":         {"+30 max HP", func(e *Effects) { e.BonusHP += 30 }},
	"hp3":         {"+50 max HP", func(e *Effects) { e.BonusHP += 50 }},
	"slow_burn":   {"six more seconds against burning foes", func(e *Effects) { e.TimeLimit += min(e.TimeLimit, 6*time.Second) }},
	"half_combo":  {"a miss only halves your combo", func(e *Effects) { e.HalfCombo = true }},
	"shield":      {"blocks one enemy hit", func(e *Effects) { e.Shields++ }},
	"reveal":      {"one more reveal (w)", func(e *Effects) { e.Reveals++ }},
	"clear_sight": {"blinding foes can't hide the pinyin", func(e *Effects) { e.HidePinyin = false }},
}

// KnownGearEffect reports whether name is a registered gear ability.
func KnownGearEffect(name string) bool {
	_, ok := gearAbilities[name]
	return ok
}

// AddGear applies worn gear on top of effects already set, and describes
// each piece for the battle intro.
func AddGear(e *Effects, worn []journey.Gear) []string {
	var lines []string
	for _, g := range worn {
		if a, ok := gearAbilities[g.Effect]; ok {
			a.apply(e)
			lines = append(lines, g.Name+": "+a.says)
		}
	}
	return lines
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

// lookAlikes maps each character to those commonly confused with it by sight.
var lookAlikes = func() map[string][]string {
	m := map[string][]string{}
	for _, p := range confusables.Pairs() {
		if p.Category != "visual" {
			continue
		}
		for _, a := range p.Chars {
			for _, b := range p.Chars {
				if a.Hanzi != b.Hanzi {
					m[a.Hanzi] = append(m[a.Hanzi], b.Hanzi)
				}
			}
		}
	}
	return m
}()

// resemblance scores how much a wrong answer looks like the right one:
// a known look-alike character, then shared characters, then equal length.
func resemblance(answer, other string) int {
	score := 0
	if slices.Contains(lookAlikes[answer], other) {
		score += 10
	}
	for _, r := range other {
		if strings.ContainsRune(answer, r) {
			score += 3
		}
	}
	if len([]rune(answer)) == len([]rune(other)) {
		score++
	}
	return score
}
