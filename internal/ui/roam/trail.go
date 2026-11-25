package roam

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/therealsiege/tuilaoshi/internal/journey"
	"github.com/therealsiege/tuilaoshi/internal/ui"
)

// The trail on the map: each step spends part of a day at the chosen pace,
// days eat and drink and move the party's health, p changes the pace and f
// the rations, a town refills the water, and a collapsed master is carried
// back to the last town.

// loadTrail reads the journey's trail, starting one for a new journey.
func (m *Model) loadTrail() {
	t, members, err := journey.LoadTrail(m.store, m.cleared)
	if err != nil {
		m.loadErr = err
	}
	m.trail, m.members = t, members
}

func (m *Model) saveTrail() {
	if err := m.store.SaveTrail(m.trail); err != nil {
		m.loadErr = err
	}
}

// weatherShort says what a term's weather does on the road.
var weatherShort = map[string]string{"cold": "more 斋饭 a day", "heat": "more water a day", "rain": "rain fills the gourds"}

// termToday is today's solar term; tests fix it.
var termToday = func() journey.SolarTerm {
	t, _ := journey.TermOn(time.Now())
	return t
}

// termWeather is today's solar term weather: cold, heat, rain or "".
func termWeather() string { return termToday().Weather }

// walked counts a step on the trail, passing a day when the walking is done.
func (m Model) walked() Model {
	m.trail.Weather = termWeather()
	if m.region.At(m.pos) == journey.TileTown {
		m.trail.Water = journey.MaxWater // wells in every town
	}
	if m.trail.Step(m.members, m.region.Region.Terrain) {
		m.note = strings.Join(append([]string{fmt.Sprintf("Day %d.", m.trail.Day)}, m.trail.Notes...), " ")
		if m.trail.Collapsed() {
			m = m.carryBack()
		} else {
			m = m.maybeEvent().maybePlot()
		}
	}
	m.saveTrail()
	return m
}

// carryBack takes the party to the last town passed (or the region's
// start) after the master collapses; he wakes weak but alive.
func (m Model) carryBack() Model {
	wake := m.region.Start
	for p := range m.region.Towns {
		if p.X >= m.pos.X && (wake == m.region.Start || p.X < wake.X) {
			wake = p
		}
	}
	m.pos = wake
	m.trail.Health[journey.Master] = 30
	m.note += " 唐僧 collapses; the disciples carry him back to rest. Rest at an inn, eat well, and slow the pace."
	m.save()
	return m
}

// trailKey changes the pace (p) or the rations (f), or rests a day (w).
func (m Model) trailKey(k tea.KeyMsg) (Model, bool) {
	switch k.String() {
	case "p":
		m.trail.Pace = journey.Next(journey.Paces, m.trail.Pace)
		m.note = "Pace: " + m.trail.Pace + paceNote[m.trail.Pace]
	case "f":
		m.trail.Rations = journey.Next(journey.Rations, m.trail.Rations)
		m.note = "Rations: " + m.trail.Rations + rationNote[m.trail.Rations]
	case "r", "w": // r, as at the inn; w kept for habit
		m.trail.Weather = termWeather()
		m.trail.PassDay(m.members, m.region.Region.Terrain, true)
		m.note = strings.Join(append([]string{fmt.Sprintf("Day %d. The party rests.", m.trail.Day)}, m.trail.Notes...), " ")
	default:
		return m, false
	}
	m.saveTrail()
	return m, true
}

var (
	paceNote   = map[string]string{"steady": " (6 steps a day; rests the party)", "brisk": " (9 steps a day; tiring)", "forced": " (12 steps a day; wearing the party down)"}
	rationNote = map[string]string{"filling": " (3 斋饭 each a day; heals)", "meager": " (2 each; holds steady)", "bare": " (1 each; weakens)"}
)

// trailView shows the day, the supplies and each member's health and spirits.
func (m Model) trailView() string {
	t := m.trail
	var b strings.Builder
	food := fmt.Sprintf("斋饭 %d", t.Food)
	if t.Food < 3*len(m.members)*2 {
		food = ui.ErrorStyle.Render(food)
	}
	water := fmt.Sprintf("水 %d", t.Water)
	if t.Water < len(m.members)*2 {
		water = ui.ErrorStyle.Render(water)
	}
	fmt.Fprintf(&b, "%s · %s · %s\n", ui.SubtitleStyle.Render(fmt.Sprintf("Day %d", t.Day)), food, water)
	b.WriteString(ui.MutedStyle.Render(fmt.Sprintf("%s pace (p) · %s rations (f)", t.Pace, t.Rations)) + "\n")
	if term := termToday(); term.Name != "" && term.Weather != "" {
		b.WriteString(ui.MutedStyle.Render(fmt.Sprintf("节气 %s %s: %s", term.Name, term.English, weatherShort[term.Weather])) + "\n")
	}
	if t.SandalsWorn() {
		b.WriteString(ui.ErrorStyle.Render("草鞋 worn out: buy new in a town") + "\n")
	} else {
		b.WriteString(ui.MutedStyle.Render(fmt.Sprintf("草鞋 good for %d more days", journey.SandalLife-t.SandalDays)) + "\n")
	}
	if t.Forewarned {
		b.WriteString(ui.SuccessStyle.Render("Forewarned: the next danger does half its harm") + "\n")
	}
	for _, name := range m.members {
		h, mo := t.Health[name], t.Morale[name]
		style := ui.SuccessStyle
		switch {
		case h == 0:
			style = ui.ErrorStyle
		case h < 40:
			style = ui.WarningStyle
		}
		status := fmt.Sprintf("♥%d ☺%d", h, mo)
		if h == 0 {
			status += " ill"
		}
		fmt.Fprintf(&b, "%s %s\n", name, style.Render(status))
	}
	return b.String()
}
