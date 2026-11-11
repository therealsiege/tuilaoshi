package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Colors — a warm Chinese-inspired palette: vermillion, gold, jade, ink.
var (
	ColorPrimary   = lipgloss.Color("#C54B3C") // vermillion/cinnabar red
	ColorSecondary = lipgloss.Color("#D4A017") // gold/amber
	ColorAccent    = lipgloss.Color("#B87333") // copper/bronze
	ColorSuccess   = lipgloss.Color("#00A86B") // jade green
	ColorWarning   = lipgloss.Color("#E5C07B") // warm yellow
	ColorMuted     = lipgloss.Color("#6B7280") // ink gray
	ColorText      = lipgloss.Color("#D1D5DB") // silk white
	ColorBg        = lipgloss.Color("#1A1A2E") // ink black
)

// Styles for reuse across views.
var (
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorPrimary).
			MarginBottom(1)

	SubtitleStyle = lipgloss.NewStyle().
			Foreground(ColorSecondary).
			Bold(true)

	MutedStyle = lipgloss.NewStyle().
			Foreground(ColorMuted)

	SuccessStyle = lipgloss.NewStyle().
			Foreground(ColorSuccess)

	WarningStyle = lipgloss.NewStyle().
			Foreground(ColorWarning)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#C54B3C"))

	BorderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorMuted).
			Padding(1, 2)

	ActiveBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(ColorSecondary).
				Padding(1, 2)

	HelpStyle = lipgloss.NewStyle().
			Foreground(ColorMuted).
			MarginTop(1)

	ChineseStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorPrimary)

	PinyinStyle = lipgloss.NewStyle().
			Foreground(ColorSecondary).
			Italic(true)

	CardStyle = lipgloss.NewStyle().
			Border(lipgloss.DoubleBorder()).
			BorderForeground(ColorPrimary).
			Padding(1, 4).
			MarginBottom(1)

	SelectedStyle = lipgloss.NewStyle().
			Foreground(ColorPrimary).
			Bold(true)

	UnselectedStyle = lipgloss.NewStyle().
			Foreground(ColorText)

	TabStyle = lipgloss.NewStyle().
			Padding(0, 2).
			Foreground(ColorMuted)

	ActiveTabStyle = lipgloss.NewStyle().
			Padding(0, 2).
			Foreground(ColorPrimary).
			Bold(true).
			Underline(true)

	StatusBarStyle = lipgloss.NewStyle().
			Foreground(ColorText).
			Background(lipgloss.Color("#2A2A3E")).
			Padding(0, 1)

	// Chat bubble styles
	ChatBubbleUserStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("#2D1F1F")).
				Foreground(lipgloss.Color("#FFFFFF")).
				Padding(0, 2).
				MarginLeft(4)

	ChatBubbleAssistStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("#1F2D1F")).
				Foreground(ColorText).
				Padding(0, 2).
				MarginRight(4)

	ChatTimestampStyle = lipgloss.NewStyle().
				Foreground(ColorMuted).
				Italic(true)

	// Difficulty styles
	DiffAgainStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#C54B3C")).
			Bold(true)

	DiffHardStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#D4A017")).
			Bold(true)

	DiffGoodStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00A86B")).
			Bold(true)

	DiffEasyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#5B9BD5")).
			Bold(true)

	// Badge style for small labels
	BadgeStyle = lipgloss.NewStyle().
			Padding(0, 1).
			Bold(true)

	// Progress bar components
	ProgressFilledStyle = lipgloss.NewStyle().
				Foreground(ColorSuccess)

	ProgressEmptyStyle = lipgloss.NewStyle().
				Foreground(ColorMuted)

	// Input mode badge styles
	InputModeNormalStyle = lipgloss.NewStyle().
				Padding(0, 1).
				Bold(true).
				Foreground(ColorText).
				Background(lipgloss.Color("#2A2A3E"))

	InputModePinyinStyle = lipgloss.NewStyle().
				Padding(0, 1).
				Bold(true).
				Foreground(lipgloss.Color("#FFFFFF")).
				Background(ColorAccent)

	InputModeChineseStyle = lipgloss.NewStyle().
				Padding(0, 1).
				Bold(true).
				Foreground(lipgloss.Color("#FFFFFF")).
				Background(ColorPrimary)
)

// ProgressBar renders a text-based progress bar.
// width is the total character width, current/total define progress.
func ProgressBar(current, total, width int) string {
	if total == 0 {
		return strings.Repeat("░", width)
	}
	filled := width * current / total
	if filled > width {
		filled = width
	}
	empty := width - filled
	return ProgressFilledStyle.Render(strings.Repeat("█", filled)) +
		ProgressEmptyStyle.Render(strings.Repeat("░", empty))
}

// Header renders a decorative Chinese-style header with box-drawing borders.
func Header(title string) string {
	titleLen := len([]rune(title))
	padTotal := 30
	if titleLen+4 > padTotal {
		padTotal = titleLen + 4
	}
	side := (padTotal - titleLen - 2) / 2
	leftPad := strings.Repeat("═", side)
	rightPad := strings.Repeat("═", padTotal-titleLen-2-side)

	top := MutedStyle.Render("╔" + strings.Repeat("═", padTotal) + "╗")
	mid := MutedStyle.Render("║") +
		MutedStyle.Render(leftPad) + " " +
		TitleStyle.Render(title) + " " +
		MutedStyle.Render(rightPad) +
		MutedStyle.Render("║")
	bot := MutedStyle.Render("╚" + strings.Repeat("═", padTotal) + "╝")

	return top + "\n" + mid + "\n" + bot
}

// Divider renders a decorative section divider.
func Divider(width int) string {
	if width <= 0 {
		width = 30
	}
	return MutedStyle.Render("  " + strings.Repeat("─", width))
}

// WrapText wraps text to fit within maxWidth characters.
func WrapText(text string, maxWidth int) string {
	if maxWidth <= 0 {
		return text
	}
	words := strings.Fields(text)
	if len(words) == 0 {
		return ""
	}
	var lines []string
	line := words[0]
	for _, w := range words[1:] {
		if len(line)+1+len(w) > maxWidth {
			lines = append(lines, line)
			line = w
		} else {
			line += " " + w
		}
	}
	lines = append(lines, line)
	return strings.Join(lines, "\n")
}
