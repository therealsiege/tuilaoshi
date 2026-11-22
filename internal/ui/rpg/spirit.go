package rpg

import (
	"math/rand"

	"github.com/therealsiege/tuilaoshi/internal/journey"
	"github.com/therealsiege/tuilaoshi/internal/model"
	rpgengine "github.com/therealsiege/tuilaoshi/internal/rpg"
)

// A radical family's hidden spirit wakes once enough of its characters are
// mastered, then sometimes takes the place of a practice battle. Every
// question uses the family's characters, and the first win pays an ingot.

const (
	// spiritChance is the chance, in percent, that a practice battle meets
	// an awake spirit.
	spiritChance = 40
	// minSpiritDeck is the fewest family words a spirit battle needs.
	minSpiritDeck = 4
)

// spiritRoll draws 0-99 against spiritChance; tests fix it.
var spiritRoll = func() int { return rand.Intn(100) }

// meetSpirit may replace a practice battle with an awake spirit's, using
// the family's words from all of cards. It reports whether it did.
func (m *Model) meetSpirit(pages []model.Page, cards []model.Card) bool {
	m.spirit = nil
	if m.store == nil {
		return false
	}
	col, err := m.store.Collection()
	if err != nil {
		m.loadErr = err
		return false
	}
	beaten := map[string]bool{}
	for _, d := range col.Defeated {
		beaten[d] = true
	}
	awake := journey.AwakeSpirits(pages, beaten)
	if len(awake) == 0 || spiritRoll() >= spiritChance {
		return false
	}
	s := awake[rand.Intn(len(awake))]
	family := journey.FamilyCards(cards, journey.RadicalIndex(pages), s.Radical)
	if len(family) < minSpiritDeck {
		return false
	}
	m.spirit, m.cards = &s, family
	m.battle = rpgengine.NewBattle(family, s.BestiaryEntry, "", m.mastered)
	return true
}
