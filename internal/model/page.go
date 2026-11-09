package model

import "time"

// Page represents a parsed wiki markdown page.
type Page struct {
	// File path relative to wiki/ directory
	Path string

	// Frontmatter fields
	Type           string    // grammar, vocabulary, character, reading, session, reference
	Level          string    // HSK1-HSK6, advanced
	Status         string    // draft, active, review-needed
	LastReviewed   time.Time
	Tags           []string
	Introduced     time.Time
	ReviewInterval int // days
	NextReview     time.Time

	// Content
	Title string // first H1 heading
	Body  string // raw markdown body (without frontmatter)

	// Character-specific (populated for character pages)
	Character string
	Pinyin    string
	Meaning   string
	Radical   string
	Compounds []string
}

// IsCharacter reports whether this page is a character analysis page.
func (p Page) IsCharacter() bool { return p.Type == "character" }

// IsVocabulary reports whether this page is a vocabulary page.
func (p Page) IsVocabulary() bool { return p.Type == "vocabulary" }

// IsGrammar reports whether this page is a grammar page.
func (p Page) IsGrammar() bool { return p.Type == "grammar" }
