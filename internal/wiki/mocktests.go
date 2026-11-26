package wiki

// MockSection is one section's result in a mock HSK test.
type MockSection struct {
	Name    string `yaml:"name"`
	Score   int    `yaml:"score"` // out of 100
	Skipped bool   `yaml:"skipped,omitempty"`
}

// MockSitting is one sitting of the 模拟考 mock HSK test.
type MockSitting struct {
	Date     string        `yaml:"date"` // YYYY-MM-DD
	Level    int           `yaml:"level"`
	Sections []MockSection `yaml:"sections"`
	Total    int           `yaml:"total"`
	Max      int           `yaml:"max"`
	Passed   bool          `yaml:"passed"`
}

// mockTestsDoc keeps the sittings in the progress file.
const mockTestsDoc = "mocktests.yaml"

// MockSittings returns every mock test sat, oldest first.
func (s *Store) MockSittings() ([]MockSitting, error) {
	var sittings []MockSitting
	err := s.readYAML(mockTestsDoc, &sittings)
	return sittings, err
}

// SaveMockSitting adds a sitting to the record.
func (s *Store) SaveMockSitting(m MockSitting) error {
	sittings, err := s.MockSittings()
	if err != nil {
		return err
	}
	return s.writeYAML(mockTestsDoc, append(sittings, m))
}
