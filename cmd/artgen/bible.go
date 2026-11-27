package main

import (
	_ "embed"
	"fmt"
	"sort"
	"strings"

	"github.com/therealsiege/tuilaoshi/internal/journey"
	"gopkg.in/yaml.v3"
)

//go:embed bible.yaml
var bibleYAML []byte

// Bible is bible.yaml: the blocks every prompt is composed from.
type Bible struct {
	Direction  string            `yaml:"direction"`
	Directions map[string]string `yaml:"directions"`
	World      string            `yaml:"world"`
	Never      string            `yaml:"never"`
	Kinds      map[string]Kind   `yaml:"kinds"`
	Terrain    map[string]struct {
		Look  string `yaml:"look"`
		Block string `yaml:"block"`
	} `yaml:"terrain"`
	Crossings  map[string]string `yaml:"crossings"`
	Party      map[string]string `yaml:"party"`
	Characters map[string]string `yaml:"characters"`
	Scenes     []Scene           `yaml:"scenes"`
}

// Scene is one battle cut scene: the tribulation it plays at, who is in it
// by Chinese name, and what happens, from the novel.
type Scene struct {
	ID    string   `yaml:"id"`
	Stage int      `yaml:"stage"`
	Cast  []string `yaml:"cast"`
	Place string   `yaml:"place"` // only when it happens away from the stage's region, as a backstory does
	Text  string   `yaml:"text"`
}

// Kind is one kind of picture: its aspect ratio and framing.
type Kind struct {
	Ratio  string `yaml:"ratio"`
	Frame  string `yaml:"frame"`
	Legend string `yaml:"legend"`
}

func loadBible() (Bible, error) {
	var b Bible
	err := yaml.Unmarshal(bibleYAML, &b)
	return b, err
}

// Job is one picture to paint: a name for its file, the prompt, and the
// aspect ratio. Layout is set for maps: the region drawn as colours.
type Job struct {
	Kind, Name, Ratio, Prompt string
	Subject                   string              // the character a ref picture is kept for, if any
	Cast                      []journey.Companion // companions shown small, whose approved pictures are attached
	Peers                     []string            // characters whose approved pictures set the style for this one
	Keep                      []figure            // characters in a scene, whose approved pictures are attached and kept exactly
	Layout                    *journey.RegionMap
}

// figure is a character by Chinese and English name.
type figure struct{ Name, English string }

// composer builds prompts from the bible and the game's content.
type composer struct {
	b         Bible
	c         journey.Content
	direction string
}

func (p composer) style() string { return p.b.Directions[p.direction] }

func (p composer) join(parts ...string) string {
	var out []string
	for _, s := range parts {
		if s = strings.TrimSpace(s); s != "" {
			out = append(out, s)
		}
	}
	return strings.Join(out, "\n\n")
}

// regionAt is the region a tribulation number falls in.
func (p composer) regionAt(stage int) journey.Region {
	for _, r := range p.c.Regions {
		if stage >= r.From && stage <= r.To {
			return r
		}
	}
	return p.c.Regions[0]
}

func (p composer) place(r journey.Region) string {
	return fmt.Sprintf("Place: %s (%s), %s.", r.English, r.Name, p.b.Terrain[r.Terrain].Look)
}

// who describes a character: their sheet if they have one, else their
// entry in the bestiary, soldiers or spirits, and where they are met.
func (p composer) who(name string) (desc string, where *journey.Region, ok bool) {
	stage := 0
	for _, b := range p.c.Bosses {
		if b.Name == name {
			stage = b.Stage
		}
	}
	for _, c := range p.c.Companions {
		if c.Name == name {
			stage = c.Stage
		}
	}
	entry := func(e journey.BestiaryEntry, extra string) string {
		s := fmt.Sprintf("Subject: %s (%s)", e.English, e.Name)
		if e.TitleEng != "" {
			s += ", " + e.TitleEng
		}
		s += ". " + strings.TrimSpace(e.Origin)
		if e.Powers != "" {
			s += " Powers: " + strings.TrimSpace(e.Powers) + "."
		}
		return s + extra
	}
	if sheet, has := p.b.Characters[name]; has {
		desc = "Subject: " + sheet
	} else {
		for _, e := range p.c.Bestiary {
			if e.Name == name {
				desc = entry(e, " A demon of the novel: design them from this, fearsome but readable.")
			}
		}
		for _, s := range p.c.Soldiers {
			if s.Name == name {
				desc, stage = entry(s.BestiaryEntry, " A "+s.Rank+" of the demon armies: rough gear, a weapon that fits their deeds."), s.From
			}
		}
		for _, s := range p.c.Spirits {
			if s.Name == name {
				desc = entry(s.BestiaryEntry, " A spirit of Chinese folklore: otherworldly rather than monstrous.")
			}
		}
	}
	if desc == "" {
		return "", nil, false
	}
	if stage > 0 {
		r := p.regionAt(stage)
		where = &r
	}
	return desc, where, true
}

func (p composer) portrait(name string) (Job, error) {
	desc, where, ok := p.who(name)
	if !ok {
		return Job{}, fmt.Errorf("no character, demon, soldier or spirit named %s", name)
	}
	k := p.b.Kinds["portrait"]
	place := ""
	if where != nil {
		place = p.place(*where)
	}
	var peers []string
	for _, c := range p.c.Companions {
		if c.Name != name {
			peers = append(peers, c.Name)
		}
	}
	return Job{Kind: "portrait", Name: name, Ratio: k.Ratio, Subject: name, Peers: peers,
		Prompt: p.join(k.Frame, desc, place, p.b.World, p.style(), p.b.Never)}, nil
}

func (p composer) region(id string) (journey.Region, int, error) {
	for i, r := range p.c.Regions {
		if r.ID == id || r.Name == id {
			return r, i, nil
		}
	}
	return journey.Region{}, 0, fmt.Errorf("no region %s", id)
}

// events lists what happens in a region, from its tribulations, so the
// picture can hint at the story told there.
func (p composer) events(r journey.Region) string {
	var names []string
	for _, t := range p.c.Tribulations {
		if t.Number >= r.From && t.Number <= r.To {
			names = append(names, t.English)
		}
	}
	return "The story here: " + strings.Join(names, "; ") + ". Hint at it in the landscape, with no figures besides the pilgrims."
}

func (p composer) vista(id string) (Job, error) {
	r, _, err := p.region(id)
	if err != nil {
		return Job{}, err
	}
	k := p.b.Kinds["vista"]
	var party []string
	var cast []journey.Companion
	for _, c := range p.c.Companions {
		if c.Stage <= r.To && p.b.Party[c.Name] != "" {
			party = append(party, p.b.Party[c.Name])
			cast = append(cast, c)
		}
	}
	frame := strings.ReplaceAll(k.Frame, "{party}", strings.Join(party, "; "))
	return Job{Kind: "vista", Name: r.ID, Ratio: k.Ratio, Cast: cast,
		Prompt: p.join(frame, p.place(r), p.events(r), p.b.World, p.style(), p.b.Never)}, nil
}

func (p composer) worldMap(id string) (Job, error) {
	r, i, err := p.region(id)
	if err != nil {
		return Job{}, err
	}
	k := p.b.Kinds["map"]
	crossing, ok := p.b.Crossings[r.Terrain]
	if !ok {
		crossing = p.b.Crossings["default"]
	}
	legend := strings.NewReplacer("{block}", p.b.Terrain[r.Terrain].Block, "{crossing}", crossing).Replace(k.Legend)
	m := journey.BuildMap(i)
	return Job{Kind: "map", Name: r.ID, Ratio: k.Ratio, Layout: &m,
		Prompt: p.join(k.Frame, legend, p.place(r), p.b.World, p.style(), p.b.Never)}, nil
}

func (p composer) town(id string) (Job, error) {
	for _, t := range p.c.Towns {
		if t.ID == id || t.Name == id {
			k := p.b.Kinds["town"]
			subject := fmt.Sprintf("Subject: %s (%s). %s", t.English, t.Name, t.DescEnglish)
			return Job{Kind: "town", Name: t.ID, Ratio: k.Ratio,
				Prompt: p.join(k.Frame, subject, p.place(p.regionAt(t.Stage)), p.b.World, p.style(), p.b.Never)}, nil
		}
	}
	return Job{}, fmt.Errorf("no town %s", id)
}

// english is a character's English name from the game's content, else
// their Chinese name.
func (p composer) english(name string) string {
	for _, c := range p.c.Companions {
		if c.Name == name {
			return c.English
		}
	}
	for _, e := range p.c.Bestiary {
		if e.Name == name {
			return e.English
		}
	}
	return name
}

func (p composer) scene(id string) (Job, error) {
	for _, s := range p.b.Scenes {
		if s.ID != id {
			continue
		}
		k := p.b.Kinds["scene"]
		place := p.place(p.regionAt(s.Stage))
		if s.Place != "" {
			place = "Place: " + strings.TrimSpace(s.Place)
		}
		var sheets []string
		var keep []figure
		for _, n := range s.Cast {
			sheet, ok := p.b.Characters[n]
			if !ok {
				return Job{}, fmt.Errorf("scene %s: no character sheet for %s", id, n)
			}
			sheets = append(sheets, "- "+strings.TrimSpace(sheet))
			keep = append(keep, figure{n, p.english(n)})
		}
		cast := "Who is in it, and how they look:\n" + strings.Join(sheets, "\n")
		return Job{Kind: "scene", Name: s.ID, Ratio: k.Ratio, Keep: keep,
			Prompt: p.join(k.Frame, "Scene: "+strings.TrimSpace(s.Text), cast, place, p.b.World, p.style(), p.b.Never)}, nil
	}
	return Job{}, fmt.Errorf("no scene %s", id)
}

// all lists every subject of a kind, in story order where there is one.
func (p composer) all(kind string) []string {
	var out []string
	switch kind {
	case "portrait":
		for _, c := range p.c.Companions {
			out = append(out, c.Name)
		}
		for _, b := range p.c.Bosses {
			out = append(out, b.Name)
		}
		seen := map[string]bool{}
		for _, n := range out {
			seen[n] = true
		}
		var rest []string
		for _, e := range p.c.Bestiary {
			rest = append(rest, e.Name)
		}
		for _, s := range p.c.Soldiers {
			rest = append(rest, s.Name)
		}
		for _, s := range p.c.Spirits {
			rest = append(rest, s.Name)
		}
		for _, n := range rest {
			if !seen[n] {
				seen[n] = true
				out = append(out, n)
			}
		}
	case "vista", "map":
		for _, r := range p.c.Regions {
			out = append(out, r.ID)
		}
	case "town":
		for _, t := range p.c.Towns {
			out = append(out, t.ID)
		}
	case "scene":
		for _, s := range p.b.Scenes {
			out = append(out, s.ID)
		}
	}
	return out
}

func (p composer) job(kind, name string) (Job, error) {
	switch kind {
	case "portrait":
		return p.portrait(name)
	case "vista":
		return p.vista(name)
	case "map":
		return p.worldMap(name)
	case "town":
		return p.town(name)
	case "scene":
		return p.scene(name)
	}
	kinds := make([]string, 0, len(p.b.Kinds))
	for k := range p.b.Kinds {
		kinds = append(kinds, k)
	}
	sort.Strings(kinds)
	return Job{}, fmt.Errorf("kind must be one of %s", strings.Join(kinds, ", "))
}
