package assessment

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/therealsiege/tuilaoshi/internal/ui/uitest"
)

var chars = map[int][]string{
	3: {"书 shū book", "水 shuǐ water", "火 huǒ fire", "山 shān mountain"},
	4: {"舟 zhōu boat", "弓 gōng bow", "矛 máo spear", "鼎 dǐng cauldron"},
}

func assessWiki(t *testing.T) Model {
	t.Helper()
	pages := map[string]string{"overview.md": "# Overview\n\n## Current Level Assessment\n_Pending_\n\n| Skill | Estimated Level | Confidence | Last Assessed |\n|---|---|---|---|\n| Speaking | ? | — | — |\n\n## Roadmap\n"}
	for level, list := range chars {
		for _, c := range list {
			f := strings.Fields(c)
			pages[fmt.Sprintf("characters/%s.md", f[1])] = fmt.Sprintf(
				"---\ntype: character\nlevel: HSK%d\n---\n\n# %s (%s)\n\n- **Character**: %s\n- **Pinyin**: %s\n- **Meaning**: %s\n", level, f[0], f[1], f[0], f[1], f[2])
		}
	}
	return New(uitest.Wiki(t, pages)).SetSize(80, 30).Refresh()
}

func TestAssessmentUpdatesOverview(t *testing.T) {
	m := assessWiki(t)
	uitest.Golden(t, "intro", m.View())

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	for i := 0; m.phase == phaseAsking && i < 50; i++ {
		m.selected = m.question.Correct // knows every character
		m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	}
	if m.phase != phaseResults || m.saveErr != nil {
		t.Fatalf("phase %v, save error %v", m.phase, m.saveErr)
	}
	data, err := os.ReadFile(filepath.Join(m.store.Dir, "overview.md"))
	if err != nil {
		t.Fatal(err)
	}
	page := string(data)
	for _, want := range []string{
		"| Character recognition | HSK4+ (highest in wiki) | 3/3 right at HSK4 |",
		"| Speaking | ? | — | — |",
		"## Roadmap",
	} {
		if !strings.Contains(page, want) {
			t.Errorf("overview.md missing %q:\n%s", want, page)
		}
	}
}

func TestAssessmentFindsLowerLevelWhenWrong(t *testing.T) {
	m := assessWiki(t)
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	for i := 0; m.phase == phaseAsking && i < 50; i++ {
		m.selected = (m.question.Correct + 1) % len(m.question.Choices) // knows nothing
		m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	}
	if got := m.results[0].Label(); got != "below HSK3" {
		t.Errorf("character result %q, want below HSK3", got)
	}
}
