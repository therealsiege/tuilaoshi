package dashboard

import (
	"strings"

	"github.com/therealsiege/tuilaoshi/internal/camp"
	"github.com/therealsiege/tuilaoshi/internal/ui"
	"github.com/therealsiege/tuilaoshi/internal/wiki"
)

// Guanyin's letter (观音来信): each Monday a short letter in Chinese about
// the week before, written from the review log: how much was reviewed and
// remembered, the words that keep slipping, what the coming week brings,
// and the Inner Demon. It is reading practice made of the learner's own
// progress. A letter comes once there has been a week with reviews.

// loadLetter writes this week's letter if it is due and returns the latest.
func (m Model) loadLetter() (*wiki.Letter, error) {
	return camp.LatestLetter(m.store, m.now())
}

// Hear is the open gazette's or letter's Chinese, to read aloud.
func (m Model) Hear() string {
	if m.gazetteOpen && m.gazette != nil {
		return m.gazette.Text
	}
	if m.reading && m.letter != nil {
		return m.letter.Text
	}
	return ""
}

// renderLetterLine says a letter is waiting, at camp.
func (m Model) renderLetterLine() string {
	if m.letter == nil {
		return ""
	}
	label := "\n  " + ui.SubtitleStyle.Render("来信") + "      "
	if !m.letter.Read {
		return label + ui.WarningStyle.Render("A letter from Guanyin has come (n)")
	}
	return label + ui.MutedStyle.Render("Guanyin's letter of "+m.letter.Week+" (n)")
}

// letterView is the letter open, with its English below.
func (m Model) letterView() string {
	var b strings.Builder
	b.WriteString(ui.TitleStyle.MarginBottom(0).Render("观音来信 A letter from Guanyin") + "  " + ui.MutedStyle.Render(m.letter.Week) + "\n\n")
	width := 76
	if m.width > 0 {
		width = m.width - 4
	}
	for _, l := range strings.Split(m.letter.Text, "\n") {
		for _, w := range ui.WrapDisplay(l, width) {
			b.WriteString("  " + ui.ChineseStyle.Render(w) + "\n")
		}
	}
	b.WriteString("\n")
	for _, l := range strings.Split(m.letter.English, "\n") {
		for _, w := range ui.WrapDisplay(l, width) {
			b.WriteString(ui.MutedStyle.Render("  "+w) + "\n")
		}
	}
	b.WriteString("\n" + ui.HelpStyle.Render("  ctrl+r hear it · enter or esc back to camp"))
	return b.String()
}
