// Package shrine is the Earth God's shrine (土地庙): a shop where coins earned
// by studying buy gear, and a character sheet showing what is worn. Gear
// changes how battles play and never touches review schedules.
package shrine

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/therealsiege/tuilaoshi/internal/audio"
	"github.com/therealsiege/tuilaoshi/internal/journey"
	"github.com/therealsiege/tuilaoshi/internal/ui"
	"github.com/therealsiege/tuilaoshi/internal/wiki"
)

// Model is the shrine view model.
type Model struct {
	store    *wiki.Store
	gkeys    ui.GlobalKeys
	items    []journey.Gear // every item, in slot order
	cleared  map[int]string
	kit      wiki.Kit
	mats     map[string]int // materials held, for crafting
	wallet   wiki.Wallet
	mastered int // level points (learned + mastered characters), which set HP
	cursor   int
	note     string // what the last action did
	noteErr  bool
	width    int
	height   int
	loadErr  error
}

// New creates a shrine model.
func New(store *wiki.Store) Model {
	var items []journey.Gear
	for _, slot := range journey.GearSlots {
		for _, g := range journey.AllGear() {
			if g.Slot == slot {
				items = append(items, g)
			}
		}
	}
	return Model{store: store, gkeys: ui.DefaultGlobalKeys(), items: items}
}

func (m Model) Init() tea.Cmd { return nil }

// SetSize updates the terminal dimensions.
func (m Model) SetSize(w, h int) Model {
	m.width, m.height = w, h
	return m
}

// Refresh reloads the purse, the gear and the stages that stock the shelves.
func (m Model) Refresh() Model {
	var errs [4]error
	m.cleared, errs[0] = m.store.ClearedStages()
	m.kit, errs[1] = m.store.Kit()
	m.wallet, errs[2] = m.store.Wallet()
	mats, matErr := m.store.Materials()
	m.mats = mats
	errs[1] = errors.Join(errs[1], matErr)
	pages, err := m.store.ListPages()
	errs[3] = err
	m.mastered = journey.CountProgress(pages).LevelPoints()
	m.loadErr = errors.Join(errs[:]...)
	m.note = ""
	return m
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	k, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	switch {
	case key.Matches(k, m.gkeys.Back):
		return m, ui.SwitchTo(ui.ViewDashboard)
	case key.Matches(k, m.gkeys.Up):
		m.cursor = max(m.cursor-1, 0)
	case key.Matches(k, m.gkeys.Down):
		m.cursor = min(m.cursor+1, len(m.items)-1)
	case key.Matches(k, m.gkeys.Enter) || k.Type == tea.KeySpace:
		return m.act(), nil
	}
	return m, nil
}

// act buys the selected item, or puts it on or takes it off once owned.
func (m Model) act() Model {
	g := m.items[m.cursor]
	switch {
	case !g.OnShelf(m.cleared):
		return m.say(true, fmt.Sprintf("%s comes to the shrine after tribulation %d.", g.Name, g.Opens))
	case !m.owns(g.ID) && g.Crafted():
		if err := m.store.Craft(g.ID, g.Recipe, wiki.Wallet(g.Price)); err != nil {
			return m.say(true, fmt.Sprintf("Can't craft %s yet: it takes %s. Materials drop from road wins that recall shaky words.", g.Name, recipeText(g.Recipe, g.Price)))
		}
		if err := m.store.Wear(g.Slot, g.ID); err != nil {
			return m.say(true, err.Error())
		}
		m = m.Refresh()
		ui.Cue(audio.Coins)
		return m.say(false, fmt.Sprintf("Crafted and put on %s %s.", g.Name, g.English))
	case !m.owns(g.ID):
		short := wiki.Wallet{Tongqian: g.Price.Tongqian - m.wallet.Tongqian, Yuanbao: g.Price.Yuanbao - m.wallet.Yuanbao}
		if short.Tongqian > 0 || short.Yuanbao > 0 {
			return m.say(true, "Not enough yet: "+shortText(short)+". Study and fight to earn more.")
		}
		if err := m.store.Buy(g.ID, wiki.Wallet(g.Price)); err != nil {
			return m.say(true, err.Error())
		}
		if err := m.store.Wear(g.Slot, g.ID); err != nil {
			return m.say(true, err.Error())
		}
		m = m.Refresh()
		ui.Cue(audio.Coins)
		return m.say(false, fmt.Sprintf("Bought and put on %s %s.", g.Name, g.English))
	case m.kit.Worn[g.Slot] == g.ID:
		if err := m.store.Wear(g.Slot, ""); err != nil {
			return m.say(true, err.Error())
		}
		m = m.Refresh()
		return m.say(false, fmt.Sprintf("Took off %s.", g.Name))
	}
	if err := m.store.Wear(g.Slot, g.ID); err != nil {
		return m.say(true, err.Error())
	}
	m = m.Refresh()
	return m.say(false, fmt.Sprintf("Put on %s.", g.Name))
}

func (m Model) say(isErr bool, s string) Model {
	m.note, m.noteErr = s, isErr
	return m
}

func (m Model) owns(id string) bool { return slices.Contains(m.kit.Owned, id) }

// shortText says how many coins are missing.
func shortText(w wiki.Wallet) string {
	switch {
	case w.Tongqian > 0 && w.Yuanbao > 0:
		return fmt.Sprintf("%d more 铜钱 and %d more 元宝", w.Tongqian, w.Yuanbao)
	case w.Yuanbao > 0:
		return fmt.Sprintf("%d more 元宝", w.Yuanbao)
	}
	return fmt.Sprintf("%d more 铜钱", w.Tongqian)
}

// recipeText lists what crafting takes: materials, then any copper.
func recipeText(recipe map[string]int, price journey.Price) string {
	var parts []string
	for _, m := range journey.Materials() {
		if n := recipe[m.ID]; n > 0 {
			parts = append(parts, fmt.Sprintf("%d %s", n, m.Name))
		}
	}
	if price.Tongqian > 0 {
		parts = append(parts, fmt.Sprintf("%d 铜钱", price.Tongqian))
	}
	return strings.Join(parts, " + ")
}
