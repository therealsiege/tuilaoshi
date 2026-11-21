package journey

// LoreEntry is an unlockable story passage from Journey to the West.
type LoreEntry struct {
	Tribulation int    `yaml:"tribulation,omitempty"` // unlocks at this tribulation number
	Title       string `yaml:"title,omitempty"`       // English title
	Chinese     string `yaml:"chinese,omitempty"`     // Chinese title
	Passage     string `yaml:"passage,omitempty"`     // 2-3 sentences in Chinese
	Pinyin      string `yaml:"pinyin,omitempty"`      // pinyin for the passage
	English     string `yaml:"english,omitempty"`     // English translation
	Unlocked    bool   `yaml:"-"`
}

// buildLore returns lore entries with unlock status based on current tribulation.
func buildLore(tribulation int) []LoreEntry {
	entries := loreData()
	for i := range entries {
		entries[i].Unlocked = tribulation >= entries[i].Tribulation
	}
	return entries
}

// loreData returns the raw lore entries without unlock status applied.
func loreData() []LoreEntry {
	return append([]LoreEntry{}, content.Lore...)
}
