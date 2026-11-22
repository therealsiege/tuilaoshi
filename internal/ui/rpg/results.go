package rpg

import (
	"time"

	"github.com/therealsiege/tuilaoshi/internal/model"
	rpgengine "github.com/therealsiege/tuilaoshi/internal/rpg"
	"github.com/therealsiege/tuilaoshi/internal/wiki"
)

// A battle's results are recorded as it is fought: each answer reschedules
// its card, and the whole fight is saved as a session when it ends.

// startSession begins recording results for a new battle.
func (m Model) startSession() Model {
	m.session = model.Session{StartedAt: time.Now(), Type: "battle"}
	m.reviewed = map[string]bool{}
	m.saved = false
	m.saveErr = nil
	m.earned, m.bonded = wiki.Wallet{}, nil
	return m
}

// recordAnswer logs an answer and reschedules its card: a miss makes it due
// again tomorrow, and a hit counts once per battle as a "Hard" review.
func (m *Model) recordAnswer(q *rpgengine.CombatQuestion, answer string, correct bool) {
	diff := model.DiffAgain
	if correct {
		diff = model.DiffHard
	}
	m.session.Items = append(m.session.Items, model.SessionItem{
		CardID:     q.Card.ID,
		UserAnswer: answer,
		Correct:    correct,
		Difficulty: diff,
	})

	if m.store == nil || (correct && m.reviewed[q.Card.ID]) {
		return
	}
	card := q.Card
	if err := m.store.RecordReview(&card, diff, time.Now()); err != nil {
		m.saveErr = err
	}
	m.reviewed[q.Card.ID] = true
}

// saveSession writes the battle to the wiki's session log once, so it counts
// toward stats, streaks, and lifetime XP.
func (m *Model) saveSession() {
	if m.saved || m.store == nil || m.battle == nil || len(m.session.Items) == 0 {
		return
	}
	m.session.EndedAt = time.Now()
	m.session.CardsTotal = m.battle.Answered
	m.session.CardsCorrect = m.battle.CorrectHits
	m.session.XP = m.battle.XPEarned
	before, err := m.store.Wallet()
	if err != nil {
		m.saveErr = err
		return
	}
	if err := m.store.WriteSession(m.session); err != nil {
		m.saveErr = err
		return
	}
	m.payWin(before)
	if m.battle.PlayerWon() {
		if err := m.store.RecordDefeat(m.battle.Enemy.Chinese); err != nil {
			m.saveErr = err
			return
		}
	}
	if m.stage != nil && m.lastFoe() && m.battle.PlayerWon() {
		if err := m.store.RecordStageCleared(m.stage.Number, time.Now()); err != nil {
			m.saveErr = err
			return
		}
	}
	m.saved = true
	m.totalXP = m.store.TotalXP()
}
