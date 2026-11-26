package assess

import (
	"math/rand"
	"strings"
	"testing"

	"github.com/therealsiege/tuilaoshi/internal/model"
	"github.com/therealsiege/tuilaoshi/internal/ui/uitest"
)

func mockFixture(t *testing.T) ([]model.Page, []model.Card) {
	t.Helper()
	store := uitest.Wiki(t, uitest.MockTestPages())
	pages, err := store.ListPages()
	if err != nil {
		t.Fatal(err)
	}
	cards, err := store.DrillableCards()
	if err != nil {
		t.Fatal(err)
	}
	return pages, cards
}

func TestBuildMockMakesAFullPaperFromTheWiki(t *testing.T) {
	pages, cards := mockFixture(t)
	for seed := int64(1); seed <= 20; seed++ {
		checkPaper(t, BuildMock(4, pages, cards, rand.New(rand.NewSource(seed))))
	}
}

func checkPaper(t *testing.T, m Mock) {
	t.Helper()
	if got := len(m.Sections[Listening]); got != ListeningItems {
		t.Errorf("%d listening items", got)
	}
	if got := len(m.Sections[Reading]); got != GapItems+PassageItems {
		t.Errorf("%d reading items", got)
	}
	if got := len(m.Sections[Writing]); got != TileItems {
		t.Errorf("%d writing items", got)
	}
	if m.Compose == nil || m.Compose.Answer == "" {
		t.Error("no sentence to compose")
	}
	seen := map[string]bool{}
	kinds := map[ItemKind]int{}
	for _, section := range m.Sections {
		for _, it := range section {
			kinds[it.Kind]++
			key := it.Answer + it.Text
			if seen[key] {
				t.Errorf("item used twice: %+v", it)
			}
			seen[key] = true
			if it.Kind == Tiles {
				if strings.Join(it.Tiles, "") == it.Answer || len(it.Tiles) < 3 || len(it.Tiles) > maxTiles {
					t.Errorf("tiles %v for %s", it.Tiles, it.Answer)
				}
				continue
			}
			if len(it.Choices) != 4 || it.Correct < 0 || it.Correct >= 4 {
				t.Fatalf("choices %v correct %d", it.Choices, it.Correct)
			}
			right := it.Choices[it.Correct]
			for i, c := range it.Choices {
				if i != it.Correct && c == right {
					t.Errorf("a wrong choice equals the right one: %v", it.Choices)
				}
			}
			switch it.Kind {
			case HeardWord:
				if !strings.Contains(it.Audio, right) {
					t.Errorf("heard %q but the answer is %q", it.Audio, right)
				}
				for i, c := range it.Choices {
					if i != it.Correct && strings.Contains(it.Audio, c) {
						t.Errorf("wrong choice %q is in the sentence %q", c, it.Audio)
					}
				}
			case Gap:
				if !strings.Contains(it.Text, "____") || strings.Replace(it.Text, "____", right, 1) != it.Answer {
					t.Errorf("gap %q filled with %q is not %q", it.Text, right, it.Answer)
				}
			}
		}
	}
	for _, k := range []ItemKind{HeardWord, HeardMeaning, Gap, Passage, Tiles} {
		if kinds[k] == 0 {
			t.Errorf("no items of kind %d: %v", k, kinds)
		}
	}
}

func TestDistractorsAreWordsOfTheSameKind(t *testing.T) {
	p := pools{words: []word{
		{zh: "推荐", en: "to recommend"}, {zh: "提醒", en: "to remind"}, {zh: "打扰", en: "to disturb"},
		{zh: "报名", en: "to sign up"}, {zh: "展览", en: "exhibition"}, {zh: "观众", en: "audience"},
	}}
	got := p.distractors("推荐", "我推荐你去。", rand.New(rand.NewSource(2)))
	for _, d := range got {
		if d == "展览" || d == "观众" {
			t.Errorf("a noun %s among distractors for a verb: %v", d, got)
		}
	}
	if len(got) != 3 {
		t.Errorf("distractors %v", got)
	}
}

func TestAThinLevelReachesDown(t *testing.T) {
	pages, cards := mockFixture(t)
	m := BuildMock(5, pages, cards, rand.New(rand.NewSource(3)))
	if len(m.Sections[Reading]) == 0 || m.Level != 5 {
		t.Errorf("an HSK5 paper on an HSK4 wiki has %d reading items", len(m.Sections[Reading]))
	}
}

func TestMockScoringAndPassMark(t *testing.T) {
	if got := (SectionScore{Points: 7, Max: 10}).Score(); got != 70 {
		t.Errorf("7/10 scored %d", got)
	}
	if got := (SectionScore{Points: 2, Max: 3}).Score(); got != 67 {
		t.Errorf("2/3 scored %d", got)
	}
	all := [NumSections]SectionScore{{Points: 6, Max: 10}, {Points: 9, Max: 15}, {Points: 2, Max: 3}}
	total, most := MockTotal(all)
	if total != 60+60+67 || most != 300 || !MockPassed(total, most) {
		t.Errorf("total %d of %d", total, most)
	}
	if MockPassed(179, 300) || !MockPassed(180, 300) {
		t.Error("the pass mark is 180 of 300")
	}
	noAudio := [NumSections]SectionScore{{Skipped: true}, {Points: 9, Max: 15}, {Points: 1, Max: 3}}
	total, most = MockTotal(noAudio)
	if total != 60+33 || most != 200 || MockPassed(total, most) {
		t.Errorf("without listening: %d of %d", total, most)
	}
	if MockPassed(0, 0) {
		t.Error("nothing sat passed")
	}
}

func TestEstimateAndVerdict(t *testing.T) {
	results := []Result{
		{Level: 4, Levels: []int{3, 4, 5}, Asked: 3}, {Level: 0, Levels: []int{3, 4}, Asked: 3},
		{Level: 5, Levels: []int{3, 4, 5}, Asked: 3}, {Levels: nil},
	}
	if got := Estimate(results); got != 4 {
		t.Errorf("estimate %d, want 4", got)
	}
	section := "| Skill | Estimated Level | Confidence | Last Assessed |\n|---|---|---|---|\n" +
		"| Character recognition | HSK5+ (highest in wiki) | 3/3 | x |\n| Grammar | below HSK3 | 0/3 | x |\n" +
		"| Vocabulary (passive) | HSK4 | 2/3 | x |\n| Speaking | ? | — | — |\n"
	if got := EstimateFromSection(section); got != 4 {
		t.Errorf("estimate from overview %d, want 4", got)
	}
	for _, c := range []struct {
		passed   bool
		estimate int
		want     string
	}{{true, 4, "bears it out"}, {true, 3, "looks low"}, {false, 4, "looks high"}, {false, 3, "agrees"}, {true, 0, "No level check"}} {
		if got := Verdict(4, c.passed, c.estimate); !strings.Contains(got, c.want) {
			t.Errorf("verdict %v %d: %q", c.passed, c.estimate, got)
		}
	}
}
