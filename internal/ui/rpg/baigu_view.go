package rpg

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/therealsiege/tuilaoshi/internal/journey"
	"github.com/therealsiege/tuilaoshi/internal/ui"
)

// baiguHeaderRows is the battle header above the disguises: its box and a
// blank line.
const baiguHeaderRows = 4

// skillEnglish names what each disguise tests.
var skillEnglish = map[string]string{"reading": "read", "listening": "listen", "writing": "write"}

// viewBaigu shows the disguise faced now. Where the screen is too short for
// everything, the blank lines go first, then the passage's English (e
// shows it in place of the Chinese).
func (m Model) viewBaigu() string {
	if m.bg.phase == baiguWon {
		return m.viewBaiguWon()
	}
	room := m.height - baiguHeaderRows - strings.Count(ui.LoadError(m.loadErr)+ui.AudioError(m.ttsErr), "\n")
	out := m.baiguLines(true, true)
	for _, try := range [][2]bool{{false, true}, {false, false}} {
		if m.height <= 0 || strings.Count(out, "\n")+1 <= room {
			break
		}
		out = m.baiguLines(try[0], try[1])
	}
	return out
}

// baiguLines lays out the round, with or without blank lines between its
// parts and with or without the passage's English beside its Chinese.
func (m Model) baiguLines(gaps, both bool) string {
	d, r := m.baiguRound()
	bg := m.bg
	answered := bg.phase == baiguRight || bg.phase == baiguWrong
	var parts []string
	add := func(s string) {
		if s != "" {
			parts = append(parts, strings.TrimSuffix(s, "\n"))
		}
	}
	add(m.renderBaiguProgress())

	// The passage: hidden while the old woman's story is heard.
	switch {
	case bg.listen && m.ttsErr == nil:
		add(ui.MutedStyle.Render(ui.WrapIndent(fmt.Sprintf("%s %s tells her story aloud. Listen: her words show once you answer (r hears it again).", d.Name, d.English), m.width)))
	case answered && !both && bg.english:
		add(ui.MutedStyle.Render(ui.WrapIndent(r.English, m.width)))
	default:
		if d.Skill == "listening" && !answered {
			add(ui.MutedStyle.Render("  (No voice to hear her by: read her story instead.)"))
		}
		text := ui.ChineseStyle.Render(ui.WrapIndent(r.Text, m.width))
		if answered && both {
			text += "\n" + ui.MutedStyle.Render(ui.WrapIndent(r.English, m.width))
		}
		add(text)
	}

	add(m.baiguQuestion(r))
	switch {
	case answered:
		add(m.renderBaiguAnswer(r))
	case bg.phase == baiguGrading:
		add("  " + ui.ChineseStyle.Render(bg.typed) + "\n" + ui.MutedStyle.Render("  The master reads what Wukong says…"))
	case m.baiguWrites():
		in := "  " + m.bgIn.View()
		if bg.note != "" {
			in += "\n" + ui.ErrorStyle.Render(ui.WrapIndent(bg.note, m.width))
		}
		add(in)
	default:
		var b strings.Builder
		for i, j := range bg.order {
			fmt.Fprintf(&b, "  %d  %s\n", i+1, r.Choices[j])
		}
		add(b.String())
	}
	if bg.phase == baiguWrong {
		add(ui.ChineseStyle.Render(ui.WrapIndent(bg.trial.Angry, m.width)) + "\n" +
			ui.MutedStyle.Render(ui.WrapIndent(bg.trial.AngryEnglish, m.width)) + "\n" +
			ui.ErrorStyle.Render(fmt.Sprintf("  The party's morale -%d. %s comes again.", journey.BaiguMorale, d.Name)))
	}
	if m.saveErr != nil {
		add(ui.ErrorStyle.Render(ui.WrapIndent("Could not save: "+m.saveErr.Error(), m.width)))
	}
	sep := "\n"
	if gaps {
		sep = "\n\n"
	}
	return strings.Join(parts, sep) + "\n" + ui.HelpStyle.Render(m.baiguHelp(both)) // the help's margin is its gap
}

// baiguQuestion is the round's question with its English, on one line
// when they fit.
func (m Model) baiguQuestion(r journey.BaiguRound) string {
	q, e := "  "+ui.SubtitleStyle.Render(r.Question), r.QuestionEnglish
	if w := m.width; w <= 0 || lipgloss.Width(q)+2+len(e) <= min(w, 80) {
		return q + "  " + ui.MutedStyle.Render(e)
	}
	return q + "\n" + ui.MutedStyle.Render(ui.WrapIndent(e, m.width))
}

// renderBaiguAnswer says what was chosen or written and whether it saw
// through the disguise.
func (m Model) renderBaiguAnswer(r journey.BaiguRound) string {
	bg := m.bg
	d, _ := m.baiguRound()
	var lines []string
	said := func(style lipgloss.Style, mark, label, zh, en string) {
		lines = append(lines, style.Render(ui.WrapIndent(mark+" "+label+zh+"  "+en, m.width)))
	}
	if bg.picked >= 0 {
		style, mark := ui.SuccessStyle, "✓"
		if bg.phase == baiguWrong {
			style, mark = ui.ErrorStyle, "✗"
		}
		said(style, mark, "You chose: ", r.Choices[bg.picked], r.EnglishChoices[bg.picked])
	} else {
		lines = append(lines, "  "+ui.ChineseStyle.Render(bg.typed))
		if bg.feedback != "" {
			lines = append(lines, ui.MutedStyle.Render(ui.WrapIndent(bg.feedback, m.width)))
		}
	}
	if bg.phase == baiguRight {
		lines = append(lines, ui.SuccessStyle.Render(ui.WrapIndent(fmt.Sprintf("Wukong sees through %s %s!", d.Name, d.English), m.width)))
	} else if bg.picked != 0 {
		said(ui.MutedStyle, "→", "It was: ", r.Choices[0], r.EnglishChoices[0])
	}
	return strings.Join(lines, "\n")
}

// baiguHelp is the round's key line.
func (m Model) baiguHelp(both bool) string {
	switch {
	case m.bg.phase == baiguRight && m.bg.disguise == len(m.bg.trial.Disguises)-1:
		return m.withEnglishKey("  enter/space strike the last time  |  esc flee", both)
	case m.bg.phase == baiguRight:
		return m.withEnglishKey("  enter/space on to her next disguise  |  esc flee", both)
	case m.bg.phase == baiguWrong:
		return m.withEnglishKey("  enter/space face her again  |  esc flee", both)
	case m.bg.phase == baiguGrading:
		return "  esc flee"
	case m.baiguWrites():
		return "  type pinyin, space/1-9 pick  |  ctrl+space mode  |  enter say it  |  esc flee"
	case m.bg.listen:
		return "  1-3 answer  |  r hear it again  |  esc flee"
	}
	return "  1-3 answer  |  esc flee"
}

// withEnglishKey adds e to the help when the passage's English is folded away.
func (m Model) withEnglishKey(help string, both bool) string {
	if both {
		return help
	}
	return help + "  |  e English"
}

// renderBaiguProgress shows the three disguises, the one faced marked, and
// the skill it tests.
func (m Model) renderBaiguProgress() string {
	names := make([]string, 0, len(m.bg.trial.Disguises))
	for i, d := range m.bg.trial.Disguises {
		seen := i < m.bg.disguise || (i == m.bg.disguise && (m.bg.phase == baiguRight || m.bg.phase == baiguWon))
		switch {
		case seen:
			names = append(names, ui.MutedStyle.Render("✓ "+d.Name))
		case i == m.bg.disguise:
			names = append(names, ui.WarningStyle.Render("▶ "+d.Name+" "+d.English))
		default:
			names = append(names, ui.MutedStyle.Render(d.Name))
		}
	}
	d, _ := m.baiguRound()
	return fmt.Sprintf("  %s  %s  %s", ui.ChineseStyle.Render("三打白骨精"), strings.Join(names, ui.MutedStyle.Render(" → ")),
		ui.MutedStyle.Render("("+skillEnglish[d.Skill]+")"))
}

// viewBaiguWon is the demon's end and what clearing the tribulation paid.
func (m Model) viewBaiguWon() string {
	var b strings.Builder
	b.WriteString(m.renderBaiguProgress() + "\n\n")
	b.WriteString(ui.SuccessStyle.Bold(true).Render("  VICTORY!  胜利!") + "\n\n")
	b.WriteString(ui.ChineseStyle.Render(ui.WrapIndent(m.bg.trial.Won, m.width)) + "\n")
	b.WriteString(ui.MutedStyle.Render(ui.WrapIndent(m.bg.trial.WonEnglish, m.width)) + "\n\n")
	if m.stage != nil {
		b.WriteString(ui.WarningStyle.Render(ui.WrapIndent(fmt.Sprintf("Tribulation %d cleared! The road west opens.", m.stage.Number), m.width)) + "\n")
	}
	b.WriteString(m.renderEarned())
	if m.saveErr != nil {
		b.WriteString(ui.ErrorStyle.Render(ui.WrapIndent("Could not save: "+m.saveErr.Error(), m.width)) + "\n")
	}
	b.WriteString(ui.HelpStyle.Render("  enter/space continue"))
	return b.String()
}

// renderBaiguIntro tells, on the tribulation's intro, how she is fought.
func (m Model) renderBaiguIntro() string {
	if !m.baiguFight() {
		return ""
	}
	return ui.WarningStyle.Render(ui.WrapIndent("三打白骨精: she comes three times in disguise. See through the village girl (read), the old woman (listen) and the old man (write) to win; a wrong unmasking angers the master.", m.width)) + "\n"
}
