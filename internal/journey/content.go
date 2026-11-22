package journey

import (
	"embed"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// The journey's story content lives in content/*.yaml, embedded in the
// binary. Adding a demon, boss, lore passage or line of dialogue is a data
// change; content_test.go validates the whole set.
//
//go:embed content/*.yaml
var contentFS embed.FS

// Content is the journey's story data.
type Content struct {
	Tribulations []Tribulation
	Bestiary     []BestiaryEntry
	Bosses       []Boss
	Companions   []Companion
	Lore         []LoreEntry
	Dialogue     []Dialogue
	Themes       []Theme
	Progression  Progression
	Story        storyBook
	Episodes     episodeBook
	Soldiers     []Soldier
}

// Progression sets which characters count toward opening a stage.
type Progression struct {
	LearningUntil    int `yaml:"learning_until"`    // last stage that counts characters being learned
	LearningInterval int `yaml:"learning_interval"` // review interval (days) that counts as learned
}

// content is loaded once at startup; the embedded files are checked by
// tests, so a parse failure is a build defect.
var content = mustLoadContent()

func mustLoadContent() Content {
	c, err := LoadContent()
	if err != nil {
		panic(err)
	}
	return c
}

// LoadContent parses the embedded content files.
func LoadContent() (Content, error) {
	var c Content
	files := []struct {
		name string
		into any
	}{
		{"tribulations", &c.Tribulations},
		{"bestiary", &c.Bestiary},
		{"bosses", &c.Bosses},
		{"companions", &c.Companions},
		{"lore", &c.Lore},
		{"dialogue", &c.Dialogue},
		{"themes", &c.Themes},
		{"progression", &c.Progression},
		{"scenes", &c.Story},
		{"episodes", &c.Episodes},
		{"soldiers", &c.Soldiers},
	}
	for _, f := range files {
		data, err := contentFS.ReadFile("content/" + f.name + ".yaml")
		if err != nil {
			return Content{}, fmt.Errorf("journey content: %w", err)
		}
		if err := yaml.Unmarshal(data, f.into); err != nil {
			return Content{}, fmt.Errorf("journey content %s.yaml: %w", f.name, err)
		}
	}
	for i := range c.Bestiary {
		c.Bestiary[i].Art, c.Bestiary[i].ArtLines = strings.Join(c.Bestiary[i].ArtLines, "\n"), nil
		if c.Bestiary[i].ImagePath == "" {
			c.Bestiary[i].ImagePath = fmt.Sprintf("assets/images/demons/%s.png", c.Bestiary[i].Name)
		}
	}
	for i := range c.Soldiers {
		s := &c.Soldiers[i]
		s.Art, s.ArtLines = strings.Join(s.ArtLines, "\n"), nil
		if s.Art == "" {
			s.Art = rankArt[s.Rank]
		}
		if s.ImagePath == "" {
			s.ImagePath = fmt.Sprintf("assets/images/soldiers/%s.png", s.Name)
		}
	}
	return c, nil
}
