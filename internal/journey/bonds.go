package journey

// Bond is a disciple's bond content: their voice for conversation, the
// assist upgrade earned at BondUpgradeRank, and a scene per rank above 0.
type Bond struct {
	Companion   string      `yaml:"companion"`
	Voice       string      `yaml:"voice"`
	Upgrade     string      `yaml:"upgrade"`
	UpgradeDesc string      `yaml:"upgrade_desc"`
	Scenes      []BondScene `yaml:"scenes"`
}

// BondScene is a short original scene opened by a bond rank.
type BondScene struct {
	Title        string `yaml:"title"`
	EnglishTitle string `yaml:"english_title"`
	Text         string `yaml:"text"`
	English      string `yaml:"english"`
}

// BondRank is a step of a bond, reached at Points.
type BondRank struct {
	Name, English string
	Points        int
}

// BondRanks are the steps of every bond; each above the first opens a scene.
var BondRanks = []BondRank{
	{"相识", "Acquainted", 0},
	{"同行", "Fellow travellers", 5},
	{"信任", "Trusted", 15},
	{"知己", "Confidants", 30},
	{"生死之交", "Sworn for life", 50},
}

// BondUpgradeRank is the rank at which a companion's assist is upgraded.
const BondUpgradeRank = 2

// BondRankFor returns the rank index reached with points.
func BondRankFor(points int) int {
	r := 0
	for i, br := range BondRanks {
		if points >= br.Points {
			r = i
		}
	}
	return r
}

// Bonds returns every disciple's bond content.
func Bonds() []Bond { return content.Bonds }

// BondFor returns a companion's bond content, or nil.
func BondFor(companion string) *Bond {
	for i := range content.Bonds {
		if content.Bonds[i].Companion == companion {
			return &content.Bonds[i]
		}
	}
	return nil
}
