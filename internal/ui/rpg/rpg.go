package rpg

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/cj/mandarin-wiki/internal/model"
	rpgengine "github.com/cj/mandarin-wiki/internal/rpg"
	"github.com/cj/mandarin-wiki/internal/ui"
	"github.com/cj/mandarin-wiki/internal/ui/imgview"
	"github.com/cj/mandarin-wiki/internal/wiki"
)

type viewState int

const (
	stateIntro viewState = iota
	stateFighting
	statePlayerAttack
	stateEnemyAttack
	stateVictory
	stateDefeat
)

// animTickMsg drives brief animation pauses between attack states.
type animTickMsg struct{}

// Model is the RPG battle view model.
type Model struct {
	store  *wiki.Store
	gkeys  ui.GlobalKeys
	battle *rpgengine.Battle
	cards  []model.Card
	state  viewState
	width  int
	height int

	// Animation timer: how many ticks remain before advancing state.
	animTicks int

	// Combat log lines (most recent last).
	log []logEntry
}

type logEntry struct {
	text  string
	style lipgloss.Style
}

// New creates a new RPG battle model.
func New(store *wiki.Store) Model {
	return Model{
		store: store,
		gkeys: ui.DefaultGlobalKeys(),
		state: stateIntro,
	}
}

// Init satisfies the Bubble Tea model interface.
func (m Model) Init() tea.Cmd {
	return nil
}

// SetSize updates terminal dimensions.
func (m Model) SetSize(w, h int) Model {
	m.width = w
	m.height = h
	return m
}

// Refresh reloads battle state with fresh cards from the wiki.
func (m Model) Refresh() Model {
	m.cards = loadCards(m.store)
	mastered := countMastered(m.store)
	m.battle = rpgengine.NewRandomBattle(m.cards, mastered)
	m.state = stateIntro
	m.log = nil
	m.animTicks = 0
	return m
}

// loadCards pulls drillable cards from the wiki store.
func loadCards(store *wiki.Store) []model.Card {
	pages, err := store.ListPages()
	if err != nil {
		return nil
	}
	var drillable []model.Page
	for _, p := range pages {
		switch p.Type {
		case "character", "vocabulary", "grammar":
			drillable = append(drillable, p)
		}
	}
	return wiki.CardsFromPages(drillable)
}

// countMastered returns the count of cards with a review interval >= 21 days.
func countMastered(store *wiki.Store) int {
	pages, err := store.ListPages()
	if err != nil {
		return 0
	}
	count := 0
	for _, p := range pages {
		if p.ReviewInterval >= 21 {
			count++
		}
	}
	return count
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case animTickMsg:
		m.animTicks--
		if m.animTicks > 0 {
			return m, tickAfter(120 * time.Millisecond)
		}
		// Animation finished — advance to the next logical state.
		return m.advanceAfterAnim()

	case tea.KeyMsg:
		// Global escape: flee battle.
		if key.Matches(msg, m.gkeys.Back) {
			return m, ui.SwitchTo(ui.ViewDashboard)
		}

		switch m.state {
		case stateIntro:
			if key.Matches(msg, m.gkeys.Enter) || msg.Type == tea.KeySpace {
				m.state = stateFighting
				m.battle.State = rpgengine.BattlePlayerTurn
				q := m.battle.GenerateQuestion(m.cards)
				m.battle.Question = q
				return m, nil
			}

		case stateFighting:
			return m.handleFightingKey(msg)

		case statePlayerAttack, stateEnemyAttack:
			// Ignore input during animations.
			return m, nil

		case stateVictory, stateDefeat:
			if key.Matches(msg, m.gkeys.Enter) || msg.Type == tea.KeySpace {
				return m, ui.SwitchTo(ui.ViewDashboard)
			}
		}
	}
	return m, nil
}

func (m Model) handleFightingKey(msg tea.KeyMsg) (Model, tea.Cmd) {
	if m.battle == nil || m.battle.Question == nil {
		return m, nil
	}
	q := m.battle.Question
	numChoices := len(q.Choices)

	switch {
	case key.Matches(msg, m.gkeys.Up):
		if m.battle.Selected > 0 {
			m.battle.Selected--
		}
	case key.Matches(msg, m.gkeys.Down):
		if m.battle.Selected < numChoices-1 {
			m.battle.Selected++
		}
	case key.Matches(msg, m.gkeys.Enter):
		return m.submitAnswer()
	default:
		// Number keys 1-5 for direct selection.
		if s := msg.String(); len(s) == 1 && s[0] >= '1' && s[0] <= '5' {
			idx := int(s[0]-'1')
			if idx < numChoices {
				m.battle.Selected = idx
				return m.submitAnswer()
			}
		}
	}
	return m, nil
}

func (m Model) submitAnswer() (Model, tea.Cmd) {
	correct, dmg := m.battle.Answer(m.battle.Selected)
	if correct {
		m.state = statePlayerAttack
		m.battle.State = rpgengine.BattlePlayerAttack
		entry := logEntry{
			text:  fmt.Sprintf("You strike %s for %d damage!", m.battle.Enemy.Chinese, dmg),
			style: ui.SuccessStyle,
		}
		m.log = appendLog(m.log, entry)
		if m.battle.Combo > 1 {
			comboEntry := logEntry{
				text:  fmt.Sprintf("COMBO x%d! %s", m.battle.Combo, comboLabel(m.battle.Combo)),
				style: ui.WarningStyle,
			}
			m.log = appendLog(m.log, comboEntry)
		}
	} else {
		m.state = stateEnemyAttack
		m.battle.State = rpgengine.BattleEnemyAttack
		entry := logEntry{
			text:  fmt.Sprintf("%s attacks! -%d HP", m.battle.Enemy.Chinese, dmg),
			style: ui.ErrorStyle,
		}
		m.log = appendLog(m.log, entry)
	}
	m.animTicks = 3
	return m, tickAfter(120 * time.Millisecond)
}

func (m Model) advanceAfterAnim() (Model, tea.Cmd) {
	if m.battle.IsOver() {
		if m.battle.PlayerWon() {
			m.state = stateVictory
			m.battle.State = rpgengine.BattleVictory
		} else {
			m.state = stateDefeat
			m.battle.State = rpgengine.BattleDefeat
		}
		return m, nil
	}
	// Back to player turn with a new question.
	m.state = stateFighting
	m.battle.State = rpgengine.BattlePlayerTurn
	m.battle.Selected = 0
	q := m.battle.GenerateQuestion(m.cards)
	m.battle.Question = q
	return m, nil
}

// tickAfter returns a Cmd that sends animTickMsg after a delay.
func tickAfter(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(_ time.Time) tea.Msg {
		return animTickMsg{}
	})
}

func (m Model) View() string {
	var b strings.Builder

	b.WriteString(ui.Header("BATTLE  战斗"))
	b.WriteString("\n\n")

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

	b.WriteString(ui.WarningStyle.Render("  A demon blocks your path!"))
	b.WriteString("\n\n")
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
	}
	b.WriteString("\n")
	b.WriteString(ui.HelpStyle.Render("  up/down navigate  |  enter/1-5 answer  |  esc flee"))
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

		// If player attack was correct, show pinyin + meaning.
		if m.state == statePlayerAttack && m.battle.Question != nil {
			q := m.battle.Question
			if q.Chinese != "" || q.Pinyin != "" {
				detail := "      "
				if q.Chinese != "" {
					detail += ui.ChineseStyle.Render(q.Chinese) + " "
				}
				if q.Pinyin != "" {
					detail += ui.PinyinStyle.Render(q.Pinyin)
				}
				b.WriteString(detail + "\n")
			}
		}
	}
	b.WriteString("\n")

	return b.String()
}

func (m Model) viewVictory() string {
	var b strings.Builder

	b.WriteString(ui.SuccessStyle.Bold(true).Render("  VICTORY!  胜利!"))
	b.WriteString("\n\n")

	b.WriteString(fmt.Sprintf("  %s defeated!\n\n",
		ui.ChineseStyle.Render(m.battle.Enemy.Chinese)))

	b.WriteString(fmt.Sprintf("  XP Earned:    %d\n", m.battle.XPEarned))
	b.WriteString(fmt.Sprintf("  Turns:        %d\n", m.battle.Turn))
	b.WriteString(fmt.Sprintf("  Max Combo:    %d\n", m.battle.MaxCombo))
	totalDmg := m.battle.DamageDealt + m.battle.DamageTaken
	if totalDmg > 0 {
		accuracy := float64(m.battle.DamageDealt) / float64(totalDmg) * 100
		b.WriteString(fmt.Sprintf("  Accuracy:     %.0f%%\n", accuracy))
	}
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

	b.WriteString(fmt.Sprintf("  Turns survived:  %d\n", m.battle.Turn))
	b.WriteString(fmt.Sprintf("  Max Combo:       %d\n", m.battle.MaxCombo))
	b.WriteString(fmt.Sprintf("  Damage Dealt:    %d\n", m.battle.DamageDealt))
	b.WriteString("\n")
	b.WriteString(ui.HelpStyle.Render("  enter/space try again  |  esc back"))

	return b.String()
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
