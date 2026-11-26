package world

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/therealsiege/tuilaoshi/internal/journey"
	"github.com/therealsiege/tuilaoshi/internal/ui"
	"github.com/therealsiege/tuilaoshi/internal/wiki"
)

// Lantern riddles (灯谜): a town hangs one character riddle a day, three at
// the Lantern Festival (元宵节), when they pay double. Type the character;
// after three wrong guesses the answer is shown and the riddle is done.

// feast reports whether today is the Lantern Festival.
func (m Model) feast() bool {
	o, on, ok := journey.FestivalNow(m.now())
	return ok && on && o.Name == journey.LanternFeast
}

// riddlesToday is how many riddles a town hangs today.
func (m Model) riddlesToday() int {
	if m.feast() {
		return journey.RiddlesAtFeast
	}
	return journey.RiddlesADay
}

// openRiddle shows the town's next riddle today, if any is left.
func (m Model) openRiddle() (Model, tea.Cmd) {
	t := m.towns[m.town]
	done, err := m.store.RiddlesDone(t.ID, m.now())
	if err != nil {
		m.loadErr = err
		return m, nil
	}
	if done >= m.riddlesToday() {
		m.note = "Today's lanterns are all read. New riddles hang tomorrow."
		return m, nil
	}
	m.riddle = journey.RiddleFor(t.ID, m.now().Format("2006-01-02"), done)
	m.screen, m.tries, m.answered, m.note = screenRiddle, 0, -1, ""
	m.input.Reset()
	return m, m.input.Focus()
}

// updateRiddle types a guess; enter tries it.
func (m Model) updateRiddle(k tea.KeyMsg) (Model, tea.Cmd) {
	switch {
	case m.answered >= 0:
		if k.Type == tea.KeyEnter || k.Type == tea.KeySpace {
			m.screen = screenTown
		}
		return m, nil
	case k.Type == tea.KeyEnter && !m.input.Composing():
		guess := m.input.Value()
		if strings.TrimSpace(guess) == "" {
			return m, nil
		}
		if m.riddle.Solves(guess) {
			return m.finishRiddle(1), nil
		}
		m.tries++
		if m.tries >= journey.RiddleTries {
			return m.finishRiddle(0), nil
		}
		m.note = fmt.Sprintf("Not %s. %d more %s.", strings.TrimSpace(guess), journey.RiddleTries-m.tries, plural(journey.RiddleTries-m.tries, "try", "tries"))
		m.input.Reset()
		return m, nil
	}
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(k)
	return m, cmd
}

// finishRiddle ends the riddle: solved (1) pays, given up (0) doesn't.
func (m Model) finishRiddle(solved int) Model {
	pay := journey.RiddlePay * solved
	if m.feast() {
		pay *= 2
	}
	if err := m.store.FinishRiddle(m.towns[m.town].ID, m.now(), wiki.Wallet{Tongqian: pay}); err != nil {
		m.loadErr = err
	}
	m.answered, m.note = solved, ""
	m.input.Blur()
	return m
}

func (m Model) riddleView() string {
	r := m.riddle
	var b strings.Builder
	title := "灯谜 A lantern riddle in " + m.towns[m.town].Name
	if m.feast() {
		title = "元宵灯会 The Lantern Festival fair in " + m.towns[m.town].Name
	}
	b.WriteString(ui.TitleStyle.MarginBottom(0).Render(title) + "\n\n")
	b.WriteString("  " + ui.MutedStyle.Render("A paper lantern sways, a riddle on it. 猜一个字: guess the character.") + "\n\n")
	b.WriteString("      " + ui.ChineseStyle.Render("谜面："+r.Riddle) + "\n\n")
	if m.answered < 0 {
		b.WriteString("  " + m.input.View() + "\n\n")
		b.WriteString(ui.HelpStyle.Render("  type the character · enter guess · esc leave"))
		return b.String()
	}
	b.WriteString("      " + ui.ChineseStyle.Render("谜底："+r.Answer) + " " + ui.Pinyin(r.Pinyin) + "\n")
	b.WriteString(m.lines(ui.MutedStyle, r.English, "  ", 76))
	b.WriteString(m.lines(ui.MutedStyle, r.Explain, "  ", 76) + "\n")
	if m.answered > 0 {
		pay := journey.RiddlePay
		if m.feast() {
			pay *= 2
		}
		b.WriteString(ui.SuccessStyle.Render(fmt.Sprintf("  猜中了！The lantern keeper pays you %d 铜钱.", pay)) + "\n")
	} else {
		b.WriteString(ui.WarningStyle.Render("  The lantern keeper smiles and tells you the answer.") + "\n")
	}
	b.WriteString("\n" + ui.HelpStyle.Render("  enter back to town"))
	return b.String()
}

func plural(n int, one, many string) string {
	if n == 1 {
		return one
	}
	return many
}
