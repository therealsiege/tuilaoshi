package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Colors — a warm palette for a Mandarin tutor.
var (
	ColorPrimary   = lipgloss.Color("#E06C75") // red (Chinese aesthetic)
	ColorSecondary = lipgloss.Color("#61AFEF") // blue
	ColorAccent    = lipgloss.Color("#C678DD") // purple
	ColorSuccess   = lipgloss.Color("#98C379") // green
	ColorWarning   = lipgloss.Color("#E5C07B") // yellow
	ColorMuted     = lipgloss.Color("#5C6370") // gray
	ColorText      = lipgloss.Color("#ABB2BF") // light gray
	ColorBg        = lipgloss.Color("#282C34") // dark background
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
			Foreground(lipgloss.Color("#E06C75"))

	BorderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorMuted).
			Padding(1, 2)

	ActiveBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(ColorPrimary).
				Padding(1, 2)

	HelpStyle = lipgloss.NewStyle().
			Foreground(ColorMuted).
			MarginTop(1)

	ChineseStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorPrimary)

	PinyinStyle = lipgloss.NewStyle().
			Foreground(ColorAccent).
			Italic(true)

	CardStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorSecondary).
			Padding(1, 3).
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
			Background(lipgloss.Color("#3E4451")).
			Padding(0, 1)

	// Chat bubble styles
	ChatBubbleUserStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("#2B4570")).
				Foreground(lipgloss.Color("#FFFFFF")).
				Padding(0, 2).
				MarginLeft(4)

	ChatBubbleAssistStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("#3E4451")).
				Foreground(ColorText).
				Padding(0, 2).
				MarginRight(4)

	ChatTimestampStyle = lipgloss.NewStyle().
				Foreground(ColorMuted).
				Italic(true)

	// Difficulty styles
	DiffAgainStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#E06C75")).
			Bold(true)

	DiffHardStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#D19A66")).
			Bold(true)

	DiffGoodStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#98C379")).
			Bold(true)

	DiffEasyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#61AFEF")).
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

	// Imagen styles
	ImagenPromptStyle = lipgloss.NewStyle().
				Foreground(ColorAccent).
				Bold(true)

	ImagenPathStyle = lipgloss.NewStyle().
			Foreground(ColorSuccess)

	// Input mode badge styles
	InputModeNormalStyle = lipgloss.NewStyle().
				Padding(0, 1).
				Bold(true).
				Foreground(ColorText).
				Background(lipgloss.Color("#3E4451"))

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
