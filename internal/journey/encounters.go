package journey

// Encounter is a random demon encounter during a drill session.
type Encounter struct {
	Demon       BestiaryEntry
	Intro       string // narrative intro line in Chinese
	IntroPinyin string
	IntroEng    string
	Challenge   string // what the player must do
}

// RandomEncounter returns an encounter based on the count (deterministic rotation).
// Returns nil if count doesn't trigger one (only triggers every ~15 reviews).
func RandomEncounter(count int) *Encounter {
	if count == 0 || count%15 != 0 {
		return nil
	}
	demons := Bestiary()
	idx := (count / 15) % len(demons)
	demon := demons[idx]
	return &Encounter{
		Demon:       demon,
		Intro:       "前方出现了妖怪！" + demon.Name + "挡住了去路！",
		IntroPinyin: "Qianfang chuxian le yaoguai! " + demon.English + " dangzhu le qulu!",
		IntroEng:    "A demon appears ahead! " + demon.English + " blocks your path!",
		Challenge:   "Answer 3 characters correctly to defeat " + demon.English + "!",
	}
}
