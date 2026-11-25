package wiki

import (
	"errors"
	"fmt"
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
// kept the old way, as journey.yaml, is moved into the saves first.
func (s *Store) Saves() ([]SaveInfo, error) {
	if err := s.adoptOldJourney(); err != nil {
		return nil, err
	}
	names, err := s.docNames(savesDir + "/")
	if err != nil {
		return nil, err
	}
	var saves []SaveInfo
	for _, name := range names {
		id, ok := strings.CutSuffix(strings.TrimPrefix(name, savesDir+"/"), ".yaml")
		if !ok || strings.Contains(id, "/") {
			continue
		}
		var rec journeyRecord
		if err := s.readYAML(name, &rec); err != nil {
			return nil, fmt.Errorf("save %s: %w", id, err)
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

// saveExists reports whether a journey is saved under id.
func (s *Store) saveExists(id string) (bool, error) {
	_, ok, err := s.readDoc(savesDir + "/" + id + ".yaml")
	return ok, err
}

// NewSave starts a journey named name and makes it the one being played.
func (s *Store) NewSave(name string, now time.Time) (string, error) {
	id := now.Format("20060102-150405")
	taken, err := s.saveExists(id)
	if err != nil {
		return "", err
	}
	if taken {
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
	ok, err := s.saveExists(id)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("no saved journey %s", id)
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
	return s.deleteDoc(savesDir + "/" + id + ".yaml")
}

// adoptOldJourney moves a journey kept as journey.yaml into the saves,
// and its study bookkeeping into study.yaml, once.
func (s *Store) adoptOldJourney() error {
	data, ok, err := s.readDoc(journeyFile)
	if err != nil || !ok {
		return err
	}
	if _, has, err := s.readDoc(studyFile); err != nil {
		return err
	} else if !has {
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
	when, err := s.docUpdated(journeyFile)
	if err != nil {
		return err
	}
	stamp := when.Format(time.RFC3339)
	rec.Created, rec.Played = stamp, stamp
	saved := s.Save
	s.Save = "journey-1"
	err = s.writeJourney(rec)
	s.Save = saved
	if err != nil {
		return err
	}
	return s.deleteDoc(journeyFile)
}
