package journey

// Gear slots; one item per slot is worn. Titles change only the name shown.
var GearSlots = []string{"weapon", "robe", "hat", "charm", "title"}

// SlotNames labels each gear slot in Chinese and English.
var SlotNames = map[string]string{
	"weapon": "兵器 weapon", "robe": "衣 robe", "hat": "冠 hat", "charm": "宝物 charm", "title": "称号 title",
}

// Price is what an item costs in copper cash (铜钱) and silver ingots (元宝).
type Price struct {
	Tongqian int `yaml:"tongqian,omitempty"`
	Yuanbao  int `yaml:"yuanbao,omitempty"`
}

// Gear is an item sold at the Earth God's shrine.
type Gear struct {
	ID      string         `yaml:"id"`
	Name    string         `yaml:"name"`
	English string         `yaml:"english"`
	Slot    string         `yaml:"slot"`
	Effect  string         `yaml:"effect,omitempty"` // a gear ability in the rpg registry; titles have none
	Desc    string         `yaml:"desc"`
	Price   Price          `yaml:"price"`
	Opens   int            `yaml:"opens,omitempty"`  // the stage that must be cleared first
	Recipe  map[string]int `yaml:"recipe,omitempty"` // crafted gear: material id -> count
}

// Crafted reports whether the item is made from materials rather than bought.
func (g Gear) Crafted() bool { return len(g.Recipe) > 0 }

// AllGear returns every item, in shelf order.
func AllGear() []Gear { return content.Gear }

// GearByID finds an item, or returns nil.
func GearByID(id string) *Gear {
	for i := range content.Gear {
		if content.Gear[i].ID == id {
			return &content.Gear[i]
		}
	}
	return nil
}

// OnShelf reports whether an item is for sale given the cleared stages.
func (g Gear) OnShelf(cleared map[int]string) bool {
	_, ok := cleared[g.Opens]
	return g.Opens == 0 || ok
}

// Worn returns the items worn, slot -> id, in slot order.
func Worn(worn map[string]string) []Gear {
	var gear []Gear
	for _, slot := range GearSlots {
		if g := GearByID(worn[slot]); g != nil {
			gear = append(gear, *g)
		}
	}
	return gear
}
