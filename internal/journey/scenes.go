package journey

import "github.com/therealsiege/tuilaoshi/internal/model"

// Scene is a short story passage in graded Chinese, shown when its stage
// first opens.
type Scene struct {
	Stage       int      `yaml:"stage"`
	Title       string   `yaml:"title"`
	Text        []string `yaml:"text"`
	English     string   `yaml:"english"`
	Speaker     string   `yaml:"speaker"` // companion who speaks the closing line
	Line        string   `yaml:"line"`
	LineEnglish string   `yaml:"line_english"`
}

// GlossEntry glosses a name or story word the decks may not have.
type GlossEntry struct {
	Word    string `yaml:"word"`
	Pinyin  string `yaml:"pinyin"`
	Meaning string `yaml:"meaning"`
}

// storyBook is content/scenes.yaml.
type storyBook struct {
	Scenes   []Scene      `yaml:"scenes"`
	Glossary []GlossEntry `yaml:"glossary"`
}

// SceneFor returns the scene for stage n, or nil when it has none.
func SceneFor(n int) *Scene {
	for i := range content.Story.Scenes {
		if content.Story.Scenes[i].Stage == n {
			return &content.Story.Scenes[i]
		}
	}
	return nil
}

// StoryGlossary returns the scene glossary as cards, for glossing scenes.
func StoryGlossary() []model.Card {
	cards := make([]model.Card, 0, len(content.Story.Glossary))
	for _, g := range content.Story.Glossary {
		cards = append(cards, model.Card{ID: "journey/scenes#" + g.Word, Chinese: g.Word, Pinyin: g.Pinyin, English: g.Meaning})
	}
	return cards
}

// JourneyEpisode is a dialogue episode tied to a stage, in content/episodes.yaml.
type JourneyEpisode struct {
	ID     string               `yaml:"id"`
	Stage  int                  `yaml:"stage"`
	Title  string               `yaml:"title"`
	Source string               `yaml:"source"`
	Lines  []model.DialogueLine `yaml:"lines"`
}

// episodeBook is content/episodes.yaml: the episodes and the cast of
// speakers who are neither companions nor demons.
type episodeBook struct {
	Cast     []string         `yaml:"cast"`
	Episodes []JourneyEpisode `yaml:"episodes"`
}

// ReachedEpisodes returns the episodes of stages up to current, in story
// order, ready for the reader.
func ReachedEpisodes(current int) []model.Episode {
	var eps []model.Episode
	for _, e := range content.Episodes.Episodes {
		if e.Stage <= current {
			eps = append(eps, model.Episode{Title: e.Title, Source: e.Source, Lines: e.Lines})
		}
	}
	return eps
}
