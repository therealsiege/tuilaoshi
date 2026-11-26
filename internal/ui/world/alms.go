package world

import (
	"fmt"
	"math/rand"
	"slices"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/therealsiege/tuilaoshi/internal/journey"
	"github.com/therealsiege/tuilaoshi/internal/ui"
)

// Asking for alms (化缘): a townsperson and three things to say. Choosing
// the fitting, polite request brings 斋饭 and lifts the party's spirits;
// either way, the town gives once a day. Trading swaps a material from the
// pack for 斋饭 at the store.

// openAlms meets a townsperson, unless alms were asked here today.
func (m Model) openAlms() Model {
	t := m.towns[m.town]
	asked, err := m.store.AlmsAsked(t.ID, m.now())
	if err != nil {
		m.loadErr = err
		return m
	}
	if asked {
		m.note = "You've asked for alms here today: try again tomorrow."
		return m
	}
	rng := rand.New(rand.NewSource(m.now().UnixNano()))
	m.alms = journey.PickAlms(rng)
	m.almsOrder = rng.Perm(len(m.alms.Choices))
	m.screen, m.answered, m.note = screenAlms, -1, ""
	return m
}

// updateAlms takes 1-3 for what to say; after answering, enter returns.
func (m Model) updateAlms(k tea.KeyMsg) Model {
	if m.answered >= 0 {
		if key.Matches(k, m.gkeys.Enter) || k.Type == tea.KeySpace {
			m.screen = screenTown
		}
		return m
	}
	s := k.String()
	if len(s) != 1 || s[0] < '1' || int(s[0]-'1') >= len(m.almsOrder) {
		return m
	}
	n := int(s[0] - '1')
	if _, err := m.store.AskAlms(m.towns[m.town].ID, m.now()); err != nil {
		m.loadErr = err
		return m
	}
	m.answered = n
	if m.almsOrder[n] == 0 {
		m.withTrail(func(t *journey.Trail, members []string) {
			t.Apply(journey.Outcome{Food: journey.AlmsFood, Morale: journey.AlmsMorale}, members)
		})
	}
	return m
}

func (m Model) almsView() string {
	a := m.alms
	var b strings.Builder
	b.WriteString(ui.TitleStyle.MarginBottom(0).Render("化缘 Asking for alms in "+m.towns[m.town].Name) + "\n\n")
	b.WriteString(m.lines(ui.ChineseStyle, a.Text, "  ", 76))
	if m.answered >= 0 {
		b.WriteString(m.lines(ui.MutedStyle, a.English, "  ", 76))
	}
	b.WriteString("\n  " + ui.SubtitleStyle.Render("你说：") + "\n")
	for i, j := range m.almsOrder {
		line := fmt.Sprintf("  %d  %s", i+1, a.Choices[j])
		switch {
		case m.answered >= 0 && j == 0:
			line = ui.SuccessStyle.Render(line + "  ✓")
		case m.answered == i:
			line = ui.ErrorStyle.Render(line + "  ✗")
		}
		b.WriteString(line + "\n")
		if m.answered >= 0 {
			b.WriteString(m.lines(ui.MutedStyle, a.EnglishChoices[j], "     ", 76))
		}
	}
	if m.answered < 0 {
		b.WriteString("\n" + ui.HelpStyle.Render("  1-3 say it · esc back"))
		return b.String()
	}
	b.WriteString("\n")
	if m.almsOrder[m.answered] == 0 {
		b.WriteString(m.lines(ui.ChineseStyle, a.Thanks, "  ", 76))
		b.WriteString(m.lines(ui.MutedStyle, a.ThanksEnglish, "  ", 76))
		b.WriteString(ui.SuccessStyle.Render(fmt.Sprintf("  +%d 斋饭, and the party's spirits lift.", journey.AlmsFood)) + "\n")
	} else {
		b.WriteString(ui.WarningStyle.Render("  They frown and turn away. A monk asks humbly; try again tomorrow.") + "\n")
	}
	b.WriteString("\n" + ui.HelpStyle.Render("  enter back to town"))
	return b.String()
}

// trade swaps one of the material held most for 斋饭.
func (m Model) trade() Model {
	held, err := m.store.Materials()
	if err != nil {
		m.loadErr = err
		return m
	}
	ids := make([]string, 0, len(held))
	for id, n := range held {
		if n > 0 {
			ids = append(ids, id)
		}
	}
	if len(ids) == 0 {
		m.note = "The storekeeper trades 斋饭 for materials, and your pack is empty: beaten demons drop them."
		return m
	}
	slices.SortFunc(ids, func(a, b string) int {
		if held[a] != held[b] {
			return held[b] - held[a]
		}
		return strings.Compare(a, b)
	})
	id := ids[0]
	if err := m.store.TakeMaterial(id); err != nil {
		m.loadErr = err
		return m
	}
	food := 0
	m.withTrail(func(t *journey.Trail, _ []string) {
		t.Food += journey.Barter
		food = t.Food
	})
	name := id
	if mat := journey.MaterialByID(id); mat != nil {
		name = mat.Name + " " + mat.English
	}
	m.note = fmt.Sprintf("Traded a %s for %d 斋饭: %d now.", name, journey.Barter, food)
	return m
}

// withTrail loads the party's trail, changes it and saves it.
func (m *Model) withTrail(change func(*journey.Trail, []string)) {
	cleared, err := m.store.ClearedStages()
	if err != nil {
		m.loadErr = err
		return
	}
	t, members, err := journey.LoadTrail(m.store, cleared)
	if err != nil {
		m.loadErr = err
		return
	}
	change(&t, members)
	if err := m.store.SaveTrail(t); err != nil {
		m.loadErr = err
	}
}

// Hear is the Chinese open in town, to read aloud: a notice, a townsperson
// met for alms (and their thanks once asked well), or the town itself.
func (m Model) Hear() string {
	t := m.towns[m.town]
	switch m.screen {
	case screenEyes:
		return m.eyesHear()
	case screenRiddle:
		return m.riddle.Riddle
	case screenHaggle:
		return m.haggle.lines[len(m.haggle.lines)-1].zh
	case screenErrand:
		return m.errandHear()
	case screenGate:
		if m.gateQ < len(m.gate.Questions) {
			return m.gate.Questions[m.gateQ].Ask
		}
		return m.gate.Greeting
	case screenQuest:
		return t.Quests[m.quest].Text
	case screenAlms:
		if m.answered >= 0 && m.almsOrder[m.answered] == 0 {
			return m.alms.Text + "\n" + m.alms.Thanks
		}
		return m.alms.Text
	case screenTown:
		return t.Desc
	}
	return ""
}
