package roam

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/therealsiege/tuilaoshi/internal/road"
	"github.com/therealsiege/tuilaoshi/internal/ui"
)

// The trail on the map, by the road package's rules (its trail.go): each
// step spends part of a day at the chosen pace, days eat and drink and
// move the party's health, p changes the pace, f the rations and r rests
// a day, a town refills the water, and a collapsed master is carried back
// to the last town.

// termToday is today's solar term; tests fix it.
var termToday = road.Today

// termWeather is today's solar term weather: cold, heat, rain or "".
func termWeather() string { return termToday().Weather }

// walked counts a step on the trail (see the road package for the
// rules), passing a day when the walking is done: the day may carry the
// master back, or bring an event or demons plotting.
func (m Model) walked() Model {
	w := m.walk()
	d, err := w.Walked(termWeather(), eventRoll, plotRoll)
	m.trail, m.pos = w.Trail, w.Pos
	if err != nil {
		m.loadErr = err
	}
	if d.Ended {
		m.note = d.Note
	}
	if d.Event != nil {
		m.event = &trailEvent{Event: *d.Event, chose: -1}
	}
	if d.Plot != nil {
		m.plot = &overheard{Plot: *d.Plot, order: d.Order, chose: -1, heard: ui.Audible(m.tts)}
	}
	return m
}

// trailKey changes the pace (p) or the rations (f), or rests a day (w).
func (m Model) trailKey(k tea.KeyMsg) (Model, bool) {
	w := m.walk()
	var err error
	switch k.String() {
	case "p":
		m.note, err = w.NextPace()
	case "f":
		m.note, err = w.NextRations()
	case "r", "w": // r, as at the inn; w kept for habit
		m.note, err = w.Rest(termWeather())
	default:
		return m, false
	}
	m.trail = w.Trail
	if err != nil {
		m.loadErr = err
	}
	return m, true
}

// trailView shows the day, the supplies and each member's health and spirits.
func (m Model) trailView() string {
	t := m.trail
	var b strings.Builder
	food := fmt.Sprintf("斋饭 %d", t.Food)
	if m.walk().FoodLow() {
		food = ui.ErrorStyle.Render(food)
	}
	water := fmt.Sprintf("水 %d", t.Water)
	if m.walk().WaterLow() {
		water = ui.ErrorStyle.Render(water)
	}
	fmt.Fprintf(&b, "%s · %s · %s\n", ui.SubtitleStyle.Render(fmt.Sprintf("Day %d", t.Day)), food, water)
	b.WriteString(ui.MutedStyle.Render(fmt.Sprintf("%s pace (p) · %s rations (f)", t.Pace, t.Rations)) + "\n")
	if term := termToday(); term.Name != "" && term.Weather != "" {
		b.WriteString(ui.MutedStyle.Render(fmt.Sprintf("节气 %s %s: %s", term.Name, term.English, road.WeatherShort[term.Weather])) + "\n")
	}
	if t.SandalsWorn() {
		b.WriteString(ui.ErrorStyle.Render("草鞋 worn out: buy new in a town") + "\n")
	} else {
		b.WriteString(ui.MutedStyle.Render(fmt.Sprintf("草鞋 good for %d more days", m.walk().SandalDaysLeft())) + "\n")
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
