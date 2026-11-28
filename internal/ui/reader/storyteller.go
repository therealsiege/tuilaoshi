package reader

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/therealsiege/tuilaoshi/internal/camp"
	"github.com/therealsiege/tuilaoshi/internal/claude"
	"github.com/therealsiege/tuilaoshi/internal/journey"
	"github.com/therealsiege/tuilaoshi/internal/ui"
	"github.com/therealsiege/tuilaoshi/internal/wiki"
)

// The teahouse storyteller (说书人) retells the novel's hundred chapters,
// one new chapter a day, in the words the learner knows. Each is written
// by Claude when first told and kept in the progress file, so it can be
// read again word by word like any other text.

// Teller writes chapter n of the novel using mostly the known words.
type Teller = camp.Teller

// toldMsg carries a chapter the storyteller has told and saved, or why not.
type toldMsg struct {
	chapter wiki.StoryChapter
	err     error
}

// nextChapter is the chapter the storyteller would tell now, and whether
// he can today.
func (m Model) nextChapter() (int, bool) {
	return wiki.NextStoryChapter(m.chapters, m.now())
}

// openShelf shows the storyteller's chapters.
func (m Model) openShelf() Model {
	m.story, m.shelfCursor, m.status = true, 0, ""
	return m
}

// shelfChapter is the told chapter at a shelf row (row 0 tells the next
// one), newest first.
func (m Model) shelfChapter(row int) wiki.StoryChapter {
	return m.chapters[len(m.chapters)-row]
}

// updateShelf handles keys on the storyteller's shelf.
func (m Model) updateShelf(k tea.KeyMsg) (Model, tea.Cmd) {
	switch {
	case key.Matches(k, m.gkeys.Back):
		m.story, m.status = false, ""
	case key.Matches(k, m.gkeys.Up) && m.shelfCursor > 0:
		m.shelfCursor--
	case key.Matches(k, m.gkeys.Down) && m.shelfCursor < len(m.chapters):
		m.shelfCursor++
	case key.Matches(k, m.gkeys.Enter) && m.shelfCursor > 0:
		m = m.openChapter(m.shelfChapter(m.shelfCursor))
	case k.String() == "l" && m.shelfCursor > 0:
		m = m.startListening(m.shelfChapter(m.shelfCursor))
	case key.Matches(k, m.gkeys.Enter):
		return m.tell()
	}
	return m, nil
}

// tell asks the storyteller for the next chapter, once a day; the chapter
// is saved as soon as it is told, even if the learner has left the screen.
func (m Model) tell() (Model, tea.Cmd) {
	n, ok := m.nextChapter()
	switch {
	case m.telling:
		return m, nil
	case n > wiki.StoryChapters:
		m.status = "The tale is told: all 100 chapters. Read them again any time."
		return m, nil
	case !ok:
		m.status = fmt.Sprintf("第%d回 is told tomorrow. Come back to the teahouse then.", n)
		return m, nil
	case m.teller == nil:
		m.status = "The storyteller " + claude.NeedsKey
		return m, nil
	}
	m.telling, m.status = true, ""
	teller, store, known, now := m.teller, m.store, m.known, m.now()
	return m, func() tea.Msg {
		told, err := camp.TellStory(context.Background(), store, teller, n, known, now)
		return toldMsg{chapter: told, err: err}
	}
}

// onTold opens a freshly told chapter.
func (m Model) onTold(msg toldMsg) Model {
	m.telling = false
	if msg.err != nil {
		m.status = "The storyteller lost his voice: " + msg.err.Error()
		return m
	}
	told, err := m.store.StoryChaptersTold()
	if err != nil {
		m.loadErr = err
	}
	m.chapters = told
	m.story = true
	return m.openChapter(msg.chapter)
}

// openChapter opens a told chapter for word-by-word reading, its English
// summary behind e.
func (m Model) openChapter(c wiki.StoryChapter) Model {
	lines := camp.StorySentences(c.Text)
	m = m.OpenText(fmt.Sprintf("说书人 第%d回 %s", c.Number, c.Title), lines, journey.StoryGlossary())
	m.english = make([]string, len(lines))
	for i := range m.english {
		m.english[i] = "This chapter: " + c.English
	}
	if c.English == "" {
		m.english = nil
	}
	m.chapterOpen = true
	return m
}

// Hear is the open chapter's text, for ctrl+r to read aloud.
func (m Model) Hear() string {
	if m.listening != nil {
		return m.listening.chapter.Text
	}
	if !m.open || !m.chapterOpen {
		return ""
	}
	var b strings.Builder
	for _, l := range m.lines {
		for _, t := range l {
			b.WriteString(t.Text)
		}
	}
	return b.String()
}

// storyRow is the storyteller's line at the top of the reading list.
func (m Model) storyRow() string {
	n, ok := m.nextChapter()
	switch {
	case n > wiki.StoryChapters:
		return "s  说书人 The Storyteller · all 100 chapters told"
	case ok:
		return fmt.Sprintf("s  说书人 The Storyteller · 第%d回 today", n)
	}
	return fmt.Sprintf("s  说书人 The Storyteller · 第%d回 tomorrow", n)
}

// viewShelf lists the chapters: the next one to tell, then those told.
func (m Model) viewShelf() string {
	var b strings.Builder
	b.WriteString(ui.SubtitleStyle.Render("  说书人 The Teahouse Storyteller"))
	b.WriteString("\n")
	b.WriteString(ui.MutedStyle.Render("  Journey to the West retold in words you know, one chapter a day."))
	b.WriteString("\n\n")
	n, ok := m.nextChapter()
	rows := []string{fmt.Sprintf("第%d回  %s", n, ui.MutedStyle.Render("tell today's chapter"))}
	switch {
	case n > wiki.StoryChapters:
		rows[0] = ui.MutedStyle.Render("The tale is told.")
	case !ok:
		rows[0] = fmt.Sprintf("第%d回  %s", n, ui.MutedStyle.Render("told tomorrow"))
	}
	for row := 1; row <= len(m.chapters); row++ {
		c := m.shelfChapter(row)
		rows = append(rows, fmt.Sprintf("第%d回  %s  %s", c.Number, c.Title, ui.MutedStyle.Render(c.Date)))
	}
	for i, r := range rows {
		if i == m.shelfCursor {
			b.WriteString(ui.SelectedStyle.Render("▸ " + r))
		} else {
			b.WriteString(ui.UnselectedStyle.Render("  " + r))
		}
		b.WriteString("\n")
	}
	b.WriteString("\n")
	if m.telling {
		b.WriteString(ui.MutedStyle.Render("  The storyteller clears his throat… 说书人清了清嗓子"))
		b.WriteString("\n\n")
	} else if m.status != "" {
		b.WriteString(ui.MutedStyle.Render("  " + m.status))
		b.WriteString("\n\n")
	}
	b.WriteString(ui.HelpStyle.Render("  ↑↓ choose · enter tell / read · l hear it first · esc readings"))
	return b.String()
}
