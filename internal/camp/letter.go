package camp

import (
	"fmt"
	"strings"
	"time"

	"github.com/therealsiege/tuilaoshi/internal/journey"
	"github.com/therealsiege/tuilaoshi/internal/wiki"
)

// Guanyin's letter (观音来信): each Monday a short letter in Chinese about
// the week before, written from the review log. Shared by the Mac's camp
// and the phone's.

// shownHardInLetter is how many hard words a letter names.
const shownHardInLetter = 3

// WeekLetter writes the letter for the week before now's Monday, or
// returns ok false when that week had no reviews.
func WeekLetter(store *wiki.Store, now time.Time) (wiki.Letter, bool, error) {
	monday := journey.WeekStart(now)
	lastMonday := monday.AddDate(0, 0, -7)
	right, total, err := store.ReviewsBetween(lastMonday, monday)
	if err != nil || total == 0 {
		return wiki.Letter{}, false, err
	}
	hard, err := store.HardWords()
	if err != nil {
		return wiki.Letter{}, false, err
	}
	var hardWords []string
	if len(hard) > 0 {
		cards, err := store.DrillableCards()
		if err != nil {
			return wiki.Letter{}, false, err
		}
		chinese := map[string]string{}
		for _, c := range cards {
			chinese[c.ID] = c.Chinese
		}
		for _, h := range hard {
			if w := chinese[h.CardID]; w != "" && len(hardWords) < shownHardInLetter {
				hardWords = append(hardWords, w)
			}
		}
	}
	missed := len(store.MissedSince(lastMonday))
	ahead, err := store.DueForecast(now, 7)
	if err != nil {
		return wiki.Letter{}, false, err
	}
	due := 0
	for _, n := range ahead {
		due += n
	}
	zh, en := ComposeLetter(total, right*100/total, hardWords, due, missed)
	return wiki.Letter{Week: monday.Format("2006-01-02"), Text: zh, English: en}, true, nil
}

// ComposeLetter writes the letter's Chinese and English from the week's
// numbers.
func ComposeLetter(reviews, pct int, hard []string, due, missed int) (string, string) {
	zh := []string{"圣僧：", fmt.Sprintf("上个星期你复习了%d次，记住了百分之%d。", reviews, pct)}
	en := []string{"Holy monk,", fmt.Sprintf("Last week you reviewed %d times and remembered %d%%.", reviews, pct)}
	switch {
	case pct >= 85:
		zh, en = append(zh, "你学得很认真，我很高兴。"), append(en, "You have studied with care, and I am glad.")
	case pct >= 70:
		zh, en = append(zh, "不错，再加把劲。"), append(en, "Not bad; push a little harder.")
	default:
		zh, en = append(zh, "别着急，慢慢来。忘了的词多复习几次就记住了。"),
			append(en, "Don't worry, take your time. Words you forget stay once you review them a few more times.")
	}
	if len(hard) > 0 {
		zh = append(zh, "有几个词常常忘记："+strings.Join(hard, "、")+"。可以去字炉看看它们是怎么组成的。")
		en = append(en, "A few words keep slipping: "+strings.Join(hard, ", ")+". Look at how they are built in the forge.")
	}
	if missed > 0 {
		zh = append(zh, fmt.Sprintf("上个星期你答错了%d个词，心魔正在等着它们。", missed))
		en = append(en, fmt.Sprintf("You missed %d words last week, and the Inner Demon is waiting for them.", missed))
	}
	zh = append(zh, fmt.Sprintf("这个星期有%d个词要复习。西天路远，每天走一点儿就好。", due), "南海 观音")
	en = append(en, fmt.Sprintf("This week %d words come up for review. The road west is long; a little each day is enough.", due), "Guanyin of the Southern Sea")
	return strings.Join(zh, "\n"), strings.Join(en, "\n")
}

// LatestLetter writes this week's letter if it is due and returns the
// latest one, or nil when there has been no week with reviews.
func LatestLetter(store *wiki.Store, now time.Time) (*wiki.Letter, error) {
	letters, err := store.Letters()
	if err != nil {
		return nil, err
	}
	week := journey.WeekStart(now).Format("2006-01-02")
	if len(letters) == 0 || letters[len(letters)-1].Week != week {
		l, ok, err := WeekLetter(store, now)
		if err != nil {
			return nil, err
		}
		if ok {
			if err := store.SaveLetter(l); err != nil {
				return nil, err
			}
			letters = append(letters, l)
		}
	}
	if len(letters) == 0 {
		return nil, nil
	}
	return &letters[len(letters)-1], nil
}
