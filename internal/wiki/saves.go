package wiki

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// savesDir holds one file per journey: wiki/saves/<id>.yaml.
const savesDir = "saves"

// SaveInfo summarises a saved journey for the title screen.
type SaveInfo struct {
	ID      string
	Name    string
	Created time.Time
	Played  time.Time
	Cleared int // tribulations cleared
	Coins   Wallet
}

// Saves lists the saved journeys, most recently played first. A journey
// kept the old way, in wiki/journey.yaml, is moved into the saves first.
func (s *Store) Saves() ([]SaveInfo, error) {
	if err := s.adoptOldJourney(); err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(filepath.Join(s.Dir, savesDir))
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("list saves: %w", err)
	}
	var saves []SaveInfo
	for _, e := range entries {
		id, ok := strings.CutSuffix(e.Name(), ".yaml")
		if e.IsDir() || !ok {
			continue
		}
		data, err := os.ReadFile(filepath.Join(s.Dir, savesDir, e.Name()))
		if err != nil {
			return nil, fmt.Errorf("read save %s: %w", id, err)
		}
		var rec journeyRecord
		if err := yaml.Unmarshal(data, &rec); err != nil {
			return nil, fmt.Errorf("parse save %s: %w", id, err)
		}
		info := SaveInfo{ID: id, Name: rec.Name, Cleared: len(rec.Cleared), Coins: rec.Coins}
		info.Created, _ = time.Parse(time.RFC3339, rec.Created) //nolint:errcheck // a missing date sorts last
		info.Played, _ = time.Parse(time.RFC3339, rec.Played)   //nolint:errcheck // a missing date sorts last
		if info.Name == "" {
			info.Name = id
		}
		saves = append(saves, info)
	}
	sort.SliceStable(saves, func(i, j int) bool { return saves[i].Played.After(saves[j].Played) })
	return saves, nil
}

// NewSave starts a journey named name and makes it the one being played.
func (s *Store) NewSave(name string, now time.Time) (string, error) {
	id := now.Format("20060102-150405")
	if _, err := os.Stat(filepath.Join(s.Dir, savesDir, id+".yaml")); err == nil {
		id += fmt.Sprintf("-%d", now.Nanosecond())
	}
	s.Save = id
	stamp := now.Format(time.RFC3339)
	return id, s.writeJourney(journeyRecord{Name: name, Created: stamp, Played: stamp, Cleared: map[int]string{}})
}

// UseSave makes a saved journey the one being played, marking it played now.
func (s *Store) UseSave(id string, now time.Time) error {
	if id == "" || strings.ContainsAny(id, `/\`) {
		return errors.New("bad save id")
	}
	if _, err := os.Stat(filepath.Join(s.Dir, savesDir, id+".yaml")); err != nil {
		return fmt.Errorf("no saved journey %s: %w", id, err)
	}
	s.Save = id
	rec, err := s.readJourney()
	if err != nil {
		return err
	}
	rec.Played = now.Format(time.RFC3339)
	return s.writeJourney(rec)
}

// DeleteSave removes a saved journey. Study progress is untouched.
func (s *Store) DeleteSave(id string) error {
	if id == "" || strings.ContainsAny(id, `/\`) {
		return errors.New("bad save id")
	}
	if s.Save == id {
		s.Save = ""
	}
	return os.Remove(filepath.Join(s.Dir, savesDir, id+".yaml"))
}

// adoptOldJourney moves a journey kept in wiki/journey.yaml into the saves,
// and its study bookkeeping into study.yaml, once.
func (s *Store) adoptOldJourney() error {
	old := filepath.Join(s.Dir, journeyFile)
	data, err := os.ReadFile(old)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("read %s: %w", journeyFile, err)
	}
	if _, err := os.Stat(filepath.Join(s.Dir, studyFile)); os.IsNotExist(err) {
		var study studyRecord
		if err := yaml.Unmarshal(data, &study); err != nil {
			return fmt.Errorf("parse %s: %w", journeyFile, err)
		}
		if err := s.writeStudy(study); err != nil {
			return err
		}
	}
	var rec journeyRecord
	if err := yaml.Unmarshal(data, &rec); err != nil {
		return fmt.Errorf("parse %s: %w", journeyFile, err)
	}
	if rec.Name == "" {
		rec.Name = "西行之路 The Road West"
	}
	info, err := os.Stat(old)
	if err != nil {
		return err
	}
	stamp := info.ModTime().Format(time.RFC3339)
	rec.Created, rec.Played = stamp, stamp
	saved := s.Save
	s.Save = "journey-1"
	err = s.writeJourney(rec)
	s.Save = saved
	if err != nil {
		return err
	}
	return os.Rename(old, old+".adopted")
}
