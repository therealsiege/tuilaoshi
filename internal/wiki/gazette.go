package wiki

// Gazette is a weekly issue of the court gazette (邸报): a short report of
// the journey's week, in Chinese with its English.
type Gazette struct {
	Week    string `yaml:"week"` // the Monday it came out, YYYY-MM-DD
	Text    string `yaml:"text"`
	English string `yaml:"english"`
	ByHand  bool   `yaml:"by_hand,omitempty"` // set from the template, without Claude
}

// gazetteDoc keeps the journey's issues in the progress file: each journey
// has its own gazette, since each has its own road.
func (s *Store) gazetteDoc() string {
	if s.Save != "" {
		return "gazette/" + s.Save + ".yaml"
	}
	return "gazette.yaml"
}

// Gazettes returns the journey's issues of the gazette, oldest first.
func (s *Store) Gazettes() ([]Gazette, error) {
	var issues []Gazette
	err := s.readYAML(s.gazetteDoc(), &issues)
	return issues, err
}

// SaveGazette adds an issue, or replaces the one for the same week.
func (s *Store) SaveGazette(g Gazette) error {
	issues, err := s.Gazettes()
	if err != nil {
		return err
	}
	for i := range issues {
		if issues[i].Week == g.Week {
			issues[i] = g
			return s.writeYAML(s.gazetteDoc(), issues)
		}
	}
	return s.writeYAML(s.gazetteDoc(), append(issues, g))
}
