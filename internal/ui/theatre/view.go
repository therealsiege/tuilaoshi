package theatre

import (
	"fmt"
	"strings"

	"github.com/therealsiege/tuilaoshi/internal/episode"
	"github.com/therealsiege/tuilaoshi/internal/ui"
	"github.com/therealsiege/tuilaoshi/internal/wiki"
)

func (m Model) View() string {
	var b strings.Builder
	switch m.screen {
	case screenListen:
		b.WriteString(m.listenView())
	case screenScene:
		b.WriteString(m.sceneView())
	case screenEpisode:
		b.WriteString(m.episodeView())
	default:
		b.WriteString(m.listView())
	}
	if m.note != "" {
		b.WriteString("\n" + ui.WarningStyle.Render("  "+m.note))
	}
	return b.String()
}

func (m Model) listView() string {
	var b strings.Builder
	b.WriteString(ui.TitleStyle.MarginBottom(0).Render("剧场 Episodes") + "\n")
	b.WriteString(ui.MutedStyle.Render("  Your own episodes, each opening at its point on the road.") + "\n\n")
	b.WriteString(ui.LoadError(m.loadErr))
	if len(m.eps) == 0 {
		b.WriteString(ui.MutedStyle.Render("  No episodes yet. Put an episode's subtitles (and its mp3) in a folder under wiki/private/episodes.") + "\n")
	}
	for i, e := range m.eps {
		status := ui.MutedStyle.Render(fmt.Sprintf("opens at tribulation %d", e.Stage))
		name := ui.MutedStyle.Render(e.Title)
		if e.Open(m.current) {
			name = ui.ChineseStyle.Render(e.Title)
			status = fmt.Sprintf("%d/%d scenes", m.done(e), len(e.Scenes()))
			if m.claimed[e.FinishID()] {
				status = ui.SuccessStyle.Render("✓ finished")
			}
		}
		marker := "  "
		if i == m.cursor {
			marker = ui.SelectedStyle.UnsetPadding().UnsetMargins().Render("▸ ")
		}
		fmt.Fprintf(&b, "  %s%s  %s\n", marker, name, status)
	}
	b.WriteString("\n" + ui.HelpStyle.Render("  ↑↓ choose · enter open · esc back"))
	return b.String()
}

func (m Model) episodeView() string {
	e := m.eps[m.ep]
	var b strings.Builder
	b.WriteString(ui.TitleStyle.MarginBottom(0).Render(e.Title) + "\n")
	words := e.Words(m.cards)
	var some []string
	for _, w := range words[:min(8, len(words))] {
		some = append(some, w.Chinese)
	}
	fmt.Fprintf(&b, "  %s\n", ui.MutedStyle.Render(fmt.Sprintf("%d lines · %d new words, first in your reviews: %s", len(e.Lines), len(words), strings.Join(some, " "))))
	b.WriteString("\n" + ui.SubtitleStyle.Render("  Scenes") + ui.MutedStyle.Render(fmt.Sprintf("  %d 铜钱 each; all of them: +%d 铜钱 +%d 元宝", episode.ScenePay, episode.FinishCopper, episode.FinishIngots)) + "\n")
	rows := max(m.height-14, 6)
	first := max(0, min(m.scene-rows/2, len(m.scenes)-rows))
	for i := first; i < min(first+rows, len(m.scenes)); i++ {
		s := m.scenes[i]
		status := ""
		switch m.states[e.QuestID(s.Index)] {
		case wiki.QuestDone:
			status = ui.SuccessStyle.Render("✓")
		case wiki.QuestLocked:
			status = ui.MutedStyle.Render("tomorrow")
		}
		first := s.Lines[0].Chinese
		line := fmt.Sprintf("Scene %d  %s", s.Index+1, ui.MutedStyle.Render(truncate(first, 24)))
		if i == m.scene {
			fmt.Fprintf(&b, "  %s%s  %s\n", ui.SelectedStyle.UnsetPadding().UnsetMargins().Render("▸ "), line, status)
		} else {
			fmt.Fprintf(&b, "    %s  %s\n", line, status)
		}
	}
	b.WriteString("\n" + ui.HelpStyle.Render("  ↑↓ scene · enter answer its question · r read & listen · l listening drill · t talk it over · esc back"))
	return b.String()
}

// sceneView shows the scene's lines in Chinese and asks what the last one
// means; the English shows once answered.
func (m Model) sceneView() string {
	e, s := m.eps[m.ep], m.scenes[m.scene]
	var b strings.Builder
	b.WriteString(ui.TitleStyle.MarginBottom(0).Render(fmt.Sprintf("%s · Scene %d", e.Title, s.Index+1)) + "\n\n")
	for i, l := range s.Lines {
		style := ui.MutedStyle
		if i == len(s.Lines)-1 {
			style = ui.ChineseStyle.Bold(true)
		}
		speaker := ""
		if l.Speaker != "" {
			speaker = l.Speaker + "："
		}
		b.WriteString("  " + style.Render(speaker+l.Chinese) + "\n")
		if m.answered >= 0 && strings.TrimSpace(l.English) != "" {
			b.WriteString(ui.MutedStyle.Render("    "+l.English) + "\n")
		}
	}
	b.WriteString("\n  " + ui.SubtitleStyle.Render("What does the last line mean?") + "\n")
	for i, c := range s.Choices {
		line := fmt.Sprintf("  %d  %s", i+1, c)
		switch {
		case m.answered >= 0 && i == s.Right:
			line = ui.SuccessStyle.Render(line + "  ✓")
		case m.answered == i:
			line = ui.ErrorStyle.Render(line + "  ✗")
		}
		b.WriteString(line + "\n")
	}
	switch {
	case m.answered < 0:
		b.WriteString("\n" + ui.HelpStyle.Render("  1-4 answer · esc back"))
	case m.answered == s.Right:
		b.WriteString("\n" + ui.SuccessStyle.Render(fmt.Sprintf("  Right! +%d 铜钱", episode.ScenePay)) + "\n" + ui.HelpStyle.Render("  enter back to the scenes"))
	default:
		b.WriteString("\n" + ui.WarningStyle.Render("  Not quite. Read or hear the scene again, and try it tomorrow.") + "\n" + ui.HelpStyle.Render("  enter back to the scenes"))
	}
	return b.String()
}

func (m Model) listenView() string {
	e, l := m.eps[m.ep], m.listen
	line := e.Lines[l.drill.Line]
	var b strings.Builder
	b.WriteString(ui.TitleStyle.MarginBottom(0).Render(e.Title+" · Listening") + "\n\n")
	if l.drill.Type {
		b.WriteString("  " + ui.SubtitleStyle.Render("Type the line you hear.") + ui.MutedStyle.Render(fmt.Sprintf("  (%d characters)", len([]rune(hanOnly(line.Chinese))))) + "\n\n")
		b.WriteString("  " + l.input.View() + "\n")
	} else {
		b.WriteString("  " + ui.SubtitleStyle.Render("What does the line you hear mean?") + "\n")
		for i, c := range l.drill.Choices {
			row := fmt.Sprintf("  %d  %s", i+1, c)
			switch {
			case l.answered && i == l.drill.Right:
				row = ui.SuccessStyle.Render(row + "  ✓")
			case l.answered && i == l.picked:
				row = ui.ErrorStyle.Render(row + "  ✗")
			}
			b.WriteString(row + "\n")
		}
	}
	if l.answered {
		verdict := ui.SuccessStyle.Render("  Caught it!")
		if !l.right {
			verdict = ui.WarningStyle.Render("  Not quite. Here it is:")
		}
		b.WriteString("\n" + verdict + "\n  " + ui.ChineseStyle.Render(line.Chinese) + "\n" + ui.MutedStyle.Render("  "+line.English) + "\n")
		b.WriteString("\n" + ui.HelpStyle.Render("  enter next line · tab hear it again · esc finish"))
		return b.String()
	}
	help := "  1-4 answer · tab hear it again · esc finish"
	if l.drill.Type {
		help = "  type pinyin, pick characters · enter answer · tab hear it again · esc finish"
	}
	b.WriteString("\n" + ui.HelpStyle.Render(help))
	return b.String()
}

func truncate(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n]) + "…"
}

func hanOnly(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r >= 0x4E00 && r <= 0x9FFF {
			b.WriteRune(r)
		}
	}
	return b.String()
}
