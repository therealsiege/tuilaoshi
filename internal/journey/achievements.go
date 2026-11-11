package journey

// Achievement is a milestone the learner has reached.
type Achievement struct {
	Name     string // English name
	Chinese  string // Chinese name
	Desc     string // What this means for your ability
	Icon     string // ASCII icon
	Earned   bool
	Hidden   bool   // Hidden achievements show as "???" until earned
	Category string // "literacy", "dedication", "mastery", "journey"
}

// buildAchievements computes all achievements from study metrics.
func buildAchievements(mastered, totalReviewed, streak, daysStudied int) []Achievement {
	// Tribulation is derived from mastered chars for journey achievements.
	// We approximate: 81 tribulations over ~5000 chars total mastery.
	tribulation := 0
	if mastered > 0 {
		tribulation = (mastered * 81) / 5000
		if tribulation > 81 {
			tribulation = 81
		}
	}

	achievements := []Achievement{
		// ── Literacy milestones ──────────────────────────────────────
		// Real-world ability markers tied to character counts.
		{
			Name:     "Menu Reader",
			Chinese:  "菜单通",
			Desc:     "You can navigate a Chinese restaurant menu",
			Icon:     "M",
			Earned:   mastered >= 100,
			Category: "literacy",
		},
		{
			Name:     "Street Signs",
			Chinese:  "路标识",
			Desc:     "You can read basic street signs and directions",
			Icon:     "S",
			Earned:   mastered >= 250,
			Category: "literacy",
		},
		{
			Name:     "Text Messenger",
			Chinese:  "短信通",
			Desc:     "You can read and write basic text messages",
			Icon:     "T",
			Earned:   mastered >= 500,
			Category: "literacy",
		},
		{
			Name:     "News Headlines",
			Chinese:  "看标题",
			Desc:     "You can understand newspaper headlines",
			Icon:     "N",
			Earned:   mastered >= 1000,
			Category: "literacy",
		},
		{
			Name:     "Novel Reader",
			Chinese:  "小说迷",
			Desc:     "You can read simplified Chinese novels",
			Icon:     "R",
			Earned:   mastered >= 2000,
			Category: "literacy",
		},
		{
			Name:     "Newspaper Literate",
			Chinese:  "报纸通",
			Desc:     "You can read a Chinese newspaper",
			Icon:     "P",
			Earned:   mastered >= 3000,
			Category: "literacy",
		},
		{
			Name:     "Scholar",
			Chinese:  "学者",
			Desc:     "Full literacy -- you can read virtually anything",
			Icon:     "W",
			Earned:   mastered >= 5000,
			Category: "literacy",
		},

		// ── Dedication achievements ─────────────────────────────────
		// Streak and consistency milestones.
		{
			Name:     "First Step",
			Chinese:  "第一步",
			Desc:     "You showed up -- the hardest part",
			Icon:     ".",
			Earned:   streak >= 1,
			Category: "dedication",
		},
		{
			Name:     "Week Warrior",
			Chinese:  "周战士",
			Desc:     "A full week of daily practice",
			Icon:     "w",
			Earned:   streak >= 7,
			Category: "dedication",
		},
		{
			Name:     "Moon Keeper",
			Chinese:  "守月者",
			Desc:     "A full month without missing a day",
			Icon:     "D",
			Earned:   streak >= 30,
			Category: "dedication",
		},
		{
			Name:     "Season Scholar",
			Chinese:  "季学者",
			Desc:     "Three months of unbroken study",
			Icon:     "Q",
			Earned:   streak >= 90,
			Category: "dedication",
		},
		{
			Name:     "Year of the Ox",
			Chinese:  "牛年功",
			Desc:     "A full year of daily study -- true perseverance",
			Icon:     "Y",
			Earned:   streak >= 365,
			Category: "dedication",
		},
		{
			Name:     "Century Club",
			Chinese:  "百日功",
			Desc:     "One hundred reviews completed",
			Icon:     "C",
			Earned:   totalReviewed >= 100,
			Category: "dedication",
		},
		{
			Name:     "Thousand Hands",
			Chinese:  "千手观音",
			Desc:     "One thousand reviews -- like the thousand-armed Guanyin",
			Icon:     "K",
			Earned:   totalReviewed >= 1000,
			Category: "dedication",
		},
		{
			Name:     "Ten Thousand",
			Chinese:  "万事通",
			Desc:     "Ten thousand reviews -- mastery through repetition",
			Icon:     "X",
			Earned:   totalReviewed >= 10000,
			Category: "dedication",
		},

		// ── Mastery achievements ────────────────────────────────────
		// Learning quality milestones.
		{
			Name:     "Perfect Day",
			Chinese:  "完美日",
			Desc:     "100%% accuracy in a review session",
			Icon:     "*",
			Earned:   false, // requires session-level accuracy tracking
			Category: "mastery",
		},
		{
			Name:     "No Leech",
			Chinese:  "无漏洞",
			Desc:     "Mastered 50 characters without failing any",
			Icon:     "O",
			Earned:   false, // requires per-character failure tracking
			Category: "mastery",
		},
		{
			Name:     "Radical Master",
			Chinese:  "部首通",
			Desc:     "Learned characters spanning 20+ different radicals",
			Icon:     "B",
			Earned:   false, // requires radical diversity tracking
			Category: "mastery",
		},
		{
			Name:     "Tone Sage",
			Chinese:  "四声圣",
			Desc:     "Mastered characters in all four tones plus neutral",
			Icon:     "~",
			Earned:   false, // requires tone diversity tracking
			Category: "mastery",
		},

		// ── Journey to the West achievements ────────────────────────
		// Narrative milestones tied to the pilgrimage.
		{
			Name:     "Left Chang'an",
			Chinese:  "离开长安",
			Desc:     "The pilgrimage has begun",
			Icon:     ">",
			Earned:   mastered >= 1,
			Category: "journey",
		},
		{
			Name:     "Freed the Monkey",
			Chinese:  "解放悟空",
			Desc:     "Reached Five Elements Mountain -- Sun Wukong joins your party",
			Icon:     "M",
			Earned:   tribulation >= 20,
			Category: "journey",
		},
		{
			Name:     "Tamed the Pig",
			Chinese:  "收服八戒",
			Desc:     "Passed Gao Village -- Zhu Bajie joins your party",
			Icon:     "P",
			Earned:   tribulation >= 40,
			Category: "journey",
		},
		{
			Name:     "Crossed the River",
			Chinese:  "渡过流沙河",
			Desc:     "Crossed the Flowing Sands River -- Sha Wujing joins your party",
			Icon:     "~",
			Earned:   tribulation >= 50,
			Category: "journey",
		},
		{
			Name:     "Survived the Flames",
			Chinese:  "火焰山过",
			Desc:     "Passed the Flaming Mountains with the Banana Leaf Fan",
			Icon:     "F",
			Earned:   tribulation >= 60,
			Category: "journey",
		},
		{
			Name:     "Heaven-Reaching River",
			Chinese:  "通天河渡",
			Desc:     "Crossed the final great river before Spirit Mountain",
			Icon:     "H",
			Earned:   tribulation >= 70,
			Category: "journey",
		},
		{
			Name:     "Reached Thunder Temple",
			Chinese:  "大雷音寺",
			Desc:     "Completed all 81 tribulations -- you stand before the Buddha",
			Icon:     "T",
			Earned:   tribulation >= 81,
			Category: "journey",
		},
		{
			Name:     "Obtained the Sutras",
			Chinese:  "取得真经",
			Desc:     "True mastery -- the scriptures are yours to carry home",
			Icon:     "E",
			Earned:   mastered >= 4000,
			Category: "journey",
		},

		// ── Hidden achievements ─────────────────────────────────────
		// Surprises that reveal themselves only when earned.

		// Study habits (need future time-of-day tracking)
		{
			Name:     "Night Owl",
			Chinese:  "夜猫子",
			Desc:     "Study after midnight -- the quiet hours belong to scholars",
			Icon:     "N",
			Earned:   false,
			Hidden:   true,
			Category: "dedication",
		},
		{
			Name:     "Early Bird",
			Chinese:  "早起鸟",
			Desc:     "Study before 6am -- catch the worm of knowledge",
			Icon:     "E",
			Earned:   false,
			Hidden:   true,
			Category: "dedication",
		},
		{
			Name:     "Weekend Warrior",
			Chinese:  "周末勇士",
			Desc:     "Study every day of a weekend -- no rest for the devoted",
			Icon:     "W",
			Earned:   false,
			Hidden:   true,
			Category: "dedication",
		},

		// Milestone surprises
		{
			Name:     "108",
			Chinese:  "一百零八",
			Desc:     "Master exactly 108 characters -- the 108 heroes of Water Margin",
			Icon:     "!",
			Earned:   mastered == 108,
			Hidden:   true,
			Category: "literacy",
		},
		{
			Name:     "Three Hundred Thousand",
			Chinese:  "三百千",
			Desc:     "A nod to the Three Character Classic -- 300 characters mastered",
			Icon:     "3",
			Earned:   mastered >= 300,
			Hidden:   true,
			Category: "literacy",
		},
		{
			Name:     "Never Quit Halfway",
			Chinese:  "半途不废",
			Desc:     "Past the halfway point of 81 tribulations -- no turning back",
			Icon:     "H",
			Earned:   tribulation >= 41,
			Hidden:   true,
			Category: "journey",
		},

		// Character discoveries (need character-specific tracking)
		{
			Name:     "Four Great Novels",
			Chinese:  "四大名著",
			Desc:     "Master the characters for all four classic novels",
			Icon:     "4",
			Earned:   false,
			Hidden:   true,
			Category: "mastery",
		},
		{
			Name:     "Five Elements Complete",
			Chinese:  "五行齐全",
			Desc:     "Master the five elements: metal, wood, water, fire, earth",
			Icon:     "5",
			Earned:   false,
			Hidden:   true,
			Category: "mastery",
		},
		{
			Name:     "Twelve Zodiac Animals",
			Chinese:  "十二生肖",
			Desc:     "Master all twelve zodiac animal characters",
			Icon:     "Z",
			Earned:   false,
			Hidden:   true,
			Category: "mastery",
		},

		// Journey milestones
		{
			Name:     "Sutra Seeker",
			Chinese:  "取经人",
			Desc:     "Reach the final tribulation -- the end is in sight",
			Icon:     "S",
			Earned:   tribulation >= 81,
			Hidden:   true,
			Category: "journey",
		},
		{
			Name:     "Great Sage Equal to Heaven",
			Chinese:  "齐天大圣",
			Desc:     "Reach the highest cultivation rank -- equal to the heavens",
			Icon:     "G",
			Earned:   mastered >= 5000,
			Hidden:   true,
			Category: "journey",
		},
		{
			Name:     "81 Tribulations",
			Chinese:  "八十一难",
			Desc:     "Complete all 81 tribulations -- the full ordeal endured",
			Icon:     "8",
			Earned:   tribulation >= 81,
			Hidden:   true,
			Category: "journey",
		},

		// Dedication secrets
		{
			Name:     "Iron Rod to Needle",
			Chinese:  "铁杵磨针",
			Desc:     "Review the same character 50 times -- persistence grinds iron to a needle",
			Icon:     "I",
			Earned:   false,
			Hidden:   true,
			Category: "dedication",
		},
		{
			Name:     "Water Drills Stone",
			Chinese:  "水滴石穿",
			Desc:     "365-day streak -- dripping water bores through stone",
			Icon:     "D",
			Earned:   streak >= 365,
			Hidden:   true,
			Category: "dedication",
		},
		{
			Name:     "Foolish Old Man Moves Mountains",
			Chinese:  "愚公移山",
			Desc:     "Master 5000 characters -- move mountains with stubbornness",
			Icon:     "F",
			Earned:   mastered >= 5000,
			Hidden:   true,
			Category: "literacy",
		},
	}

	return achievements
}

// EarnedCount returns how many achievements have been earned.
func EarnedCount(achievements []Achievement) int {
	n := 0
	for _, a := range achievements {
		if a.Earned {
			n++
		}
	}
	return n
}

// HiddenUnearnedCount returns how many hidden achievements have not yet been earned.
func HiddenUnearnedCount(achievements []Achievement) int {
	n := 0
	for _, a := range achievements {
		if a.Hidden && !a.Earned {
			n++
		}
	}
	return n
}

// EarnedInCategory returns earned achievements in a specific category.
func EarnedInCategory(achievements []Achievement, category string) []Achievement {
	var earned []Achievement
	for _, a := range achievements {
		if a.Category == category && a.Earned {
			earned = append(earned, a)
		}
	}
	return earned
}
