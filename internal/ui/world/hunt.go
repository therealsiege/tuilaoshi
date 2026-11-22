package world

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/therealsiege/tuilaoshi/internal/journey"
	"github.com/therealsiege/tuilaoshi/internal/ui"
	"github.com/therealsiege/tuilaoshi/internal/wiki"
)

// A hunt quest sends the pilgrim after one of the region's soldiers: accept
// it here, beat the soldier on the road (wild ground or a gauntlet), then
// come back to collect.

// huntState is where a hunt stands: not taken, under way, ready to turn in.
func (m Model) huntState(q journey.SideQuest) (taken bool, beaten int) {
	n, ok := m.hunts[q.ID]
	return ok, n
}

// updateHunt accepts the hunt, or pays it once the soldiers are beaten.
func (m Model) updateHunt(k tea.KeyMsg) (Model, tea.Cmd) {
	if k.Type != tea.KeyEnter && k.Type != tea.KeySpace {
		return m, nil
	}
	q := m.towns[m.town].Quests[m.quest]
	taken, beaten := m.huntState(q)
	switch {
	case !taken:
		if err := m.store.AcceptHunt(q.ID); err != nil {
			m.loadErr = err
			return m, nil
		}
		m.note = fmt.Sprintf("Accepted. Look for %s on the road's wild ground.", q.Target)
	case beaten >= q.Count:
		if err := m.store.FinishQuest(q.ID, wiki.Wallet{Tongqian: q.Pay(1)}); err != nil {
			m.loadErr = err
			return m, nil
		}
		m.note = fmt.Sprintf("%s thanks you and pays %d 铜钱.", q.Giver, q.Pay(1))
	default:
		m.screen = screenTown
		return m.Refresh(), nil
	}
	m.screen = screenTown
	note := m.note
	m = m.Refresh()
	m.note = note
	return m, nil
}

func (m Model) huntView() string {
	q := m.towns[m.town].Quests[m.quest]
	taken, beaten := m.huntState(q)
	var b strings.Builder
	b.WriteString(ui.TitleStyle.MarginBottom(0).Render(q.Title) + "  " + ui.MutedStyle.Render(q.Giver+" · "+q.GiverEnglish) + "\n\n")
	b.WriteString("  " + ui.ChineseStyle.Render(ui.WrapText(q.Text, 36)) + "\n\n")
	b.WriteString(ui.MutedStyle.Render("  "+ui.WrapText(q.English, 72)) + "\n\n")
	target := q.Target
	if e := journey.EnemyLookup(q.Target); e != nil {
		target += " " + e.English
	}
	fmt.Fprintf(&b, "  Hunt: %s × %d · reward %d 铜钱\n\n", target, q.Count, q.Pay(1))
	help := "  enter accept · esc back"
	switch {
	case taken && beaten >= q.Count:
		b.WriteString(ui.SuccessStyle.Render("  Done! Tell them so.") + "\n")
		help = "  enter collect your reward · esc back"
	case taken:
		b.WriteString(ui.WarningStyle.Render(fmt.Sprintf("  Under way: %d of %d beaten.", beaten, q.Count)) + "\n")
		help = "  enter back to the notice board"
	}
	b.WriteString("\n" + ui.HelpStyle.Render(help))
	return b.String()
}
