package journey

import "github.com/therealsiege/tuilaoshi/internal/model"

// Bosses lists the journey's bosses in order. Mastering a boss's Required
// characters unlocks the fight; victories (Chinese name -> date) mark the
// ones beaten in battle.
func Bosses(mastered int, victories map[string]string) []Boss {
	bosses := append([]Boss{}, content.Bosses...)
	for i := range bosses {
		bosses[i].Unlocked = mastered >= bosses[i].Required
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
