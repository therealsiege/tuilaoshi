package journey

import (
	"fmt"

	"github.com/cj/mandarin-wiki/internal/wiki"
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
	Name        string // Chinese name
	English     string
	Emoji       string // Simple ASCII representation
	Role        string // What they help with in the app
	Unlocked    bool
	UnlockAt    int    // Characters needed to unlock
	Ability     string // ability name
	AbilityDesc string // what the ability does
	AbilityIcon string // single ASCII char
}

// Boss is a demon king encounter.
type Boss struct {
	Name     string // Chinese name
	English  string
	Theme    string // What type of characters
	Required int    // Characters mastered to defeat
	Defeated bool
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
	Artifacts    []Artifact
	Quests       []Quest
	Achievements []Achievement
	Lore         []LoreEntry
}

// Calculate computes the complete gamification state from the wiki store.
func Calculate(store *wiki.Store) State {
	pages, _ := store.ListPages()

	// Count character pages and mastered characters
	totalChars := 0
	mastered := 0
	for _, p := range pages {
		if p.Type == "character" {
			totalChars++
			if p.ReviewInterval >= 7 {
				mastered++
			}
		}
	}

	streak := store.Streak()

	nextBoss := findNextBoss(mastered)

	// totalReviewed and daysStudied are not yet tracked — pass 0 for now.
	// Achievements depending on those metrics will simply show as unearned.
	totalReviewed := 0
	daysStudied := 0

	progress := calculateProgress(mastered, totalChars)

	return State{
		Rank:         calculateRank(mastered),
		Progress:     progress,
		Companions:   buildCompanions(mastered),
		Bosses:       buildBosses(mastered),
		NextBoss:     nextBoss,
		Artifacts:    buildArtifacts(mastered, streak),
		Quests:       buildQuests(mastered, streak, nextBoss),
		Achievements: buildAchievements(mastered, totalReviewed, streak, daysStudied),
		Lore:         buildLore(progress.Tribulation),
	}
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

func calculateProgress(mastered, totalChars int) Progress {
	// Calculate tribulation (1-81)
	tribulation := 1
	if totalChars > 0 {
		tribulation = (mastered * 81) / totalChars
		if tribulation < 1 {
			tribulation = 1
		}
		if tribulation > 81 {
			tribulation = 81
		}
	}

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

	percent := 0.0
	if totalChars > 0 {
		percent = (float64(mastered) / float64(totalChars)) * 100.0
		if percent > 100.0 {
			percent = 100.0
		}
	}

	return Progress{
		Tribulation: tribulation,
		Location:    location,
		LocationEn:  locationEn,
		Mastered:    mastered,
		TotalChars:  totalChars,
		Percent:     percent,
	}
}

func buildCompanions(mastered int) []Companion {
	companions := []Companion{
		{
			Name:     "唐僧",
			English:  "Tang Sanzang",
			Emoji:    "@",
			Role:     "Guides your journey",
			UnlockAt: 0,
			Unlocked: mastered >= 0,
		},
		{
			Name:        "孙悟空",
			English:     "Sun Wukong",
			Emoji:       "M",
			Role:        "Battles difficult characters",
			UnlockAt:    50,
			Unlocked:    mastered >= 50,
			Ability:     "火眼金睛",
			AbilityDesc: "Reveals mnemonic decompositions for difficult characters",
			AbilityIcon: "◎",
		},
		{
			Name:        "猪八戒",
			English:     "Zhu Bajie",
			Emoji:       "P",
			Role:        "Helps with radicals",
			UnlockAt:    200,
			Unlocked:    mastered >= 200,
			Ability:     "懒人有懒福",
			AbilityDesc: "Reduces streak penalty: mistakes only reset half the streak",
			AbilityIcon: "~",
		},
		{
			Name:        "沙悟净",
			English:     "Sha Wujing",
			Emoji:       "S",
			Role:        "Organizes reviews",
			UnlockAt:    500,
			Unlocked:    mastered >= 500,
			Ability:     "勤勤恳恳",
			AbilityDesc: "Marks items for review: highlights characters due for review",
			AbilityIcon: "·",
		},
		{
			Name:        "白龙马",
			English:     "Bai Longma",
			Emoji:       "H",
			Role:        "Speeds up learning",
			UnlockAt:    1000,
			Unlocked:    mastered >= 1000,
			Ability:     "日行千里",
			AbilityDesc: "Speeds up journey: progress counts double on review days",
			AbilityIcon: "»",
		},
	}

	return companions
}

func buildBosses(mastered int) []Boss {
	bosses := []Boss{
		{
			Name:     "混世魔王",
			English:  "Havoc Monster",
			Theme:    "Basic radicals",
			Required: 100,
			Defeated: mastered >= 100,
		},
		{
			Name:     "白骨精",
			English:  "White Bone Demon",
			Theme:    "Common characters",
			Required: 250,
			Defeated: mastered >= 250,
		},
		{
			Name:     "红孩儿",
			English:  "Red Boy",
			Theme:    "Fire radicals",
			Required: 500,
			Defeated: mastered >= 500,
		},
		{
			Name:     "蜘蛛精",
			English:  "Spider Demon",
			Theme:    "Complex characters",
			Required: 800,
			Defeated: mastered >= 800,
		},
		{
			Name:     "金角大王",
			English:  "Gold Horn King",
			Theme:    "Metal radicals",
			Required: 1200,
			Defeated: mastered >= 1200,
		},
		{
			Name:     "牛魔王",
			English:  "Bull Demon King",
			Theme:    "Advanced vocabulary",
			Required: 1800,
			Defeated: mastered >= 1800,
		},
		{
			Name:     "大鹏金翅鸟",
			English:  "Golden-Winged Roc",
			Theme:    "Literary characters",
			Required: 2500,
			Defeated: mastered >= 2500,
		},
		{
			Name:     "如来佛祖",
			English:  "Buddha (Final Trial)",
			Theme:    "Complete mastery",
			Required: 4000,
			Defeated: mastered >= 4000,
		},
	}

	return bosses
}

func findNextBoss(mastered int) *Boss {
	bosses := buildBosses(mastered)
	for i := range bosses {
		if !bosses[i].Defeated {
			return &bosses[i]
		}
	}
	return nil
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
			Desc:    fmt.Sprintf("Defeat %s %s: master %d characters", nextBoss.Name, nextBoss.English, nextBoss.Required),
			Target:  nextBoss.Required,
			Current: mastered,
			Done:    mastered >= nextBoss.Required,
		})
	}

	return quests
}
