package journey

import (
	"testing"

	"github.com/therealsiege/tuilaoshi/internal/model"
)

func TestEnemySetsCoverTheRoad(t *testing.T) {
	sets := EnemySets()
	seen := map[string]bool{}
	for _, s := range sets {
		if len(s.Members) == 0 || s.Reward.Yuanbao == 0 {
			t.Errorf("set %s is empty or pays no ingot", s.Name)
		}
		for _, m := range s.Members {
			if seen[m] {
				t.Errorf("%s is in two sets", m)
			}
			seen[m] = true
			if EnemyLookup(m) == nil {
				t.Errorf("set %s names unknown enemy %s", s.Name, m)
			}
		}
	}
	if seen["如来佛祖"] {
		t.Error("the Buddha is not a collectible enemy")
	}
	for _, d := range []string{"黑熊精", "黄风怪", "红孩儿", "金翅大鹏", "青牛精", "六耳猕猴"} {
		if !seen[d] {
			t.Errorf("demon %s belongs to no region set", d)
		}
	}
}

func TestRadicalSetsGroupCharacters(t *testing.T) {
	page := func(ch, rad string, interval int) model.Page {
		return model.Page{Type: "character", Character: ch, Radical: rad, ReviewInterval: interval}
	}
	pages := []model.Page{page("河", "氵 water", 7), page("海", "氵 water", 1), page("火", "火 fire", 7)}
	sets := RadicalSets(pages)
	if len(sets) != 2 || sets[0].ID != "radical-氵" || len(sets[0].Members) != 2 || sets[0].Reward.Tongqian != 20 {
		t.Fatalf("sets %+v", sets)
	}
	got := MasteredCharacters(pages)
	if sets[0].Have(got) != 1 || sets[1].Have(got) != 1 {
		t.Errorf("have: water %d, fire %d", sets[0].Have(got), sets[1].Have(got))
	}
}
