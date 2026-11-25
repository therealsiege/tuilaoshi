package journey

import (
	"math"
	"time"
)

// The solar terms (节气) follow the sun: a term begins when the sun's
// ecliptic longitude reaches a multiple of 15°. The longitude is computed
// (a low-precision formula, good to well under an hour), so no table of
// dates is needed, and dates are taken in China time.

// SolarTerm is one of the 24 terms.
type SolarTerm struct {
	Name        string  `yaml:"name"`
	Pinyin      string  `yaml:"pinyin"`
	English     string  `yaml:"english"`
	Longitude   float64 `yaml:"longitude"` // degrees
	Weather     string  `yaml:"weather"`   // cold, heat, rain, or ""
	Text        string  `yaml:"text"`
	TextEnglish string  `yaml:"text_english"`
}

// chinaTime is the zone the terms' days are counted in.
var chinaTime = time.FixedZone("CST", 8*3600)

// sunLongitude is the sun's apparent ecliptic longitude in degrees at t.
func sunLongitude(t time.Time) float64 {
	n := float64(t.UTC().Unix())/86400 + 2440587.5 - 2451545.0 // days from J2000
	l := 280.460 + 0.9856474*n
	g := (357.528 + 0.9856003*n) * math.Pi / 180
	lambda := l + 1.915*math.Sin(g) + 0.020*math.Sin(2*g)
	return math.Mod(math.Mod(lambda, 360)+360, 360)
}

// TermOn returns the solar term in effect on now's day (in China time),
// and whether it begins that day.
func TermOn(now time.Time) (term SolarTerm, starts bool) {
	d := now.In(chinaTime)
	start := time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, chinaTime)
	at := func(t time.Time) int { return int(sunLongitude(t) / 15) }
	end := at(start.Add(24 * time.Hour))
	return termAt(end), at(start) != end
}

// termAt is the term beginning at the given multiple of 15°.
func termAt(step int) SolarTerm {
	for _, t := range content.SolarTerms {
		if int(t.Longitude)/15 == step {
			return t
		}
	}
	return SolarTerm{}
}

// Weather effects on a day of the trail: cold eats more, heat drinks more,
// rain refills the gourds.
const (
	coldFood  = 1 // more 斋饭 for each member
	heatWater = 1 // more water for each member
	rainWater = 3 // water the rain adds
)
