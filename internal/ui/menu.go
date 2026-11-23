package ui

import "strings"

// MenuItem is one dashboard action: its key and what it does, in plain words.
type MenuItem struct {
	Key, Label string
}

// MenuGroup is a titled set of dashboard actions.
type MenuGroup struct {
	Title string
	Items []MenuItem
}

// DashboardMenu is the main menu, grouped by purpose. The dashboard and its
// help screen both render it, and every DashboardKeys binding appears here.
var DashboardMenu = []MenuGroup{
	{"Study", []MenuItem{
		{"r", "Review due words"},
		{"d", "Practice flashcards"},
		{"t", "Read & listen"},
		{"c", "Chat with the tutor"},
	}},
	{"Journey", []MenuItem{
		{"f", "Spar (practice battle)"},
		{"j", "Journey log"},
		{"g", "Shrine: gear & coins"},
		{"o", "Collection cards"},
		{"p", "Party & bonds"},
		{"s", "My progress"},
		{"a", "Check my level"},
	}},
	{"More", []MenuItem{
		{"b", "Browse the wiki"},
		{"/", "Search the wiki"},
		{"x", "Radicals"},
		{"v", "Voices"},
	}},
}

// MenuLabel returns the label of a dashboard key, or "".
func MenuLabel(key string) string {
	for _, g := range DashboardMenu {
		for _, it := range g.Items {
			if it.Key == key {
				return it.Label
			}
		}
	}
	return ""
}

// menuHelp renders the menu as plain text for the help screen.
func menuHelp() string {
	var b strings.Builder
	b.WriteString("Camp Keys\n")
	for _, g := range DashboardMenu {
		b.WriteString("\n" + g.Title + "\n")
		for _, it := range g.Items {
			b.WriteString("  " + it.Key + "  " + it.Label + "\n")
		}
	}
	b.WriteString("\n  ?  Toggle help\n  q  Quit")
	return b.String()
}
