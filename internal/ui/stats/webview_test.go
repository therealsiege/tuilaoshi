package stats

import (
	"fmt"
	"os"
	"path/filepath"
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
	// The pages' review dates only mark which words the game has reviewed:
	// save those as the reviews they stand for.
	var saved strings.Builder
	for _, c := range cards {
		if !c.LastReviewed.IsZero() {
			fmt.Fprintf(&saved, "%q:\n    last_reviewed: %q\n    review_interval: %d\n    next_review: %q\n",
				c.ID, wiki.FormatDate(c.LastReviewed), c.ReviewInterval, wiki.FormatDate(c.NextReview))
		}
	}
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "schedules.yaml"), []byte(saved.String()), 0644); err != nil {
		t.Fatal(err)
	}
	if err := store.Import(dir); err != nil {
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
	want := len(m.spider.g.Cluster(m.spider.centre))
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
	sel := m.spider.g.Words[m.strands()[m.spider.sel]].Card.Chinese
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
