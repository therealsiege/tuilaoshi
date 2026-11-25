package journey

// Prices and wear on the road west. Supplies cost more the farther west a
// town is, as on any long trail, and straw sandals (草鞋) wear out: after
// SandalLife days of walking, a worn-out pair slows the party and blisters
// its feet until new ones are bought in a town. A temple offering (香火钱)
// lifts the party's spirits, once a day in each town.

// Base prices in 铜钱, at the start of the road.
const (
	FoodPrice     = 15 // a bundle of FoodBundle 斋饭
	FoodBundle    = 10
	InnPrice      = 10
	SandalPrice   = 6 // a pair, for each member
	OfferingPrice = 20
)

// SandalLife is how many days of walking a pair of sandals lasts.
const SandalLife = 30

// What worn-out sandals cost each day: fewer steps and some health.
const (
	wornSteps  = 2
	wornHealth = 3
)

// OfferingMorale is how much a temple offering lifts each member's spirits.
const OfferingMorale = 15

// Price is a base price at this town: it rises steadily going west, to
// four times the base at the last tribulation.
func (t Town) Price(base int) int { return base + base*t.Stage*3/81 }

// SandalsWorn reports whether the party's sandals are worn out.
func (t Trail) SandalsWorn() bool { return t.SandalDays > SandalLife }

// NewSandals puts a fresh pair on every member.
func (t *Trail) NewSandals() { t.SandalDays = 0 }

// wearSandals counts a day of walking on the party's sandals and returns
// the health the day costs once they are worn out.
func (t *Trail) wearSandals() int {
	t.SandalDays++
	if !t.SandalsWorn() {
		return 0
	}
	t.Notes = append(t.Notes, "Worn-out sandals: feet blister and the party slows. Buy 草鞋 in a town.")
	return wornHealth
}

// stepsToday is how far a day goes at the pace, less with worn sandals.
func (t Trail) stepsToday() int {
	if t.SandalsWorn() {
		return stepsPerDay[t.Pace] - wornSteps
	}
	return stepsPerDay[t.Pace]
}
