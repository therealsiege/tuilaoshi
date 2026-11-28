package road

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/therealsiege/tuilaoshi/internal/journey"
	"github.com/therealsiege/tuilaoshi/internal/rpg"
	"github.com/therealsiege/tuilaoshi/internal/wiki"
)

// The trail on the road: each step spends part of a day at the chosen
// pace, days eat and drink and move the party's health, the pace and the
// rations can be changed and a day rested, a town refills the water, and
// a collapsed master is carried back to the last town. A day's end may
// bring an event (events.go) or, failing that, demons overheard plotting
// (plots.go).

// WeatherShort says what a term's weather does on the road.
var WeatherShort = map[string]string{"cold": "more 斋饭 a day", "heat": "more water a day", "rain": "rain fills the gourds"}

// PaceNote and RationNote say what each pace and ration does, after its name.
var (
	PaceNote   = map[string]string{"steady": " (6 steps a day; rests the party)", "brisk": " (9 steps a day; tiring)", "forced": " (12 steps a day; wearing the party down)"}
	RationNote = map[string]string{"filling": " (3 斋饭 each a day; heals)", "meager": " (2 each; holds steady)", "bare": " (1 each; weakens)"}
)

// Today is today's solar term.
func Today() journey.SolarTerm {
	t, _ := journey.TermOn(time.Now())
	return t
}

// Day is what a step on the trail came to.
type Day struct {
	Note      string         // "Day 4." and what the day brought; "" until a day ends
	Ended     bool           // the day's walking is done
	Collapsed bool           // the master collapsed and was carried back
	Event     *journey.Event // the day's event, met before anything else
	Plot      *journey.Plot  // or demons overheard plotting, in Order
	Order     []int          // the plot's choices as shown: indices into Choices, 0 the right one
}

// Walked counts a step on the trail at today's weather (cold, heat, rain
// or ""), passing a day when the walking is done: a collapsed master is
// carried back, and otherwise the day may bring an event or a plot
// (eventRoll and plotRoll draw 0-99 against their chances; nil rolls at
// random). The trail is saved; the error is the saves'.
func (w *Walk) Walked(weather string, eventRoll, plotRoll func() int) (Day, error) {
	var err error
	w.Trail.Weather = weather
	if w.Region.At(w.Pos) == journey.TileTown {
		w.Trail.Water = journey.MaxWater // wells in every town
	}
	var d Day
	if w.Trail.Step(w.Members, w.Region.Region.Terrain) {
		d.Ended = true
		d.Note = strings.Join(append([]string{fmt.Sprintf("Day %d.", w.Trail.Day)}, w.Trail.Notes...), " ")
		if w.Trail.Collapsed() {
			d.Collapsed = true
			d.Note += w.carryBack()
			err = w.Save()
		} else {
			d.Event = w.pickEvent(eventRoll)
			if d.Event == nil {
				d.Plot, d.Order = pickPlot(w.Trail.Forewarned, plotRoll)
			}
		}
	}
	if serr := w.SaveTrail(); err == nil {
		err = serr
	}
	return d, err
}

// carryBack takes the party to the last town passed (or the region's
// start) after the master collapses; he wakes weak but alive. It returns
// what to say of it.
func (w *Walk) carryBack() string {
	wake := w.Region.Start
	for p := range w.Region.Towns {
		if p.X >= w.Pos.X && (wake == w.Region.Start || p.X < wake.X) {
			wake = p
		}
	}
	w.Pos = wake
	w.Trail.Health[journey.Master] = 30
	return " 唐僧 collapses; the disciples carry him back to rest. Rest at an inn, eat well, and slow the pace."
}

// NextPace changes to the next pace and says what it does.
func (w *Walk) NextPace() (string, error) {
	w.Trail.Pace = journey.Next(journey.Paces, w.Trail.Pace)
	return "Pace: " + w.Trail.Pace + PaceNote[w.Trail.Pace], w.SaveTrail()
}

// NextRations changes to the next rations and says what they do.
func (w *Walk) NextRations() (string, error) {
	w.Trail.Rations = journey.Next(journey.Rations, w.Trail.Rations)
	return "Rations: " + w.Trail.Rations + RationNote[w.Trail.Rations], w.SaveTrail()
}

// Rest passes a day without walking, at today's weather, and says what it brought.
func (w *Walk) Rest(weather string) (string, error) {
	w.Trail.Weather = weather
	w.Trail.PassDay(w.Members, w.Region.Region.Terrain, true)
	return strings.Join(append([]string{fmt.Sprintf("Day %d. The party rests.", w.Trail.Day)}, w.Trail.Notes...), " "), w.SaveTrail()
}

// SaveTrail records the trail.
func (w Walk) SaveTrail() error { return w.Store.SaveTrail(w.Trail) }

// FoodLow reports whether the 斋饭 is down to two days' filling rations.
func (w Walk) FoodLow() bool { return w.Trail.Food < 3*len(w.Members)*2 }

// WaterLow reports whether the water is down to two days' drinking.
func (w Walk) WaterLow() bool { return w.Trail.Water < len(w.Members)*2 }

// SandalDaysLeft is how many more days the party's sandals last; 0 or
// less when worn out.
func (w Walk) SandalDaysLeft() int { return journey.SandalLife - w.Trail.SandalDays }

// PilgrimHP is the pilgrim's HP on the road: the most it can be by level,
// gear and the party's bonds, less the wounds carried from the last
// battle (never below 1).
func PilgrimHP(store *wiki.Store) (hp, most int, err error) {
	pages, err := store.ListPages()
	if err != nil {
		return 0, 0, err
	}
	cleared, err := store.ClearedStages()
	if err != nil {
		return 0, 0, err
	}
	kit, err := store.Kit()
	if err != nil {
		return 0, 0, err
	}
	bonds, err := store.BondPoints()
	if err != nil {
		return 0, 0, err
	}
	place, err := store.Place()
	if err != nil {
		return 0, 0, err
	}
	var e rpg.Effects
	rpg.AddGear(&e, journey.Worn(kit.Worn))
	rpg.AddUpgrades(&e, journey.Party(cleared), bonds)
	most = rpg.PlayerHP(journey.CountProgress(pages).LevelPoints()) + e.BonusHP
	return max(most-place.Wounds, 1), most, nil
}

// fresh is a random source seeded at random, for a pick or a shuffle.
func fresh() *rand.Rand { return rand.New(rand.NewSource(rand.Int63())) }

// orRandom is roll, or a random draw of 0-99 when nil.
func orRandom(roll func() int) func() int {
	if roll == nil {
		return func() int { return rand.Intn(100) }
	}
	return roll
}
