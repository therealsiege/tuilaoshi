package wiki

import (
	"regexp"
	"strings"
	"time"

	"github.com/cj/mandarin-wiki/internal/model"
	"gopkg.in/yaml.v3"
)

// frontmatter is the raw YAML deserialization target.
type frontmatter struct {
	Type           string   `yaml:"type"`
	Level          string   `yaml:"level"`
	Status         string   `yaml:"status"`
	LastReviewed   string   `yaml:"last_reviewed"`
	Tags           []string `yaml:"tags"`
	Introduced     string   `yaml:"introduced"`
	ReviewInterval int      `yaml:"review_interval"`
	NextReview     string   `yaml:"next_review"`
}

const dateFmt = "2006-01-02"

// ParsePage parses a wiki markdown file into a Page.
// relPath is the path relative to the wiki directory (e.g. "characters/好.md").
func ParsePage(relPath string, data []byte) (model.Page, error) {
	fm, body := splitFrontmatter(data)

	p := model.Page{
		Path:  relPath,
		Body:  body,
		Title: extractTitle(body),
	}

	if fm != nil {
		var f frontmatter
		if err := yaml.Unmarshal(fm, &f); err != nil {
			return p, err
		}
		p.Type = f.Type
		p.Level = f.Level
		p.Status = f.Status
		p.Tags = f.Tags
		p.ReviewInterval = f.ReviewInterval
		p.LastReviewed = parseDate(f.LastReviewed)
		p.Introduced = parseDate(f.Introduced)
		p.NextReview = parseDate(f.NextReview)
	}

	if p.Type == "character" {
		parseCharacterFields(&p)
	}

	return p, nil
}

// splitFrontmatter separates YAML frontmatter (between --- delimiters) from body.
func splitFrontmatter(data []byte) (fm []byte, body string) {
	s := string(data)
	if !strings.HasPrefix(s, "---") {
		return nil, s
	}
	// Find the closing ---
	rest := s[3:]
	idx := strings.Index(rest, "\n---")
	if idx < 0 {
		return nil, s
	}
	fm = []byte(strings.TrimSpace(rest[:idx]))
	body = strings.TrimLeft(rest[idx+4:], "\n")
	return fm, body
}

// extractTitle returns the text of the first H1 heading.
func extractTitle(body string) string {
	for _, line := range strings.Split(body, "\n") {
		if strings.HasPrefix(line, "# ") {
			return strings.TrimPrefix(line, "# ")
		}
	}
	return ""
}

var pinyinRe = regexp.MustCompile(`\(([^)]+)\)`)

// parseCharacterFields extracts character-specific data from the body.
func parseCharacterFields(p *model.Page) {
	lines := strings.Split(p.Body, "\n")
	for _, line := range lines {
		lower := strings.ToLower(strings.TrimSpace(line))
		switch {
		case strings.HasPrefix(lower, "- **character"):
			p.Character = extractFieldValue(line)
		case strings.HasPrefix(lower, "- **pinyin"):
			p.Pinyin = extractFieldValue(line)
		case strings.HasPrefix(lower, "- **meaning"):
			p.Meaning = extractFieldValue(line)
		case strings.HasPrefix(lower, "- **radical"):
			p.Radical = extractFieldValue(line)
		}
	}

	// Extract compounds from a ## Compounds or ## Common Compounds section
	inCompounds := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "## ") && strings.Contains(strings.ToLower(trimmed), "compound") {
			inCompounds = true
			continue
		}
		if strings.HasPrefix(trimmed, "## ") {
			inCompounds = false
			continue
		}
		if inCompounds && strings.HasPrefix(trimmed, "- ") {
			p.Compounds = append(p.Compounds, strings.TrimPrefix(trimmed, "- "))
		}
	}

	// Fallback: use title as character if not found in body fields
	if p.Character == "" && p.Title != "" {
		parts := strings.Fields(p.Title)
		if len(parts) > 0 {
			p.Character = parts[0]
		}
	}
}

// extractFieldValue extracts the value after **: in a "- **Key**: value" line.
func extractFieldValue(line string) string {
	idx := strings.Index(line, "**:")
	if idx < 0 {
		return ""
	}
	return strings.TrimSpace(line[idx+3:])
}

// parseDate parses a YYYY-MM-DD string; returns zero time on failure.
func parseDate(s string) time.Time {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}
	}
	t, _ := time.Parse(dateFmt, s)
	return t
}

// FormatDate formats a time as YYYY-MM-DD, or empty string if zero.
func FormatDate(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(dateFmt)
}
