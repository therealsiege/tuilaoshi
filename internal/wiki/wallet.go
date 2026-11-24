package wiki

import (
	"slices"
	"time"

	"github.com/therealsiege/tuilaoshi/internal/model"
)

// Wallet holds the journey's two currencies: copper cash (铜钱 tóngqián),
// earned by studying and fighting, and silver ingots (元宝 yuánbǎo), the
// rare reward for bosses and milestones.
type Wallet struct {
	Tongqian int `yaml:"tongqian,omitempty"`
	Yuanbao  int `yaml:"yuanbao,omitempty"`
}

const (
	// DailyGoal is how many answers in a day earn the daily bonus.
	DailyGoal = 20
	// dailyBonus is the copper paid for reaching the goal, plus one coin per
	// day of the streak, up to streakBonusCap.
	dailyBonus     = 10
	streakBonusCap = 10
)

// Wallet returns the coins saved so far.
func (s *Store) Wallet() (Wallet, error) {
	rec, err := s.readJourney()
	return rec.Coins, err
}

// Earn adds coins to the wallet.
func (s *Store) Earn(w Wallet) error {
	if w == (Wallet{}) {
		return nil
	}
	rec, err := s.readJourney()
	if err != nil {
		return err
	}
	rec.Coins.Tongqian += w.Tongqian
	rec.Coins.Yuanbao += w.Yuanbao
	return s.writeJourney(rec)
}

// payForSession pays a copper coin for each card answered right in a
// session just saved, once per card per day, so repeating easy words earns
// nothing; and the daily bonus the first time today's answers reach the goal.
// Each of those fresh answers also heals the pilgrim's wounds.
//
// Which words have paid is shared by every journey (study.yaml); the coins
// go to the journey being played.
func (s *Store) payForSession(sess model.Session) error {
	study, err := s.readStudy()
	if err != nil {
		return err
	}
	rec, err := s.readJourney()
	if err != nil {
		return err
	}
	now := time.Now()
	today := now.Format(dateFmt)
	if study.PaidDay != today {
		// The day's first session: shields cover any days missed since the last.
		s.shieldGap(&study, now)
		study.PaidDay, study.PaidCards = today, nil
	}
	pay := Wallet{}
	for _, it := range sess.Items {
		if it.Correct && !slices.Contains(study.PaidCards, it.CardID) {
			study.PaidCards = append(study.PaidCards, it.CardID)
			pay.Tongqian++
		}
	}
	fresh := pay.Tongqian
	tally(&study, &rec, today, sess.Type+"_right", fresh)
	if study.DailyPaid != today {
		// Write first so the stats below count this session.
		if err := s.writeStudy(study); err != nil {
			return err
		}
		if s.TodayStats().CardsReviewed >= DailyGoal {
			pay.Tongqian += dailyBonus + min(s.Streak(), streakBonusCap)
			study.DailyPaid = today
		}
	}
	rec.Wounds = max(rec.Wounds-studyHeal*fresh, 0) // recall restores the pilgrim
	rec.Coins.Tongqian += pay.Tongqian
	if err := s.writeStudy(study); err != nil {
		return err
	}
	return s.writeJourney(rec)
}

// Spend takes coins from the purse, or returns ErrCantAfford.
func (s *Store) Spend(w Wallet) error {
	rec, err := s.readJourney()
	if err != nil {
		return err
	}
	if rec.Coins.Tongqian < w.Tongqian || rec.Coins.Yuanbao < w.Yuanbao {
		return ErrCantAfford
	}
	rec.Coins.Tongqian -= w.Tongqian
	rec.Coins.Yuanbao -= w.Yuanbao
	return s.writeJourney(rec)
}
