package wiki

import (
	"slices"
	"time"
)

// QuestState is where a side quest stands.
type QuestState int

const (
	QuestOpen   QuestState = iota
	QuestLocked            // answered wrong today; open again tomorrow
	QuestDone
)

// Quests returns the state of every side quest touched, by id, on now's day.
func (s *Store) Quests(now time.Time) (map[string]QuestState, error) {
	rec, err := s.readJourney()
	states := map[string]QuestState{}
	for _, id := range rec.QuestsDone {
		states[id] = QuestDone
	}
	today := now.Format(dateFmt)
	for id, day := range rec.QuestsLocked {
		if day == today && states[id] != QuestDone {
			states[id] = QuestLocked
		}
	}
	return states, err
}

// FinishQuest records a quest done and pays its reward, once.
func (s *Store) FinishQuest(id string, reward Wallet) error {
	rec, err := s.readJourney()
	if err != nil || slices.Contains(rec.QuestsDone, id) {
		return err
	}
	rec.QuestsDone = append(rec.QuestsDone, id)
	delete(rec.QuestsLocked, id)
	rec.Coins.Tongqian += reward.Tongqian
	rec.Coins.Yuanbao += reward.Yuanbao
	return s.writeJourney(rec)
}

// FailQuest locks a quest for the rest of now's day.
func (s *Store) FailQuest(id string, now time.Time) error {
	rec, err := s.readJourney()
	if err != nil {
		return err
	}
	if rec.QuestsLocked == nil {
		rec.QuestsLocked = map[string]string{}
	}
	rec.QuestsLocked[id] = now.Format(dateFmt)
	return s.writeJourney(rec)
}
