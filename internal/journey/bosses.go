package journey

import "github.com/therealsiege/tuilaoshi/internal/model"

// Bosses lists the journey's bosses in order. Mastering a boss's Required
// characters unlocks the fight; victories (Chinese name -> date) mark the
// ones beaten in battle.
func Bosses(mastered int, victories map[string]string) []Boss {
	bosses := []Boss{
		{
			Name:     "混世魔王",
			English:  "Havoc Monster",
			Theme:    "Basic radicals",
			Required: 100,
			Unlocked: mastered >= 100,
		},
		{
			Name:     "白骨精",
			English:  "White Bone Demon",
			Theme:    "Common characters",
			Required: 250,
			Unlocked: mastered >= 250,
		},
		{
			Name:     "红孩儿",
			English:  "Red Boy",
			Theme:    "Fire radicals",
			Required: 500,
			Unlocked: mastered >= 500,
		},
		{
			Name:     "蜘蛛精",
			English:  "Spider Demon",
			Theme:    "Complex characters",
			Required: 800,
			Unlocked: mastered >= 800,
		},
		{
			Name:     "金角大王",
			English:  "Gold Horn King",
			Theme:    "Metal radicals",
			Required: 1200,
			Unlocked: mastered >= 1200,
		},
		{
			Name:     "牛魔王",
			English:  "Bull Demon King",
			Theme:    "Advanced vocabulary",
			Required: 1800,
			Unlocked: mastered >= 1800,
		},
		{
			Name:     "金翅大鹏",
			English:  "Golden-Winged Roc",
			Theme:    "Literary characters",
			Required: 2500,
			Unlocked: mastered >= 2500,
		},
		{
			Name:     "如来佛祖",
			English:  "Buddha (Final Trial)",
			Theme:    "Complete mastery",
			Required: 4000,
			Unlocked: mastered >= 4000,
		},
	}

	for i := range bosses {
		_, bosses[i].Defeated = victories[bosses[i].Name]
	}
	return bosses
}

// NextBoss returns the first boss not yet beaten, or nil when all are.
// Bosses are fought in order.
func NextBoss(bosses []Boss) *Boss {
	for i := range bosses {
		if !bosses[i].Defeated {
			return &bosses[i]
		}
	}
	return nil
}

// BossEntry returns the bestiary entry used to fight boss, matching by
// Chinese or English name; bosses missing from the bestiary get a bare entry.
func BossEntry(boss Boss) BestiaryEntry {
	for _, d := range Bestiary() {
		if d.Name == boss.Name || d.English == boss.English {
			return d
		}
	}
	return BestiaryEntry{Name: boss.Name, English: boss.English, Chapter: "Chapter 100"}
}

// Pilgrimage reports how far along the journey the learner is.
func Pilgrimage(pages []model.Page) Progress {
	totalChars := 0
	for _, p := range pages {
		if p.Type == "character" {
			totalChars++
		}
	}
	return calculateProgress(CountMastered(pages), totalChars)
}
