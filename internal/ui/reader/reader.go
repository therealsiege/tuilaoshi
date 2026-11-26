// Package reader shows the wiki's readings and dialogue episodes word by
// word: move through the
// text to see each word's pinyin and meaning, and mark the ones you didn't
// know so they come back for review.
package reader

import (
	"errors"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/therealsiege/tuilaoshi/internal/audio"
	"github.com/therealsiege/tuilaoshi/internal/claude"
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

// clip is where a line is spoken in the episode's audio.
type clip struct{ start, end time.Duration }

// clipDoneMsg reports that a line's audio finished, or why it failed.
type clipDoneMsg struct{ err error }

// playClip plays part of an audio file; tests replace it.
var playClip = func(path string, c clip) error { return audio.PlayClip(path, c.start, c.end) }

// playLine plays the selected line from the episode's original audio.
func (m Model) playLine() tea.Cmd {
	if m.audio == "" || len(m.words) == 0 {
		return nil
	}
	c := m.clips[m.words[m.cursor].line]
	if c.end <= c.start {
		return nil
	}
	audio := m.audio
	return func() tea.Msg { return clipDoneMsg{playClip(audio, c)} }
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
	audio       string                  // the episode's original audio, if any
	clips       []clip                  // where each line is spoken in it
	status      string                  // result of the last mark

	// The teahouse storyteller (see storyteller.go).
	teller      Teller              // nil without an API key
	now         func() time.Time    // today, for one chapter a day
	known       []string            // a sample of the learner's known words
	chapters    []wiki.StoryChapter // chapters told so far, in order
	story       bool                // on the storyteller's shelf
	shelfCursor int                 // 0 tells the next chapter; then told ones, newest first
	telling     bool                // waiting for the storyteller
	chapterOpen bool                // the open text is a told chapter
	listening   *listening          // a chapter being heard before it is read (listen.go)

	loadErr, saveErr error
	width, height    int
}

// New creates a reading view.
func New(store *wiki.Store) Model {
	return Model{store: store, gkeys: ui.DefaultGlobalKeys(), now: time.Now}
}

// WithStoryteller lets the storyteller tell new chapters through Claude;
// without a ready client he says he needs a key.
func (m Model) WithStoryteller(client *claude.Client) Model {
	if client.Ready() {
		m.teller = client.TellChapter
	}
	return m
}

// Init satisfies the Bubble Tea model interface.
func (m Model) Init() tea.Cmd { return nil }

// SetSize updates terminal dimensions.
func (m Model) SetSize(w, h int) Model {
	m.width, m.height = w, h
	return m
}

// Open reports whether a text or the storyteller's shelf is open, so esc
// returns to the list.
func (m Model) Open() bool { return m.open || m.story }

// Refresh reloads the reading list and the decks used for glosses.
func (m Model) Refresh() Model {
	readings, readErr := m.store.ListPagesByType("reading")
	cards, cardsErr := m.store.DrillableCards()
	private, privErr := m.store.PrivateEpisodes()
	pages, pagesErr := m.store.ListPages()
	cleared, clearedErr := m.store.ClearedStages()
	told, toldErr := m.store.StoryChaptersTold()
	m.loadErr = errors.Join(readErr, cardsErr, privErr, pagesErr, clearedErr, toldErr)
	// A chapter told while the learner was away was saved, though its reply
	// never reached this screen, so stop waiting for it.
	m.chapters, m.known, m.story, m.telling = told, knownWords(cards, pages), false, false
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
	m.audio, m.clips, m.chapterOpen = "", nil, false
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
		if l.Original {
			english[i] = "原文, from the novel: " + l.English
		}
		if l.Speaker != "" {
			lines[i] = l.Speaker + "：" + l.Chinese
		}
	}
	m = m.OpenText(ep.Title, lines, journey.StoryGlossary())
	m.english, m.audio = english, ep.Audio
	m.clips = make([]clip, len(ep.Lines))
	for i, l := range ep.Lines {
		m.clips[i] = clip{l.Start, l.End}
	}
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
		if err := m.store.RecordForgotten(&card, time.Now()); err != nil {
			m.saveErr = err
			return m
		}
	}
	m.marked[w] = true
	m.status = w + " will come back for review tomorrow"
	return m
}

// sentenceCard keeps the selected word with the line it is in, as a card
// whose review shows the line with the word blanked.
func (m Model) sentenceCard() Model {
	w := m.current()
	gl := m.glosses(w)
	if w == "" || len(gl) == 0 {
		m.status = "Pick a word with a meaning to make a sentence card."
		return m
	}
	var line strings.Builder
	for _, t := range m.lines[m.words[m.cursor].line] {
		line.WriteString(t.Text)
	}
	added, err := m.store.AddSentenceCard(wiki.Word{Chinese: w, Pinyin: gl[0].Pinyin, Meaning: gl[0].English}, line.String(), time.Now())
	switch {
	case err != nil:
		m.saveErr = err
	case added:
		m.status = w + " joins your reviews with this sentence"
	default:
		m.status = w + " already has a sentence card"
	}
	return m
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if done, ok := msg.(clipDoneMsg); ok {
		if done.err != nil {
			m.status = done.err.Error()
		}
		return m, nil
	}
	if told, ok := msg.(toldMsg); ok {
		return m.onTold(told), nil
	}
	k, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	if m.listening != nil {
		return m.updateListening(k)
	}
	if m.story && !m.open {
		return m.updateShelf(k)
	}
	if !m.open {
		switch {
		case key.Matches(k, m.gkeys.Up) && m.listCursor > 0:
			m.listCursor--
		case key.Matches(k, m.gkeys.Down) && m.listCursor < len(m.items)-1:
			m.listCursor++
		case k.String() == "s":
			m = m.openShelf()
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
	case k.String() == "c":
		m = m.sentenceCard()
	case k.String() == "e":
		m.showEnglish = !m.showEnglish
	case k.String() == "p":
		return m, m.playLine()
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
