package journey

import (
	"fmt"
	"strings"

	"github.com/therealsiege/tuilaoshi/internal/model"
)

// Spirit is a hidden enemy tied to a radical family.
type Spirit struct {
	BestiaryEntry `yaml:",inline"`
	Radical       string `yaml:"radical"`
}

// SpiritWakes is how many of a family's characters must be mastered for
// its spirit to appear.
const SpiritWakes = 10

// Spirits returns every hidden spirit.
func Spirits() []Spirit { return content.Spirits }

// AwakeSpirits returns the spirits whose family has SpiritWakes characters
// mastered and who have not been beaten yet.
func AwakeSpirits(pages []model.Page, beaten map[string]bool) []Spirit {
	radicals := RadicalIndex(pages)
	mastered := MasteredCharacters(pages)
	count := map[string]int{}
	for ch := range mastered {
		if rad, ok := radicals[ch]; ok {
			count[radicalKey(rad)]++
		}
	}
	var awake []Spirit
	for _, s := range content.Spirits {
		if count[s.Radical] >= SpiritWakes && !beaten[s.Name] {
			awake = append(awake, s)
		}
	}
	return awake
}

// FamilyCards returns the cards with a character of the radical family.
func FamilyCards(cards []model.Card, radicals map[string]string, radical string) []model.Card {
	var out []model.Card
	for _, c := range cards {
		for _, r := range c.Chinese {
			if rad, ok := radicals[string(r)]; ok && radicalKey(rad) == radical {
				out = append(out, c)
				break
			}
		}
	}
	return out
}

// radicalKey is a radical entry's first character: "氵 (water)" -> "氵".
func radicalKey(rad string) string {
	if r := []rune(strings.TrimSpace(rad)); len(r) > 0 {
		return string(r[0])
	}
	return ""
}

func loadSpirits(c *Content) {
	for i := range c.Spirits {
		s := &c.Spirits[i]
		if s.Art == "" {
			s.Art = rankArt[RankCaptain]
		}
		if s.ImagePath == "" {
			s.ImagePath = fmt.Sprintf("assets/images/spirits/%s.png", s.Name)
		}
	}
}
