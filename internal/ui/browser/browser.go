// Package browser lists the wiki's pages, sorted by type and level, with
// search over titles, pinyin (tones optional), meanings and tags, and shows
// the selected page as rendered Markdown.
package browser

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"

	"github.com/therealsiege/tuilaoshi/internal/ui"
	"github.com/therealsiege/tuilaoshi/internal/wiki"
)

type focus int

const (
	focusList focus = iota
	focusContent
)

// renderCacheSize bounds how many rendered pages are kept.
const renderCacheSize = 200

// Model is the browser view model.
type Model struct {
	store   *wiki.Store
	gkeys   ui.GlobalKeys
	entries []entry
	shown   []int // indices into entries that match the search
	cursor  int   // index into shown
	offset  int   // first shown row visible in the list
	focus   focus

	searching bool   // typing a search
	query     string // current search

	viewport  viewport.Model
	renderer  *glamour.TermRenderer
	rendererW int
	rendered  map[string]string // page path -> rendered Markdown, at rendererW

	width, height int
	loadErr       error
}

// New creates a new browser model.
func New(store *wiki.Store) Model {
	return Model{store: store, gkeys: ui.DefaultGlobalKeys(), viewport: viewport.New(60, 20), rendered: map[string]string{}}
}

func (m Model) Init() tea.Cmd { return nil }

// SetSize updates the terminal dimensions.
func (m Model) SetSize(w, h int) Model {
	m.width, m.height = w, h
	m.layoutViewport()
	m.showCurrent()
	return m
}

// Refresh reloads and indexes the pages, clearing any search.
func (m Model) Refresh() Model {
	pages, err := m.store.ListPages()
	m.loadErr = err
	m.entries = buildIndex(pages)
	m.query, m.searching, m.focus = "", false, focusList
	m.rendered = map[string]string{}
	m.applySearch()
	return m
}

// StartSearch opens the browser with the search box active.
func (m Model) StartSearch() Model {
	m.searching, m.focus = true, focusList
	return m
}

// Searching reports whether the search box has the keyboard, so esc closes
// it rather than leaving the browser.
func (m Model) Searching() bool { return m.searching }

func (m *Model) applySearch() {
	if m.query == "" {
		m.shown = make([]int, len(m.entries))
		for i := range m.shown {
			m.shown[i] = i
		}
	} else {
		m.shown = match(m.entries, m.query)
	}
	m.cursor, m.offset = 0, 0
	m.showCurrent()
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	k, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	if m.searching {
		return m.updateSearch(k)
	}
	if m.focus == focusContent {
		return m.updateContent(k)
	}
	switch {
	case k.String() == "/":
		m.searching = true
	case key.Matches(k, m.gkeys.Up):
		m.move(-1)
	case key.Matches(k, m.gkeys.Down):
		m.move(1)
	case k.Type == tea.KeyPgUp:
		m.move(-m.listRows())
	case k.Type == tea.KeyPgDown:
		m.move(m.listRows())
	case k.Type == tea.KeyHome || k.String() == "g":
		m.move(-len(m.shown))
	case k.Type == tea.KeyEnd || k.String() == "G":
		m.move(len(m.shown))
	case key.Matches(k, m.gkeys.Enter) || key.Matches(k, m.gkeys.Tab) || key.Matches(k, m.gkeys.Right):
		if len(m.shown) > 0 {
			m.focus = focusContent
		}
	}
	return m, nil
}

// updateSearch edits the query: every printable key types, arrows still
// move through the results, enter reads the selected page, esc clears.
func (m Model) updateSearch(k tea.KeyMsg) (Model, tea.Cmd) {
	switch k.Type {
	case tea.KeyRunes, tea.KeySpace:
		m.query += string(k.Runes)
		if k.Type == tea.KeySpace {
			m.query += " "
		}
		m.applySearch()
	case tea.KeyBackspace:
		if r := []rune(m.query); len(r) > 0 {
			m.query = string(r[:len(r)-1])
			m.applySearch()
		}
	case tea.KeyUp:
		m.move(-1)
	case tea.KeyDown:
		m.move(1)
	case tea.KeyEnter:
		m.searching = false
		if len(m.shown) > 0 {
			m.focus = focusContent
		}
	case tea.KeyEsc:
		m.searching, m.query = false, ""
		m.applySearch()
	}
	return m, nil
}

// updateContent scrolls the page; tab, left or backspace go back to the list.
func (m Model) updateContent(k tea.KeyMsg) (Model, tea.Cmd) {
	switch {
	case key.Matches(k, m.gkeys.Tab) || key.Matches(k, m.gkeys.Left) || k.Type == tea.KeyBackspace:
		m.focus = focusList
		return m, nil
	case k.String() == "/":
		m.focus, m.searching = focusList, true
		return m, nil
	}
	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(k)
	return m, cmd
}

// move shifts the selection, keeping it inside the visible list window.
func (m *Model) move(by int) {
	if len(m.shown) == 0 {
		return
	}
	m.cursor = min(max(m.cursor+by, 0), len(m.shown)-1)
	rows := m.listRows()
	if m.cursor < m.offset {
		m.offset = m.cursor
	} else if m.cursor >= m.offset+rows {
		m.offset = m.cursor - rows + 1
	}
	m.showCurrent()
}

// showCurrent renders the selected page into the viewport, from the cache
// when it has been shown before at this width.
func (m *Model) showCurrent() {
	if len(m.shown) == 0 {
		m.viewport.SetContent("")
		return
	}
	p := m.entries[m.shown[m.cursor]].page
	out, ok := m.rendered[p.Path]
	if !ok {
		out = p.Body
		if r := m.markdown(); r != nil {
			if s, err := r.Render(p.Body); err == nil {
				out = s
			}
		}
		if len(m.rendered) >= renderCacheSize {
			m.rendered = map[string]string{}
		}
		m.rendered[p.Path] = out
	}
	m.viewport.SetContent(out)
	m.viewport.GotoTop()
}

// markdown returns a renderer for the content width, made once per width.
// It uses a fixed dark style: the "auto" style queries the terminal, which
// is slow and can hang over mosh.
func (m *Model) markdown() *glamour.TermRenderer {
	w := max(m.viewport.Width-2, 20)
	if m.renderer != nil && m.rendererW == w {
		return m.renderer
	}
	r, err := glamour.NewTermRenderer(glamour.WithStandardStyle("dark"), glamour.WithWordWrap(w))
	if err != nil {
		return nil
	}
	m.renderer, m.rendererW, m.rendered = r, w, map[string]string{}
	return r
}
