package wiki

// Letter is a weekly letter from Guanyin about the week before it, in
// Chinese with its English.
type Letter struct {
	Week    string `yaml:"week"` // the Monday it arrived, YYYY-MM-DD
	Text    string `yaml:"text"`
	English string `yaml:"english"`
	Read    bool   `yaml:"read,omitempty"`
}

// lettersDoc keeps the letters in the progress file.
const lettersDoc = "letters.yaml"

// Letters returns Guanyin's letters, oldest first.
func (s *Store) Letters() ([]Letter, error) {
	var letters []Letter
	err := s.readYAML(lettersDoc, &letters)
	return letters, err
}

// SaveLetter adds a letter, or replaces the one for the same week.
func (s *Store) SaveLetter(l Letter) error {
	letters, err := s.Letters()
	if err != nil {
		return err
	}
	for i := range letters {
		if letters[i].Week == l.Week {
			letters[i] = l
			return s.writeYAML(lettersDoc, letters)
		}
	}
	return s.writeYAML(lettersDoc, append(letters, l))
}
