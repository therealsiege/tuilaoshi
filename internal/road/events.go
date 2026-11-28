package road

import (
	"errors"
	"fmt"
	"strings"

	"github.com/therealsiege/tuilaoshi/internal/journey"
	"github.com/therealsiege/tuilaoshi/internal/wiki"
)

// The day's events and the region's crossing: an event is met by one of
// its choices, which may pay or charge 铜钱 and moves the trail's supplies
// and the party's health (halved harm when forewarned). A crossing is an
// event too, made from the obstacle's condition today (see Move); once a
// way across is chosen the party is on the far side for good (Cross).

// pickEvent may bring an event at the end of a day (roll as for Walked).
func (w Walk) pickEvent(roll func() int) *journey.Event {
	if orRandom(roll)() >= journey.EventChance {
		return nil
	}
	if e, ok := journey.PickEvent(w.Region.Region.Terrain, w.Members, fresh()); ok {
		return &e
	}
	return nil
}

// Choose meets event e by its choice i: paying or taking its 铜钱, then
// putting its outcome into the trail (halved and the warning spent when
// forewarned) and saving it. It returns what the choice changed, in short;
// when the purse can't pay, nothing comes of it and it says so. The error
// is the store's.
func (w *Walk) Choose(e journey.Event, i int) (string, error) {
	if i < 0 || i >= len(e.Choices) {
		return "", fmt.Errorf("no choice %d of %d", i, len(e.Choices))
	}
	o := e.Choices[i].Outcome
	var earnErr error
	switch {
	case o.Coins < 0:
		if err := w.Store.Spend(wiki.Wallet{Tongqian: -o.Coins}); err != nil {
			return "Not enough 铜钱 to pay: nothing comes of it.", nil
		}
	case o.Coins > 0:
		earnErr = w.Store.Earn(wiki.Wallet{Tongqian: o.Coins})
	}
	spared := false
	if w.Trail.Forewarned {
		if o, spared = journey.Forewarn(o); spared {
			w.Trail.Forewarned = false // the warning is spent on this danger
		}
	}
	w.Trail.Apply(o, w.Members)
	effect := EffectText(o)
	if spared {
		effect += " (forewarned: harm halved)"
	}
	if err := w.SaveTrail(); err != nil {
		return effect, err
	}
	return effect, earnErr
}

// EffectText says what an outcome changed: "斋饭 +12 · health -4".
func EffectText(o journey.Outcome) string {
	var parts []string
	add := func(label string, v int) {
		if v != 0 {
			parts = append(parts, fmt.Sprintf("%s %+d", label, v))
		}
	}
	add("斋饭", o.Food)
	add("水", o.Water)
	add("铜钱", o.Coins)
	who := "party"
	if o.Who != "" {
		who = o.Who
	}
	add(who+" health", o.Health)
	add(who+" morale", o.Morale)
	if len(parts) == 0 {
		return "Nothing changes."
	}
	return strings.Join(parts, " · ")
}

// Cross puts the party on the far side of the region's obstacle, for good.
func (w *Walk) Cross() error {
	w.Trail.Crossed = append(w.Trail.Crossed, w.Region.Region.ID)
	w.Pos = w.Region.CrossAt
	return errors.Join(w.Save(), w.SaveTrail())
}
