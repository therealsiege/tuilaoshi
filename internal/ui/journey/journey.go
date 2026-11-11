package journey

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/cj/mandarin-wiki/internal/journey"
	"github.com/cj/mandarin-wiki/internal/ui"
	"github.com/cj/mandarin-wiki/internal/ui/imgview"
	"github.com/cj/mandarin-wiki/internal/wiki"
)

// Model is the Journey to the West gamification view.
type Model struct {
	store          *wiki.Store
	state          journey.State
	tab            int // 0=journey, 1=party, 2=bosses, 3=artifacts, 4=quests, 5=achievements, 6=lore, 7=bestiary
	width          int
	height         int
	bestiaryCursor int // selected demon index in the bestiary tab
}

// New creates a new journey view model.
func New(store *wiki.Store) Model {
	return Model{
		store: store,
		state: journey.Calculate(store),
		tab:   0,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab":
			m.tab = (m.tab + 1) % 8
			return m, nil
		case "left", "h":
			m.tab = (m.tab - 1 + 8) % 8
			return m, nil
		case "right", "l":
			m.tab = (m.tab + 1) % 8
			return m, nil
		case "up", "k":
			if m.tab == 7 {
				entries := journey.Bestiary()
				if len(entries) > 0 {
					m.bestiaryCursor = (m.bestiaryCursor - 1 + len(entries)) % len(entries)
				}
				return m, nil
			}
		case "down", "j":
			if m.tab == 7 {
				entries := journey.Bestiary()
				if len(entries) > 0 {
					m.bestiaryCursor = (m.bestiaryCursor + 1) % len(entries)
				}
				return m, nil
			}
		case "esc":
			return m, ui.SwitchTo(ui.ViewDashboard)
		}
	}
	return m, nil
}

func (m Model) View() string {
	var b strings.Builder

	// Header
	b.WriteString(ui.Header("西游记 Journey to the West"))
	b.WriteString("\n\n")

	// Tab bar
	b.WriteString(m.renderTabBar())
	b.WriteString("\n\n")

	// Content based on active tab
	switch m.tab {
	case 0:
		b.WriteString(m.renderJourney())
	case 1:
		b.WriteString(m.renderParty())
	case 2:
		b.WriteString(m.renderBosses())
	case 3:
		b.WriteString(m.renderArtifacts())
	case 4:
		b.WriteString(m.renderQuests())
	case 5:
		b.WriteString(m.renderAchievements())
	case 6:
		b.WriteString(m.renderLore())
	case 7:
		b.WriteString(m.renderBestiary())
	}

	return b.String()
}

// SetSize updates the terminal dimensions.
func (m Model) SetSize(w, h int) Model {
	m.width = w
	m.height = h
	return m
}

// Refresh reloads data from the wiki store.
func (m Model) Refresh() Model {
	m.state = journey.Calculate(m.store)
	return m
}

// tabFooter returns the shared footer for all tab views.
func (m Model) tabFooter() string {
	return "\n" + ui.Divider(40) + "\n" + ui.HelpStyle.Render("  tab/←→: switch tabs | esc: dashboard")
}

// renderTabBar renders the tab navigation bar.
func (m Model) renderTabBar() string {
	tabs := []string{"journey", "party", "bosses", "artifacts", "quests", "achievements", "lore", "bestiary"}
	var rendered []string

	for i, tab := range tabs {
		if i == m.tab {
			rendered = append(rendered, ui.ActiveTabStyle.Render(tab))
		} else {
			rendered = append(rendered, ui.TabStyle.Render(tab))
		}
	}

	return "  " + strings.Join(rendered, " | ")
}

// renderJourney renders the main journey map view (tab 0).
func (m Model) renderJourney() string {
	var b strings.Builder

	trib := m.state.Progress.Tribulation
	current := journey.TribulationAt(trib)

	// Current tribulation name prominently at the top
	if current != nil {
		tribTitle := fmt.Sprintf("  第%d难: %s (%s)",
			current.Number, current.Name, current.English)
		b.WriteString(ui.ChineseStyle.Render(tribTitle))
		b.WriteString("\n")
		b.WriteString(ui.MutedStyle.Render("  " + current.Desc))
		b.WriteString("\n\n")
	}

	// Cultivation Rank section
	b.WriteString(ui.SubtitleStyle.Render("  修为 Cultivation Rank"))
	b.WriteString("\n")

	rankBox := m.renderRankBox()
	b.WriteString(ui.BorderStyle.Width(50).Render(rankBox))
	b.WriteString("\n\n")

	// Pilgrimage progress section — scrollable tribulation list
	b.WriteString(ui.SubtitleStyle.Render("  八十一难 The 81 Tribulations"))
	b.WriteString("\n\n")

	all := journey.AllTribulations()
	windowSize := 10

	// Center the window around the current tribulation (0-indexed)
	currentIdx := trib - 1
	if currentIdx < 0 {
		currentIdx = 0
	}
	startIdx := currentIdx - windowSize/2
	if startIdx < 0 {
		startIdx = 0
	}
	if startIdx+windowSize > len(all) {
		startIdx = len(all) - windowSize
	}
	if startIdx < 0 {
		startIdx = 0
	}
	endIdx := startIdx + windowSize
	if endIdx > len(all) {
		endIdx = len(all)
	}

	// Ellipsis for earlier tribulations
	if startIdx > 0 {
		b.WriteString(ui.MutedStyle.Render(fmt.Sprintf("       ... %d more above ...", startIdx)))
		b.WriteString("\n")
	}

	for i := startIdx; i < endIdx; i++ {
		t := all[i]
		line := fmt.Sprintf("%2d. %s — %s", t.Number, t.Name, t.English)

		if t.Number < trib {
			// Past: completed
			b.WriteString(ui.SuccessStyle.Render("  \u2713 " + line))
		} else if t.Number == trib {
			// Current: highlighted
			b.WriteString(ui.SelectedStyle.Render("  \u25b8 " + line))
		} else {
			// Future: muted
			b.WriteString(ui.MutedStyle.Render("  \u00b7 " + line))
		}
		b.WriteString("\n")
	}

	// Ellipsis for later tribulations
	remaining := len(all) - endIdx
	if remaining > 0 {
		b.WriteString(ui.MutedStyle.Render(fmt.Sprintf("       ... %d more below ...", remaining)))
		b.WriteString("\n")
	}

	b.WriteString("\n")

	// Progress summary
	completed := trib - 1
	if completed < 0 {
		completed = 0
	}
	progressText := fmt.Sprintf("  %d / 81 tribulations completed  \u00b7  %d/%d characters mastered (%.0f%%)",
		completed,
		m.state.Progress.Mastered,
		m.state.Progress.TotalChars,
		m.state.Progress.Percent)
	b.WriteString(ui.MutedStyle.Render(progressText))
	b.WriteString("\n\n")

	b.WriteString(m.tabFooter())

	return b.String()
}

// renderRankBox renders the cultivation rank progress box.
func (m Model) renderRankBox() string {
	var b strings.Builder

	// Current rank with Chinese and English
	rankLine := fmt.Sprintf("%s \u00b7 %s",
		ui.ChineseStyle.Render(m.state.Rank.Name),
		m.state.Rank.English)
	b.WriteString(rankLine)
	b.WriteString("\n")

	// Progress bar
	barWidth := 40
	progressBar := ui.ProgressBar(m.state.Rank.Current, m.state.Rank.Required, barWidth)
	b.WriteString(progressBar)
	b.WriteString(ui.MutedStyle.Render(fmt.Sprintf(" %d/%d", m.state.Rank.Current, m.state.Rank.Required)))
	b.WriteString("\n")

	// Next rank
	if m.state.Rank.NextName != "" {
		nextText := fmt.Sprintf("Next: %s %s",
			m.state.Rank.NextName,
			ui.MutedStyle.Render(m.state.Rank.NextEnglish))
		b.WriteString(ui.MutedStyle.Render(nextText))
	} else {
		b.WriteString(ui.SuccessStyle.Render("\u2713 Max cultivation achieved!"))
	}

	return b.String()
}

// renderParty renders the companions/party view (tab 1).
func (m Model) renderParty() string {
	var b strings.Builder

	b.WriteString(ui.SubtitleStyle.Render("  同伴 Companions"))
	b.WriteString("\n\n")

	if len(m.state.Companions) == 0 {
		b.WriteString(ui.MutedStyle.Render("  No companions yet..."))
		b.WriteString(m.tabFooter())
		return b.String()
	}

	for _, companion := range m.state.Companions {
		b.WriteString("  ")
		if companion.Unlocked {
			// Unlocked companion - show in success color
			b.WriteString(ui.SuccessStyle.Render("✓ "))
			b.WriteString(companion.Emoji + " ")
			b.WriteString(ui.ChineseStyle.Render(companion.Name))
			b.WriteString(" · " + companion.English)
			b.WriteString("\n")
			b.WriteString(ui.MutedStyle.Render(fmt.Sprintf("     %s", companion.Role)))
			if companion.Ability != "" {
				b.WriteString("\n")
				b.WriteString(fmt.Sprintf("     %s %s: %s",
					companion.AbilityIcon,
					ui.SuccessStyle.Render(companion.Ability),
					companion.AbilityDesc))
			}
		} else {
			// Locked companion
			b.WriteString(ui.MutedStyle.Render("○ "))
			b.WriteString(ui.MutedStyle.Render(companion.Emoji + " "))
			b.WriteString(ui.MutedStyle.Render(companion.Name + " · " + companion.English))
			b.WriteString("\n")
			b.WriteString(ui.MutedStyle.Render(fmt.Sprintf("     Unlock at %d characters", companion.UnlockAt)))
			if companion.Ability != "" {
				b.WriteString("\n")
				b.WriteString(ui.MutedStyle.Render(fmt.Sprintf("     %s %s: %s",
					companion.AbilityIcon, companion.Ability, companion.AbilityDesc)))
			}
		}
		b.WriteString("\n")
	}

	b.WriteString(m.tabFooter())

	return b.String()
}

// renderBosses renders the boss challenges view (tab 2).
func (m Model) renderBosses() string {
	var b strings.Builder

	b.WriteString(ui.SubtitleStyle.Render("  妖怪 Boss Challenges"))
	b.WriteString("\n\n")

	if len(m.state.Bosses) == 0 {
		b.WriteString(ui.MutedStyle.Render("  No bosses yet..."))
		b.WriteString(m.tabFooter())
		return b.String()
	}

	mastered := m.state.Progress.Mastered
	for i, boss := range m.state.Bosses {
		if i > 0 {
			b.WriteString("\n")
		}
		b.WriteString("  ")
		if boss.Defeated {
			// Defeated boss
			b.WriteString(ui.SuccessStyle.Render("✓ "))
			b.WriteString(ui.ChineseStyle.Render(boss.Name))
			b.WriteString(" · " + boss.English)
			b.WriteString("\n")
			b.WriteString(ui.SuccessStyle.Render(fmt.Sprintf("     Defeated! Theme: %s", boss.Theme)))
		} else {
			// Upcoming or current boss
			isNext := m.state.NextBoss != nil && m.state.NextBoss.Name == boss.Name
			if isNext {
				b.WriteString(ui.WarningStyle.Render("⚔ "))
				b.WriteString(ui.WarningStyle.Render(boss.Name + " · " + boss.English))
				b.WriteString("\n")
				b.WriteString(ui.WarningStyle.Render(fmt.Sprintf("     NEXT BOSS! Theme: %s", boss.Theme)))
				b.WriteString("\n")
				b.WriteString("     " + ui.ProgressBar(mastered, boss.Required, 20))
				b.WriteString(ui.MutedStyle.Render(fmt.Sprintf(" %d/%d", mastered, boss.Required)))
				remaining := boss.Required - mastered
				if remaining > 0 {
					b.WriteString("\n")
					b.WriteString(ui.MutedStyle.Render(fmt.Sprintf("     Need %d more characters", remaining)))
				}
			} else {
				b.WriteString(ui.MutedStyle.Render("○ "))
				b.WriteString(ui.MutedStyle.Render(boss.Name + " · " + boss.English))
				b.WriteString("\n")
				b.WriteString(ui.MutedStyle.Render(fmt.Sprintf("     Unlock at %d characters · Theme: %s", boss.Required, boss.Theme)))
			}
		}
		b.WriteString("\n")
	}

	b.WriteString(m.tabFooter())

	return b.String()
}

// renderArtifacts renders the artifacts/achievements view (tab 3).
func (m Model) renderArtifacts() string {
	var b strings.Builder

	b.WriteString(ui.SubtitleStyle.Render("  法宝 Sacred Artifacts"))
	b.WriteString("\n\n")

	if len(m.state.Artifacts) == 0 {
		b.WriteString(ui.MutedStyle.Render("  No artifacts yet..."))
		b.WriteString(m.tabFooter())
		return b.String()
	}

	for _, artifact := range m.state.Artifacts {
		b.WriteString("  ")
		if artifact.Earned {
			// Earned artifact
			b.WriteString(ui.SuccessStyle.Render("✓ "))
			b.WriteString(ui.ChineseStyle.Render(artifact.Chinese))
			b.WriteString(" · " + artifact.Name)
			b.WriteString("\n")
			b.WriteString(ui.MutedStyle.Render(fmt.Sprintf("     %s", artifact.Desc)))
		} else {
			// Not yet earned
			b.WriteString(ui.MutedStyle.Render("○ "))
			b.WriteString(ui.MutedStyle.Render(artifact.Chinese + " · " + artifact.Name))
			b.WriteString("\n")
			b.WriteString(ui.MutedStyle.Render(fmt.Sprintf("     %s", artifact.Desc)))
			b.WriteString("\n")
			if artifact.Kind == "streak" {
				b.WriteString(ui.MutedStyle.Render(fmt.Sprintf("     Earn with %d-day streak", artifact.EarnAt)))
			} else {
				b.WriteString(ui.MutedStyle.Render(fmt.Sprintf("     Earn at %d characters", artifact.EarnAt)))
			}
		}
		b.WriteString("\n")
	}

	b.WriteString(m.tabFooter())

	return b.String()
}

// renderQuests renders the daily quests view (tab 4).
func (m Model) renderQuests() string {
	var b strings.Builder

	b.WriteString(ui.SubtitleStyle.Render("  每日任务 Daily Quests"))
	b.WriteString("\n\n")

	if len(m.state.Quests) == 0 {
		b.WriteString(ui.MutedStyle.Render("  No quests available..."))
		b.WriteString(m.tabFooter())
		return b.String()
	}

	barWidth := 40
	for i, quest := range m.state.Quests {
		if i > 0 {
			b.WriteString("\n")
		}
		b.WriteString("  ")

		// Determine quest style and icon
		if quest.Done {
			b.WriteString(ui.SuccessStyle.Render("✓ " + quest.Desc))
			b.WriteString("\n")
			b.WriteString(ui.SuccessStyle.Render(fmt.Sprintf("    %s · %d/%d",
				quest.Chinese, quest.Current, quest.Target)))
			b.WriteString("\n")
			b.WriteString("    " + ui.ProgressBar(quest.Current, quest.Target, barWidth))
			b.WriteString(ui.MutedStyle.Render(fmt.Sprintf(" %d/%d", quest.Current, quest.Target)))
		} else if quest.Name == "Demon Slayer" {
			b.WriteString(ui.WarningStyle.Render("⚔ " + quest.Desc))
			b.WriteString("\n")
			b.WriteString(ui.WarningStyle.Render(fmt.Sprintf("    %s · %d/%d characters",
				quest.Chinese, quest.Current, quest.Target)))
			b.WriteString("\n")
			b.WriteString("    " + ui.ProgressBar(quest.Current, quest.Target, barWidth))
			b.WriteString(ui.MutedStyle.Render(fmt.Sprintf(" %d/%d", quest.Current, quest.Target)))
		} else {
			b.WriteString(ui.MutedStyle.Render("○ " + quest.Desc))
			b.WriteString("\n")
			unit := "reviews"
			if quest.Name == "Steady Pilgrim" {
				unit = "days"
			}
			b.WriteString(ui.MutedStyle.Render(fmt.Sprintf("    %s · %d/%d %s",
				quest.Chinese, quest.Current, quest.Target, unit)))
			b.WriteString("\n")
			b.WriteString("    " + ui.ProgressBar(quest.Current, quest.Target, barWidth))
			b.WriteString(ui.MutedStyle.Render(fmt.Sprintf(" %d/%d", quest.Current, quest.Target)))
		}
		b.WriteString("\n")
	}

	b.WriteString(m.tabFooter())

	return b.String()
}

// renderAchievements renders the achievements view (tab 5).
func (m Model) renderAchievements() string {
	var b strings.Builder

	b.WriteString(ui.SubtitleStyle.Render("  成就 Achievements"))
	b.WriteString("\n\n")

	achievements := m.state.Achievements
	earned := journey.EarnedCount(achievements)
	total := len(achievements)

	// Summary line
	b.WriteString(fmt.Sprintf("  %d / %d earned", earned, total))
	b.WriteString("\n\n")

	// Group by category and render each group
	categories := []struct {
		key   string
		label string
	}{
		{"literacy", "Literacy"},
		{"dedication", "Dedication"},
		{"mastery", "Mastery"},
		{"journey", "Journey"},
	}

	for _, cat := range categories {
		var group []journey.Achievement
		for _, a := range achievements {
			if a.Category == cat.key {
				group = append(group, a)
			}
		}
		if len(group) == 0 {
			continue
		}

		catEarned := 0
		for _, a := range group {
			if a.Earned {
				catEarned++
			}
		}
		header := fmt.Sprintf("  -- %s (%d/%d) --", cat.label, catEarned, len(group))
		b.WriteString(ui.SubtitleStyle.Render(header))
		b.WriteString("\n\n")

		for _, a := range group {
			b.WriteString("  ")
			if a.Earned {
				b.WriteString(ui.SuccessStyle.Render("✓ "))
				b.WriteString(ui.ChineseStyle.Render(a.Chinese))
				b.WriteString(" · " + a.Name)
				b.WriteString("\n")
				b.WriteString(ui.MutedStyle.Render(fmt.Sprintf("     %s", a.Desc)))
			} else if a.Hidden {
				b.WriteString(ui.MutedStyle.Render("? "))
				b.WriteString(ui.MutedStyle.Render("??? — ???"))
			} else {
				b.WriteString(ui.MutedStyle.Render("· "))
				b.WriteString(ui.MutedStyle.Render(a.Chinese + " · " + a.Name))
				b.WriteString("\n")
				b.WriteString(ui.MutedStyle.Render(fmt.Sprintf("     %s", a.Desc)))
			}
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}

	// Hidden achievements remaining count
	hiddenRemaining := journey.HiddenUnearnedCount(achievements)
	if hiddenRemaining > 0 {
		b.WriteString(ui.MutedStyle.Render(fmt.Sprintf("  %d hidden achievements remaining", hiddenRemaining)))
		b.WriteString("\n\n")
	}

	b.WriteString(m.tabFooter())

	return b.String()
}

// renderLore renders the story passages view (tab 6).
func (m Model) renderLore() string {
	var b strings.Builder

	b.WriteString(ui.SubtitleStyle.Render("  西游故事 Story Passages"))
	b.WriteString("\n\n")

	if len(m.state.Lore) == 0 {
		b.WriteString(ui.MutedStyle.Render("  No lore entries yet..."))
		b.WriteString(m.tabFooter())
		return b.String()
	}

	unlocked := 0
	for _, entry := range m.state.Lore {
		if entry.Unlocked {
			unlocked++
		}
	}
	summary := fmt.Sprintf("  %d/%d passages unlocked", unlocked, len(m.state.Lore))
	b.WriteString(ui.MutedStyle.Render(summary))
	b.WriteString("\n\n")

	for i, entry := range m.state.Lore {
		if i > 0 {
			b.WriteString("\n")
			b.WriteString(ui.Divider(50))
			b.WriteString("\n\n")
		}

		if entry.Unlocked {
			// Title line: Chinese title + English title
			titleLine := fmt.Sprintf("  %s  %s",
				ui.ChineseStyle.Render(entry.Chinese),
				entry.Title)
			b.WriteString(titleLine)
			b.WriteString("\n")
			b.WriteString(ui.MutedStyle.Render(fmt.Sprintf("  Tribulation %d", entry.Tribulation)))
			b.WriteString("\n\n")

			// Chinese passage
			b.WriteString("  " + ui.ChineseStyle.Render(entry.Passage))
			b.WriteString("\n\n")

			// Pinyin
			b.WriteString("  " + ui.PinyinStyle.Render(entry.Pinyin))
			b.WriteString("\n\n")

			// English translation
			b.WriteString("  " + ui.MutedStyle.Render(entry.English))
			b.WriteString("\n")
		} else {
			// Locked entry
			lockText := fmt.Sprintf("  ??? -- Reach tribulation %d to unlock", entry.Tribulation)
			b.WriteString(ui.MutedStyle.Render(lockText))
			b.WriteString("\n")
		}
	}

	b.WriteString(m.tabFooter())

	return b.String()
}

// renderBestiary renders the demon catalog view (tab 7).
func (m Model) renderBestiary() string {
	var b strings.Builder

	b.WriteString(ui.TitleStyle.Render("  妖怪图鉴 Demon Bestiary"))
	b.WriteString("\n\n")

	entries := journey.Bestiary()
	if len(entries) == 0 {
		b.WriteString(ui.MutedStyle.Render("  No demons cataloged..."))
		b.WriteString(m.tabFooter())
		return b.String()
	}

	// Build a set of defeated boss names for quick lookup.
	defeated := make(map[string]bool)
	for _, boss := range m.state.Bosses {
		if boss.Defeated {
			defeated[boss.Name] = true
		}
	}

	// Clamp cursor.
	cursor := m.bestiaryCursor
	if cursor >= len(entries) {
		cursor = len(entries) - 1
	}
	if cursor < 0 {
		cursor = 0
	}

	// LEFT pane: scrollable demon name list.
	var left strings.Builder
	for i, entry := range entries {
		mark := "·"
		if defeated[entry.Name] {
			mark = ui.SuccessStyle.Render("✓")
		}

		name := fmt.Sprintf("%s %s  %s", mark, entry.Name, entry.English)
		if i == cursor {
			left.WriteString(ui.SelectedStyle.Render(" > " + name))
		} else {
			left.WriteString("   " + name)
		}
		left.WriteString("\n")
	}

	// RIGHT pane: detail panel for the selected demon.
	selected := entries[cursor]
	var right strings.Builder

	// Demon image: try terminal image first, fall back to ASCII art.
	rendered := imgview.RenderWithFallback(selected.ImagePath, 30, 15, selected.Art)
	if rendered != "" {
		for _, line := range strings.Split(rendered, "\n") {
			right.WriteString("  " + line + "\n")
		}
		right.WriteString("\n")
	}

	// Title / epithet
	right.WriteString("  ")
	right.WriteString(ui.ChineseStyle.Render(selected.Title))
	right.WriteString("  ")
	right.WriteString(ui.MutedStyle.Render(selected.TitleEng))
	right.WriteString("\n\n")

	// Origin
	right.WriteString(ui.SubtitleStyle.Render("  Origin"))
	right.WriteString("\n")
	right.WriteString("  " + ui.MutedStyle.Render(ui.WrapText(selected.Origin, 44)))
	right.WriteString("\n\n")

	// Powers
	right.WriteString(ui.SubtitleStyle.Render("  Powers"))
	right.WriteString("\n")
	right.WriteString("  " + ui.MutedStyle.Render(ui.WrapText(selected.Powers, 44)))
	right.WriteString("\n\n")

	// Weakness
	right.WriteString(ui.SubtitleStyle.Render("  Weakness"))
	right.WriteString("\n")
	right.WriteString("  " + ui.MutedStyle.Render(ui.WrapText(selected.Weakness, 44)))
	right.WriteString("\n\n")

	// Chapter
	right.WriteString(ui.SubtitleStyle.Render("  Chapter"))
	right.WriteString("\n")
	right.WriteString("  " + ui.MutedStyle.Render(selected.Chapter))
	right.WriteString("\n")

	// Compose the two-pane layout side by side.
	leftLines := strings.Split(strings.TrimRight(left.String(), "\n"), "\n")
	rightLines := strings.Split(strings.TrimRight(right.String(), "\n"), "\n")

	// Fixed left column width for alignment.
	leftWidth := 36

	maxLines := len(leftLines)
	if len(rightLines) > maxLines {
		maxLines = len(rightLines)
	}

	sep := ui.MutedStyle.Render(" │ ")
	for i := 0; i < maxLines; i++ {
		lc := ""
		if i < len(leftLines) {
			lc = leftLines[i]
		}
		rc := ""
		if i < len(rightLines) {
			rc = rightLines[i]
		}
		// Pad the left column to a fixed width (using rune count for CJK).
		padded := padToWidth(lc, leftWidth)
		b.WriteString(padded)
		b.WriteString(sep)
		b.WriteString(rc)
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(ui.HelpStyle.Render("  ↑↓/jk: scroll demons | tab/←→: switch tabs | esc: dashboard"))

	return b.String()
}

// padToWidth pads a string with spaces to reach the target visible width.
// It counts runes rather than bytes so CJK characters are handled reasonably.
func padToWidth(s string, width int) string {
	runes := []rune(s)
	if len(runes) >= width {
		return s
	}
	return s + strings.Repeat(" ", width-len(runes))
}
