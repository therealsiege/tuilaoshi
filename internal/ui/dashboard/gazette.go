package dashboard

import (
	"cmp"
	"context"
	"fmt"
	"slices"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/therealsiege/tuilaoshi/internal/claude"
	"github.com/therealsiege/tuilaoshi/internal/journey"
	"github.com/therealsiege/tuilaoshi/internal/wiki"
)

// The court gazette (邸报): once a week, from Monday, a short newspaper
// report of the learner's own journey in Chinese at HSK4-5 with its
// English: the kingdoms that have stamped the travel papers, the week's
// tribulations cleared and rumours settled, errands done and sacks
// haggled for, and the diary's latest words. Claude writes it from those
// facts; without a key the court's clerk sets it from a template. Each
// issue is kept with the journey.

// GazetteWriter writes an issue from the week's facts, returning its
// Chinese and English (claude.Client.WriteGazette).
type GazetteWriter func(ctx context.Context, facts string) (string, string, error)

// WithGazette lets Claude write the gazette; a nil client leaves it to the
// template.
func (m Model) WithGazette(c *claude.Client) Model {
	if c != nil {
		m.writer = c.WriteGazette
	}
	return m
}

// gazetteMsg is an issue written by Claude.
type gazetteMsg struct {
	week    string
	zh, en  string
	err     error
	byClerk wiki.Gazette // the template's issue, in case Claude fails
}

// weekFacts is what the gazette reports for the week before monday.
type weekFacts struct {
	Kingdoms  []string // every kingdom that has stamped the papers, in order
	Cleared   []string // tribulations cleared in the week
	Mysteries []string // towns where a rumour was settled in the week
	Unmasked  []string // demons unmasked in the week, as 名字（城）
	Errands   []string // errands done in the week, by their goal
	Haggles   []string // sacks haggled for in the week
	HaggledEN []string // the same in English
	Diary     string   // the diary's latest entry
}

// gatherFacts reads the week before monday from the journey.
func gatherFacts(store *wiki.Store, monday time.Time) (weekFacts, error) {
	from, to := monday.AddDate(0, 0, -7).Format("2006-01-02"), monday.Format("2006-01-02")
	in := func(day string) bool { return day >= from && day < to }
	names := map[string]string{}
	for _, t := range journey.Towns() {
		names[t.ID] = t.Name
	}
	var f weekFacts
	stamps, err := store.Stamps()
	if err != nil {
		return f, err
	}
	for _, s := range stamps {
		f.Kingdoms = append(f.Kingdoms, names[s])
	}
	cleared, err := store.ClearedStages()
	if err != nil {
		return f, err
	}
	var numbers []int
	for n, day := range cleared {
		if journey.TribulationAt(n) != nil && in(day) {
			numbers = append(numbers, n)
		}
	}
	slices.Sort(numbers) // by number: 第2难 before 第10难
	for _, n := range numbers {
		f.Cleared = append(f.Cleared, fmt.Sprintf("第%d难「%s」", n, journey.TribulationAt(n).Name))
	}
	settled, err := store.MysteriesSettled()
	if err != nil {
		return f, err
	}
	for town, day := range settled {
		if in(day) {
			f.Mysteries = append(f.Mysteries, names[town])
		}
	}
	slices.Sort(f.Mysteries)
	deeds, err := store.Deeds()
	if err != nil {
		return f, err
	}
	for _, d := range deeds {
		switch {
		case !in(d.Date):
		case d.Kind == wiki.DeedUnmask:
			f.Unmasked = append(f.Unmasked, fmt.Sprintf("%s（%s）", d.What, names[d.Town]))
		case d.Kind == wiki.DeedErrand:
			f.Errands = append(f.Errands, strings.TrimRight(d.What, "。"))
		case d.Kind == wiki.DeedHaggle:
			f.Haggles = append(f.Haggles, fmt.Sprintf("在%s用%d个铜钱买了%s", names[d.Town], d.Coins, d.What))
			f.HaggledEN = append(f.HaggledEN, fmt.Sprintf("%s for %d copper in %s", d.What, d.Coins, names[d.Town]))
		}
	}
	diary, err := store.Diary()
	if len(diary) > 0 {
		f.Diary = diary[len(diary)-1].Text
	}
	return f, err
}

// prompt lists the facts for Claude.
func (f weekFacts) prompt(week string) string {
	none := func(s []string) string {
		if len(s) == 0 {
			return "none"
		}
		return strings.Join(s, "；")
	}
	return fmt.Sprintf("Issue of the week starting %s, reporting the week before.\n"+
		"Kingdoms that have stamped the travel papers so far: %s\n"+
		"Tribulations cleared last week: %s\n"+
		"Towns where a strange rumour (a demon in disguise among villagers) was looked into last week: %s\n"+
		"Demons in disguise the pilgrims unmasked last week (the disguise's name, and the town): %s\n"+
		"Errands done for townsfolk last week: %s\n"+
		"Haggling at the market last week: %s\n"+
		"The monk's latest diary entry: %s",
		week, none(f.Kingdoms), none(f.Cleared), none(f.Mysteries), none(f.Unmasked), none(f.Errands), none(f.Haggles), cmp.Or(f.Diary, "none"))
}

// composeGazette is the clerk's issue, set from the facts by template.
func composeGazette(f weekFacts) (string, string) {
	zh := []string{"本报讯：自东土大唐前往西天取经的圣僧一行，上周继续向西而行。"}
	en := []string{"Our report: the holy monk and his party, travelling from the Tang lands in the east to the west for the scriptures, went on west last week."}
	if len(f.Kingdoms) > 0 {
		zh = append(zh, fmt.Sprintf("至今已有%d国在通关文牒上盖了印：%s。", len(f.Kingdoms), strings.Join(f.Kingdoms, "、")))
		en = append(en, fmt.Sprintf("So far %d %s stamped their travel papers: %s.", len(f.Kingdoms), plural(len(f.Kingdoms), "kingdom has", "kingdoms have"), strings.Join(f.Kingdoms, ", ")))
	} else {
		zh, en = append(zh, "他们的通关文牒上还没有盖过印。"), append(en, "No kingdom has stamped their travel papers yet.")
	}
	if len(f.Cleared) > 0 {
		zh = append(zh, fmt.Sprintf("上周他们渡过了%d难：%s，可喜可贺。", len(f.Cleared), strings.Join(f.Cleared, "、")))
		en = append(en, fmt.Sprintf("Last week they came through %d %s: %s. Congratulations are in order.", len(f.Cleared), plural(len(f.Cleared), "tribulation", "tribulations"), strings.Join(f.Cleared, ", ")))
	}
	if len(f.Unmasked) > 0 {
		zh = append(zh, fmt.Sprintf("圣僧的大徒弟用火眼金睛识破了扮成百姓的妖怪：%s。", strings.Join(f.Unmasked, "、")))
		en = append(en, fmt.Sprintf("The master's eldest disciple saw through demons disguised as villagers with his Fiery Eyes: %s.", strings.Join(f.Unmasked, ", ")))
	} else if len(f.Mysteries) > 0 {
		zh = append(zh, fmt.Sprintf("据悉，有妖怪扮成百姓，圣僧的大徒弟在%s用火眼金睛查了这件怪事。", strings.Join(f.Mysteries, "、")))
		en = append(en, fmt.Sprintf("We hear demons disguised themselves as villagers, and in %s the master's eldest disciple looked into it with his Fiery Eyes.", strings.Join(f.Mysteries, ", ")))
	}
	if len(f.Errands) > 0 {
		zh = append(zh, fmt.Sprintf("一路上，他们还帮百姓办了%d件事：%s。百姓们都说，这些和尚的中文说得越来越好了。", len(f.Errands), strings.Join(f.Errands, "；")))
		en = append(en, fmt.Sprintf("Along the way they also ran %d %s for the townsfolk, who all say these monks' Chinese gets better and better.", len(f.Errands), plural(len(f.Errands), "errand", "errands")))
	}
	if len(f.Haggles) > 0 {
		zh = append(zh, "在市场上，他们讨价还价："+strings.Join(f.Haggles, "；")+"。")
		en = append(en, "At the market they haggled: "+strings.Join(f.HaggledEN, "; ")+".")
	}
	if len(f.Cleared)+len(f.Mysteries)+len(f.Errands)+len(f.Haggles) == 0 {
		zh, en = append(zh, "上周路上比较平静，没有什么大事。"), append(en, "The road was quiet last week, with nothing of note.")
	}
	if f.Diary != "" {
		zh, en = append(zh, "圣僧在日记里写道：「"+f.Diary+"」"), append(en, "The holy monk wrote in his diary: \""+f.Diary+"\"")
	}
	zh, en = append(zh, "西天路远，本报将继续报道。"), append(en, "The road west is long; this gazette will keep reporting.")
	return strings.Join(zh, "\n"), strings.Join(en, "\n")
}

// weekGazette is this week's issue if it has come out.
func (m Model) weekGazette() (*wiki.Gazette, error) {
	issues, err := m.store.Gazettes()
	week := journey.WeekStart(m.now()).Format("2006-01-02")
	if err != nil || len(issues) == 0 || issues[len(issues)-1].Week != week {
		return nil, err
	}
	return &issues[len(issues)-1], nil
}

// openGazette shows this week's issue, printing it first if it hasn't come
// out: by Claude when there is a writer, else from the template.
func (m Model) openGazette() (Model, tea.Cmd) {
	if m.printing {
		return m, nil
	}
	issue, err := m.weekGazette()
	if err != nil {
		m.loadErr = err
		return m, nil
	}
	if issue != nil {
		m.gazette, m.gazetteOpen = issue, true
		return m, nil
	}
	monday := journey.WeekStart(m.now())
	facts, err := gatherFacts(m.store, monday)
	if err != nil {
		m.loadErr = err
		return m, nil
	}
	week := monday.Format("2006-01-02")
	zh, en := composeGazette(facts)
	clerk := wiki.Gazette{Week: week, Text: zh, English: en, ByHand: true}
	if m.writer == nil {
		return m.printGazette(clerk), nil
	}
	m.printing, m.note = true, "The court gazette is being printed…"
	write, prompt := m.writer, facts.prompt(week)
	return m, func() tea.Msg {
		zh, en, err := write(context.Background(), prompt)
		return gazetteMsg{week: week, zh: zh, en: en, err: err, byClerk: clerk}
	}
}

// onGazette keeps Claude's issue, or the clerk's when Claude failed.
func (m Model) onGazette(msg gazetteMsg) Model {
	m.printing, m.note = false, ""
	if msg.err != nil {
		m = m.printGazette(msg.byClerk)
		m.note = "Couldn't reach Claude (" + msg.err.Error() + "): the court's clerk set this issue by hand."
		return m
	}
	return m.printGazette(wiki.Gazette{Week: msg.week, Text: msg.zh, English: msg.en})
}

// printGazette keeps an issue and opens it.
func (m Model) printGazette(g wiki.Gazette) Model {
	if err := m.store.SaveGazette(g); err != nil {
		m.loadErr = err
		return m
	}
	m.gazette, m.gazetteOpen = &g, true
	return m
}
