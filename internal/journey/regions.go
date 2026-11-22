package journey

// Region is a stretch of the road walked as one tile map (see regionmap.go).
type Region struct {
	ID      string `yaml:"id"`
	Name    string `yaml:"name"`
	English string `yaml:"english"`
	From    int    `yaml:"from"`
	To      int    `yaml:"to"`
	Terrain string `yaml:"terrain"`
}

// Regions returns the regions in road order.
func Regions() []Region { return content.Regions }

// RegionOf returns the index of the region holding stage n.
func RegionOf(n int) int {
	for i, r := range content.Regions {
		if r.From <= n && n <= r.To {
			return i
		}
	}
	return len(content.Regions) - 1
}

// RegionByID returns a region's index, or -1.
func RegionByID(id string) int {
	for i, r := range content.Regions {
		if r.ID == id {
			return i
		}
	}
	return -1
}

// TownsIn returns the towns reached on a region's stages, in road order.
func (r Region) TownsIn() []Town {
	var towns []Town
	for _, t := range content.Towns {
		if r.From <= t.Stage && t.Stage <= r.To {
			towns = append(towns, t)
		}
	}
	return towns
}
