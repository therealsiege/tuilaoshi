package road

import (
	"math/rand"
	"strings"
	"testing"

	"github.com/therealsiege/tuilaoshi/internal/journey"
)

func event(choices ...journey.Outcome) journey.Event {
	e := journey.Event{ID: "test", Text: "有事。", English: "Something."}
	for _, o := range choices {
		e.Choices = append(e.Choices, journey.EventChoice{Text: "走", English: "Walk", Outcome: o, Result: "好。", ResultEnglish: "Good."})
	}
	return e
}

func TestChoosingMeetsAnEvent(t *testing.T) {
	w := open(t, 0)
	food := w.Trail.Food
	e := event(journey.Outcome{Food: 12, Health: -4}, journey.Outcome{Coins: -50})
	effect, err := w.Choose(e, 0)
	if err != nil || effect != "斋饭 +12 · party health -4" || w.Trail.Food != food+12 || w.Trail.Health[journey.Master] != 96 {
		t.Fatalf("choice 1: %q %v, food %d, health %d", effect, err, w.Trail.Food, w.Trail.Health[journey.Master])
	}
	if reloaded(t, w).Trail.Food != food+12 {
		t.Error("the choice wasn't saved")
	}
	if effect, err := w.Choose(e, 1); err != nil || !strings.HasPrefix(effect, "Not enough 铜钱") {
		t.Errorf("a price the purse can't pay: %q %v", effect, err)
	}
	if _, err := w.Choose(e, 2); err == nil {
		t.Error("a choice out of range should be refused")
	}
}

func TestChoosingPaysAndEarns(t *testing.T) {
	w := open(t, 0)
	e := event(journey.Outcome{Coins: 20}, journey.Outcome{Coins: -5})
	if _, err := w.Choose(e, 0); err != nil {
		t.Fatal(err)
	}
	if effect, err := w.Choose(e, 1); err != nil || effect != "铜钱 -5" {
		t.Fatalf("paying: %q %v", effect, err)
	}
	if wallet, err := w.Store.Wallet(); err != nil || wallet.Tongqian != 15 {
		t.Errorf("purse %+v %v, want 15", wallet, err)
	}
}

func TestAForewarnedPartyTakesHalfTheHarm(t *testing.T) {
	w := open(t, 0)
	w.Trail.Forewarned = true
	food := w.Trail.Food
	effect, err := w.Choose(event(journey.Outcome{Food: -10, Health: -20}), 0)
	if err != nil || w.Trail.Food != food-5 || w.Trail.Health[journey.Master] != 90 || w.Trail.Forewarned || !strings.HasSuffix(effect, "(forewarned: harm halved)") {
		t.Errorf("forewarned: %q %v, food %d, health %d, still forewarned %v", effect, err, w.Trail.Food, w.Trail.Health[journey.Master], w.Trail.Forewarned)
	}
}

func TestEffectText(t *testing.T) {
	for o, want := range map[journey.Outcome]string{
		{}:                                "Nothing changes.",
		{Water: 5, Coins: -3}:             "水 +5 · 铜钱 -3",
		{Morale: 10, Who: "孙悟空"}:          "孙悟空 morale +10",
		{Food: -1, Health: 2, Morale: -2}: "斋饭 -1 · party health +2 · party morale -2",
	} {
		if got := EffectText(o); got != want {
			t.Errorf("%+v: %q, want %q", o, got, want)
		}
	}
}

func TestCrossingPutsThePartyOnTheFarSide(t *testing.T) {
	w := open(t, 0)
	e := w.Region.Crossing.Event(rand.New(rand.NewSource(1)))
	if _, err := w.Choose(e, 2); err != nil { // wait a day
		t.Fatal(err)
	}
	if err := w.Cross(); err != nil || w.Pos != w.Region.CrossAt || !w.HasCrossed() {
		t.Fatalf("crossed: at %v, crossed %v, %v", w.Pos, w.HasCrossed(), err)
	}
	again := reloaded(t, w)
	if !again.HasCrossed() || again.Pos != w.Region.CrossAt {
		t.Errorf("the crossing wasn't saved: %v at %v", again.HasCrossed(), again.Pos)
	}
	if _, o := from(t, again, w.Region.CrossAt); o.Kind != Moved {
		t.Error("a crossed river barred the road again")
	}
}
