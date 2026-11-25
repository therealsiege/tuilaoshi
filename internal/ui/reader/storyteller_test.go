package reader

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/therealsiege/tuilaoshi/internal/claude"
	"github.com/therealsiege/tuilaoshi/internal/model"
	"github.com/therealsiege/tuilaoshi/internal/ui/uitest"
)

const chapterOne = "很久以前，有一座花果山。山上有一块石头。\n石头里出来一只猴子，他看了看菜单。\n欲知后事如何，且听下回分解。"

func typeKey(m Model, s string) (Model, tea.Cmd) {
	return m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)})
}

// storyteller is a reader on 2026-11-25 whose storyteller is a fake that
// counts the chapters asked for.
func storyteller(t *testing.T, asked *[]int) Model {
	t.Helper()
	m := New(uitest.Wiki(t, map[string]string{"vocabulary/food.md": food})).SetSize(80, 30)
	m.now = func() time.Time { return time.Date(2026, 11, 25, 9, 0, 0, 0, time.Local) }
	m.teller = func(_ context.Context, n int, known []string) (claude.Chapter, error) {
		*asked = append(*asked, n)
		if !reflect.DeepEqual(known, []string{"菜单"}) {
			return claude.Chapter{}, errors.New("known words not passed")
		}
		return claude.Chapter{Title: "石猴出世", Text: chapterOne, English: "A stone monkey is born."}, nil
	}
	return m.Refresh()
}

func TestStorytellerTellsAChapterADay(t *testing.T) {
	var asked []int
	m := storyteller(t, &asked)
	if !strings.Contains(m.View(), "第1回 today") {
		t.Errorf("list doesn't offer today's chapter:\n%s", m.View())
	}
	m, _ = typeKey(m, "s")
	if !m.Open() {
		t.Fatal("s did not open the storyteller")
	}
	uitest.Golden(t, "storyteller", m.View())

	m, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil || !strings.Contains(m.View(), "clears his throat") {
		t.Fatalf("enter didn't start the story:\n%s", m.View())
	}
	m, _ = m.Update(cmd())
	if !m.open || !m.chapterOpen || m.status != "" {
		t.Fatalf("the told chapter isn't open: status %q", m.status)
	}
	uitest.Golden(t, "chapter", m.View())
	if m.Hear() != strings.ReplaceAll(chapterOne, "\n", "") {
		t.Errorf("Hear() = %q", m.Hear())
	}
	for m.current() != "菜单" && m.cursor < len(m.words)-1 {
		m = press(m, tea.KeyRight)
	}
	if m = press(m, tea.KeyEnter); !m.marked["菜单"] {
		t.Errorf("enter didn't send 菜单 to the reviews: %q", m.status)
	}

	// The chapter is kept, and no second chapter is told today.
	m = press(m, tea.KeyEsc)
	if m.open || !m.story {
		t.Fatal("esc didn't return to the storyteller's chapters")
	}
	m = m.Refresh().openShelf()
	if len(m.chapters) != 1 || m.chapters[0].Date != "2026-11-25" {
		t.Fatalf("chapters kept: %+v", m.chapters)
	}
	m, cmd = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd != nil || !strings.Contains(m.status, "第2回 is told tomorrow") {
		t.Errorf("a second chapter today: status %q", m.status)
	}
	m = press(m, tea.KeyDown)
	if m = press(m, tea.KeyEnter); !m.chapterOpen || !strings.Contains(m.View(), "石猴出世") {
		t.Errorf("chapter 1 doesn't open again:\n%s", m.View())
	}

	m = m.Refresh().openShelf()
	m.now = func() time.Time { return time.Date(2026, 11, 26, 9, 0, 0, 0, time.Local) }
	_, cmd = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("the next day's chapter wasn't told")
	}
	cmd()
	if !reflect.DeepEqual(asked, []int{1, 2}) {
		t.Errorf("chapters asked for %v; want 1 then 2", asked)
	}
}

func TestStorytellerNeedsAKey(t *testing.T) {
	m := New(uitest.Wiki(t, map[string]string{"vocabulary/food.md": food})).SetSize(80, 30).Refresh()
	m, _ = typeKey(m, "s")
	m, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd != nil || !strings.Contains(m.View(), "ANTHROPIC_API_KEY") {
		t.Errorf("without a key:\n%s", m.View())
	}
	if m.Hear() != "" {
		t.Error("Hear() with no chapter open should be empty")
	}
}

func TestStorytellerErrorIsShown(t *testing.T) {
	var asked []int
	m := storyteller(t, &asked)
	m.teller = func(context.Context, int, []string) (claude.Chapter, error) {
		return claude.Chapter{}, errors.New("overloaded")
	}
	m, _ = typeKey(m, "s")
	m, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m, _ = m.Update(cmd())
	if m.telling || !strings.Contains(m.status, "overloaded") || len(m.chapters) != 0 {
		t.Errorf("after an error: telling %v, status %q, chapters %d", m.telling, m.status, len(m.chapters))
	}
}

func TestSentences(t *testing.T) {
	got := sentences("他说：“走吧！”我们走了。真的？\n好。")
	want := []string{"他说：“走吧！”", "我们走了。", "真的？", "好。"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("sentences %q, want %q", got, want)
	}
}

func TestKnownWordsPreferLearnedAndEasyWords(t *testing.T) {
	pages := []model.Page{{Path: "vocabulary/easy.md", Level: "HSK2"}, {Path: "vocabulary/hard.md", Level: "HSK6"}}
	cards := []model.Card{
		{ID: "vocabulary/easy.md#吃饭", Chinese: "吃饭"},
		{ID: "vocabulary/hard.md#踌躇", Chinese: "踌躇"},
		{ID: "vocabulary/hard.md#斟酌", Chinese: "斟酌", ReviewInterval: 365},
		{ID: "grammar/ba.md", Type: model.CardGrammar, Chinese: "把", ReviewInterval: 365},
	}
	got := knownWords(cards, pages)
	if !reflect.DeepEqual(got, []string{"斟酌", "吃饭"}) {
		t.Errorf("known words %q; want the learned word, then the HSK2 one", got)
	}
}
