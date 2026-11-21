package journey

import "github.com/therealsiege/tuilaoshi/internal/model"

// StageStatus is where a stage stands on the pilgrimage.
type StageStatus int

const (
	StageLocked  StageStatus = iota // needs more mastered characters, or an earlier stage
	StageOpen                       // ready: winning its battle clears it
	StageCleared                    // won
)

// Stage is a tribulation as a step of the game. It opens once enough
// characters are mastered and every earlier stage is cleared, and it is
// cleared by winning its battle, so the story moves only when the learner
// both studies and plays.
type Stage struct {
	Tribulation
	Status   StageStatus
	Opponent BestiaryEntry // the demon or boss fought here
	Boss     *Boss         // set when this is a boss stage
}

// Stages lays out all 81 stages for a learner with mastered characters
// and cleared stages (tribulation number -> date).
func Stages(mastered int, cleared map[int]string) []Stage {
	bosses := map[int]*Boss{}
	list := Bosses(mastered, cleared)
	for i := range list {
		bosses[list[i].Stage] = &list[i]
	}
	stages := make([]Stage, len(content.Tribulations))
	opponent := content.Bestiary[0]
	allBefore := true // every earlier stage cleared
	for i, t := range content.Tribulations {
		if d := BestiaryLookup(t.Demon); d != nil {
			opponent = *d
		}
		s := Stage{Tribulation: t, Opponent: opponent, Boss: bosses[t.Number]}
		if s.Boss != nil {
			s.Opponent = BossEntry(*s.Boss)
		}
		_, won := cleared[t.Number]
		switch {
		case won:
			s.Status = StageCleared
		case allBefore && mastered >= t.Unlock:
			s.Status = StageOpen
		}
		allBefore = allBefore && won
		stages[i] = s
	}
	return stages
}

// CurrentStage returns the first stage not yet cleared, or nil when the
// pilgrimage is complete.
func CurrentStage(stages []Stage) *Stage {
	for i := range stages {
		if stages[i].Status != StageCleared {
			return &stages[i]
		}
	}
	return nil
}

// Bosses lists the journey's bosses in story order. A boss is unlocked when
// its stage's characters are mastered and defeated when its stage is cleared.
func Bosses(mastered int, cleared map[int]string) []Boss {
	bosses := append([]Boss{}, content.Bosses...)
	for i := range bosses {
		if t := TribulationAt(bosses[i].Stage); t != nil {
			bosses[i].Required = t.Unlock
		}
		bosses[i].Unlocked = mastered >= bosses[i].Required
		_, bosses[i].Defeated = cleared[bosses[i].Stage]
	}
	return bosses
}

// NextBoss returns the first boss not yet beaten, or nil when all are.
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

// Pilgrimage reports how far along the journey the learner is: at the
// first uncleared stage, with the share of stages cleared.
func Pilgrimage(pages []model.Page, cleared map[int]string) Progress {
	totalChars := 0
	for _, p := range pages {
		if p.Type == "character" {
			totalChars++
		}
	}
	mastered := CountMastered(pages)
	current := len(content.Tribulations)
	if s := CurrentStage(Stages(mastered, cleared)); s != nil {
		current = s.Number
	}
	return calculateProgress(current, len(cleared), mastered, totalChars)
}
