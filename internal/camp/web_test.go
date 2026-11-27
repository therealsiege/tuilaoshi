package camp

import (
	"strings"
	"testing"
	"time"

	"github.com/therealsiege/tuilaoshi/internal/model"
	"github.com/therealsiege/tuilaoshi/internal/wiki"
)

// webNow is the test day: firm words are due after it, slipping ones before.
var webNow = time.Date(2026, 11, 25, 9, 0, 0, 0, time.Local)

// studied makes a studied word due in dueIn days with the given interval.
func studied(chinese string, interval, dueIn, lapses int) wiki.StudiedWord {
	return wiki.StudiedWord{Card: model.Card{
		ID: "vocabulary/x.md#" + chinese, Type: model.CardVocabulary, Chinese: chinese,
		LastReviewed: webNow.AddDate(0, 0, -1), ReviewInterval: interval, NextReview: webNow.AddDate(0, 0, dueIn),
	}, Lapses: lapses}
}

func TestWebLinksWordsSharingACharacter(t *testing.T) {
	g := BuildWeb([]wiki.StudiedWord{
		studied("学生", 10, 5, 0), studied("学习", 10, 5, 0), studied("习惯", 10, 5, 0),
		studied("电脑", 10, 5, 0), studied("学学", 10, 5, 0),
	}, webNow)
	linked := func(a, b string) bool {
		for _, j := range g.Links[g.Index[a]] {
			if g.Words[j].Card.Chinese == b {
				return true
			}
		}
		return false
	}
	for _, pair := range [][2]string{{"学生", "学习"}, {"学习", "习惯"}, {"学学", "学生"}} {
		if !linked(pair[0], pair[1]) || !linked(pair[1], pair[0]) {
			t.Errorf("%s and %s share a character but are not tied", pair[0], pair[1])
		}
	}
	if linked("学生", "习惯") || len(g.Links[g.Index["电脑"]]) != 0 {
		t.Errorf("words with no character in common are tied: %v", g.Links)
	}
	// 学生-学习, 学生-学学, 学习-学学, 学习-习惯: each pair once, however many characters it shares
	if g.Strands != 4 || g.WeakStrands != 0 {
		t.Errorf("strands %d weak %d, want 4 and 0", g.Strands, g.WeakStrands)
	}
}

func TestWebMarksSlippingWords(t *testing.T) {
	g := BuildWeb([]wiki.StudiedWord{
		studied("学生", 10, 5, 0),                    // firm
		studied("学习", 10, -2, 0),                   // overdue
		studied("学校", 2, 1, 0),                     // under 4 days
		studied("学期", 30, 20, wiki.HardWordLapses), // a hard word
		studied("生日", 10, 5, 0),                    // firm, tied only to 学生
	}, webNow)
	for word, weak := range map[string]bool{"学生": false, "学习": true, "学校": true, "学期": true, "生日": false} {
		if got := g.Words[g.Index[word]].Weak; got != weak {
			t.Errorf("%s weak = %v, want %v", word, got, weak)
		}
	}
	// six strands among the four 学 words, each touching a slipping word; 学生-生日 is firm
	if g.Strands != 7 || g.WeakStrands != 6 {
		t.Errorf("strands %d weak %d, want 7 and 6", g.Strands, g.WeakStrands)
	}
	if w := g.Words[g.Weakest()].Card.Chinese; w != "学期" {
		t.Errorf("the web opens on %s, want the hard word 学期", w)
	}
	var revive []string
	for _, c := range g.Cluster(g.Index["学生"]) {
		revive = append(revive, c.Chinese)
	}
	if strings.Join(revive, " ") != "学习 学期 学校" {
		t.Errorf("walking from 学生 revives %v, want its three slipping neighbours", revive)
	}
}

func TestWebKeepsTheWeakerCardOfATwiceStudiedWord(t *testing.T) {
	g := BuildWeb([]wiki.StudiedWord{studied("书", 20, 5, 0), studied("书", 1, 0, 0)}, webNow)
	if len(g.Words) != 1 || g.Words[0].Card.ReviewInterval != 1 {
		t.Errorf("书 twice should appear once as its weaker card: %+v", g.Words)
	}
}
