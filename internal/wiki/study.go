package wiki

// studyFile keeps the day's study bookkeeping, shared by every journey:
// which words have already paid today, the daily bonus, bounty progress
// and streak shields. Study belongs to the learner, not to one journey.
const studyFile = "study.yaml"

type studyRecord struct {
	DailyPaid    string         `yaml:"daily_paid,omitempty"`    // the day the daily bonus was last paid
	PaidDay      string         `yaml:"paid_day,omitempty"`      // the day PaidCards covers
	PaidCards    []string       `yaml:"paid_cards,omitempty"`    // cards already paid for that day
	TallyDay     string         `yaml:"tally_day,omitempty"`     // the day Tally counts
	Tally        map[string]int `yaml:"tally,omitempty"`         // fresh progress toward today's bounties
	BountiesPaid []string       `yaml:"bounties_paid,omitempty"` // today's bounties already paid
	Shields      int            `yaml:"shields,omitempty"`       // streak shields held
	Shielded     []string       `yaml:"shielded,omitempty"`      // missed days a shield covered
	FreqDay      string         `yaml:"freq_day,omitempty"`      // the day frequency words were last introduced
}

func (s *Store) readStudy() (studyRecord, error) {
	var rec studyRecord
	err := s.readYAML(studyFile, &rec)
	return rec, err
}

func (s *Store) writeStudy(rec studyRecord) error { return s.writeYAML(studyFile, rec) }
