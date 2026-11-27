package main

import (
	"strings"
	"testing"

	"github.com/therealsiege/tuilaoshi/internal/journey"
)

func testComposer(t *testing.T) composer {
	t.Helper()
	b, err := loadBible()
	if err != nil {
		t.Fatal(err)
	}
	c, err := journey.LoadContent()
	if err != nil {
		t.Fatal(err)
	}
	return composer{b: b, c: c, direction: b.Direction}
}

func TestEveryCompanionAndBossHasASheet(t *testing.T) {
	p := testComposer(t)
	for _, c := range p.c.Companions {
		if p.b.Characters[c.Name] == "" {
			t.Errorf("companion %s has no character sheet", c.Name)
		}
	}
	for _, b := range p.c.Bosses {
		if p.b.Characters[b.Name] == "" {
			t.Errorf("boss %s has no character sheet", b.Name)
		}
	}
}

func TestEveryTerrainIsDescribed(t *testing.T) {
	p := testComposer(t)
	for _, r := range p.c.Regions {
		if tr := p.b.Terrain[r.Terrain]; tr.Look == "" || tr.Block == "" {
			t.Errorf("region %s: terrain %q has no look or block", r.ID, r.Terrain)
		}
	}
}

// Every subject of every kind composes a whole prompt: framing, subject,
// the world rules, the direction and the never list, with no placeholder
// left unfilled.
func TestEverySubjectComposes(t *testing.T) {
	p := testComposer(t)
	for _, dir := range []string{"real", "clean", "blend"} {
		p.direction = dir
		for kind := range p.b.Kinds {
			names := p.all(kind)
			if len(names) == 0 {
				t.Fatalf("kind %s has no subjects", kind)
			}
			for _, n := range names {
				j, err := p.job(kind, n)
				if err != nil {
					t.Fatalf("%s %s: %v", kind, n, err)
				}
				for _, must := range []string{p.b.World, p.style(), p.b.Never} {
					if !strings.Contains(j.Prompt, strings.TrimSpace(must)) {
						t.Fatalf("%s %s (%s): prompt is missing a block", kind, n, dir)
					}
				}
				if strings.Contains(j.Prompt, "{") || strings.Contains(j.Prompt, "Place: ()") {
					t.Fatalf("%s %s: unfilled placeholder in %q", kind, n, j.Prompt)
				}
				if kind == "map" && j.Layout == nil {
					t.Fatalf("map %s has no layout", n)
				}
			}
		}
	}
}

func TestPortraitsPlaceCharactersWhereTheyAreMet(t *testing.T) {
	p := testComposer(t)
	j, err := p.job("portrait", "白骨精")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(j.Prompt, "White Tiger Ridge") {
		t.Errorf("the White Bone Demon should be painted on White Tiger Ridge:\n%s", j.Prompt)
	}
	if _, err := p.job("portrait", "nobody"); err == nil {
		t.Error("an unknown name should be an error")
	}
}

// Vistas paint only the companions who have joined by the end of the region.
func TestVistasPaintThePartyOfTheTime(t *testing.T) {
	p := testComposer(t)
	early, err := p.job("vista", "liangjie") // stages 8-10: Wukong and the horse, no Pigsy yet
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(early.Prompt, "golden monkey") || strings.Contains(early.Prompt, "pig-faced") {
		t.Errorf("Mountain of Two Frontiers should show Wukong but not Pigsy:\n%s", early.Prompt)
	}
	late, err := p.job("vista", "baihu")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(late.Prompt, "blue-faced") {
		t.Errorf("White Tiger Ridge should show all the pilgrims:\n%s", late.Prompt)
	}
}
