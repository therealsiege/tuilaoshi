package journey

import (
	"fmt"

	"github.com/therealsiege/tuilaoshi/internal/model"
	"github.com/therealsiege/tuilaoshi/internal/wiki"
)

// Rank represents the player's cultivation level, inspired by xianxia novels.
type Rank struct {
	Name        string // Chinese name: 凡人, 练气, 筑基, 金丹, 元婴, 化神, 大乘, 斗战胜佛
	English     string // English translation
	Level       int    // 0-7
	Current     int    // characters mastered at current rank
	Required    int    // total needed for next rank
	NextName    string // Chinese name of the next rank (empty at max rank)
	NextEnglish string // English name of the next rank (empty at max rank)
}

// Progress tracks position on the Journey to the West pilgrimage.
type Progress struct {
	Tribulation int     // 1-81 (the 81 tribulations)
	Location    string  // Current location name (Chinese)
	LocationEn  string  // English name
	Mastered    int     // Total characters with review_interval >= 7
	TotalChars  int     // Total character pages
	Percent     float64 // Overall journey completion 0-100
}

// Companion is a party member from Journey to the West.
type Companion struct {
	Name        string `yaml:"name,omitempty"` // Chinese name
	English     string `yaml:"english,omitempty"`
	Emoji       string `yaml:"emoji,omitempty"` // Simple ASCII representation
	Role        string `yaml:"role,omitempty"`  // What they help with in the app
	Unlocked    bool   `yaml:"-"`
	Stage       int    `yaml:"stage"`                  // tribulation at which they join the party
	Ability     string `yaml:"ability,omitempty"`      // ability name
	AbilityDesc string `yaml:"ability_desc,omitempty"` // what the ability does
	AbilityIcon string `yaml:"ability_icon,omitempty"` // single ASCII char
	Effect      string `yaml:"effect,omitempty"`       // battle ability by name (see rpg.Effects)
}

// Boss is a demon king encounter.
type Boss struct {
	Name     string `yaml:"name,omitempty"` // Chinese name
	English  string `yaml:"english,omitempty"`
	Theme    string `yaml:"-"`     // name of its stage's vocabulary theme
	Stage    int    `yaml:"stage"` // tribulation where it is fought
	Required int    `yaml:"-"`     // characters mastered to open that stage
	Unlocked bool   `yaml:"-"`     // enough characters mastered to challenge it
	Defeated bool   `yaml:"-"`     // beaten in battle (recorded in wiki/bosses.yaml)
}

// Artifact is a collectible power-up earned through study.
type Artifact struct {
	Name    string
	Chinese string
	Desc    string
	Earned  bool
	EarnAt  int    // streak days or chars needed
	Kind    string // "mastery" or "streak"
}

// Quest is a daily objective for the player.
type Quest struct {
	Name    string // e.g. "Flame Quencher"
	Chinese string // e.g. "灭火者"
	Desc    string // e.g. "Master 3 fire-radical characters"
	Target  int    // number needed
	Current int    // current progress
	Done    bool
}

// State is the complete gamification state.
type State struct {
	Rank         Rank
	Progress     Progress
	Companions   []Companion
	Bosses       []Boss
	NextBoss     *Boss
	Stage        *Stage // the stage being worked on; nil once all are cleared
	Counts       Counts // characters being learned and mastered, which open stages
	Artifacts    []Artifact
	Quests       []Quest
	Achievements []Achievement
	Lore         []LoreEntry
}

// masteredInterval is the review interval, in days, at which a character
// counts as mastered for ranks, bosses, and battle stats.
const masteredInterval = 7

// IsMastered reports whether a page counts toward mastery.
func IsMastered(p model.Page) bool {
	return p.Type == "character" && p.ReviewInterval >= masteredInterval
}

// Counts are the learner's characters that open stages: those being learned
// (answered correctly at least once) and those mastered.
type Counts struct {
	Learned, Mastered int
}

// CountProgress counts the character pages being learned and mastered.
func CountProgress(pages []model.Page) Counts {
	c := Counts{Mastered: CountMastered(pages)}
	for _, p := range pages {
		if p.Type == "character" && p.ReviewInterval >= content.Progression.LearningInterval {
			c.Learned++
		}
	}
	return c
}

// CountMastered returns how many pages count toward mastery.
func CountMastered(pages []model.Page) int {
	count := 0
	for _, p := range pages {
		if IsMastered(p) {
			count++
		}
	}
	return count
}

// Calculate computes the complete gamification state from the wiki store.
func Calculate(store *wiki.Store) (State, error) {
	pages, err := store.ListPages()
	if err != nil {
		return State{}, err
	}

	counts := CountProgress(pages)
	mastered := counts.Mastered

	streak := store.Streak()

	cleared, err := store.ClearedStages()
	if err != nil {
		return State{}, err
	}
	bosses := Bosses(counts, cleared)
	nextBoss := NextBoss(bosses)
	stages := Stages(counts, cleared)

	// totalReviewed and daysStudied are not yet tracked — pass 0 for now.
	// Achievements depending on those metrics will simply show as unearned.
	totalReviewed := 0
	daysStudied := 0

	progress := Pilgrimage(pages, cleared)

	return State{
		Rank:         calculateRank(mastered),
		Progress:     progress,
		Stage:        CurrentStage(stages),
		Counts:       counts,
		Companions:   buildCompanions(cleared),
		Bosses:       bosses,
		NextBoss:     nextBoss,
		Artifacts:    buildArtifacts(mastered, streak),
		Quests:       buildQuests(mastered, streak, nextBoss),
		Achievements: buildAchievements(mastered, totalReviewed, streak, daysStudied),
		Lore:         buildLore(progress.Tribulation),
	}, nil
}

func calculateRank(mastered int) Rank {
	ranks := []struct {
		threshold int
		name      string
		english   string
		level     int
		next      int
	}{
		{0, "凡人", "Mortal", 0, 100},
		{100, "练气", "Qi Refining", 1, 300},
		{300, "筑基", "Foundation Building", 2, 600},
		{600, "金丹", "Golden Core", 3, 1200},
		{1200, "元婴", "Nascent Soul", 4, 2000},
		{2000, "化神", "Spirit Severing", 5, 3000},
		{3000, "大乘", "Mahayana", 6, 5000},
		{5000, "斗战胜佛", "Victorious Fighting Buddha", 7, 0},
	}

	for i := len(ranks) - 1; i >= 0; i-- {
		if mastered >= ranks[i].threshold {
			current := mastered - ranks[i].threshold
			required := ranks[i].next
			if required == 0 {
				required = ranks[i].threshold
				current = mastered
			}
			r := Rank{
				Name:     ranks[i].name,
				English:  ranks[i].english,
				Level:    ranks[i].level,
				Current:  current,
				Required: required,
			}
			if i+1 < len(ranks) {
				r.NextName = ranks[i+1].name
				r.NextEnglish = ranks[i+1].english
			}
			return r
		}
	}

	return Rank{
		Name:        "凡人",
		English:     "Mortal",
		Level:       0,
		Current:     mastered,
		Required:    100,
		NextName:    "练气",
		NextEnglish: "Qi Refining",
	}
}

// calculateProgress places the learner at tribulation (the current stage)
// with cleared stages behind them.
func calculateProgress(tribulation, cleared, mastered, totalChars int) Progress {

	// Journey waypoints
	locations := []struct {
		name    string
		english string
		start   int
	}{
		{"长安", "Chang'an (Capital)", 0},
		{"双叉岭", "Twin Fork Ridge", 10},
		{"五行山", "Five Elements Mountain", 20},
		{"观音院", "Guanyin Temple", 30},
		{"高老庄", "Gao Village", 40},
		{"流沙河", "Flowing Sands River", 50},
		{"火焰山", "Flaming Mountains", 60},
		{"通天河", "Heaven-Reaching River", 70},
		{"灵山", "Spirit Mountain", 80},
		{"大雷音寺", "Great Thunder Monastery", 81},
	}

	location := "长安"
	locationEn := "Chang'an (Capital)"
	for i := len(locations) - 1; i >= 0; i-- {
		if tribulation >= locations[i].start {
			location = locations[i].name
			locationEn = locations[i].english
			break
		}
	}

	percent := float64(cleared) / float64(len(content.Tribulations)) * 100.0

	return Progress{
		Tribulation: tribulation,
		Location:    location,
		LocationEn:  locationEn,
		Mastered:    mastered,
		TotalChars:  totalChars,
		Percent:     percent,
	}
}

// Party returns the companions, marking who has joined for cleared stages.
func Party(cleared map[int]string) []Companion { return buildCompanions(cleared) }

// buildCompanions marks who has joined: each companion joins when the
// stage of their story is cleared (Tang Sanzang travels from the start).
func buildCompanions(cleared map[int]string) []Companion {
	companions := append([]Companion{}, content.Companions...)
	for i := range companions {
		_, done := cleared[companions[i].Stage]
		companions[i].Unlocked = companions[i].Stage <= 1 || done
	}
	return companions
}

func buildArtifacts(mastered, streak int) []Artifact {
	artifacts := []Artifact{
		{
			Name:    "Ruyi Jingu Bang",
			Chinese: "金箍棒",
			Desc:    "Legendary staff of Sun Wukong",
			EarnAt:  100,
			Earned:  mastered >= 100,
			Kind:    "mastery",
		},
		{
			Name:    "Headband Spell",
			Chinese: "紧箍咒",
			Desc:    "Controls the Monkey King",
			EarnAt:  7,
			Earned:  streak >= 7,
			Kind:    "streak",
		},
		{
			Name:    "Banana Leaf Fan",
			Chinese: "芭蕉扇",
			Desc:    "Extinguishes the Flaming Mountains",
			EarnAt:  500,
			Earned:  mastered >= 500,
			Kind:    "mastery",
		},
		{
			Name:    "Ginseng Fruit",
			Chinese: "人参果",
			Desc:    "Grants immortality and wisdom",
			EarnAt:  1000,
			Earned:  mastered >= 1000,
			Kind:    "mastery",
		},
		{
			Name:    "Fiery Eyes",
			Chinese: "火眼金睛",
			Desc:    "See through any deception",
			EarnAt:  2000,
			Earned:  mastered >= 2000,
			Kind:    "mastery",
		},
		{
			Name:    "Sea-Calming Needle",
			Chinese: "定海神针",
			Desc:    "Controls the oceans",
			EarnAt:  3000,
			Earned:  mastered >= 3000,
			Kind:    "mastery",
		},
	}

	return artifacts
}

func buildQuests(mastered, streak int, nextBoss *Boss) []Quest {
	quests := []Quest{
		{
			Name:    "Diligent Scholar",
			Chinese: "勤学者",
			Desc:    "Complete 10 reviews today",
			Target:  10,
			Current: 0,
			Done:    false,
		},
		{
			Name:    "Steady Pilgrim",
			Chinese: "保持连续",
			Desc:    "Maintain your study streak",
			Target:  streak + 1,
			Current: streak,
			Done:    streak > 0,
		},
	}

	if nextBoss != nil {
		quests = append(quests, Quest{
			Name:    "Demon Slayer",
			Chinese: "击败妖怪",
			Desc:    fmt.Sprintf("Unlock %s %s: master %d characters, then win the fight", nextBoss.Name, nextBoss.English, nextBoss.Required),
			Target:  nextBoss.Required,
			Current: mastered,
			Done:    mastered >= nextBoss.Required,
		})
	}

	return quests
}

// NeedText says what a locked stage still needs, e.g. "learn 6 more characters".
func NeedText(s Stage, c Counts) string {
	verb := "master"
	if CountsLearning(s.Number) {
		verb = "learn"
	}
	return fmt.Sprintf("%s %d more characters", verb, s.Needed(c))
}
