package journey

import (
	"strings"
	"unicode"

	"github.com/therealsiege/tuilaoshi/internal/model"
)

// Theme is a set of words that fits a part of the story, chosen by page
// tags, keywords in the English meaning, characters the word contains, or
// (for character pages) the character's radical.
// Levels, when set, restrict the theme to those HSK levels.
type Theme struct {
	ID       string   `yaml:"id"`
	Name     string   `yaml:"name"`
	Levels   []string `yaml:"levels,omitempty"`
	Tags     []string `yaml:"tags,omitempty"`
	Keywords []string `yaml:"keywords,omitempty"`
	Contains []string `yaml:"contains,omitempty"`
	Radicals []string `yaml:"radicals,omitempty"`
}

// ThemeByID returns the theme with id, or nil.
func ThemeByID(id string) *Theme {
	for i := range content.Themes {
		if content.Themes[i].ID == id {
			return &content.Themes[i]
		}
	}
	return nil
}

// stageTheme is the theme in force at tribulation n: its own, or the last
// one named before it.
func stageTheme(n int) Theme {
	theme := content.Themes[0]
	for _, t := range content.Tribulations {
		if t.Number > n {
			break
		}
		if th := ThemeByID(t.Theme); th != nil {
			theme = *th
		}
	}
	return theme
}

// Matches reports whether a card from page belongs to the theme.
func (t Theme) Matches(c model.Card, page model.Page) bool {
	if len(t.Levels) > 0 && !contains(t.Levels, page.Level) {
		return false
	}
	if len(t.Tags) == 0 && len(t.Keywords) == 0 && len(t.Contains) == 0 && len(t.Radicals) == 0 {
		return true // a level-only theme
	}
	for _, tag := range page.Tags {
		if contains(t.Tags, tag) {
			return true
		}
	}
	// A page's radical reads like "氵 (water)" or "扌(hand)": match its first character.
	if r := []rune(page.Radical); len(r) > 0 && contains(t.Radicals, string(r[0])) {
		return true
	}
	for _, ch := range t.Contains {
		if strings.Contains(c.Chinese, ch) {
			return true
		}
	}
	for _, w := range strings.FieldsFunc(strings.ToLower(c.English), func(r rune) bool { return !unicode.IsLetter(r) }) {
		if contains(t.Keywords, w) || contains(t.Keywords, strings.TrimSuffix(w, "s")) {
			return true
		}
	}
	return false
}

func contains(list []string, s string) bool {
	for _, x := range list {
		if x == s {
			return true
		}
	}
	return false
}

// MinStageWords is the fewest words a stage's battle draws from before it
// widens the pool.
const MinStageWords = 12

// stageMaxLevel caps a stage's words at an HSK level that grows along the
// journey (HSK2 at the start, HSK6 by the end), so early fights use early words.
func stageMaxLevel(n int) int { return min(6, 2+(n-1)/16) }

// StageCards picks the words a stage's battle quizzes: its theme's words up
// to the stage's HSK cap. With too few, it lifts the cap, then drops the
// theme but keeps the cap, then uses every card.
func StageCards(s Stage, pages []model.Page, cards []model.Card) []model.Card {
	byPath := map[string]model.Page{}
	for _, p := range pages {
		byPath[p.Path] = p
	}
	capLevel := stageMaxLevel(s.Number)
	pick := func(themed, capped bool) []model.Card {
		var out []model.Card
		for _, c := range cards {
			p := byPath[strings.SplitN(c.ID, "#", 2)[0]]
			if c.Chinese == "" || (capped && hsk(p.Level) > capLevel) || (themed && !s.Words.Matches(c, p)) {
				continue
			}
			out = append(out, c)
		}
		return out
	}
	for _, try := range [][2]bool{{true, true}, {true, false}, {false, true}} {
		if got := pick(try[0], try[1]); len(got) >= MinStageWords {
			return got
		}
	}
	return cards
}

// hsk reads "HSK3" as 3; anything else (unlevelled, "advanced") as 7.
func hsk(level string) int {
	if len(level) == 4 && strings.HasPrefix(level, "HSK") && level[3] >= '1' && level[3] <= '6' {
		return int(level[3] - '0')
	}
	return 7
}

// RadicalIndex maps each character with a page to its radical, written
// "火 (fire)", for explaining words through their characters.
func RadicalIndex(pages []model.Page) map[string]string {
	idx := map[string]string{}
	for _, p := range pages {
		r := []rune(strings.TrimSpace(p.Radical))
		if p.Type != "character" || p.Character == "" || len(r) == 0 {
			continue
		}
		rad := string(r[0])
		if gloss := strings.TrimSpace(string(r[1:])); gloss != "" {
			rad += " " + gloss
		}
		idx[p.Character] = rad
	}
	return idx
}

// WordProgress is how many of a stage's words the learner has learned.
type WordProgress struct {
	Learned, Total int
}

// StageProgress counts the learned words among those a stage's battle uses.
func StageProgress(s Stage, pages []model.Page, cards []model.Card) WordProgress {
	var wp WordProgress
	for _, c := range StageCards(s, pages, cards) {
		wp.Total++
		if c.ReviewInterval >= content.Progression.LearningInterval {
			wp.Learned++
		}
	}
	return wp
}
