package journey

import (
	"math/rand"
	"strings"
	"testing"
)

func TestMysteriesAreComplete(t *testing.T) {
	if len(content.Mysteries) < 6 {
		t.Fatalf("only %d mysteries", len(content.Mysteries))
	}
	ids := map[string]bool{}
	for _, m := range content.Mysteries {
		if m.ID == "" || ids[m.ID] {
			t.Errorf("mystery id %q empty or repeated", m.ID)
		}
		ids[m.ID] = true
		if m.Title == "" || m.English == "" || m.Situation == "" || m.SituationEnglish == "" {
			t.Errorf("%s: missing title or situation", m.ID)
		}
		if len(m.Facts) < 2 || len(m.FactsEnglish) != len(m.Facts) {
			t.Errorf("%s: want 2+ facts each with English; got %d and %d", m.ID, len(m.Facts), len(m.FactsEnglish))
		}
		if len(m.Villagers) != 3 {
			t.Errorf("%s: want 3 villagers, got %d", m.ID, len(m.Villagers))
		}
		for _, v := range m.Villagers {
			if v.Name == "" || v.Role == "" || v.English == "" || v.Persona == "" ||
				v.Story == "" || v.Lie == "" || v.Flaw == "" || v.FlawEnglish == "" {
				t.Errorf("%s: villager %q is missing a field", m.ID, v.Name)
			}
		}
	}
}

func TestVillagerPromptHidesTheTruthFromTheInnocent(t *testing.T) {
	m := content.Mysteries[0]
	demon := m.VillagerPrompt(1, 1)
	if !strings.Contains(demon, m.Villagers[1].Lie) || strings.Contains(demon, m.Villagers[1].Story) {
		t.Errorf("the demon should tell its lie, not the true story:\n%s", demon)
	}
	honest := m.VillagerPrompt(0, 1)
	if !strings.Contains(honest, m.Villagers[0].Story) || strings.Contains(honest, "demon in disguise") {
		t.Errorf("an innocent villager should tell the truth:\n%s", honest)
	}
	if !strings.Contains(honest, m.Facts[0]) || !strings.Contains(honest, "EN:") {
		t.Errorf("the prompt should give the facts and ask for an EN: line:\n%s", honest)
	}
}

func TestPickMysteryVariesTheDemon(t *testing.T) {
	seen := map[int]bool{}
	rng := rand.New(rand.NewSource(1))
	for range 60 {
		_, d := PickMystery(rng)
		seen[d] = true
	}
	if len(seen) != 3 {
		t.Errorf("the demon should be any of the three villagers: %v", seen)
	}
}

func TestSplitReply(t *testing.T) {
	zh, en := SplitReply("我昨天在家。\n雨很大。\nEN: I was home yesterday. It rained hard.")
	if zh != "我昨天在家。雨很大。" || en != "I was home yesterday. It rained hard." {
		t.Errorf("got %q / %q", zh, en)
	}
}
