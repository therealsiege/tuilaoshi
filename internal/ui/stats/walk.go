package stats

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/therealsiege/tuilaoshi/internal/model"
	"github.com/therealsiege/tuilaoshi/internal/ui"
)

// Walking the web: a quick self-rated review of the slipping words tied to
// the centre, saved through RecordReview like any review and kept as a
// review session, so it counts for the day.

// walk is a quick review of one cluster of the web.
type walk struct {
	centre  string       // the word walked from
	cards   []model.Card // the slipping words, the centre first
	i       int          // the card being asked
	shown   bool         // its answer is showing
	done    bool         // every card is rated; the summary is showing
	started time.Time
	items   []model.SessionItem
	right   int
}

// active reports whether a walk is under way.
func (w walk) active() bool { return len(w.cards) > 0 }

// startWalk begins reviving the slipping words around the centre.
func (m Model) startWalk() Model {
	cards := m.spider.g.Cluster(m.spider.centre)
	if len(cards) == 0 {
		m.spider.note = fmt.Sprintf("Nothing slipping around %s: every strand holds.", m.centreWord())
		return m
	}
	m.spider.walk = walk{centre: m.centreWord(), cards: cards, started: m.now()}
	return m
}

// updateWalk shows each word, then saves the learner's rating of it.
func (m Model) updateWalk(msg tea.KeyMsg) (Model, tea.Cmd) {
	w := &m.spider.walk
	if w.done {
		return m.endWalk(), nil
	}
	key := msg.String()
	if key == "esc" {
		return m.saveWalk().endWalk(), nil
	}
	if !w.shown {
		if key == "enter" || key == " " {
			w.shown = true
		}
		return m, nil
	}
	diff, ok := map[string]model.DrillDifficulty{
		"1": model.DiffAgain, "2": model.DiffHard, "3": model.DiffGood, "4": model.DiffEasy,
		"enter": model.DiffGood, " ": model.DiffGood,
	}[key]
	if !ok {
		return m, nil
	}
	card := &w.cards[w.i]
	if err := m.store.RecordReview(card, diff, m.now()); err != nil && m.spider.err == nil {
		m.spider.err = err
	}
	w.items = append(w.items, model.SessionItem{CardID: card.ID, Correct: diff != model.DiffAgain, Difficulty: diff})
	if diff != model.DiffAgain {
		w.right++
	}
	w.i++
	w.shown = false
	if w.i >= len(w.cards) {
		m = m.saveWalk()
		m.spider.walk.done = true
	}
	return m, nil
}

// saveWalk keeps the answers given so far as a review session.
func (m Model) saveWalk() Model {
	w := m.spider.walk
	if len(w.items) == 0 {
		return m
	}
	err := m.store.WriteSession(model.Session{
		StartedAt: w.started, EndedAt: m.now(), Type: "review",
		CardsTotal: len(w.items), CardsCorrect: w.right, Items: w.items,
	})
	if err != nil && m.spider.err == nil {
		m.spider.err = err
	}
	return m
}

// endWalk goes back to the web, redrawn with the new schedules.
func (m Model) endWalk() Model {
	err := m.spider.err
	m.spider.walk = walk{}
	m = m.Refresh()
	if err != nil {
		m.spider.err = err
	}
	return m
}

// walkView asks one word of the cluster, or sums up the walk.
func (m Model) walkView() string {
	w := m.spider.walk
	width, _ := m.size()
	var b strings.Builder
	title := ui.TitleStyle.MarginBottom(0).Render("盘丝洞 Walking the web from " + w.centre)
	b.WriteString(fit(title+ui.MutedStyle.Render(fmt.Sprintf("  %d of %d", min(w.i+1, len(w.cards)), len(w.cards))), width) + "\n\n")
	if m.spider.err != nil {
		b.WriteString(fit(ui.ErrorStyle.Render("  "+m.spider.err.Error()), width) + "\n\n")
	}
	if w.done {
		fmt.Fprintf(&b, "  Revived %d of %d slipping words around %s.\n", w.right, len(w.cards), w.centre)
		if missed := len(w.cards) - w.right; missed > 0 {
			b.WriteString(ui.WarningStyle.Render(fmt.Sprintf("  %d still slipping: they come back tomorrow.", missed)) + "\n")
		} else {
			b.WriteString(ui.SuccessStyle.Render("  The whole cluster holds again.") + "\n")
		}
		b.WriteString(ui.HelpStyle.Render("  any key: back to the web"))
		return b.String()
	}
	card := w.cards[w.i]
	b.WriteString("    " + ui.ChineseStyle.Render(card.Chinese) + "\n\n")
	if !w.shown {
		b.WriteString(ui.MutedStyle.Render("  Say its sound and meaning to yourself, then look.") + "\n")
		b.WriteString(ui.HelpStyle.Render(fit("  space show · esc back to the web", width)))
		return b.String()
	}
	b.WriteString(fit("    "+ui.Pinyin(card.Pinyin)+"  "+card.English, width) + "\n")
	if len(card.Examples) > 0 {
		b.WriteString(fit(ui.MutedStyle.Render("    "+card.Examples[0]), width) + "\n")
	}
	b.WriteString("\n  " + ui.DiffAgainStyle.Render("[1] Again") + "  " + ui.DiffHardStyle.Render("[2] Hard") + "  " +
		ui.DiffGoodStyle.Render("[3] Good") + "  " + ui.DiffEasyStyle.Render("[4] Easy") + "\n")
	b.WriteString(ui.HelpStyle.Render(fit("  1-4 rate · enter Good · esc back to the web", width)))
	return b.String()
}
