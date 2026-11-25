package title

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/therealsiege/tuilaoshi/internal/claude"
	"github.com/therealsiege/tuilaoshi/internal/ui"
	"github.com/therealsiege/tuilaoshi/internal/wiki"
)

var banner = []string{
	`        /\          /\/\                     /\`,
	`   /\  /  \   /\   /    \      ~~~~     /\  /  \  /\`,
	`  /  \/    \_/  \_/      \___ ~~~~ ___/  \/    \/  \___`,
}

func (m Model) View() string {
	var b strings.Builder
	b.WriteString("\n")
	for _, l := range banner {
		b.WriteString(ui.MutedStyle.Render("  "+l) + "\n")
	}
	b.WriteString("\n")
	b.WriteString(ui.TitleStyle.MarginBottom(0).Render("      西 行 · 推 老 师") + "\n")
	b.WriteString(ui.MutedStyle.Render("      TuīLǎoShī · a Mandarin journey to the West") + "\n\n")
	b.WriteString(ui.LoadError(m.loadErr))
	if !claude.KeyPresent() {
		b.WriteString(ui.MutedStyle.Render("      No Claude key: reviews, flashcards, battles and the road work offline;") + "\n")
		b.WriteString(ui.MutedStyle.Render("      chat, talks, duels, writing and the diary need ANTHROPIC_API_KEY in ~/.config/tuilaoshi/env.") + "\n\n")
	}
	switch m.screen {
	case screenName:
		b.WriteString(ui.SubtitleStyle.Render("      Name your journey") + "\n\n")
		b.WriteString("      " + m.name.View() + "\n\n")
		b.WriteString(ui.HelpStyle.Render("      enter begin · ctrl+space pinyin/English · esc back"))
	case screenLoad:
		b.WriteString(m.loadView())
	default:
		b.WriteString(m.menuView())
	}
	return b.String()
}

func (m Model) menuView() string {
	var b strings.Builder
	for i, c := range m.menu {
		label := map[choice]string{choiceContinue: "Continue", choiceNew: "New journey", choiceLoad: "Load journey", choiceQuit: "Quit"}[c]
		detail := ""
		switch c {
		case choiceContinue:
			detail = m.saves[0].Name + " · " + describe(m.saves[0], m.now())
		case choiceLoad:
			detail = fmt.Sprintf("%d saved", len(m.saves))
		case choiceNew:
			detail = "your words and reviews come with you"
		}
		line := fmt.Sprintf("%-14s", label)
		if i == m.cursor {
			line = ui.SelectedStyle.UnsetPadding().UnsetMargins().Render("▸ "+line) + "  " + ui.MutedStyle.Render(detail)
		} else {
			line = "  " + line + "  " + ui.MutedStyle.Render(detail)
		}
		b.WriteString("      " + line + "\n")
	}
	b.WriteString("\n" + ui.HelpStyle.Render("      ↑↓ choose · enter select · q quit"))
	return b.String()
}

func (m Model) loadView() string {
	var b strings.Builder
	b.WriteString(ui.SubtitleStyle.Render("      Saved journeys") + "\n\n")
	for i, s := range m.saves {
		line := lipgloss.NewStyle().Width(24).Render(s.Name) + "  " + ui.MutedStyle.Render(describe(s, m.now()))
		if i == m.pick {
			line = ui.SelectedStyle.UnsetPadding().UnsetMargins().Render("▸ ") + line
		} else {
			line = "  " + line
		}
		b.WriteString("      " + line + "\n")
	}
	b.WriteString("\n")
	if m.confirm {
		b.WriteString(ui.WarningStyle.Render(fmt.Sprintf("      Press d again to delete %q. Your study progress stays.", m.saves[m.pick].Name)) + "\n")
	}
	b.WriteString(ui.HelpStyle.Render("      ↑↓ choose · enter load · d delete · esc back"))
	return b.String()
}

// describe summarises a save: how far it has come and when it was played.
func describe(s wiki.SaveInfo, now time.Time) string {
	return fmt.Sprintf("%d/81 tribulations · %d 铜钱 · played %s", s.Cleared, s.Coins.Tongqian, ago(s.Played, now))
}

// ago says how long ago t was, roughly.
func ago(t, now time.Time) string {
	d := now.Sub(t)
	switch {
	case t.IsZero():
		return "never"
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		return fmt.Sprintf("%d min ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%d h ago", int(d.Hours()))
	}
	return t.Format("2 Jan")
}
