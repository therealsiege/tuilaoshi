package world

import (
	"fmt"
	"strings"

	"github.com/therealsiege/tuilaoshi/internal/journey"
	"github.com/therealsiege/tuilaoshi/internal/ui"
)

// eyesShown is how many of a villager's latest answers are on screen.
const eyesShown = 3

// eyesView shows the rumour, the facts, the villagers and the talk with
// the one being questioned, or how the accusation went.
func (m Model) eyesView() string {
	e := m.eyes
	my := e.mystery
	var b strings.Builder
	b.WriteString(ui.TitleStyle.MarginBottom(0).Render("火眼金睛 Fiery Eyes in "+m.towns[m.town].Name) + "  " + ui.MutedStyle.Render(my.Title+" · "+my.English) + "\n\n")
	b.WriteString("  " + ui.ChineseStyle.Render(ui.WrapText(my.Situation, 36)) + "\n")
	if e.english {
		b.WriteString(mutedWrap(my.SituationEnglish, "  ", 72))
	}
	b.WriteString("\n  " + ui.SubtitleStyle.Render("大家都知道：") + "\n")
	for i, f := range my.Facts {
		b.WriteString("  · " + ui.ChineseStyle.Render(f) + "\n")
		if e.english {
			b.WriteString(mutedWrap(my.FactsEnglish[i], "    ", 72))
		}
	}
	b.WriteString("\n")
	for i, v := range my.Villagers {
		marker := "  "
		if i == e.who && (e.asking || len(e.talks[i]) > 0) {
			marker = ui.SelectedStyle.UnsetPadding().UnsetMargins().Render("▸ ")
		}
		fmt.Fprintf(&b, "  %s%d  %s %s  %s\n", marker, i+1, ui.ChineseStyle.Render(v.Name), v.Role, ui.MutedStyle.Render(v.English))
	}
	if e.accused >= 0 {
		b.WriteString(m.verdictView())
		return b.String()
	}
	b.WriteString(m.talkView())
	left := journey.MysteryQuestions - e.asked
	help := fmt.Sprintf("  1-3 question someone · a accuse · e English · esc back to town · %d %s left", left, plural(left, "question", "questions"))
	switch {
	case e.accusing:
		help = "  Which one is the demon? 1-3 accuse · esc not yet"
	case e.asking:
		help = fmt.Sprintf("  enter ask · ctrl+space pinyin/English · esc stop asking · %d left", left)
	}
	b.WriteString("\n" + ui.HelpStyle.Render(help))
	return b.String()
}

// talkView shows the latest questions put to the villager being questioned
// and the question box.
func (m Model) talkView() string {
	e := m.eyes
	talks := e.talks[e.who]
	if !e.asking && len(talks) == 0 {
		return ""
	}
	v := e.mystery.Villagers[e.who]
	var b strings.Builder
	b.WriteString("\n")
	for _, x := range talks[max(len(talks)-eyesShown, 0):] {
		b.WriteString("  " + ui.MutedStyle.Render("你：") + x.question + "\n")
		switch {
		case x.answer != "":
			b.WriteString("  " + ui.ChineseStyle.Render(v.Name+"："+ui.WrapText(x.answer, 34)) + "\n")
			if e.english && x.english != "" {
				b.WriteString(mutedWrap(x.english, "      ", 70))
			}
		case e.waiting:
			b.WriteString(ui.MutedStyle.Render("  "+v.Name+" thinks…") + "\n")
		}
	}
	if e.asking {
		b.WriteString("\n  " + ui.MutedStyle.Render("问"+v.Name+"：") + m.input.View() + "\n")
	}
	return b.String()
}

// verdictView shows whom the pilgrim accused and what the demon got wrong.
func (m Model) verdictView() string {
	e := m.eyes
	d := e.mystery.Villagers[e.demon]
	var b strings.Builder
	b.WriteString("\n")
	if e.accused == e.demon {
		b.WriteString(ui.SuccessStyle.Render(fmt.Sprintf("  火眼金睛！%s's disguise splits open: a demon! +%d 铜钱", d.Name, journey.MysteryBounty)) + "\n")
	} else {
		b.WriteString(ui.WarningStyle.Render(fmt.Sprintf("  %s is only a villager. In the fuss, %s slips away: it was the demon. The party's spirits sink.", e.mystery.Villagers[e.accused].Name, d.Name)) + "\n")
	}
	b.WriteString("\n  " + ui.SubtitleStyle.Render("破绽 The flaw in its story") + "\n")
	b.WriteString("  " + ui.ChineseStyle.Render(d.Lie) + "\n")
	b.WriteString("  " + ui.ChineseStyle.Render(d.Flaw) + "\n")
	b.WriteString(mutedWrap(d.FlawEnglish, "  ", 72))
	help := "  enter back to town"
	switch {
	case e.foe != nil:
		b.WriteString("\n" + ui.WarningStyle.Render(fmt.Sprintf("  It shows its true form: %s %s. Fight!", e.foe.Name, e.foe.English)) + "\n")
		help = "  enter fight"
	case e.materials:
		b.WriteString("\n" + ui.SuccessStyle.Render("  It flees unmasked, dropping a demon core (妖丹).") + "\n")
	}
	b.WriteString("\n" + ui.HelpStyle.Render(help))
	return b.String()
}

// mutedWrap wraps English to width, each line indented by pad and dimmed.
func mutedWrap(text, pad string, width int) string {
	var b strings.Builder
	for _, line := range strings.Split(ui.WrapText(text, width), "\n") {
		b.WriteString(ui.MutedStyle.Render(pad+line) + "\n")
	}
	return b.String()
}
