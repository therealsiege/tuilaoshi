package journey

// BestiaryEntry describes a demon from Journey to the West.
type BestiaryEntry struct {
	Name      string   `yaml:"name,omitempty"`       // Chinese name
	English   string   `yaml:"english,omitempty"`    // English name
	Title     string   `yaml:"title,omitempty"`      // demon's title/epithet in Chinese
	TitleEng  string   `yaml:"title_eng,omitempty"`  // English title
	Origin    string   `yaml:"origin,omitempty"`     // 1-sentence origin story
	Powers    string   `yaml:"powers,omitempty"`     // what the demon can do
	Weakness  string   `yaml:"weakness,omitempty"`   // how they were defeated
	Chapter   string   `yaml:"chapter,omitempty"`    // approximate chapter reference
	Art       string   `yaml:"-"`                    // 3-5 line ASCII art of the demon
	ArtLines  []string `yaml:"art,omitempty"`        // Art as written in content/bestiary.yaml, one quoted line each (block scalars lose leading spaces)
	ImagePath string   `yaml:"image_path,omitempty"` // path to a terminal-renderable image (defaults to assets/images/demons/NAME.png)
}

// Bestiary returns all demon entries in the catalog.
func Bestiary() []BestiaryEntry { return content.Bestiary }

// BestiaryLookup finds a demon by its Chinese name, or returns nil.
func BestiaryLookup(name string) *BestiaryEntry {
	for i := range content.Bestiary {
		if content.Bestiary[i].Name == name {
			return &content.Bestiary[i]
		}
	}
	return nil
}
