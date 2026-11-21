// Package reader shows the wiki's readings and dialogue episodes word by
// word: move through the
// text to see each word's pinyin and meaning, and mark the ones you didn't
// know so they come back for review.
package reader

import (
	"errors"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/therealsiege/tuilaoshi/internal/journey"
	"github.com/therealsiege/tuilaoshi/internal/model"
	"github.com/therealsiege/tuilaoshi/internal/segment"
	"github.com/therealsiege/tuilaoshi/internal/ui"
	"github.com/therealsiege/tuilaoshi/internal/wiki"
)

// item is one entry in the reading list: a wiki reading or an episode.
type item struct {
	title, tag string
	page       *model.Page
	episode    *model.Episode
}

// place is the position of a word token: its line and index in that line.
type place struct{ line, tok int }

// Model is the reading view model.
type Model struct {
	store *wiki.Store
	gkeys ui.GlobalKeys

	items      []item
	listCursor int

	open   bool // reading a text rather than choosing one
	title  string
	lines  [][]segment.Token
	words  []place // every selectable word, in reading order
	cursor int     // index into words

	decks       map[string][]model.Card // drillable cards by Chinese, for glosses and scheduling
	vocab       map[string][]model.Card // the open reading's own vocabulary table
	marked      map[string]bool         // words marked as not known this visit
	english     []string                // each line's translation, when known
	showEnglish bool                    // show the current line's translation
	status      string                  // result of the last mark

	loadErr, saveErr error
	width, height    int
}

// New creates a reading view.
func New(store *wiki.Store) Model {
	return Model{store: store, gkeys: ui.DefaultGlobalKeys()}
}

// Init satisfies the Bubble Tea model interface.
func (m Model) Init() tea.Cmd { return nil }

// SetSize updates terminal dimensions.
func (m Model) SetSize(w, h int) Model {
	m.width, m.height = w, h
	return m
}

// Open reports whether a text is open, so esc returns to the list.
func (m Model) Open() bool { return m.open }

// Refresh reloads the reading list and the decks used for glosses.
func (m Model) Refresh() Model {
	readings, readErr := m.store.ListPagesByType("reading")
	cards, cardsErr := m.store.DrillableCards()
	private, privErr := m.store.PrivateEpisodes()
	pages, pagesErr := m.store.ListPages()
	cleared, clearedErr := m.store.ClearedStages()
	m.loadErr = errors.Join(readErr, cardsErr, privErr, pagesErr, clearedErr)
	m.items = nil
	for i := range readings {
		m.items = append(m.items, item{title: readings[i].Title, tag: readings[i].Level, page: &readings[i]})
	}
	// Journey episodes unlock as the pilgrimage reaches their stage.
	reached := journey.ReachedEpisodes(journey.Pilgrimage(pages, cleared).Tribulation)
	for i := range reached {
		m.items = append(m.items, item{title: reached[i].Title, tag: "journey episode", episode: &reached[i]})
	}
	for i := range private {
		m.items = append(m.items, item{title: private[i].Title, tag: "your episode", episode: &private[i]})
	}
	m.decks = map[string][]model.Card{}
	for _, c := range cards {
		m.decks[c.Chinese] = append(m.decks[c.Chinese], c)
	}
	m.open, m.saveErr, m.status = false, nil, ""
	if m.listCursor >= len(m.items) {
		m.listCursor = 0
	}
	return m
}

// openReading opens a wiki reading's text.
func (m Model) openReading(p model.Page) Model {
	return m.OpenText(p.Title, textSection(p.Body), wiki.ReadingVocab(p))
}

// OpenText opens lines of Chinese for word-by-word reading, glossed by vocab
// first and then the learner's decks (call Refresh first to load them).
func (m Model) OpenText(title string, lines []string, vocab []model.Card) Model {
	m.vocab = map[string][]model.Card{}
	for _, c := range vocab {
		m.vocab[c.Chinese] = append(m.vocab[c.Chinese], c)
	}
	known := map[string]bool{}
	for w := range m.decks {
		known[w] = true
	}
	for w := range m.vocab {
		known[w] = true
	}
	m.title, m.lines, m.words, m.english, m.showEnglish = title, nil, nil, nil, false
	for _, l := range lines {
		toks := segment.Split(l, known)
		for i, t := range toks {
			if t.Word {
				m.words = append(m.words, place{len(m.lines), i})
			}
		}
		m.lines = append(m.lines, toks)
	}
	m.open, m.cursor, m.marked, m.status = true, 0, map[string]bool{}, ""
	return m
}

// OpenEpisode opens a dialogue: each line reads "Speaker：line", glossed by
// the story glossary (names) and the decks, with its translation behind e.
func (m Model) OpenEpisode(ep model.Episode) Model {
	lines := make([]string, len(ep.Lines))
	english := make([]string, len(ep.Lines))
	for i, l := range ep.Lines {
		lines[i], english[i] = l.Chinese, l.English
		if l.Speaker != "" {
			lines[i] = l.Speaker + "：" + l.Chinese
		}
	}
	m = m.OpenText(ep.Title, lines, journey.StoryGlossary())
	m.english = english
	return m
}

// currentEnglish is the translation of the selected word's line, if any.
func (m Model) currentEnglish() string {
	if len(m.words) == 0 || len(m.english) == 0 {
		return ""
	}
	return m.english[m.words[m.cursor].line]
}

// current returns the selected word, or "" when the text has none.
func (m Model) current() string {
	if len(m.words) == 0 {
		return ""
	}
	p := m.words[m.cursor]
	return m.lines[p.line][p.tok].Text
}

// glosses returns the meanings of a word: the reading's own first, then
// the learner's decks.
func (m Model) glosses(word string) []model.Card {
	return append(append([]model.Card{}, m.vocab[word]...), m.decks[word]...)
}

// mark records that the learner didn't know the selected word: each deck
// card for it is reviewed as "Again", so it comes back tomorrow.
func (m Model) mark() Model {
	w := m.current()
	if w == "" || m.marked[w] {
		return m
	}
	cards := m.decks[w]
	if len(cards) == 0 {
		m.status = w + " is not in your decks yet"
		return m
	}
	for _, c := range cards {
		card := c
		if err := m.store.RecordReview(&card, model.DiffAgain, time.Now()); err != nil {
			m.saveErr = err
			return m
		}
	}
	m.marked[w] = true
	m.status = w + " will come back for review tomorrow"
	return m
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	k, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	if !m.open {
		switch {
		case key.Matches(k, m.gkeys.Up) && m.listCursor > 0:
			m.listCursor--
		case key.Matches(k, m.gkeys.Down) && m.listCursor < len(m.items)-1:
			m.listCursor++
		case key.Matches(k, m.gkeys.Enter) && len(m.items) > 0:
			if it := m.items[m.listCursor]; it.page != nil {
				m = m.openReading(*it.page)
			} else {
				m = m.OpenEpisode(*it.episode)
			}
		}
		return m, nil
	}
	switch {
	case key.Matches(k, m.gkeys.Back):
		m.open = false
	case key.Matches(k, m.gkeys.Right) && m.cursor < len(m.words)-1:
		m.cursor++
		m.status = ""
	case key.Matches(k, m.gkeys.Left) && m.cursor > 0:
		m.cursor--
		m.status = ""
	case key.Matches(k, m.gkeys.Down):
		m.cursor = m.lineStart(+1)
	case key.Matches(k, m.gkeys.Up):
		m.cursor = m.lineStart(-1)
	case key.Matches(k, m.gkeys.Enter):
		m = m.mark()
	case k.String() == "e":
		m.showEnglish = !m.showEnglish
	}
	return m, nil
}

// lineStart returns the first word of the next (dir +1) or previous (-1)
// line that has words, or the current word when there is none.
func (m Model) lineStart(dir int) int {
	if len(m.words) == 0 {
		return 0
	}
	line := m.words[m.cursor].line
	for i := range m.words {
		j := i
		if dir < 0 {
			j = len(m.words) - 1 - i
		}
		if (dir > 0 && m.words[j].line > line) || (dir < 0 && m.words[j].line < line) {
			for j > 0 && m.words[j-1].line == m.words[j].line {
				j--
			}
			return j
		}
	}
	return m.cursor
}
