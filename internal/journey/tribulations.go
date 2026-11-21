package journey

// Tribulation represents one of the 81 tribulations (八十一难) from Journey to the West.
type Tribulation struct {
	Number   int    `yaml:"number,omitempty"`
	Name     string `yaml:"name,omitempty"`     // Chinese name
	English  string `yaml:"english,omitempty"`  // English name
	Desc     string `yaml:"desc,omitempty"`     // 1-sentence description of what happens at this trial
	Location string `yaml:"location,omitempty"` // place name in Chinese
}

// AllTribulations returns the complete list of 81 tribulations.
func AllTribulations() []Tribulation { return content.Tribulations }

// TribulationAt returns the tribulation at the given number (1-81), or nil if not found.
func TribulationAt(n int) *Tribulation {
	for i := range content.Tribulations {
		if content.Tribulations[i].Number == n {
			return &content.Tribulations[i]
		}
	}
	return nil
}
