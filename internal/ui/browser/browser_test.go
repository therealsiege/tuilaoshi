package browser

import (
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/therealsiege/tuilaoshi/internal/ui/uitest"
)

func char(c, pinyin, meaning, level string) string {
	return fmt.Sprintf("---\ntype: character\nlevel: %s\n---\n\n# %s %s\n\n- **Character**: %s\n- **Pinyin**: %s\n- **Meaning**: %s\n", level, c, meaning, c, pinyin, meaning)
}

// testBrowser has a reading, a grammar page and many characters, so the list
// is longer than the screen.
func testBrowser(t *testing.T, w, h int) Model {
	t.Helper()
	pages := map[string]string{
		"readings/r.md":    "---\ntype: reading\nlevel: HSK2\n---\n\n# 在餐厅 At the Restaurant\n\n## Text\n\n你好。\n",
		"grammar/ba.md":    "---\ntype: grammar\nlevel: HSK3\n---\n\n# 把 Construction\n\nBody.\n",
		"characters/a.md":  char("书", "shū", "book", "HSK1"),
		"characters/b.md":  char("家", "jiā", "home, family", "HSK1"),
		"characters/c.md":  char("看", "kàn", "to look", "HSK1"),
		"characters/zz.md": char("龘", "dá", "the appearance of a dragon flying, a very long meaning indeed", "HSK6"),
	}
	for i := 0; i < 60; i++ {
		pages[fmt.Sprintf("characters/x%02d.md", i)] = char(string(rune(0x4E00+i)), fmt.Sprintf("p%d", i), fmt.Sprintf("meaning %d", i), "HSK4")
	}
	return New(uitest.Wiki(t, pages)).SetSize(w, h).Refresh()
}

func titles(m Model) []string {
	var out []string
	for _, i := range m.shown {
		out = append(out, m.entries[i].title)
	}
	return out
}

func press(m Model, k tea.KeyMsg) Model {
	m, _ = m.Update(k)
	return m
}

func typeKeys(m Model, s string) Model {
	for _, r := range s {
		m = press(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	return m
}

func TestViewFitsTheTerminalAndFollowsTheCursor(t *testing.T) {
	m := testBrowser(t, 100, 24)
	if h := lipgloss.Height(m.View()); h > 24 {
		t.Fatalf("view is %d rows, taller than the 24-row terminal", h)
	}
	for i := 0; i < 50; i++ {
		m = press(m, tea.KeyMsg{Type: tea.KeyDown})
	}
	if !strings.Contains(uitest.Plain(m.View()), "▸") {
		t.Error("the selection scrolled out of view")
	}
	if h := lipgloss.Height(m.View()); h > 24 {
		t.Errorf("after scrolling the view is %d rows", h)
	}
}

func TestPagesSortByTypeThenLevel(t *testing.T) {
	got := titles(testBrowser(t, 100, 24))
	if got[0] != "在餐厅 At the Restaurant" || got[1] != "把 Construction" || !strings.HasPrefix(got[2], "书") && !strings.HasPrefix(got[2], "家") && !strings.HasPrefix(got[2], "看") {
		t.Errorf("order starts %v; want the reading, the grammar page, then HSK1 characters", got[:3])
	}
	if last := got[len(got)-1]; !strings.HasPrefix(last, "龘") {
		t.Errorf("last %q; want the HSK6 character last", last)
	}
}

func TestSearchFindsPinyinWithoutTonesAndTypesAnyKey(t *testing.T) {
	m := testBrowser(t, 100, 24)
	m = typeKeys(press(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/")}), "shu")
	if got := titles(m); len(got) != 1 || !strings.HasPrefix(got[0], "书") {
		t.Errorf("'shu' found %v; want 书 (shū)", got)
	}
	m = press(m, tea.KeyMsg{Type: tea.KeyEsc})
	m = typeKeys(press(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/")}), "jia")
	if got := titles(m); len(got) != 1 || !strings.HasPrefix(got[0], "家") || m.cursor != 0 {
		t.Errorf("'jia' found %v (j must type, not move); want 家", got)
	}
	m = typeKeys(press(m, tea.KeyMsg{Type: tea.KeyBackspace}), "")
	m = press(m, tea.KeyMsg{Type: tea.KeyEsc})
	m = typeKeys(press(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/")}), "hsk1 look")
	if got := titles(m); len(got) != 1 || !strings.HasPrefix(got[0], "看") {
		t.Errorf("'hsk1 look' found %v; every word must match", got)
	}
	if m = press(m, tea.KeyMsg{Type: tea.KeyEnter}); m.Searching() || m.focus != focusContent {
		t.Error("enter should close the search and open the page")
	}
}

func TestSearchRanksWholeWordsFirst(t *testing.T) {
	entries := buildIndex(nil)
	for _, e := range []struct{ title, pinyin string }{{"小说 novel", "xiǎoshuō"}, {"水 water", "shuǐ"}, {"书 book", "shū"}} {
		entries = append(entries, entry{title: e.title, key: fold(e.title + " " + e.pinyin)})
	}
	var got []string
	for _, i := range match(entries, "shu") {
		got = append(got, entries[i].title)
	}
	if strings.Join(got, ",") != "书 book,水 water,小说 novel" {
		t.Errorf("ranked %v; want 书 (shu), then 水 (starts shui), then 小说 (shu inside xiaoshuo)", got)
	}
}

func TestEscClearsTheSearch(t *testing.T) {
	m := testBrowser(t, 100, 24).StartSearch()
	m = typeKeys(m, "zzz")
	if len(m.shown) != 0 || !strings.Contains(uitest.Plain(m.View()), "No pages match.") {
		t.Fatalf("shown %d for a search with no matches", len(m.shown))
	}
	if m = press(m, tea.KeyMsg{Type: tea.KeyEsc}); m.Searching() || m.query != "" || len(m.shown) != len(m.entries) {
		t.Errorf("esc: searching %v, query %q, %d shown", m.Searching(), m.query, len(m.shown))
	}
}

func TestLongTitlesTruncateCleanly(t *testing.T) {
	m := testBrowser(t, 80, 24)
	m = press(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("G")})
	for _, line := range strings.Split(uitest.Plain(m.listView()), "\n") {
		if lipgloss.Width(line) > m.listWidth() || strings.ContainsRune(line, '�') {
			t.Errorf("list line %q is too wide or broken", line)
		}
	}
}

func TestBrowserScreen(t *testing.T) {
	uitest.Golden(t, "browser", testBrowser(t, 100, 24).View())
}
