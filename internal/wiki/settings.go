package wiki

// settingsFile keeps the learner's settings, shared by every journey.
const settingsFile = "settings.yaml"

// Settings are the learner's preferences.
type Settings struct {
	Speech      bool   `yaml:"speech"`       // speak answers and lines aloud
	Effects     bool   `yaml:"effects"`      // short sound cues for game events
	Input       string `yaml:"input"`        // answer boxes start in: chinese, pinyin, english
	Difficulty  string `yaml:"difficulty"`   // battles: adaptive, gentler, sterner
	Encounters  string `yaml:"encounters"`   // wild encounters: normal, fewer, off
	Portraits   string `yaml:"portraits"`    // enemy portraits: large, small, off
	ReviewChunk int    `yaml:"review_chunk"` // due words "Start here" suggests at once
	NewWords    int    `yaml:"new_words"`    // frequency-list words introduced a day
	ClaudeCap   int    `yaml:"claude_cap"`   // Claude calls allowed a day; 0 for no limit
	ToneColours bool   `yaml:"tone_colours"` // pinyin drawn in its tones' colours
}

// DefaultSettings are the settings before the learner changes any.
func DefaultSettings() Settings {
	return Settings{Speech: true, Effects: true, Input: "chinese", Difficulty: "adaptive", Encounters: "normal", Portraits: "large", ReviewChunk: 20, NewWords: 10, ToneColours: true}
}

// Settings reads the learner's settings, filling any missing with defaults.
func (s *Store) Settings() (Settings, error) {
	set := DefaultSettings()
	if err := s.readYAML(settingsFile, &set); err != nil {
		return DefaultSettings(), err
	}
	if set.NewWords < 0 {
		set.NewWords = 0
	}
	if set.ReviewChunk <= 0 {
		set.ReviewChunk = DefaultSettings().ReviewChunk
	}
	return set, nil
}

// SaveSettings writes the learner's settings.
func (s *Store) SaveSettings(set Settings) error {
	return s.writeYAML(settingsFile, set)
}
