package wiki

import "strings"

// ReadingText returns the lines under a reading page's "## Text" heading,
// without blank lines.
func ReadingText(body string) []string {
	var lines []string
	in := false
	for _, l := range strings.Split(body, "\n") {
		if strings.HasPrefix(l, "## ") {
			in = strings.TrimSpace(l) == "## Text"
			continue
		}
		if in && strings.TrimSpace(l) != "" {
			lines = append(lines, strings.TrimSpace(l))
		}
	}
	return lines
}
