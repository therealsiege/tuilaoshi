package radicals

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/cj/mandarin-wiki/internal/ui"
)

// RadicalGroup holds a radical and its family of example characters.
type RadicalGroup struct {
	Radical  string   // e.g. "氵"
	Name     string   // e.g. "water"
	Chinese  string   // e.g. "三点水"
	Strokes  int
	Examples []string // e.g. ["河", "海", "湖", "洗", "流"]
}

// Model is the radical decomposition view model.
type Model struct {
	keys      ui.GlobalKeys
	groups    []RadicalGroup
	cursor    int
	scrollOff int // scroll offset for the list
	width     int
	height    int
}

// New creates a new radicals view model.
func New() Model {
	return Model{
		keys:   ui.DefaultGlobalKeys(),
		groups: defaultRadicals(),
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Up):
			if m.cursor > 0 {
				m.cursor--
				if m.cursor < m.scrollOff {
					m.scrollOff = m.cursor
				}
			}
			return m, nil
		case key.Matches(msg, m.keys.Down):
			if m.cursor < len(m.groups)-1 {
				m.cursor++
				visible := m.visibleRows()
				if m.cursor >= m.scrollOff+visible {
					m.scrollOff = m.cursor - visible + 1
				}
			}
			return m, nil
		case key.Matches(msg, m.keys.Back):
			return m, ui.SwitchTo(ui.ViewDashboard)
		}
	}
	return m, nil
}

// visibleRows returns how many list rows fit in the available height.
func (m Model) visibleRows() int {
	// Header (3) + spacing (2) + footer (3) = 8 lines of chrome
	available := m.height - 8
	if available < 5 {
		available = 5
	}
	return available
}

func (m Model) View() string {
	var b strings.Builder

	// Header
	b.WriteString(ui.Header("部首 Radical Families"))
	b.WriteString("\n\n")

	// Two-pane layout: list on the left, detail on the right
	leftPane := m.renderList()
	rightPane := m.renderDetail()

	leftBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ui.ColorMuted).
		Padding(0, 1).
		Width(28).
		Render(leftPane)

	rightBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ui.ColorSecondary).
		Padding(1, 2).
		Width(38).
		Render(rightPane)

	b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, leftBox, "  ", rightBox))
	b.WriteString("\n\n")

	// Footer
	b.WriteString(ui.Divider(40))
	b.WriteString("\n")
	b.WriteString(ui.HelpStyle.Render("  ↑↓ select radical | esc dashboard | ? help"))

	return b.String()
}

// renderList renders the scrollable radical list on the left.
func (m Model) renderList() string {
	var b strings.Builder

	visible := m.visibleRows()
	end := m.scrollOff + visible
	if end > len(m.groups) {
		end = len(m.groups)
	}

	for i := m.scrollOff; i < end; i++ {
		g := m.groups[i]
		label := fmt.Sprintf("%s %s", g.Radical, g.Name)
		if i == m.cursor {
			b.WriteString(ui.SelectedStyle.Render("▸ " + label))
		} else {
			b.WriteString(ui.UnselectedStyle.Render("  " + label))
		}
		b.WriteString("\n")
	}

	// Scroll indicators
	if m.scrollOff > 0 {
		b.WriteString(ui.MutedStyle.Render("  ↑ more"))
		b.WriteString("\n")
	}
	if end < len(m.groups) {
		b.WriteString(ui.MutedStyle.Render("  ↓ more"))
		b.WriteString("\n")
	}

	return b.String()
}

// renderDetail renders the detail panel for the selected radical.
func (m Model) renderDetail() string {
	if m.cursor < 0 || m.cursor >= len(m.groups) {
		return ui.MutedStyle.Render("No radical selected")
	}

	g := m.groups[m.cursor]
	var b strings.Builder

	// Radical character, large and prominent
	b.WriteString(ui.ChineseStyle.Render(g.Radical))
	b.WriteString("\n\n")

	// Name and Chinese name
	b.WriteString(ui.SubtitleStyle.Render(g.Name))
	b.WriteString("  ")
	b.WriteString(ui.MutedStyle.Render(g.Chinese))
	b.WriteString("\n\n")

	// Stroke count
	b.WriteString(ui.MutedStyle.Render(fmt.Sprintf("Strokes: %d", g.Strokes)))
	b.WriteString("\n\n")

	// Example characters
	b.WriteString(ui.SubtitleStyle.Render("Examples"))
	b.WriteString("\n")
	for _, ch := range g.Examples {
		b.WriteString("  ")
		b.WriteString(ui.ChineseStyle.Render(ch))
	}
	b.WriteString("\n")

	return b.String()
}

// SetSize updates the terminal dimensions.
func (m Model) SetSize(w, h int) Model {
	m.width = w
	m.height = h
	return m
}

// Refresh reloads data (no-op for hardcoded data, kept for interface consistency).
func (m Model) Refresh() Model {
	return m
}

// defaultRadicals returns 30 common radical groups.
func defaultRadicals() []RadicalGroup {
	return []RadicalGroup{
		{Radical: "氵", Name: "water", Chinese: "三点水", Strokes: 3,
			Examples: []string{"河", "海", "湖", "洗", "流"}},
		{Radical: "木", Name: "tree", Chinese: "木字旁", Strokes: 4,
			Examples: []string{"林", "森", "树", "村", "桥"}},
		{Radical: "火", Name: "fire", Chinese: "火字旁", Strokes: 4,
			Examples: []string{"灯", "烧", "炒", "烟", "灭"}},
		{Radical: "土", Name: "earth", Chinese: "提土旁", Strokes: 3,
			Examples: []string{"地", "场", "城", "块", "坐"}},
		{Radical: "钅", Name: "metal", Chinese: "金字旁", Strokes: 5,
			Examples: []string{"铁", "银", "钱", "铜", "锅"}},
		{Radical: "亻", Name: "person", Chinese: "单人旁", Strokes: 2,
			Examples: []string{"他", "你", "们", "住", "作"}},
		{Radical: "口", Name: "mouth", Chinese: "口字旁", Strokes: 3,
			Examples: []string{"吃", "喝", "叫", "听", "吗"}},
		{Radical: "忄", Name: "heart", Chinese: "竖心旁", Strokes: 3,
			Examples: []string{"快", "慢", "情", "怕", "忙"}},
		{Radical: "扌", Name: "hand", Chinese: "提手旁", Strokes: 3,
			Examples: []string{"打", "拉", "把", "找", "拿"}},
		{Radical: "日", Name: "sun", Chinese: "日字旁", Strokes: 4,
			Examples: []string{"明", "时", "早", "晚", "暖"}},
		{Radical: "月", Name: "moon", Chinese: "月字旁", Strokes: 4,
			Examples: []string{"朋", "胖", "脸", "脚", "期"}},
		{Radical: "目", Name: "eye", Chinese: "目字旁", Strokes: 5,
			Examples: []string{"看", "眼", "睡", "眉", "睛"}},
		{Radical: "讠", Name: "speech", Chinese: "言字旁", Strokes: 2,
			Examples: []string{"说", "话", "请", "谢", "读"}},
		{Radical: "饣", Name: "food", Chinese: "食字旁", Strokes: 3,
			Examples: []string{"饭", "饿", "饮", "饺", "馒"}},
		{Radical: "女", Name: "woman", Chinese: "女字旁", Strokes: 3,
			Examples: []string{"妈", "姐", "妹", "好", "她"}},
		{Radical: "子", Name: "child", Chinese: "子字旁", Strokes: 3,
			Examples: []string{"学", "孩", "孙", "字", "孤"}},
		{Radical: "山", Name: "mountain", Chinese: "山字旁", Strokes: 3,
			Examples: []string{"岁", "岛", "岸", "崇", "峰"}},
		{Radical: "石", Name: "stone", Chinese: "石字旁", Strokes: 5,
			Examples: []string{"破", "砖", "碗", "硬", "确"}},
		{Radical: "⺮", Name: "bamboo", Chinese: "竹字头", Strokes: 6,
			Examples: []string{"笔", "第", "等", "答", "筷"}},
		{Radical: "纟", Name: "silk", Chinese: "绞丝旁", Strokes: 3,
			Examples: []string{"红", "绿", "给", "经", "练"}},
		{Radical: "虫", Name: "insect", Chinese: "虫字旁", Strokes: 6,
			Examples: []string{"蛇", "蝴", "蝶", "蚊", "蜜"}},
		{Radical: "贝", Name: "shell", Chinese: "贝字旁", Strokes: 4,
			Examples: []string{"财", "败", "贵", "赚", "购"}},
		{Radical: "马", Name: "horse", Chinese: "马字旁", Strokes: 3,
			Examples: []string{"骑", "驾", "驻", "验", "骗"}},
		{Radical: "鱼", Name: "fish", Chinese: "鱼字旁", Strokes: 8,
			Examples: []string{"鲜", "鲤", "鲸", "鳄", "鳗"}},
		{Radical: "鸟", Name: "bird", Chinese: "鸟字旁", Strokes: 5,
			Examples: []string{"鸡", "鸭", "鹅", "鸽", "鹰"}},
		{Radical: "刂", Name: "knife", Chinese: "立刀旁", Strokes: 2,
			Examples: []string{"到", "别", "利", "刻", "剧"}},
		{Radical: "力", Name: "power", Chinese: "力字旁", Strokes: 2,
			Examples: []string{"动", "办", "功", "加", "助"}},
		{Radical: "雨", Name: "rain", Chinese: "雨字头", Strokes: 8,
			Examples: []string{"雪", "雷", "零", "霜", "露"}},
		{Radical: "门", Name: "gate", Chinese: "门字框", Strokes: 3,
			Examples: []string{"问", "间", "闪", "闻", "闭"}},
		{Radical: "艹", Name: "grass", Chinese: "草字头", Strokes: 3,
			Examples: []string{"花", "草", "茶", "药", "菜"}},
	}
}
