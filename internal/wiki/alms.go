package wiki

import "time"

// AskAlms records asking for alms in a town today; it reports false when
// alms were already asked there today (a town gives once a day).
func (s *Store) AskAlms(town string, now time.Time) (bool, error) {
	rec, err := s.readJourney()
	if err != nil {
		return false, err
	}
	day := now.Format(dateFmt)
	if rec.Alms[town] == day {
		return false, nil
	}
	if rec.Alms == nil {
		rec.Alms = map[string]string{}
	}
	rec.Alms[town] = day
	return true, s.writeJourney(rec)
}

// AlmsAsked reports whether alms were asked in a town today.
func (s *Store) AlmsAsked(town string, now time.Time) (bool, error) {
	rec, err := s.readJourney()
	return rec.Alms[town] == now.Format(dateFmt), err
}

// TakeMaterial removes one of a material from the pack, for trading.
func (s *Store) TakeMaterial(id string) error {
	rec, err := s.readJourney()
	if err != nil {
		return err
	}
	if rec.Materials[id] < 1 {
		return ErrMissingMaterials
	}
	rec.Materials[id]--
	if rec.Materials[id] == 0 {
		delete(rec.Materials, id)
	}
	return s.writeJourney(rec)
}

// DiaryEntry is a day's entry in the travel diary, written in Chinese at
// camp and graded 0-2.
type DiaryEntry struct {
	Date     string `yaml:"date"`
	Day      int    `yaml:"day"` // the day on the trail
	Text     string `yaml:"text"`
	Score    int    `yaml:"score"`
	Feedback string `yaml:"feedback,omitempty"`
}

// Diary returns the journey's travel diary, oldest first.
func (s *Store) Diary() ([]DiaryEntry, error) {
	rec, err := s.readJourney()
	return rec.Diary, err
}

// WriteDiary adds an entry to the travel diary.
func (s *Store) WriteDiary(e DiaryEntry) error {
	rec, err := s.readJourney()
	if err != nil {
		return err
	}
	rec.Diary = append(rec.Diary, e)
	return s.writeJourney(rec)
}
