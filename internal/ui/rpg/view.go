package rpg

import (
	"fmt"
	"slices"
	"strings"

	"github.com/charmbracelet/lipgloss"

	rpgengine "github.com/therealsiege/tuilaoshi/internal/rpg"
	"github.com/therealsiege/tuilaoshi/internal/ui"
	"github.com/therealsiege/tuilaoshi/internal/ui/imgview"
)

func (m Model) View() string {
	var b strings.Builder

	b.WriteString(ui.Header("BATTLE  战斗"))
	b.WriteString("\n\n")
	b.WriteString(ui.LoadError(m.loadErr))
	b.WriteString(ui.AudioError(m.ttsErr))

	if m.battle == nil {
		b.WriteString(ui.MutedStyle.Render("  No battle loaded. Press esc to return."))
		return b.String()
	}

	switch m.state {
	case stateIntro:
		b.WriteString(m.viewIntro())
	case stateFighting:
		b.WriteString(m.viewFighting())
	case statePlayerAttack, stateEnemyAttack:
		b.WriteString(m.viewAttackAnim())
	case stateVictory:
		b.WriteString(m.viewVictory())
	case stateDefeat:
		b.WriteString(m.viewDefeat())
	}

	return b.String()
}

func (m Model) viewIntro() string {
	var b strings.Builder
	enemy := m.battle.Enemy

	// Enemy image or ASCII art.
	art := imgview.RenderWithFallback(enemy.ImagePath, 40, 15, enemy.Art)
	for _, line := range strings.Split(art, "\n") {
		b.WriteString(ui.ErrorStyle.Render("      " + line))
		b.WriteString("\n")
	}
	b.WriteString("\n")

	// Enemy name and title.
	b.WriteString("      ")
	b.WriteString(ui.ChineseStyle.Render(enemy.Chinese))
	b.WriteString("  ")
	b.WriteString(ui.MutedStyle.Render(enemy.Name))
	b.WriteString("\n\n")

	b.WriteString(m.renderStakes())
	b.WriteString(ui.HelpStyle.Render("  enter/space begin battle  |  esc flee"))

	return b.String()
}

func (m Model) viewFighting() string {
	var b strings.Builder
	b.WriteString(m.renderHPBars())
	b.WriteString("\n")
	art := imgview.RenderWithFallback(m.battle.Enemy.ImagePath, 30, 10, m.battle.Enemy.Art)
	for _, line := range strings.Split(art, "\n") {
		b.WriteString("      " + line + "\n")
	}
	b.WriteString("\n")
	b.WriteString(m.renderLog())
	b.WriteString("\n")
	if m.battle.Question != nil {
		b.WriteString(m.renderQuestion())
	} else {
		b.WriteString(ui.MutedStyle.Render("  Not enough study cards to form a question. Press esc to flee."))
		b.WriteString("\n")
	}
	b.WriteString("\n")
	if q := m.battle.Question; q != nil && q.Category == "write" {
		b.WriteString(ui.HelpStyle.Render("  type pinyin, space/1-9 pick  |  ctrl+space input mode  |  enter answer  |  esc flee"))
	} else if q != nil && q.Category == "order" {
		b.WriteString(ui.HelpStyle.Render("  1-9 add the next word  |  backspace undo  |  esc flee"))
	} else {
		b.WriteString(ui.HelpStyle.Render("  up/down navigate  |  enter/1-5 answer  |  esc flee"))
	}
	return b.String()
}

func (m Model) viewAttackAnim() string {
	var b strings.Builder
	b.WriteString(m.renderHPBars())
	b.WriteString("\n")
	if len(m.log) > 0 {
		last := m.log[len(m.log)-1]
		b.WriteString("\n")
		b.WriteString(last.style.Bold(true).Render("      " + last.text))
		b.WriteString("\n")

		if q := m.battle.Question; q != nil {
			if m.state == stateEnemyAttack {
				b.WriteString(m.renderMissedAnswer(q))
			}
			b.WriteString(renderWordDetail(q))
		}
	}
	b.WriteString("\n")

	if m.state == stateEnemyAttack {
		b.WriteString(ui.HelpStyle.Render("  enter/space continue  |  esc flee"))
	}

	return b.String()
}

// renderMissedAnswer shows what the player chose beside the correct answer.
func (m Model) renderMissedAnswer(q *rpgengine.CombatQuestion) string {
	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(ui.SubtitleStyle.Render("      " + q.Prompt))
	b.WriteString("\n")
	if m.lastAnswer != "" {
		b.WriteString(ui.ErrorStyle.Render("      ✗ " + m.lastAnswer))
		b.WriteString("\n")
	}
	b.WriteString(ui.SuccessStyle.Render("      ✓ " + q.CorrectAnswer()))
	b.WriteString("\n\n")
	return b.String()
}

// renderWordDetail shows the tested word with its pinyin and meaning.
func renderWordDetail(q *rpgengine.CombatQuestion) string {
	parts := make([]string, 0, 3)
	if q.Chinese != "" {
		parts = append(parts, ui.ChineseStyle.Render(q.Chinese))
	}
	if q.Pinyin != "" {
		parts = append(parts, ui.PinyinStyle.Render(q.Pinyin))
	}
	if q.English != "" {
		parts = append(parts, ui.MutedStyle.Render(q.English))
	}
	if len(parts) == 0 {
		return ""
	}
	return "      " + strings.Join(parts, "  ") + "\n"
}

func (m Model) viewVictory() string {
	var b strings.Builder

	b.WriteString(ui.SuccessStyle.Bold(true).Render("  VICTORY!  胜利!"))
	b.WriteString("\n\n")

	fmt.Fprintf(&b, "  %s defeated!\n\n",
		ui.ChineseStyle.Render(m.battle.Enemy.Chinese))
	if m.stage != nil {
		b.WriteString(ui.WarningStyle.Render(fmt.Sprintf("  Tribulation %d cleared! The road west opens.", m.stage.Number)))
		b.WriteString("\n\n")
	}

	fmt.Fprintf(&b, "  XP Earned:    %d\n", m.battle.XPEarned)
	fmt.Fprintf(&b, "  Turns:        %d\n", m.battle.Turn)
	fmt.Fprintf(&b, "  Max Combo:    %d\n", m.battle.MaxCombo)
	fmt.Fprintf(&b, "  Accuracy:     %d%%\n", m.battle.Accuracy())
	b.WriteString(m.renderSaveStatus())
	b.WriteString("\n")
	b.WriteString(ui.HelpStyle.Render("  enter/space continue  |  esc back"))

	return b.String()
}

func (m Model) viewDefeat() string {
	var b strings.Builder

	b.WriteString(ui.ErrorStyle.Bold(true).Render("  DEFEAT  败北"))
	b.WriteString("\n\n")

	b.WriteString(ui.MutedStyle.Render("  The demon overwhelms you..."))
	b.WriteString("\n")
	b.WriteString(ui.PinyinStyle.Render("  失败乃成功之母 -- Failure is the mother of success."))
	b.WriteString("\n\n")

	fmt.Fprintf(&b, "  Turns survived:  %d\n", m.battle.Turn)
	fmt.Fprintf(&b, "  Max Combo:       %d\n", m.battle.MaxCombo)
	fmt.Fprintf(&b, "  Damage Dealt:    %d\n", m.battle.TotalDealt)
	fmt.Fprintf(&b, "  XP Earned:       %d\n", m.battle.XPEarned)
	b.WriteString(m.renderSaveStatus())
	b.WriteString("\n")
	b.WriteString(ui.HelpStyle.Render("  enter/space try again  |  esc back"))

	return b.String()
}

// renderSaveStatus reports lifetime XP once the battle is saved, or why
// saving failed.
func (m Model) renderSaveStatus() string {
	if m.saveErr != nil {
		return ui.ErrorStyle.Render(fmt.Sprintf("  Could not save results: %v", m.saveErr)) + "\n"
	}
	if !m.saved {
		return ""
	}
	return ui.SuccessStyle.Render(fmt.Sprintf("  Total XP:     %d", m.totalXP)) + "\n"
}

func (m Model) renderHPBars() string {
	var b strings.Builder
	p := m.battle.Player
	e := m.battle.Enemy
	barWidth := 20

	b.WriteString("  ")
	b.WriteString(hpBar(p.Chinese, p.HP, p.MaxHP, barWidth))
	b.WriteString("    ")
	b.WriteString(hpBar(e.Chinese, e.HP, e.MaxHP, barWidth))
	b.WriteString("\n")

	return b.String()
}

func hpBar(name string, hp, maxHP, width int) string {
	if maxHP <= 0 {
		maxHP = 1
	}
	if hp < 0 {
		hp = 0
	}
	ratio := float64(hp) / float64(maxHP)
	filled := int(ratio * float64(width))
	if filled > width {
		filled = width
	}
	empty := width - filled

	// Color by HP percentage.
	var barStyle lipgloss.Style
	switch {
	case ratio > 0.5:
		barStyle = lipgloss.NewStyle().Foreground(ui.ColorSuccess)
	case ratio > 0.25:
		barStyle = lipgloss.NewStyle().Foreground(ui.ColorWarning)
	default:
		barStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#C54B3C"))
	}

	bar := barStyle.Render(strings.Repeat("█", filled)) +
		ui.MutedStyle.Render(strings.Repeat("░", empty))

	label := ui.ChineseStyle.Render(name)
	hpText := ui.MutedStyle.Render(fmt.Sprintf(" %d/%d", hp, maxHP))

	return label + " " + bar + hpText
}

func (m Model) renderLog() string {
	var b strings.Builder
	start := 0
	if len(m.log) > 3 {
		start = len(m.log) - 3
	}
	for _, entry := range m.log[start:] {
		b.WriteString("  ")
		b.WriteString(entry.style.Render(entry.text))
		b.WriteString("\n")
	}
	return b.String()
}

func appendLog(log []logEntry, entry logEntry) []logEntry {
	const maxLog = 20
	log = append(log, entry)
	if len(log) > maxLog {
		log = log[len(log)-maxLog:]
	}
	return log
}

func (m Model) renderQuestion() string {
	q := m.battle.Question
	var b strings.Builder

	b.WriteString(ui.SubtitleStyle.Render("  " + q.Prompt))
	b.WriteString("\n\n")
	if q.Category == "write" {
		return b.String() + "  " + m.input.View() + "\n"
	}
	if q.Category == "order" {
		built := m.orderText()
		if built == "" {
			built = "…"
		}
		fmt.Fprintf(&b, "  %s %s\n\n  ", ui.MutedStyle.Render("Your sentence:"), ui.ChineseStyle.Render(built))
		for i, p := range q.Pieces {
			piece := fmt.Sprintf("%d.%s  ", i+1, p)
			if slices.Contains(m.order, i) {
				b.WriteString(ui.MutedStyle.Render(piece))
			} else {
				b.WriteString(ui.ChineseStyle.Render(piece))
			}
		}
		return b.String() + "\n"
	}

	for i, choice := range q.Choices {
		prefix := fmt.Sprintf("    %d. %s", i+1, choice)
		if i == m.battle.Selected {
			b.WriteString(ui.SelectedStyle.Render("  ▸ "+prefix) + "\n")
		} else {
			b.WriteString(ui.UnselectedStyle.Render("    "+prefix) + "\n")
		}
	}

	return b.String()
}

func comboLabel(combo int) string {
	switch {
	case combo >= 10:
		return "连击!太厉害了!"
	case combo >= 7:
		return "连击!"
	case combo >= 5:
		return "连击!"
	case combo >= 3:
		return "连击!"
	default:
		return ""
	}
}

// renderStakes says what the battle is for: clearing the open stage (a
// boss's, maybe), or practice while the next stage still needs mastery.
func (m Model) renderStakes() string {
	var b strings.Builder
	switch {
	case m.stage != nil:
		fmt.Fprintf(&b, "  %s %s %s\n", ui.MutedStyle.Render(fmt.Sprintf("Tribulation %d ·", m.stage.Number)),
			ui.ChineseStyle.Render(m.stage.Name), m.stage.English)
		if m.stage.Boss != nil {
			b.WriteString(ui.WarningStyle.Render(fmt.Sprintf("  BOSS! It will test your weakest words. Theme: %s", m.stage.Boss.Theme)))
			b.WriteString("\n")
		}
		b.WriteString(ui.WarningStyle.Render("  Win to clear this tribulation."))
	case m.next != nil:
		b.WriteString(ui.WarningStyle.Render("  A demon blocks your path! (practice)"))
		fmt.Fprintf(&b, "\n  %s", ui.MutedStyle.Render(fmt.Sprintf("Master %d more characters to open tribulation %d, %s.",
			m.next.Unlock-m.mastered, m.next.Number, m.next.English)))
	default:
		b.WriteString(ui.WarningStyle.Render("  All 81 tribulations cleared. Practice freely."))
	}
	return b.String() + "\n\n"
}
