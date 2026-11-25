// Package assessment is the Assess screen: a short adaptive diagnostic per
// skill whose results update the assessment table in wiki/overview.md.
package assessment

import (
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/therealsiege/tuilaoshi/internal/assess"
	"github.com/therealsiege/tuilaoshi/internal/model"
	rpgengine "github.com/therealsiege/tuilaoshi/internal/rpg"
	"github.com/therealsiege/tuilaoshi/internal/ui"
	"github.com/therealsiege/tuilaoshi/internal/wiki"
)

const (
	overviewPage    = "overview.md"
	overviewSection = "Current Level Assessment"
	// questionTries limits how many cards are tried to build one question.
	questionTries = 20
)

type phase int

const (
	phaseIntro phase = iota
	phaseAsking
	phaseResults
)

// Model is the Assess screen.
type Model struct {
	store *wiki.Store
	gkeys ui.GlobalKeys
	phase phase

	skills   []assess.Skill
	stairs   []*assess.Staircase
	skill    int // index of the skill being assessed
	question *rpgengine.CombatQuestion
	selected int
	last     string // "✓" or "✗ …" for the previous answer

	results          []assess.Result
	known            []model.Card // characters and words at or below the levels found, to mark known
	marked           int          // how many were marked known, once done (-1 before)
	loadErr, saveErr error
	width, height    int
}

// New creates the Assess screen.
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

// Refresh loads the skills to assess and returns to the intro.
func (m Model) Refresh() Model {
	pages, pagesErr := m.store.ListPages()
	cards, cardsErr := m.store.DrillableCards()
	m.loadErr = errors.Join(pagesErr, cardsErr)
	m.skills = assess.Skills(pages, cards, wiki.ReadingVocab)
	m.phase, m.results, m.saveErr, m.last = phaseIntro, nil, nil, ""
	return m
}

// start begins a fresh diagnostic.
func (m Model) start() Model {
	m.stairs = make([]*assess.Staircase, len(m.skills))
	for i, s := range m.skills {
		m.stairs[i] = assess.NewStaircase(s.Levels())
	}
	m.skill, m.phase, m.last = 0, phaseAsking, ""
	return m.advance()
}

// advance asks the next question, moving through skills as each finishes,
// and saves the results after the last.
func (m Model) advance() Model {
	for m.skill < len(m.skills) {
		st := m.stairs[m.skill]
		if !st.Done() {
			if q := m.ask(m.skills[m.skill], st.Level()); q != nil {
				m.question, m.selected = q, 0
				return m
			}
			st.Record(false) // a level with no askable question can't be passed
			continue
		}
		m.skill++
	}
	return m.finish()
}

// ask builds a meaning question from a random card at level.
func (m Model) ask(skill assess.Skill, level int) *rpgengine.CombatQuestion {
	cards, pool := skill.ByLevel[level], skill.Pool()
	for i := 0; i < questionTries && len(cards) > 0; i++ {
		if q := rpgengine.MeaningQuestion(cards[rand.Intn(len(cards))], pool); q != nil && len(q.Choices) > 1 {
			return q
		}
	}
	return nil
}

// answer scores the selected choice.
func (m Model) answer() Model {
	q := m.question
	correct := m.selected == q.Correct
	m.stairs[m.skill].Record(correct)
	m.last = "✓"
	if !correct {
		m.last = fmt.Sprintf("✗ %s means %s", q.Chinese, q.CorrectAnswer())
	}
	return m.advance()
}

// finish summarizes every skill and writes the results to overview.md.
func (m Model) finish() Model {
	m.phase, m.question = phaseResults, nil
	m.results = nil
	for i, s := range m.skills {
		m.results = append(m.results, assess.FromStaircase(s, m.stairs[i]))
	}
	date := time.Now().Format("2006-01-02")
	m.saveErr = errors.Join(m.store.UpdateSection(overviewPage, overviewSection, func(old string) string {
		return assess.UpdateSection(old, m.results, date)
	}), m.store.MarkAssessed(time.Now()))
	m.known, m.marked = knownCards(m.skills, m.results), -1
	return m
}

// knownCards are the character and vocabulary cards at or below the level
// each of those two skills was found at.
func knownCards(skills []assess.Skill, results []assess.Result) []model.Card {
	var known []model.Card
	for i, s := range skills {
		if i >= len(results) || (s.Name != "Character recognition" && s.Name != "Vocabulary (passive)") {
			continue
		}
		for level, cards := range s.ByLevel {
			if level <= results[i].Level {
				known = append(known, cards...)
			}
		}
	}
	return known
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	k, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	switch m.phase {
	case phaseIntro:
		if key.Matches(k, m.gkeys.Enter) {
			return m.start(), nil
		}
	case phaseAsking:
		n := len(m.question.Choices)
		switch {
		case key.Matches(k, m.gkeys.Up) && m.selected > 0:
			m.selected--
		case key.Matches(k, m.gkeys.Down) && m.selected < n-1:
			m.selected++
		case key.Matches(k, m.gkeys.Enter):
			return m.answer(), nil
		default:
			if s := k.String(); len(s) == 1 && s[0] >= '1' && int(s[0]-'1') < n {
				m.selected = int(s[0] - '1')
				return m.answer(), nil
			}
		}
	case phaseResults:
		switch {
		case key.Matches(k, m.gkeys.Enter):
			return m, ui.SwitchTo(ui.ViewDashboard)
		case k.String() == "k" && m.marked < 0 && len(m.known) > 0:
			n, err := m.store.MarkKnown(m.known, time.Now())
			m.marked, m.saveErr = n, errors.Join(m.saveErr, err)
		}
	}
	return m, nil
}

func (m Model) View() string {
	var b strings.Builder
	b.WriteString(ui.TitleStyle.Render("Assess 测"))
	b.WriteString("\n\n")
	b.WriteString(ui.LoadError(m.loadErr))
	switch m.phase {
	case phaseIntro:
		b.WriteString("  A short diagnostic: for each skill it starts at HSK3, moves up a level\n")
		fmt.Fprintf(&b, "  after %d of %d right and down otherwise, and finds where you are.\n", assess.PassMark, assess.PerLevel)
		b.WriteString("  Results replace the assessment table in wiki/overview.md.\n\n")
		for _, s := range m.skills {
			levels := s.Levels()
			if len(levels) == 0 {
				fmt.Fprintf(&b, "  %-24s %s\n", s.Name, ui.MutedStyle.Render("not enough cards"))
				continue
			}
			fmt.Fprintf(&b, "  %-24s %s\n", s.Name, ui.MutedStyle.Render(fmt.Sprintf("HSK%d–%d", levels[0], levels[len(levels)-1])))
		}
		b.WriteString("\n")
		b.WriteString(ui.HelpStyle.Render("  enter start · esc back"))
	case phaseAsking:
		b.WriteString(m.viewQuestion())
	case phaseResults:
		b.WriteString(m.viewResults())
	}
	return b.String()
}

func (m Model) viewQuestion() string {
	var b strings.Builder
	s := m.skills[m.skill]
	fmt.Fprintf(&b, "  %s  %s\n\n", ui.SubtitleStyle.Render(fmt.Sprintf("%s · skill %d of %d", s.Name, m.skill+1, len(m.skills))),
		ui.MutedStyle.Render(fmt.Sprintf("HSK%d", m.stairs[m.skill].Level())))
	b.WriteString(ui.SubtitleStyle.Render("  " + m.question.Prompt))
	b.WriteString("\n\n")
	for i, c := range m.question.Choices {
		line := fmt.Sprintf("%d. %s", i+1, c)
		if i == m.selected {
			b.WriteString(ui.SelectedStyle.Render("  ▸ " + line))
		} else {
			b.WriteString(ui.UnselectedStyle.Render("    " + line))
		}
		b.WriteString("\n")
	}
	if m.last != "" {
		b.WriteString("\n" + ui.MutedStyle.Render("  last: "+m.last) + "\n")
	}
	b.WriteString("\n")
	b.WriteString(ui.HelpStyle.Render("  ↑↓ choose · enter/1-4 answer · esc stop (nothing saved)"))
	return b.String()
}

func (m Model) viewResults() string {
	var b strings.Builder
	for _, r := range m.results {
		conf := ""
		if r.Asked > 0 {
			conf = ui.MutedStyle.Render(fmt.Sprintf("  %d/%d right", r.Right, r.Asked))
		}
		fmt.Fprintf(&b, "  %-24s %s%s\n", r.Skill, ui.ChineseStyle.Render(r.Label()), conf)
	}
	b.WriteString("\n")
	if m.saveErr != nil {
		b.WriteString(ui.ErrorStyle.Render(fmt.Sprintf("  Could not update overview.md: %v", m.saveErr)))
	} else {
		b.WriteString(ui.SuccessStyle.Render("  Saved to wiki/overview.md (Current Level Assessment)."))
	}
	b.WriteString("\n\n")
	help := "  enter dashboard"
	switch {
	case m.marked >= 0:
		b.WriteString(ui.SuccessStyle.Render(fmt.Sprintf("  %d words marked known: they leave today's reviews and come back over the next two months.", m.marked)) + "\n\n")
	case len(m.known) > 0:
		fmt.Fprintf(&b, "  %s  Mark the %d characters and words at or below your level as known.\n", ui.ChineseStyle.Render("k"), len(m.known))
		b.WriteString(ui.MutedStyle.Render("     They leave today's reviews and come back spread over the next two months,") + "\n")
		b.WriteString(ui.MutedStyle.Render("     so they stay fresh while you work on the rest. Words you've reviewed keep their schedule.") + "\n\n")
		help = "  k mark them known · enter dashboard"
	}
	b.WriteString(ui.HelpStyle.Render(help))
	return b.String()
}
