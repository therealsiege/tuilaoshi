package wiki

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cj/mandarin-wiki/internal/model"
	"gopkg.in/yaml.v3"
)

// wordScheduleFile holds review schedules for individual vocabulary words.
// Words on a vocabulary page share its frontmatter, so each word's own
// schedule lives here, keyed by card ID (e.g. "vocabulary/food.md#米饭").
const wordScheduleFile = "word-schedule.yaml"

// wordSchedule is one word's entry in the schedule file.
type wordSchedule struct {
	Introduced     string `yaml:"introduced,omitempty"`
	LastReviewed   string `yaml:"last_reviewed,omitempty"`
	ReviewInterval int    `yaml:"review_interval"`
	NextReview     string `yaml:"next_review,omitempty"`
}

// isWordCard reports whether a card is one row of a vocabulary table rather
// than a whole page.
func isWordCard(id string) bool {
	return strings.Contains(id, "#")
}

// readWordSchedules loads the word schedule file. A missing file means no
// word has been reviewed on its own yet.
func (s *Store) readWordSchedules() (map[string]wordSchedule, error) {
	schedules := map[string]wordSchedule{}
	data, err := os.ReadFile(filepath.Join(s.Dir, wordScheduleFile))
	if os.IsNotExist(err) {
		return schedules, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read word schedules: %w", err)
	}
	if err := yaml.Unmarshal(data, &schedules); err != nil {
		return nil, fmt.Errorf("parse %s: %w", wordScheduleFile, err)
	}
	return schedules, nil
}

func (s *Store) writeWordSchedules(schedules map[string]wordSchedule) error {
	data, err := yaml.Marshal(schedules)
	if err != nil {
		return fmt.Errorf("encode word schedules: %w", err)
	}
	return os.WriteFile(filepath.Join(s.Dir, wordScheduleFile), data, 0644)
}

// DrillableCards returns flashcards for every character, vocabulary, and
// grammar page. Vocabulary words reviewed on their own carry their own
// schedule; the rest inherit their page's.
func (s *Store) DrillableCards() ([]model.Card, error) {
	pages, err := s.ListPages()
	if err != nil {
		return nil, err
	}

	var drillable []model.Page
	for _, p := range pages {
		switch p.Type {
		case "character", "vocabulary", "grammar":
			drillable = append(drillable, p)
		}
	}
	cards := CardsFromPages(drillable)

	schedules, err := s.readWordSchedules()
	if err != nil {
		return nil, err
	}
	for i := range cards {
		sch, ok := schedules[cards[i].ID]
		if !ok {
			continue
		}
		cards[i].Introduced = parseDate(sch.Introduced)
		cards[i].LastReviewed = parseDate(sch.LastReviewed)
		cards[i].ReviewInterval = sch.ReviewInterval
		cards[i].NextReview = parseDate(sch.NextReview)
	}
	return cards, nil
}

// RecordReview reschedules a card after a review and saves the result:
// vocabulary words to the word schedule file, other cards to their page.
func (s *Store) RecordReview(card *model.Card, difficulty model.DrillDifficulty, now time.Time) error {
	updateSRS(card, difficulty, now)

	if isWordCard(card.ID) {
		schedules, err := s.readWordSchedules()
		if err != nil {
			return err
		}
		schedules[card.ID] = wordSchedule{
			Introduced:     FormatDate(card.Introduced),
			LastReviewed:   FormatDate(card.LastReviewed),
			ReviewInterval: card.ReviewInterval,
			NextReview:     FormatDate(card.NextReview),
		}
		return s.writeWordSchedules(schedules)
	}

	page, err := s.ReadPage(card.ID)
	if err != nil {
		return err
	}
	applySRSToPage(&page, *card)
	return s.writeFrontmatter(page)
}
