package world

import (
	"fmt"
	"slices"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/therealsiege/tuilaoshi/internal/journey"
	"github.com/therealsiege/tuilaoshi/internal/ui"
	"github.com/therealsiege/tuilaoshi/internal/wiki"
)

// The travel rescript (通关文牒): at a kingdom's gate the officer asks four
// questions; type each answer in Chinese. A miss shows the question's
// English, a second miss a sample answer. Four answered earns the
// kingdom's stamp and the court's gift. p opens the gate, or the passport
// with every stamp collected.

// openPassport opens the town's gate if its kingdom hasn't stamped the
// rescript yet, else the passport.
func (m Model) openPassport() (Model, tea.Cmd) {
	t := m.towns[m.town]
	g, ok := journey.GateFor(t.ID)
	if !ok || slices.Contains(m.stamps, t.ID) {
		m.screen, m.note = screenPassport, ""
		return m, nil
	}
	m.gate, m.gateQ, m.misses, m.answered, m.note = g, 0, 0, -1, ""
	m.screen = screenGate
	m.input.Reset()
	return m, m.input.Focus()
}

// updateGate types an answer; enter gives it to the officer.
func (m Model) updateGate(k tea.KeyMsg) (Model, tea.Cmd) {
	switch {
	case m.answered >= 0:
		if k.Type == tea.KeyEnter || k.Type == tea.KeySpace {
			m.screen = screenTown
			return m.Refresh(), nil
		}
		return m, nil
	case k.Type == tea.KeyEnter && !m.input.Composing():
		answer := strings.TrimSpace(m.input.Value())
		if answer == "" {
			return m, nil
		}
		if !m.gate.Questions[m.gateQ].Accepts(answer) {
			m.misses++
			m.note = m.gate.Officer + "皱了皱眉：“我没听明白，再说一遍。”"
			return m, nil
		}
		m.gateQ, m.misses, m.note = m.gateQ+1, 0, ""
		m.input.Reset()
		if m.gateQ < len(m.gate.Questions) {
			return m, nil
		}
		if err := m.store.Stamp(m.gate.Town, wiki.Wallet{Tongqian: journey.StampPay}); err != nil {
			m.loadErr = err
		}
		m.answered = 1
		m.input.Blur()
		return m, nil
	}
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(k)
	return m, cmd
}

func (m Model) gateView() string {
	g, t := m.gate, m.towns[m.town]
	var b strings.Builder
	b.WriteString(ui.TitleStyle.MarginBottom(0).Render("通关文牒 At the gate of "+t.Name) + "  " + ui.MutedStyle.Render(g.Officer+" · "+g.OfficerEnglish) + "\n\n")
	b.WriteString(m.lines(ui.ChineseStyle, g.Greeting, "  ", 76))
	b.WriteString(m.lines(ui.MutedStyle, g.GreetingEnglish, "  ", 76) + "\n")
	if m.answered > 0 {
		b.WriteString("      " + ui.ErrorStyle.Render("【 "+t.Name+" 印 】") + "\n\n")
		b.WriteString(ui.SuccessStyle.Render(fmt.Sprintf("  The officer stamps your rescript with %s's seal, and the court sends %d 铜钱.", t.Name, journey.StampPay)) + "\n")
		b.WriteString("\n" + ui.HelpStyle.Render("  enter into the town"))
		return b.String()
	}
	for i, q := range g.Questions[:m.gateQ] {
		b.WriteString(ui.MutedStyle.Render(fmt.Sprintf("  %d. %s ✓", i+1, q.Ask)) + "\n")
	}
	q := g.Questions[m.gateQ]
	fmt.Fprintf(&b, "  %d. %s\n", m.gateQ+1, ui.ChineseStyle.Render(q.Ask))
	if m.misses > 0 {
		b.WriteString(m.lines(ui.MutedStyle, q.English, "     ", 76))
	}
	if m.misses > 1 {
		b.WriteString(ui.MutedStyle.Render("     e.g. "+q.Sample()) + "\n")
	}
	b.WriteString("\n  " + m.input.View() + "\n\n")
	b.WriteString(ui.HelpStyle.Render("  answer in Chinese · enter give it · ctrl+space pinyin/English · esc leave"))
	return b.String()
}

// passportView shows every kingdom's seal collected, in order along the road.
func (m Model) passportView() string {
	var b strings.Builder
	got := 0
	for _, g := range journey.Gates() {
		if slices.Contains(m.stamps, g.Town) {
			got++
		}
	}
	b.WriteString(ui.TitleStyle.MarginBottom(0).Render("通关文牒 The travel rescript") + "  " + ui.MutedStyle.Render(fmt.Sprintf("%d of %d seals", got, len(journey.Gates()))) + "\n\n")
	names := map[string]journey.Town{}
	for _, t := range m.towns {
		names[t.ID] = t
	}
	for _, g := range journey.Gates() {
		t := names[g.Town]
		if slices.Contains(m.stamps, g.Town) {
			fmt.Fprintf(&b, "  %s  %s\n", ui.ErrorStyle.Render("【"+t.Name+"印】"), ui.MutedStyle.Render(t.English))
		} else if t.Stage <= m.cleared+1 {
			fmt.Fprintf(&b, "  %s  %s\n", ui.WarningStyle.Render("【 "+t.Name+" 】"), ui.MutedStyle.Render(t.English+": passed without its seal; go back to its gate (p)"))
		} else {
			fmt.Fprintf(&b, "  %s  %s\n", ui.MutedStyle.Render("【 ？？ 】"), ui.MutedStyle.Render(fmt.Sprintf("a kingdom ahead (tribulation %d)", t.Stage)))
		}
	}
	b.WriteString("\n" + ui.HelpStyle.Render("  esc back to town"))
	return b.String()
}
