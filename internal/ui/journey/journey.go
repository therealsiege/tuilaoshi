package journey

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/cj/mandarin-wiki/internal/journey"
	"github.com/cj/mandarin-wiki/internal/ui"
	"github.com/cj/mandarin-wiki/internal/wiki"
)

// Model is the Journey to the West gamification view.
type Model struct {
	store  *wiki.Store
	state  journey.State
	tab    int // 0=journey, 1=party, 2=bosses, 3=artifacts, 4=quests, 5=achievements, 6=lore
	width  int
	height int
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
			m.tab = (m.tab + 1) % 7
			return m, nil
		case "left", "h":
			m.tab = (m.tab - 1 + 7) % 7
			return m, nil
		case "right", "l":
			m.tab = (m.tab + 1) % 7
			return m, nil
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
	tabs := []string{"journey", "party", "bosses", "artifacts", "quests", "achievements", "lore"}
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

	// Cultivation Rank section
	b.WriteString(ui.SubtitleStyle.Render("  修为 Cultivation Rank"))
	b.WriteString("\n")

	rankBox := m.renderRankBox()
	b.WriteString(ui.BorderStyle.Width(50).Render(rankBox))
	b.WriteString("\n\n")

	// Pilgrimage progress section
	b.WriteString(ui.SubtitleStyle.Render("  取经路 The Pilgrimage"))
	b.WriteString("\n\n")

	// Journey map line
	journeyMap := m.renderJourneyMap()
	b.WriteString("  " + journeyMap)
	b.WriteString("\n")

	// Current position
	posText := fmt.Sprintf("  You are at %s %s — Tribulation %d of 81",
		ui.ChineseStyle.Render(m.state.Progress.Location),
		ui.MutedStyle.Render("("+m.state.Progress.LocationEn+")"),
		m.state.Progress.Tribulation)
	b.WriteString(posText)
	b.WriteString("\n\n")

	// Progress summary
	progressText := fmt.Sprintf("  Progress: %d/%d characters mastered (%.0f%%)",
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
	rankLine := fmt.Sprintf("%s · %s",
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
		b.WriteString(ui.SuccessStyle.Render("✓ Max cultivation achieved!"))
	}

	return b.String()
}

// renderJourneyMap renders the named-location journey progress line.
func (m Model) renderJourneyMap() string {
	locations := []struct {
		name  string
		start int
	}{
		{"长安", 0},
		{"双叉岭", 10},
		{"五行山", 20},
		{"观音院", 30},
		{"高老庄", 40},
		{"流沙河", 50},
		{"火焰山", 60},
		{"通天河", 70},
		{"灵山", 80},
		{"天竺", 81},
	}

	trib := m.state.Progress.Tribulation
	var parts []string
	for i, loc := range locations {
		if trib >= loc.start {
			parts = append(parts, ui.SuccessStyle.Render(loc.name))
		} else if i > 0 && trib >= locations[i-1].start {
			parts = append(parts, ui.WarningStyle.Render(loc.name))
		} else {
			parts = append(parts, ui.MutedStyle.Render(loc.name))
		}
		if i < len(locations)-1 {
			parts = append(parts, ui.MutedStyle.Render(" → "))
		}
	}
	return strings.Join(parts, "")
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
