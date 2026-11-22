package wiki

import (
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

// payForSession pays a copper coin for each right answer in a session just
// saved, and the daily bonus the first time today's answers reach the goal.
func (s *Store) payForSession(sess model.Session) error {
	pay := Wallet{}
	for _, it := range sess.Items {
		if it.Correct {
			pay.Tongqian++
		}
	}
	rec, err := s.readJourney()
	if err != nil {
		return err
	}
	today := time.Now().Format(dateFmt)
	if rec.DailyPaid != today && s.TodayStats().CardsReviewed >= DailyGoal {
		pay.Tongqian += dailyBonus + min(s.Streak(), streakBonusCap)
		rec.DailyPaid = today
	}
	if pay == (Wallet{}) {
		return nil
	}
	rec.Coins.Tongqian += pay.Tongqian
	return s.writeJourney(rec)
}
