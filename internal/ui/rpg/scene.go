package rpg

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/therealsiege/tuilaoshi/internal/journey"
	"github.com/therealsiege/tuilaoshi/internal/ui"
	"github.com/therealsiege/tuilaoshi/internal/ui/reader"
)

// openScene starts on the open stage's story scene the first time the stage
// is fought; otherwise the battle starts at the intro.
func (m Model) openScene() Model {
	if m.stage == nil {
		return m
	}
	sc := journey.SceneFor(m.stage.Number)
	if sc == nil {
		return m
	}
	seen, err := m.store.SceneSeen(m.stage.Number)
	if err != nil {
		m.loadErr = err
		return m
	}
	if seen {
		return m
	}
	return m.showScene(sc)
}

// showScene opens sc in the reader, glossed by the story glossary.
func (m Model) showScene(sc *journey.Scene) Model {
	m.sceneOf, m.showEnglish = sc, false
	m.scene = reader.New(m.store).Refresh().OpenText(sc.Title, sc.Text, journey.StoryGlossary())
	m.state = stateScene
	return m
}

// updateScene reads the scene word by word; tab goes on to the battle and
// records the scene as seen, e shows the translation.
func (m Model) updateScene(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "tab":
		if m.stage != nil {
			if err := m.store.MarkSceneSeen(m.stage.Number); err != nil {
				m.saveErr = err
			}
		}
		m.state = stateIntro
		return m, nil
	case "e":
		m.showEnglish = !m.showEnglish
		return m, nil
	}
	var cmd tea.Cmd
	m.scene, cmd = m.scene.Update(msg)
	return m, cmd
}

func (m Model) viewScene() string {
	var b strings.Builder
	sc := m.sceneOf
	b.WriteString(m.scene.TextView(""))
	if m.showEnglish {
		b.WriteString("\n")
		for _, line := range strings.Split(ui.WrapText(sc.English, 70), "\n") {
			b.WriteString(ui.MutedStyle.Render("  "+line) + "\n")
		}
	}
	fmt.Fprintf(&b, "\n  %s %s\n  %s\n\n", ui.ChineseStyle.Render(sc.Speaker+"："), ui.ChineseStyle.Render(sc.Line),
		ui.MutedStyle.Render(sc.LineEnglish))
	b.WriteString(ui.HelpStyle.Render("  ←→ word · ↑↓ line · enter didn't know it · e translation · tab to battle · esc flee"))
	return b.String()
}
