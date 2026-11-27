package dashboard

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/therealsiege/tuilaoshi/internal/camp"
	"github.com/therealsiege/tuilaoshi/internal/ui/uitest"
	"github.com/therealsiege/tuilaoshi/internal/wiki"
)

// gazetteMonday is a Monday; its issue reports 2026-11-23 to 29.
var gazetteMonday = time.Date(2026, 11, 30, 9, 0, 0, 0, time.Local)

// busyWeek is camp after a week on the road: a kingdom's stamp, a
// tribulation cleared that week (and one before it), a rumour settled, an
// errand, a haggle and a diary entry.
func busyWeek(t *testing.T) Model {
	t.Helper()
	store := uitest.Wiki(t, map[string]string{"characters/shu1.md": uitest.OverdueCharacter})
	inWeek, before := gazetteMonday.AddDate(0, 0, -3), gazetteMonday.AddDate(0, 0, -10)
	for _, err := range []error{
		store.Stamp("baoxiang", wiki.Wallet{}),
		store.RecordStageCleared(1, before),
		store.RecordStageCleared(2, inWeek),
		store.SettleMystery("gao-village", inWeek, wiki.Wallet{}),
		store.FinishErrand("changan", "在客栈订一间五个人住的房间，再要一些热水。", inWeek, wiki.Wallet{Tongqian: 30}),
		store.FinishErrand("changan", "去药铺买退烧的药。", before, wiki.Wallet{Tongqian: 30}),
		store.AddDeed(wiki.Deed{Date: inWeek.Format("2006-01-02"), Kind: wiki.DeedHaggle, Town: "changan", What: "20份斋饭", Coins: 25}),
		store.WriteDiary(wiki.DiaryEntry{Date: "2026-11-27", Day: 3, Text: "今天我们走了很远的路，八戒说他饿了。"}),
	} {
		if err != nil {
			t.Fatal(err)
		}
	}
	m := New(store, nil)
	m.now = func() time.Time { return gazetteMonday }
	return m.SetSize(120, 45).Refresh()
}

func TestGatherFactsReadsTheWeek(t *testing.T) {
	m := busyWeek(t)
	f, err := camp.GatherFacts(m.store, gazetteMonday)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Join(f.Kingdoms, "") != "宝象国" || len(f.Cleared) != 1 || !strings.HasPrefix(f.Cleared[0], "第2难") ||
		strings.Join(f.Mysteries, "") != "高老庄" || len(f.Errands) != 1 || !strings.Contains(f.Errands[0], "热水") ||
		len(f.Haggles) != 1 || !strings.Contains(f.Haggles[0], "25个铜钱") || !strings.Contains(f.Diary, "八戒") {
		t.Errorf("facts %+v", f)
	}
	if p := f.Prompt("2026-11-30"); !strings.Contains(p, "宝象国") || !strings.Contains(p, "第2难") || strings.Contains(p, "退烧") {
		t.Errorf("prompt:\n%s", p)
	}
}

func TestComposeGazette(t *testing.T) {
	zh, en := camp.ComposeGazette(camp.Facts{Kingdoms: []string{"宝象国"}, Cleared: []string{"第2难出城逢虎"}, Errands: []string{"订房"}, Diary: "累了。"})
	for _, want := range []string{"本报讯", "1国在通关文牒上盖了印：宝象国", "渡过了1难：第2难出城逢虎", "办了1件事：订房", "「累了。」", "本报将继续报道"} {
		if !strings.Contains(zh, want) {
			t.Errorf("issue lacks %q:\n%s", want, zh)
		}
	}
	if strings.Count(zh, "\n") != strings.Count(en, "\n") {
		t.Errorf("Chinese and English have different lines:\n%s\n%s", zh, en)
	}
	if zh, _ := camp.ComposeGazette(camp.Facts{}); !strings.Contains(zh, "还没有盖过印") || !strings.Contains(zh, "平静") || strings.Contains(zh, "日记") {
		t.Errorf("a quiet week:\n%s", zh)
	}
}

func press(s string) tea.KeyMsg { return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)} }

func TestTheClerkSetsAWeeklyGazetteWithoutAKey(t *testing.T) {
	m := busyWeek(t)
	m, cmd := m.Update(press("u"))
	if cmd != nil || !m.gazetteOpen || !m.gazette.ByHand || !strings.Contains(m.View(), "邸报 The court gazette") || !strings.Contains(m.Hear(), "宝象国") {
		t.Fatalf("no gazette without a key:\n%s", m.View())
	}
	uitest.Golden(t, "gazette-120x45", m.View())
	uitest.Golden(t, "gazette-80x24", m.SetSize(80, 24).View())
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if m.gazetteOpen || m.Hear() != "" {
		t.Error("esc should close the gazette")
	}
	m, _ = m.Update(press("u"))
	m.now = func() time.Time { return gazetteMonday.AddDate(0, 0, 3) } // later the same week
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m, _ = m.Update(press("u"))
	issues, err := m.store.Gazettes()
	if err != nil || len(issues) != 1 || issues[0].Week != "2026-11-30" {
		t.Fatalf("one issue a week: %+v %v", issues, err)
	}
	m.now = func() time.Time { return gazetteMonday.AddDate(0, 0, 7) }
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m, _ = m.Update(press("u"))
	if issues, err = m.store.Gazettes(); err != nil || len(issues) != 2 || !strings.Contains(issues[1].Text, "平静") {
		t.Errorf("next week's issue: %+v %v", issues, err)
	}
}

func TestClaudeWritesTheGazette(t *testing.T) {
	m := busyWeek(t)
	m.writer = func(_ context.Context, facts string) (string, string, error) {
		if !strings.Contains(facts, "宝象国") || !strings.Contains(facts, "25个铜钱") {
			t.Errorf("facts sent: %s", facts)
		}
		return "本报讯：圣僧一行抵达宝象国。", "Our report: the holy monk reached Baoxiang.", nil
	}
	m, cmd := m.Update(press("u"))
	if cmd == nil || m.gazetteOpen || !m.printing {
		t.Fatal("Claude wasn't asked")
	}
	if again, cmd := m.Update(press("u")); cmd != nil || again.gazetteOpen {
		t.Error("asked twice while printing")
	}
	m, _ = m.Update(cmd())
	issues, err := m.store.Gazettes()
	if !m.gazetteOpen || m.gazette.ByHand || err != nil || len(issues) != 1 || issues[0].Text != "本报讯：圣僧一行抵达宝象国。" {
		t.Fatalf("Claude's issue: %+v %v\n%s", issues, err, m.View())
	}

	m = busyWeek(t)
	m.writer = func(context.Context, string) (string, string, error) { return "", "", errors.New("offline") }
	m, cmd = m.Update(press("u"))
	m, _ = m.Update(cmd())
	if !m.gazetteOpen || !m.gazette.ByHand || !strings.Contains(m.View(), "offline") {
		t.Errorf("a failed call should fall back to the clerk:\n%s", m.View())
	}
}

func TestTheGazetteFitsEightyByTwentyFour(t *testing.T) {
	m := busyWeek(t).SetSize(80, 24)
	long := strings.Repeat("本报讯：圣僧一行抵达宝象国，国王亲自在通关文牒上盖了印，百姓们都出来看热闹。\n", 8)
	m.gazette = &wiki.Gazette{Week: "2026-11-30", Text: long, English: strings.Repeat("Our report: the holy monk reached Baoxiang, where the king stamped the papers himself.\n", 8)}
	m.gazetteOpen, m.note = true, "Couldn't reach Claude (a long error from the network that runs well past the edge of the window): the clerk set it."
	for _, keys := range []string{"", "e", "jjjjjjjjjjjjjjjjjjjjjjjjjjjjjjjjjjjjj", "e"} {
		for _, k := range keys {
			m, _ = m.Update(press(string(k)))
		}
		out := m.View()
		for _, l := range strings.Split(out, "\n") {
			if lipgloss.Width(l) > 80 {
				t.Errorf("after %q a line %d wide: %s", keys, lipgloss.Width(l), l)
			}
		}
		if n := strings.Count(out, "\n") + 1; n > 24 {
			t.Errorf("after %q the gazette takes %d lines:\n%s", keys, n, out)
		}
	}
}
