package celebrate

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/cj/mandarin-wiki/internal/ui"
)

// Event types for celebrations.
const (
	RankUp       = "rank_up"
	BossDefeated = "boss_defeated"
	StreakMile   = "streak"
	Achievement  = "achievement"
	CompanionNew = "companion"
)

// boxWidth is the inner width of celebration boxes (between the border chars).
const boxWidth = 40

// Celebration holds the data for a celebration display.
type Celebration struct {
	Event   string
	Title   string
	Chinese string
	Desc    string
	Art     string // pre-composed ASCII art lines (no border)
}

// RankUpCelebration creates a rank-up celebration.
func RankUpCelebration(rankName, rankEnglish string) Celebration {
	art := strings.Join([]string{
		"            /\\",
		"           /  \\",
		"          / ** \\",
		"         /      \\",
		"        /________\\",
	}, "\n")

	return Celebration{
		Event:   RankUp,
		Title:   "CULTIVATION BREAKTHROUGH",
		Chinese: rankName,
		Desc:    rankName + " * " + rankEnglish,
		Art:     art,
	}
}

// BossDefeatCelebration creates a boss defeat celebration.
func BossDefeatCelebration(bossName, bossEnglish string) Celebration {
	art := strings.Join([]string{
		"            X",
		"           / \\",
		"          /   \\",
	}, "\n")

	return Celebration{
		Event:   BossDefeated,
		Title:   "DEMON VANQUISHED",
		Chinese: bossName,
		Desc:    bossName + " * " + bossEnglish,
		Art:     art,
	}
}

// StreakCelebration creates a streak milestone celebration.
func StreakCelebration(days int) Celebration {
	var art string
	var chinese string

	switch {
	case days >= 100:
		chinese = fmt.Sprintf("连续%d天！百日苦修！", days)
		art = strings.Join([]string{
			"        ( ( (",
			"         ) ) )",
			"        ( ( (",
			"      .------.",
			"      |      |",
			"      '------'",
		}, "\n")
	case days >= 30:
		chinese = fmt.Sprintf("连续%d天！", days)
		art = strings.Join([]string{
			"         ) ) )",
			"        ( ( (",
			"      .------.",
			"      |      |",
			"      '------'",
		}, "\n")
	case days >= 7:
		chinese = fmt.Sprintf("连续%d天！", days)
		art = strings.Join([]string{
			"        ( ( (",
			"      .------.",
			"      |      |",
			"      '------'",
		}, "\n")
	default:
		chinese = fmt.Sprintf("连续%d天！", days)
		art = strings.Join([]string{
			"         ( (",
			"      .----.",
			"      |    |",
			"      '----'",
		}, "\n")
	}

	return Celebration{
		Event:   StreakMile,
		Title:   "STREAK MILESTONE",
		Chinese: chinese,
		Desc:    fmt.Sprintf("%d-day streak!", days),
		Art:     art,
	}
}

// CompanionCelebration creates a new companion celebration.
func CompanionCelebration(name, english string) Celebration {
	art := strings.Join([]string{
		"            O",
		"           /|\\",
		"           / \\",
	}, "\n")

	return Celebration{
		Event:   CompanionNew,
		Title:   "NEW COMPANION",
		Chinese: name,
		Desc:    name + " * " + english,
		Art:     art,
	}
}

// AchievementCelebration creates a generic achievement celebration.
func AchievementCelebration(title, chinese, desc string) Celebration {
	art := strings.Join([]string{
		"         .---.",
		"         | * |",
		"        /_____\\",
		"        |     |",
		"        '-----'",
	}, "\n")

	return Celebration{
		Event:   Achievement,
		Title:   title,
		Chinese: chinese,
		Desc:    desc,
		Art:     art,
	}
}

// Render formats a celebration for terminal display with lipgloss colors.
func (c Celebration) Render() string {
	var b strings.Builder

	// Select styles based on event type.
	titleStyle := ui.SubtitleStyle
	accentStyle := ui.SuccessStyle
	borderStyle := ui.MutedStyle

	switch c.Event {
	case BossDefeated:
		accentStyle = ui.WarningStyle
		titleStyle = ui.WarningStyle.Bold(true)
	case RankUp:
		accentStyle = ui.SuccessStyle
		titleStyle = ui.SuccessStyle.Bold(true)
	case StreakMile:
		accentStyle = ui.SubtitleStyle
		titleStyle = ui.SubtitleStyle.Bold(true)
	case CompanionNew:
		accentStyle = ui.ChineseStyle
		titleStyle = ui.ChineseStyle.Bold(true)
	case Achievement:
		accentStyle = ui.SuccessStyle
		titleStyle = ui.SuccessStyle.Bold(true)
	}

	// Build the decorative header symbol per event.
	var headerDecor string
	switch c.Event {
	case RankUp:
		headerDecor = "**"
	case BossDefeated:
		headerDecor = "X"
	case StreakMile:
		headerDecor = "~"
	case CompanionNew:
		headerDecor = "*"
	default:
		headerDecor = "*"
	}

	// Top border.
	top := borderStyle.Render("    " + "\u2554" + strings.Repeat("\u2550", boxWidth) + "\u2557")
	b.WriteString(top)
	b.WriteString("\n")

	// Title line.
	titleText := fmt.Sprintf(" %s %s %s ", headerDecor, c.Title, headerDecor)
	titleRendered := titleStyle.Render(titleText)
	b.WriteString(boxLine(borderStyle, titleRendered, boxWidth))
	b.WriteString("\n")

	// Empty line.
	b.WriteString(boxLine(borderStyle, "", boxWidth))
	b.WriteString("\n")

	// Art lines.
	for _, line := range strings.Split(c.Art, "\n") {
		if line == "" {
			continue
		}
		artRendered := accentStyle.Render(line)
		b.WriteString(boxLine(borderStyle, artRendered, boxWidth))
		b.WriteString("\n")
	}

	// Empty line after art.
	b.WriteString(boxLine(borderStyle, "", boxWidth))
	b.WriteString("\n")

	// Chinese text line.
	chineseRendered := ui.ChineseStyle.Render("    " + c.Chinese)
	b.WriteString(boxLine(borderStyle, chineseRendered, boxWidth))
	b.WriteString("\n")

	// Description line.
	descRendered := accentStyle.Render("    " + c.Desc)
	b.WriteString(boxLine(borderStyle, descRendered, boxWidth))
	b.WriteString("\n")

	// Bottom border.
	bot := borderStyle.Render("    " + "\u255A" + strings.Repeat("\u2550", boxWidth) + "\u255D")
	b.WriteString(bot)
	b.WriteString("\n")

	return b.String()
}

// boxLine renders a single line inside the celebration box.
// content is already styled; rawWidth is the inner width between borders.
func boxLine(borderStyle lipgloss.Style, content string, rawWidth int) string {
	left := borderStyle.Render("    \u2551")
	right := borderStyle.Render("\u2551")

	if content == "" {
		return left + strings.Repeat(" ", rawWidth) + right
	}

	// Pad to rawWidth using the visible (unstyled) length of content.
	vis := visibleLen(content)
	pad := rawWidth - vis
	if pad < 0 {
		pad = 0
	}
	return left + content + strings.Repeat(" ", pad) + right
}

// visibleLen estimates the visible character count of a styled string by
// stripping ANSI escape sequences. Counts each rune as 1 for simplicity;
// good enough for box alignment with ASCII art.
func visibleLen(s string) int {
	inEsc := false
	n := 0
	for _, r := range s {
		if r == '\033' {
			inEsc = true
			continue
		}
		if inEsc {
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
				inEsc = false
			}
			continue
		}
		n++
	}
	return n
}
