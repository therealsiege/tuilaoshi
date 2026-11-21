package reader

import "strings"

// textSection returns the lines under a page's "## Text" heading, without
// blank lines.
func textSection(body string) []string {
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
