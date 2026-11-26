package wiki

import (
	"slices"
	"time"

	"gopkg.in/yaml.v3"
)

// A journey's game state (story, purse, gear, bonds, quests, collection)
// is kept in a save, saves/<id>.yaml in the progress file, so several
// journeys can be played; with no save chosen (Store.Save empty) it is
// journey.yaml.
// Study progress is shared by every journey: word schedules live on the
// pages, and the day's study bookkeeping in study.yaml (see study.go).
const journeyFile = "journey.yaml"

type journeyRecord struct {
	Name    string `yaml:"name,omitempty"`    // the journey's name, shown when loading
	Created string `yaml:"created,omitempty"` // RFC 3339
	Played  string `yaml:"played,omitempty"`  // last played, RFC 3339

	Cleared      map[int]string    `yaml:"cleared"`
	Seen         []int             `yaml:"seen,omitempty"` // stages whose story scene has been shown
	Coins        Wallet            `yaml:"coins,omitempty"`
	Owned        []string          `yaml:"owned,omitempty"`         // gear bought at the shrine, by id
	Worn         map[string]string `yaml:"worn,omitempty"`          // slot -> gear id
	Defeated     []string          `yaml:"defeated,omitempty"`      // enemies beaten at least once, for the collection
	Claimed      []string          `yaml:"claimed,omitempty"`       // collection sets whose reward has been paid
	Bonds        map[string]int    `yaml:"bonds,omitempty"`         // companion -> bond points
	TalkDay      string            `yaml:"talk_day,omitempty"`      // the day TalkPoints counts
	TalkPoints   map[string]int    `yaml:"talk_points,omitempty"`   // bond points from talking today, by companion
	QuestsDone   []string          `yaml:"quests_done,omitempty"`   // side quests completed
	QuestsLocked map[string]string `yaml:"quests_locked,omitempty"` // quest id -> the day it was failed
	Stones       []string          `yaml:"stones,omitempty"`        // town stones opened, by town id
	Hunts        map[string]int    `yaml:"hunts,omitempty"`         // hunt quests accepted: id -> soldiers beaten
	Nightmares   []string          `yaml:"nightmares,omitempty"`    // bosses beaten in a nightmare rematch
	Materials    map[string]int    `yaml:"materials,omitempty"`     // material id -> count held, for crafting
	Region       string            `yaml:"region,omitempty"`        // where the pilgrim stands: region id and cell
	X            int               `yaml:"x,omitempty"`
	Y            int               `yaml:"y,omitempty"`
	Wounds       int               `yaml:"wounds,omitempty"`        // HP lost on the road; rest and study heal it
	Trail        yaml.Node         `yaml:"trail,omitempty"`         // the trail's state (journey.Trail), kept opaque here
	Alms         map[string]string `yaml:"alms,omitempty"`          // town id -> the day alms were last asked there
	Diary        []DiaryEntry      `yaml:"diary,omitempty"`         // the travel diary, oldest first
	Offered      map[string]string `yaml:"offered,omitempty"`       // town id -> the day an offering was last made there
	Riddles      map[string]int    `yaml:"riddles,omitempty"`       // town@day -> lantern riddles finished there today
	Stamps       []string          `yaml:"stamps,omitempty"`        // kingdoms (by town) that stamped the travel rescript
	Mysteries    map[string]string `yaml:"eyes,omitempty"`          // town id -> the day its Fiery Eyes mystery was settled
	GrammarGates []string          `yaml:"grammar_gates,omitempty"` // regions whose grammar gate (语法关) is passed
}

// journeyDoc names the save of the journey being played (see progress.go).
func (s *Store) journeyDoc() string {
	if s.Save != "" {
		return savesDir + "/" + s.Save + ".yaml"
	}
	return journeyFile
}

func (s *Store) readJourney() (journeyRecord, error) {
	rec := journeyRecord{}
	err := s.readYAML(s.journeyDoc(), &rec)
	if rec.Cleared == nil {
		rec.Cleared = map[int]string{}
	}
	return rec, err
}

func (s *Store) writeJourney(rec journeyRecord) error {
	return s.writeYAML(s.journeyDoc(), rec)
}

// ClearedStages returns the date each stage was cleared, by tribulation
// number. A missing file means none has been cleared yet.
func (s *Store) ClearedStages() (map[int]string, error) {
	rec, err := s.readJourney()
	return rec.Cleared, err
}

// RecordStageCleared saves a win that clears a stage. Winning again keeps
// the date of the first clear.
func (s *Store) RecordStageCleared(stage int, when time.Time) error {
	rec, err := s.readJourney()
	if err != nil {
		return err
	}
	if _, ok := rec.Cleared[stage]; ok {
		return nil
	}
	rec.Cleared[stage] = when.Format(dateFmt)
	return s.writeJourney(rec)
}

// SceneSeen reports whether a stage's story scene has been shown.
func (s *Store) SceneSeen(stage int) (bool, error) {
	rec, err := s.readJourney()
	return slices.Contains(rec.Seen, stage), err
}

// MarkSceneSeen records that a stage's story scene has been shown.
func (s *Store) MarkSceneSeen(stage int) error {
	rec, err := s.readJourney()
	if err != nil || slices.Contains(rec.Seen, stage) {
		return err
	}
	rec.Seen = append(rec.Seen, stage)
	return s.writeJourney(rec)
}
