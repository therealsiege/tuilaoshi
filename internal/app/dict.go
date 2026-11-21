package app

import (
	"github.com/therealsiege/tuilaoshi/internal/model"
	"github.com/therealsiege/tuilaoshi/internal/ui/input"
	"github.com/therealsiege/tuilaoshi/internal/wiki"
)

// buildDict constructs a pinyin-to-character dictionary from all wiki pages.
func buildDict(pages []model.Page) *input.Dict {
	var drillable []model.Page
	for _, p := range pages {
		switch p.Type {
		case "character", "vocabulary", "grammar":
			drillable = append(drillable, p)
		}
	}
	cards := wiki.CardsFromPages(drillable)
	return input.NewFromCards(cards)
}
