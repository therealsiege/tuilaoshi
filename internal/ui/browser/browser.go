package browser

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/cj/mandarin-wiki/internal/model"
	"github.com/cj/mandarin-wiki/internal/ui"
	"github.com/cj/mandarin-wiki/internal/wiki"
)

type focus int

const (
	focusList focus = iota
	focusContent
)

// Model is the browser view model.
type Model struct {
	store    *wiki.Store
	pages    []model.Page
	cursor   int
	focus    focus
	viewport viewport.Model
	rendered string
	search   string
	filtered []int // indices into pages
	width    int
	height   int
	gkeys    ui.GlobalKeys
}

// New creates a new browser model.
func New(store *wiki.Store) Model {
	vp := viewport.New(60, 20)
	return Model{
		store:    store,
		viewport: vp,
		gkeys:    ui.DefaultGlobalKeys(),
	}
}

func (m Model) Init() tea.Cmd {
	return func() tea.Msg { return loadPagesMsg{} }
}

type loadPagesMsg struct{}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case loadPagesMsg:
		pages, _ := m.store.ListPages()
		m.pages = pages
		m.resetFilter()
		if len(m.filtered) > 0 {
			m.renderCurrentPage()
		}
		return m, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.gkeys.Tab):
			if m.focus == focusList {
				m.focus = focusContent
			} else {
				m.focus = focusList
			}
			return m, nil

		case key.Matches(msg, m.gkeys.Up):
			if m.focus == focusList {
				if m.cursor > 0 {
					m.cursor--
					m.renderCurrentPage()
				}
			} else {
				m.viewport.LineUp(1)
			}
			return m, nil

		case key.Matches(msg, m.gkeys.Down):
			if m.focus == focusList {
				if m.cursor < len(m.filtered)-1 {
					m.cursor++
					m.renderCurrentPage()
				}
			} else {
				m.viewport.LineDown(1)
			}
			return m, nil

		case key.Matches(msg, m.gkeys.Enter):
			if m.focus == focusList {
				m.focus = focusContent
			}
			return m, nil

		case key.Matches(msg, m.gkeys.Search):
			// TODO: implement search input
			return m, nil

		case msg.Type == tea.KeyRunes:
			// Quick filter by typing
			if m.focus == focusList {
				m.search += msg.String()
				m.applyFilter()
				m.cursor = 0
				if len(m.filtered) > 0 {
					m.renderCurrentPage()
				}
				return m, nil
			}

		case msg.Type == tea.KeyBackspace:
			if m.focus == focusList && len(m.search) > 0 {
				m.search = m.search[:len(m.search)-1]
				m.applyFilter()
				m.cursor = 0
				if len(m.filtered) > 0 {
					m.renderCurrentPage()
				}
				return m, nil
			}
			if m.focus == focusContent {
				m.focus = focusList
				return m, nil
			}
		}
	}

	if m.focus == focusContent {
		var cmd tea.Cmd
		m.viewport, cmd = m.viewport.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m Model) View() string {
	listWidth := 30
	if m.width > 0 {
		listWidth = m.width / 3
		if listWidth < 20 {
			listWidth = 20
		}
		if listWidth > 40 {
			listWidth = 40
		}
	}

	// Build file list
	var listItems []string
	for i, idx := range m.filtered {
		p := m.pages[idx]
		name := p.Path
		if p.Title != "" {
			name = p.Title
		}
		if len(name) > listWidth-4 {
			name = name[:listWidth-7] + "..."
		}
		if i == m.cursor {
			listItems = append(listItems, ui.SelectedStyle.Render("▸ "+name))
		} else {
			listItems = append(listItems, ui.UnselectedStyle.Render("  "+name))
		}
	}

	listContent := strings.Join(listItems, "\n")
	if m.search != "" {
		listContent = ui.MutedStyle.Render("filter: "+m.search) + "\n\n" + listContent
	}

	listStyle := ui.BorderStyle
	contentStyle := ui.BorderStyle
	if m.focus == focusList {
		listStyle = ui.ActiveBorderStyle
	} else {
		contentStyle = ui.ActiveBorderStyle
	}

	contentWidth := 60
	contentHeight := 20
	if m.width > 0 {
		contentWidth = m.width - listWidth - 8
		if contentWidth < 30 {
			contentWidth = 30
		}
	}
	if m.height > 0 {
		contentHeight = m.height - 6
		if contentHeight < 10 {
			contentHeight = 10
		}
	}

	m.viewport.Width = contentWidth - 4
	m.viewport.Height = contentHeight - 2

	list := listStyle.Width(listWidth).Height(contentHeight).Render(listContent)
	content := contentStyle.Width(contentWidth).Height(contentHeight).Render(m.viewport.View())

	main := lipgloss.JoinHorizontal(lipgloss.Top, list, content)

	title := ui.TitleStyle.Render("Wiki Browser 浏览")
	help := ui.HelpStyle.Render("  tab switch pane · ↑↓ navigate · esc back · type to filter")

	return fmt.Sprintf("%s\n%s\n%s", title, main, help)
}

// SetSize updates the terminal dimensions.
func (m Model) SetSize(w, h int) Model {
	m.width = w
	m.height = h
	return m
}

// Refresh reloads pages from store.
func (m Model) Refresh() Model {
	pages, _ := m.store.ListPages()
	m.pages = pages
	m.resetFilter()
	m.cursor = 0
	if len(m.filtered) > 0 {
		m.renderCurrentPage()
	}
	return m
}

func (m *Model) resetFilter() {
	m.filtered = make([]int, len(m.pages))
	for i := range m.pages {
		m.filtered[i] = i
	}
	m.search = ""
}

func (m *Model) applyFilter() {
	if m.search == "" {
		m.resetFilter()
		return
	}
	m.filtered = nil
	lower := strings.ToLower(m.search)
	for i, p := range m.pages {
		if strings.Contains(strings.ToLower(p.Path), lower) ||
			strings.Contains(strings.ToLower(p.Title), lower) {
			m.filtered = append(m.filtered, i)
		}
	}
}

func (m *Model) renderCurrentPage() {
	if len(m.filtered) == 0 || m.cursor >= len(m.filtered) {
		m.rendered = ""
		m.viewport.SetContent("")
		return
	}

	page := m.pages[m.filtered[m.cursor]]

	width := 60
	if m.width > 0 {
		width = m.width - 40
		if width < 40 {
			width = 40
		}
	}

	renderer, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(width),
	)
	if err != nil {
		m.viewport.SetContent(page.Body)
		return
	}

	rendered, err := renderer.Render(page.Body)
	if err != nil {
		m.viewport.SetContent(page.Body)
		return
	}

	m.rendered = rendered
	m.viewport.SetContent(rendered)
	m.viewport.GotoTop()
}
