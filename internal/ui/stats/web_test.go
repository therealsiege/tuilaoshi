package stats

import (
	"strconv"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/therealsiege/tuilaoshi/internal/model"
	"github.com/therealsiege/tuilaoshi/internal/ui/uitest"
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
	g := buildWeb([]wiki.StudiedWord{
		studied("学生", 10, 5, 0), studied("学习", 10, 5, 0), studied("习惯", 10, 5, 0),
		studied("电脑", 10, 5, 0), studied("学学", 10, 5, 0),
	}, webNow)
	linked := func(a, b string) bool {
		for _, j := range g.links[g.index[a]] {
			if g.words[j].card.Chinese == b {
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
	if linked("学生", "习惯") || len(g.links[g.index["电脑"]]) != 0 {
		t.Errorf("words with no character in common are tied: %v", g.links)
	}
	// 学生-学习, 学生-学学, 学习-学学, 学习-习惯: each pair once, however many characters it shares
	if g.strands != 4 || g.weakStrands != 0 {
		t.Errorf("strands %d weak %d, want 4 and 0", g.strands, g.weakStrands)
	}
}

func TestWebMarksSlippingWords(t *testing.T) {
	g := buildWeb([]wiki.StudiedWord{
		studied("学生", 10, 5, 0),                    // firm
		studied("学习", 10, -2, 0),                   // overdue
		studied("学校", 2, 1, 0),                     // under 4 days
		studied("学期", 30, 20, wiki.HardWordLapses), // a hard word
		studied("生日", 10, 5, 0),                    // firm, tied only to 学生
	}, webNow)
	for word, weak := range map[string]bool{"学生": false, "学习": true, "学校": true, "学期": true, "生日": false} {
		if got := g.words[g.index[word]].weak; got != weak {
			t.Errorf("%s weak = %v, want %v", word, got, weak)
		}
	}
	// six strands among the four 学 words, each touching a slipping word; 学生-生日 is firm
	if g.strands != 7 || g.weakStrands != 6 {
		t.Errorf("strands %d weak %d, want 7 and 6", g.strands, g.weakStrands)
	}
	if w := g.words[g.weakest()].card.Chinese; w != "学期" {
		t.Errorf("the web opens on %s, want the hard word 学期", w)
	}
	var revive []string
	for _, c := range g.cluster(g.index["学生"]) {
		revive = append(revive, c.Chinese)
	}
	if strings.Join(revive, " ") != "学习 学期 学校" {
		t.Errorf("walking from 学生 revives %v, want its three slipping neighbours", revive)
	}
}

func TestWebKeepsTheWeakerCardOfATwiceStudiedWord(t *testing.T) {
	g := buildWeb([]wiki.StudiedWord{studied("书", 20, 5, 0), studied("书", 1, 0, 0)}, webNow)
	if len(g.words) != 1 || g.words[0].card.ReviewInterval != 1 {
		t.Errorf("书 twice should appear once as its weaker card: %+v", g.words)
	}
}

// webWiki is a small wiki: firm words, slipping words, a hard character,
// a word tied to nothing, and words never reviewed (left off the web).
func webWiki(t *testing.T) *wiki.Store {
	t.Helper()
	page := func(interval int, next, rows string) string {
		return "---\ntype: vocabulary\nlast_reviewed: 2026-11-20\nreview_interval: " + strconv.Itoa(interval) +
			"\nnext_review: " + next + "\n---\n\n# Words\n\n| Word | Pinyin | Meaning |\n|------|--------|---------|\n" + rows
	}
	store := uitest.Wiki(t, map[string]string{
		"vocabulary/firm.md": page(12, "2026-12-02", "| 学生 | xuésheng | student |\n| 生活 | shēnghuó | life |\n"+
			"| 习惯 | xíguàn | habit |\n| 校长 | xiàozhǎng | headmaster |\n| 先生 | xiānsheng | sir |\n| 电脑 | diànnǎo | computer |\n"),
		"vocabulary/slipping.md": page(2, "2026-11-22", "| 学习 | xuéxí | to study |\n| 学校 | xuéxiào | school |\n"+
			"| 复习 | fùxí | to revise |\n| 生日 | shēngrì | birthday |\n"),
		// enough 学 words to overflow a 24-line window
		"vocabulary/school.md": page(9, "2026-11-30", "| 大学 | dàxué | university |\n| 数学 | shùxué | maths |\n"+
			"| 化学 | huàxué | chemistry |\n| 同学 | tóngxué | classmate |\n| 留学 | liúxué | study abroad |\n"+
			"| 学费 | xuéfèi | tuition |\n| 中学 | zhōngxué | middle school |\n| 小学 | xiǎoxué | primary school |\n"+
			"| 文学 | wénxué | literature |\n| 科学 | kēxué | science |\n| 学问 | xuéwen | learning |\n| 学者 | xuézhě | scholar |\n"+
			"| 学院 | xuéyuàn | college |\n| 学术 | xuéshù | scholarship |\n"+
			"| 大夫 | dàifu | doctor |\n| 中文 | Zhōngwén | Chinese |\n"),
		"vocabulary/unseen.md": "---\ntype: vocabulary\n---\n\n| Word | Pinyin | Meaning |\n|---|---|---|\n| 学期 | xuéqī | term |\n",
		"characters/xue2.md":   "---\ntype: character\n---\n\n# 学\n\n- **Character**: 学\n- **Pinyin**: xué\n- **Meaning**: to learn\n",
	})
	cards, err := store.DrillableCards()
	if err != nil {
		t.Fatal(err)
	}
	for i := range cards {
		if cards[i].Chinese == "学" {
			for range wiki.HardWordLapses {
				if err := store.RecordReview(&cards[i], model.DiffAgain, webNow.AddDate(0, 0, -3)); err != nil {
					t.Fatal(err)
				}
			}
		}
	}
	return store
}

// webModel opens the web on the test wiki at a window size.
func webModel(t *testing.T, w, h int) Model {
	t.Helper()
	m := New(webWiki(t))
	m.now = func() time.Time { return webNow }
	m = m.SetSize(w, h).Refresh()
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("w")})
	return m
}

func TestWebScreens(t *testing.T) {
	for _, size := range []struct {
		name string
		w, h int
	}{{"web-80x24", 80, 24}, {"web-120x45", 120, 45}} {
		out := webModel(t, size.w, size.h).View()
		uitest.Golden(t, size.name, out)
		lines := strings.Split(out, "\n")
		if len(lines) > size.h {
			t.Errorf("%s takes %d lines", size.name, len(lines))
		}
		for _, l := range lines {
			if lipgloss.Width(l) > size.w {
				t.Errorf("%s line too wide (%d): %q", size.name, lipgloss.Width(l), uitest.Plain(l))
			}
		}
		for _, want := range []string{"盘丝洞", "学 xué", "missed 6×", "╳", "学习", "enter revive"} {
			if !strings.Contains(uitest.Plain(out), want) {
				t.Errorf("%s lacks %q:\n%s", size.name, want, uitest.Plain(out))
			}
		}
		if strings.Contains(out, "学期") {
			t.Errorf("a word never reviewed is on the web:\n%s", uitest.Plain(out))
		}
	}
}

func TestWalkingTheWebRevivesTheCluster(t *testing.T) {
	m := webModel(t, 80, 24)
	key := func(k string) {
		t.Helper()
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(k)}
		if k == "enter" {
			msg = tea.KeyMsg{Type: tea.KeyEnter}
		}
		m, _ = m.Update(msg)
	}
	want := len(m.spider.g.cluster(m.spider.centre))
	if want < 2 {
		t.Fatalf("the hard character 学 should have slipping neighbours, got %d", want)
	}
	key("enter")
	if !m.Walking() {
		t.Fatal("enter should start walking the web")
	}
	uitest.Golden(t, "walk-80x24", m.View())
	for range want {
		key("enter") // show
		key("3")     // good
	}
	if out := uitest.Plain(m.View()); !strings.Contains(out, "Revived") {
		t.Fatalf("no summary after the walk:\n%s", out)
	}
	key("x")
	if m.Walking() || !strings.Contains(uitest.Plain(m.View()), "盘丝洞 The Spider's Web") {
		t.Fatalf("a key after the summary should return to the web:\n%s", uitest.Plain(m.View()))
	}
	if right, total, err := m.store.Retention(webNow, 1); err != nil || right != want || total != want {
		t.Errorf("the walk logged %d of %d reviews right (%v), want %d", right, total, err, want)
	}
	words, err := m.store.StudiedWords()
	if err != nil {
		t.Fatal(err)
	}
	for _, w := range words {
		if w.Card.Chinese == "学习" && w.Card.IsDue(webNow) {
			t.Errorf("学习 was not rescheduled: %+v", w.Card)
		}
	}
}

func TestFollowingAStrandAndBack(t *testing.T) {
	m := webModel(t, 80, 24)
	press := func(k tea.KeyType) { m, _ = m.Update(tea.KeyMsg{Type: k}) }
	for range 16 {
		press(tea.KeyDown)
	}
	sel := m.spider.g.words[m.strands()[m.spider.sel]].card.Chinese
	if !strings.Contains(uitest.Plain(m.View()), "▸"+sel) {
		t.Errorf("the chosen strand %s scrolled out of sight:\n%s", sel, uitest.Plain(m.View()))
	}
	press(tea.KeyRight)
	if m.centreWord() != sel || !strings.Contains(uitest.Plain(m.View()), "centre "+sel) {
		t.Errorf("→ should make %s the centre, got %s", sel, m.centreWord())
	}
	press(tea.KeyLeft)
	if m.centreWord() != "学" {
		t.Errorf("← should go back to 学, got %s", m.centreWord())
	}
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("w")})
	if !strings.Contains(uitest.Plain(m.View()), "Study Statistics") {
		t.Errorf("w should go back to the numbers")
	}
}
