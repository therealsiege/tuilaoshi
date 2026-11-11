// Package phonetic maps Chinese phonetic components to their character families.
//
// Roughly 80% of Chinese characters are phono-semantic compounds: one part
// hints at meaning (the radical) and another hints at sound (the phonetic
// component). Learning these families lets a reader predict the pronunciation
// of unfamiliar characters.
//
// Usage:
//
//	for _, f := range phonetic.Families() {
//	    fmt.Printf("%s (%s) -> %d members\n", f.Component, f.Pinyin, len(f.Members))
//	}
package phonetic

// Family groups characters that share a phonetic component.
type Family struct {
	Component string   // the shared phonetic element: 青, 方, 包, etc.
	Pinyin    string   // base pronunciation of the component
	Members   []Member // characters containing this component
}

// Member is a single character within a phonetic family.
type Member struct {
	Hanzi   string // the full character: 清, 请, 情 ...
	Pinyin  string // pronunciation with tone mark
	Meaning string // concise English gloss
	Radical string // semantic radical: 氵, 讠, 忄 ...
}

// Families returns all phonetic family data.
func Families() []Family {
	return families
}

// Lookup returns the Family whose Component equals the given string, or
// nil if no match is found.
func Lookup(component string) *Family {
	for i := range families {
		if families[i].Component == component {
			return &families[i]
		}
	}
	return nil
}

var families = []Family{
	// ── 1. 青 ──────────────────────────────────────────────
	{
		Component: "青",
		Pinyin:    "qīng",
		Members: []Member{
			{Hanzi: "清", Pinyin: "qīng", Meaning: "clear, clean", Radical: "氵"},
			{Hanzi: "请", Pinyin: "qǐng", Meaning: "please, invite", Radical: "讠"},
			{Hanzi: "情", Pinyin: "qíng", Meaning: "emotion, feeling", Radical: "忄"},
			{Hanzi: "晴", Pinyin: "qíng", Meaning: "sunny, clear sky", Radical: "日"},
			{Hanzi: "睛", Pinyin: "jīng", Meaning: "eye, eyeball", Radical: "目"},
			{Hanzi: "精", Pinyin: "jīng", Meaning: "essence, refined", Radical: "米"},
		},
	},
	// ── 2. 方 ──────────────────────────────────────────────
	{
		Component: "方",
		Pinyin:    "fāng",
		Members: []Member{
			{Hanzi: "房", Pinyin: "fáng", Meaning: "room, house", Radical: "户"},
			{Hanzi: "放", Pinyin: "fàng", Meaning: "release, put", Radical: "攵"},
			{Hanzi: "防", Pinyin: "fáng", Meaning: "prevent, defend", Radical: "阝"},
			{Hanzi: "访", Pinyin: "fǎng", Meaning: "visit, interview", Radical: "讠"},
			{Hanzi: "纺", Pinyin: "fǎng", Meaning: "spin thread", Radical: "纟"},
			{Hanzi: "芳", Pinyin: "fāng", Meaning: "fragrant", Radical: "艹"},
		},
	},
	// ── 3. 包 ──────────────────────────────────────────────
	{
		Component: "包",
		Pinyin:    "bāo",
		Members: []Member{
			{Hanzi: "抱", Pinyin: "bào", Meaning: "hug, embrace", Radical: "扌"},
			{Hanzi: "跑", Pinyin: "pǎo", Meaning: "run", Radical: "足"},
			{Hanzi: "泡", Pinyin: "pào", Meaning: "bubble, soak", Radical: "氵"},
			{Hanzi: "炮", Pinyin: "pào", Meaning: "cannon, firecracker", Radical: "火"},
			{Hanzi: "饱", Pinyin: "bǎo", Meaning: "full (eaten enough)", Radical: "饣"},
			{Hanzi: "苞", Pinyin: "bāo", Meaning: "bud", Radical: "艹"},
		},
	},
	// ── 4. 工 ──────────────────────────────────────────────
	{
		Component: "工",
		Pinyin:    "gōng",
		Members: []Member{
			{Hanzi: "功", Pinyin: "gōng", Meaning: "merit, achievement", Radical: "力"},
			{Hanzi: "攻", Pinyin: "gōng", Meaning: "attack", Radical: "攵"},
			{Hanzi: "空", Pinyin: "kōng", Meaning: "empty, sky", Radical: "穴"},
			{Hanzi: "红", Pinyin: "hóng", Meaning: "red", Radical: "纟"},
			{Hanzi: "江", Pinyin: "jiāng", Meaning: "river", Radical: "氵"},
			{Hanzi: "虹", Pinyin: "hóng", Meaning: "rainbow", Radical: "虫"},
		},
	},
	// ── 5. 寺 ──────────────────────────────────────────────
	{
		Component: "寺",
		Pinyin:    "sì",
		Members: []Member{
			{Hanzi: "诗", Pinyin: "shī", Meaning: "poem, poetry", Radical: "讠"},
			{Hanzi: "待", Pinyin: "dài", Meaning: "wait, treat", Radical: "彳"},
			{Hanzi: "持", Pinyin: "chí", Meaning: "hold, maintain", Radical: "扌"},
			{Hanzi: "时", Pinyin: "shí", Meaning: "time", Radical: "日"},
			{Hanzi: "特", Pinyin: "tè", Meaning: "special", Radical: "牜"},
			{Hanzi: "等", Pinyin: "děng", Meaning: "wait, equal", Radical: "竹"},
		},
	},
	// ── 6. 分 ──────────────────────────────────────────────
	{
		Component: "分",
		Pinyin:    "fēn",
		Members: []Member{
			{Hanzi: "份", Pinyin: "fèn", Meaning: "portion, share", Radical: "亻"},
			{Hanzi: "纷", Pinyin: "fēn", Meaning: "numerous, confused", Radical: "纟"},
			{Hanzi: "芬", Pinyin: "fēn", Meaning: "fragrance", Radical: "艹"},
			{Hanzi: "粉", Pinyin: "fěn", Meaning: "powder, pink", Radical: "米"},
			{Hanzi: "盆", Pinyin: "pén", Meaning: "basin, pot", Radical: "皿"},
			{Hanzi: "贫", Pinyin: "pín", Meaning: "poor", Radical: "贝"},
		},
	},
	// ── 7. 马 ──────────────────────────────────────────────
	{
		Component: "马",
		Pinyin:    "mǎ",
		Members: []Member{
			{Hanzi: "妈", Pinyin: "mā", Meaning: "mother", Radical: "女"},
			{Hanzi: "吗", Pinyin: "ma", Meaning: "question particle", Radical: "口"},
			{Hanzi: "码", Pinyin: "mǎ", Meaning: "number, code", Radical: "石"},
			{Hanzi: "骂", Pinyin: "mà", Meaning: "scold, curse", Radical: "口"},
			{Hanzi: "玛", Pinyin: "mǎ", Meaning: "agate (transliteration)", Radical: "王"},
			{Hanzi: "蚂", Pinyin: "mǎ", Meaning: "ant (蚂蚁)", Radical: "虫"},
		},
	},
	// ── 8. 果 ──────────────────────────────────────────────
	{
		Component: "果",
		Pinyin:    "guǒ",
		Members: []Member{
			{Hanzi: "课", Pinyin: "kè", Meaning: "class, lesson", Radical: "讠"},
			{Hanzi: "棵", Pinyin: "kē", Meaning: "MW for trees", Radical: "木"},
			{Hanzi: "颗", Pinyin: "kē", Meaning: "MW for small round things", Radical: "页"},
			{Hanzi: "裸", Pinyin: "luǒ", Meaning: "bare, naked", Radical: "衤"},
		},
	},
	// ── 9. 交 ──────────────────────────────────────────────
	{
		Component: "交",
		Pinyin:    "jiāo",
		Members: []Member{
			{Hanzi: "校", Pinyin: "xiào", Meaning: "school", Radical: "木"},
			{Hanzi: "较", Pinyin: "jiào", Meaning: "compare, relatively", Radical: "车"},
			{Hanzi: "饺", Pinyin: "jiǎo", Meaning: "dumpling", Radical: "饣"},
			{Hanzi: "郊", Pinyin: "jiāo", Meaning: "suburb, outskirts", Radical: "阝"},
			{Hanzi: "胶", Pinyin: "jiāo", Meaning: "glue, rubber", Radical: "月"},
			{Hanzi: "效", Pinyin: "xiào", Meaning: "effect, imitate", Radical: "攵"},
		},
	},
	// ── 10. 反 ─────────────────────────────────────────────
	{
		Component: "反",
		Pinyin:    "fǎn",
		Members: []Member{
			{Hanzi: "饭", Pinyin: "fàn", Meaning: "cooked rice, meal", Radical: "饣"},
			{Hanzi: "返", Pinyin: "fǎn", Meaning: "return", Radical: "辶"},
			{Hanzi: "板", Pinyin: "bǎn", Meaning: "board, plank", Radical: "木"},
			{Hanzi: "版", Pinyin: "bǎn", Meaning: "edition, printing plate", Radical: "片"},
			{Hanzi: "叛", Pinyin: "pàn", Meaning: "betray, rebel", Radical: "又"},
		},
	},
	// ── 11. 占 ─────────────────────────────────────────────
	{
		Component: "占",
		Pinyin:    "zhàn",
		Members: []Member{
			{Hanzi: "站", Pinyin: "zhàn", Meaning: "stand, station", Radical: "立"},
			{Hanzi: "点", Pinyin: "diǎn", Meaning: "dot, point, o'clock", Radical: "灬"},
			{Hanzi: "店", Pinyin: "diàn", Meaning: "shop, store", Radical: "广"},
			{Hanzi: "战", Pinyin: "zhàn", Meaning: "war, battle", Radical: "戈"},
			{Hanzi: "粘", Pinyin: "zhān", Meaning: "sticky, paste", Radical: "米"},
		},
	},
	// ── 12. 令 ─────────────────────────────────────────────
	{
		Component: "令",
		Pinyin:    "lìng",
		Members: []Member{
			{Hanzi: "领", Pinyin: "lǐng", Meaning: "lead, collar", Radical: "页"},
			{Hanzi: "零", Pinyin: "líng", Meaning: "zero", Radical: "雨"},
			{Hanzi: "铃", Pinyin: "líng", Meaning: "bell", Radical: "钅"},
			{Hanzi: "邻", Pinyin: "lín", Meaning: "neighbor", Radical: "阝"},
			{Hanzi: "怜", Pinyin: "lián", Meaning: "pity, sympathize", Radical: "忄"},
		},
	},
	// ── 13. 生 ─────────────────────────────────────────────
	{
		Component: "生",
		Pinyin:    "shēng",
		Members: []Member{
			{Hanzi: "星", Pinyin: "xīng", Meaning: "star", Radical: "日"},
			{Hanzi: "性", Pinyin: "xìng", Meaning: "nature, gender", Radical: "忄"},
			{Hanzi: "姓", Pinyin: "xìng", Meaning: "surname", Radical: "女"},
			{Hanzi: "胜", Pinyin: "shèng", Meaning: "win, victory", Radical: "月"},
			{Hanzi: "牲", Pinyin: "shēng", Meaning: "livestock, sacrifice", Radical: "牜"},
		},
	},
	// ── 14. 兆 ─────────────────────────────────────────────
	{
		Component: "兆",
		Pinyin:    "zhào",
		Members: []Member{
			{Hanzi: "跳", Pinyin: "tiào", Meaning: "jump", Radical: "足"},
			{Hanzi: "挑", Pinyin: "tiāo", Meaning: "pick, carry on shoulder", Radical: "扌"},
			{Hanzi: "桃", Pinyin: "táo", Meaning: "peach", Radical: "木"},
			{Hanzi: "逃", Pinyin: "táo", Meaning: "escape, flee", Radical: "辶"},
		},
	},
	// ── 15. 合 ─────────────────────────────────────────────
	{
		Component: "合",
		Pinyin:    "hé",
		Members: []Member{
			{Hanzi: "给", Pinyin: "gěi", Meaning: "give", Radical: "纟"},
			{Hanzi: "哈", Pinyin: "hā", Meaning: "laugh (onomatopoeia)", Radical: "口"},
			{Hanzi: "答", Pinyin: "dá", Meaning: "answer, reply", Radical: "竹"},
			{Hanzi: "鸽", Pinyin: "gē", Meaning: "pigeon, dove", Radical: "鸟"},
			{Hanzi: "盒", Pinyin: "hé", Meaning: "box, case", Radical: "皿"},
		},
	},
	// ── 16. 巴 ─────────────────────────────────────────────
	{
		Component: "巴",
		Pinyin:    "bā",
		Members: []Member{
			{Hanzi: "把", Pinyin: "bǎ", Meaning: "hold, handle, BA-particle", Radical: "扌"},
			{Hanzi: "吧", Pinyin: "ba", Meaning: "suggestion particle", Radical: "口"},
			{Hanzi: "爸", Pinyin: "bà", Meaning: "father", Radical: "父"},
			{Hanzi: "芭", Pinyin: "bā", Meaning: "banana (芭蕉)", Radical: "艹"},
			{Hanzi: "疤", Pinyin: "bā", Meaning: "scar", Radical: "疒"},
		},
	},
	// ── 17. 皮 ─────────────────────────────────────────────
	{
		Component: "皮",
		Pinyin:    "pí",
		Members: []Member{
			{Hanzi: "被", Pinyin: "bèi", Meaning: "passive marker, quilt", Radical: "衤"},
			{Hanzi: "坡", Pinyin: "pō", Meaning: "slope, hillside", Radical: "土"},
			{Hanzi: "波", Pinyin: "bō", Meaning: "wave", Radical: "氵"},
			{Hanzi: "破", Pinyin: "pò", Meaning: "break, broken", Radical: "石"},
			{Hanzi: "疲", Pinyin: "pí", Meaning: "tired, fatigued", Radical: "疒"},
		},
	},
	// ── 18. 且 ─────────────────────────────────────────────
	{
		Component: "且",
		Pinyin:    "qiě",
		Members: []Member{
			{Hanzi: "姐", Pinyin: "jiě", Meaning: "older sister", Radical: "女"},
			{Hanzi: "组", Pinyin: "zǔ", Meaning: "group, organize", Radical: "纟"},
			{Hanzi: "祖", Pinyin: "zǔ", Meaning: "ancestor", Radical: "礻"},
			{Hanzi: "租", Pinyin: "zū", Meaning: "rent", Radical: "禾"},
			{Hanzi: "阻", Pinyin: "zǔ", Meaning: "block, hinder", Radical: "阝"},
		},
	},
	// ── 19. 可 ─────────────────────────────────────────────
	{
		Component: "可",
		Pinyin:    "kě",
		Members: []Member{
			{Hanzi: "河", Pinyin: "hé", Meaning: "river", Radical: "氵"},
			{Hanzi: "何", Pinyin: "hé", Meaning: "what, which", Radical: "亻"},
			{Hanzi: "荷", Pinyin: "hé", Meaning: "lotus", Radical: "艹"},
			{Hanzi: "哥", Pinyin: "gē", Meaning: "older brother", Radical: "口"},
		},
	},
	// ── 20. 各 ─────────────────────────────────────────────
	{
		Component: "各",
		Pinyin:    "gè",
		Members: []Member{
			{Hanzi: "格", Pinyin: "gé", Meaning: "pattern, grid, standard", Radical: "木"},
			{Hanzi: "客", Pinyin: "kè", Meaning: "guest, customer", Radical: "宀"},
			{Hanzi: "路", Pinyin: "lù", Meaning: "road, path", Radical: "足"},
			{Hanzi: "落", Pinyin: "luò", Meaning: "fall, drop", Radical: "艹"},
			{Hanzi: "阁", Pinyin: "gé", Meaning: "pavilion, cabinet", Radical: "门"},
		},
	},
	// ── 21. 古 ─────────────────────────────────────────────
	{
		Component: "古",
		Pinyin:    "gǔ",
		Members: []Member{
			{Hanzi: "故", Pinyin: "gù", Meaning: "reason, old, therefore", Radical: "攵"},
			{Hanzi: "固", Pinyin: "gù", Meaning: "solid, firm", Radical: "囗"},
			{Hanzi: "苦", Pinyin: "kǔ", Meaning: "bitter, suffering", Radical: "艹"},
			{Hanzi: "湖", Pinyin: "hú", Meaning: "lake", Radical: "氵"},
			{Hanzi: "估", Pinyin: "gū", Meaning: "estimate, guess", Radical: "亻"},
		},
	},
	// ── 22. 半 ─────────────────────────────────────────────
	{
		Component: "半",
		Pinyin:    "bàn",
		Members: []Member{
			{Hanzi: "伴", Pinyin: "bàn", Meaning: "companion, partner", Radical: "亻"},
			{Hanzi: "判", Pinyin: "pàn", Meaning: "judge, decide", Radical: "刂"},
			{Hanzi: "胖", Pinyin: "pàng", Meaning: "fat, plump", Radical: "月"},
			{Hanzi: "拌", Pinyin: "bàn", Meaning: "mix, stir", Radical: "扌"},
		},
	},
	// ── 23. 申 ─────────────────────────────────────────────
	{
		Component: "申",
		Pinyin:    "shēn",
		Members: []Member{
			{Hanzi: "伸", Pinyin: "shēn", Meaning: "stretch, extend", Radical: "亻"},
			{Hanzi: "神", Pinyin: "shén", Meaning: "god, spirit", Radical: "礻"},
			{Hanzi: "审", Pinyin: "shěn", Meaning: "examine, trial", Radical: "宀"},
			{Hanzi: "绅", Pinyin: "shēn", Meaning: "gentry", Radical: "纟"},
		},
	},
	// ── 24. 昌 ─────────────────────────────────────────────
	{
		Component: "昌",
		Pinyin:    "chāng",
		Members: []Member{
			{Hanzi: "唱", Pinyin: "chàng", Meaning: "sing", Radical: "口"},
			{Hanzi: "猖", Pinyin: "chāng", Meaning: "rampant, wild", Radical: "犭"},
			{Hanzi: "倡", Pinyin: "chàng", Meaning: "advocate, initiate", Radical: "亻"},
			{Hanzi: "娼", Pinyin: "chāng", Meaning: "prostitute", Radical: "女"},
		},
	},
	// ── 25. 仑 ─────────────────────────────────────────────
	{
		Component: "仑",
		Pinyin:    "lún",
		Members: []Member{
			{Hanzi: "论", Pinyin: "lùn", Meaning: "theory, discuss", Radical: "讠"},
			{Hanzi: "轮", Pinyin: "lún", Meaning: "wheel, turn", Radical: "车"},
			{Hanzi: "伦", Pinyin: "lún", Meaning: "ethics, order", Radical: "亻"},
			{Hanzi: "抡", Pinyin: "lūn", Meaning: "swing, brandish", Radical: "扌"},
			{Hanzi: "沦", Pinyin: "lún", Meaning: "sink, decline", Radical: "氵"},
		},
	},
	// ── 26. 卖 ─────────────────────────────────────────────
	{
		Component: "卖",
		Pinyin:    "mài",
		Members: []Member{
			{Hanzi: "读", Pinyin: "dú", Meaning: "read", Radical: "讠"},
			{Hanzi: "续", Pinyin: "xù", Meaning: "continue", Radical: "纟"},
			{Hanzi: "赎", Pinyin: "shú", Meaning: "redeem, ransom", Radical: "贝"},
			{Hanzi: "犊", Pinyin: "dú", Meaning: "calf", Radical: "牜"},
		},
	},
	// ── 27. 甫 ─────────────────────────────────────────────
	{
		Component: "甫",
		Pinyin:    "fǔ",
		Members: []Member{
			{Hanzi: "辅", Pinyin: "fǔ", Meaning: "assist, supplement", Radical: "车"},
			{Hanzi: "铺", Pinyin: "pù", Meaning: "shop, spread", Radical: "钅"},
			{Hanzi: "捕", Pinyin: "bǔ", Meaning: "catch, arrest", Radical: "扌"},
			{Hanzi: "浦", Pinyin: "pǔ", Meaning: "riverbank, shore", Radical: "氵"},
		},
	},
	// ── 28. 奇 ─────────────────────────────────────────────
	{
		Component: "奇",
		Pinyin:    "qí",
		Members: []Member{
			{Hanzi: "骑", Pinyin: "qí", Meaning: "ride (horse/bike)", Radical: "马"},
			{Hanzi: "寄", Pinyin: "jì", Meaning: "send, mail", Radical: "宀"},
			{Hanzi: "椅", Pinyin: "yǐ", Meaning: "chair", Radical: "木"},
			{Hanzi: "崎", Pinyin: "qí", Meaning: "rugged, uneven", Radical: "山"},
			{Hanzi: "琦", Pinyin: "qí", Meaning: "fine jade", Radical: "王"},
		},
	},
	// ── 29. 莫 ─────────────────────────────────────────────
	{
		Component: "莫",
		Pinyin:    "mò",
		Members: []Member{
			{Hanzi: "模", Pinyin: "mó", Meaning: "model, pattern", Radical: "木"},
			{Hanzi: "摸", Pinyin: "mō", Meaning: "touch, feel", Radical: "扌"},
			{Hanzi: "膜", Pinyin: "mó", Meaning: "membrane, film", Radical: "月"},
			{Hanzi: "漠", Pinyin: "mò", Meaning: "desert, indifferent", Radical: "氵"},
			{Hanzi: "幕", Pinyin: "mù", Meaning: "curtain, screen", Radical: "巾"},
		},
	},
	// ── 30. 乍 ─────────────────────────────────────────────
	{
		Component: "乍",
		Pinyin:    "zhà",
		Members: []Member{
			{Hanzi: "作", Pinyin: "zuò", Meaning: "do, make, write", Radical: "亻"},
			{Hanzi: "昨", Pinyin: "zuó", Meaning: "yesterday", Radical: "日"},
			{Hanzi: "炸", Pinyin: "zhà", Meaning: "explode, deep-fry", Radical: "火"},
			{Hanzi: "诈", Pinyin: "zhà", Meaning: "cheat, swindle", Radical: "讠"},
		},
	},
	// ── 31. 监 ─────────────────────────────────────────────
	{
		Component: "监",
		Pinyin:    "jiān",
		Members: []Member{
			{Hanzi: "蓝", Pinyin: "lán", Meaning: "blue", Radical: "艹"},
			{Hanzi: "篮", Pinyin: "lán", Meaning: "basket", Radical: "竹"},
			{Hanzi: "鉴", Pinyin: "jiàn", Meaning: "mirror, inspect", Radical: "金"},
			{Hanzi: "槛", Pinyin: "kǎn", Meaning: "threshold, railing", Radical: "木"},
		},
	},
	// ── 32. 成 ─────────────────────────────────────────────
	{
		Component: "成",
		Pinyin:    "chéng",
		Members: []Member{
			{Hanzi: "城", Pinyin: "chéng", Meaning: "city, wall", Radical: "土"},
			{Hanzi: "诚", Pinyin: "chéng", Meaning: "sincere, honest", Radical: "讠"},
			{Hanzi: "盛", Pinyin: "shèng", Meaning: "flourishing, abundant", Radical: "皿"},
			{Hanzi: "程", Pinyin: "chéng", Meaning: "journey, procedure", Radical: "禾"},
		},
	},
	// ── 33. 丁 ─────────────────────────────────────────────
	{
		Component: "丁",
		Pinyin:    "dīng",
		Members: []Member{
			{Hanzi: "顶", Pinyin: "dǐng", Meaning: "top, crown", Radical: "页"},
			{Hanzi: "订", Pinyin: "dìng", Meaning: "order, subscribe", Radical: "讠"},
			{Hanzi: "盯", Pinyin: "dīng", Meaning: "stare at", Radical: "目"},
			{Hanzi: "叮", Pinyin: "dīng", Meaning: "sting, urge repeatedly", Radical: "口"},
			{Hanzi: "灯", Pinyin: "dēng", Meaning: "lamp, light", Radical: "火"},
		},
	},
	// ── 34. 东 ─────────────────────────────────────────────
	{
		Component: "东",
		Pinyin:    "dōng",
		Members: []Member{
			{Hanzi: "冻", Pinyin: "dòng", Meaning: "freeze", Radical: "冫"},
			{Hanzi: "栋", Pinyin: "dòng", Meaning: "MW for buildings, ridgepole", Radical: "木"},
			{Hanzi: "陈", Pinyin: "chén", Meaning: "display, surname Chen", Radical: "阝"},
			{Hanzi: "练", Pinyin: "liàn", Meaning: "practice, drill", Radical: "纟"},
		},
	},
	// ── 35. 非 ─────────────────────────────────────────────
	{
		Component: "非",
		Pinyin:    "fēi",
		Members: []Member{
			{Hanzi: "飞", Pinyin: "fēi", Meaning: "fly", Radical: "飞"},
			{Hanzi: "排", Pinyin: "pái", Meaning: "row, arrange", Radical: "扌"},
			{Hanzi: "悲", Pinyin: "bēi", Meaning: "sad, sorrow", Radical: "心"},
			{Hanzi: "辈", Pinyin: "bèi", Meaning: "generation, lifetime", Radical: "车"},
			{Hanzi: "菲", Pinyin: "fēi", Meaning: "luxuriant, fragrant", Radical: "艹"},
		},
	},
	// ── 36. 戋 ─────────────────────────────────────────────
	{
		Component: "戋",
		Pinyin:    "jiān",
		Members: []Member{
			{Hanzi: "浅", Pinyin: "qiǎn", Meaning: "shallow", Radical: "氵"},
			{Hanzi: "钱", Pinyin: "qián", Meaning: "money", Radical: "钅"},
			{Hanzi: "线", Pinyin: "xiàn", Meaning: "line, thread", Radical: "纟"},
			{Hanzi: "贱", Pinyin: "jiàn", Meaning: "cheap, lowly", Radical: "贝"},
			{Hanzi: "践", Pinyin: "jiàn", Meaning: "trample, carry out", Radical: "足"},
		},
	},
	// ── 37. 尧 ─────────────────────────────────────────────
	{
		Component: "尧",
		Pinyin:    "yáo",
		Members: []Member{
			{Hanzi: "烧", Pinyin: "shāo", Meaning: "burn, cook", Radical: "火"},
			{Hanzi: "浇", Pinyin: "jiāo", Meaning: "water, pour", Radical: "氵"},
			{Hanzi: "绕", Pinyin: "rào", Meaning: "wind around, detour", Radical: "纟"},
			{Hanzi: "挠", Pinyin: "náo", Meaning: "scratch, hinder", Radical: "扌"},
		},
	},
	// ── 38. 良 ─────────────────────────────────────────────
	{
		Component: "良",
		Pinyin:    "liáng",
		Members: []Member{
			{Hanzi: "浪", Pinyin: "làng", Meaning: "wave, unrestrained", Radical: "氵"},
			{Hanzi: "娘", Pinyin: "niáng", Meaning: "mother, young woman", Radical: "女"},
			{Hanzi: "粮", Pinyin: "liáng", Meaning: "grain, food", Radical: "米"},
			{Hanzi: "狼", Pinyin: "láng", Meaning: "wolf", Radical: "犭"},
			{Hanzi: "郎", Pinyin: "láng", Meaning: "young man, husband", Radical: "阝"},
		},
	},
	// ── 39. 也 ─────────────────────────────────────────────
	{
		Component: "也",
		Pinyin:    "yě",
		Members: []Member{
			{Hanzi: "他", Pinyin: "tā", Meaning: "he, him", Radical: "亻"},
			{Hanzi: "她", Pinyin: "tā", Meaning: "she, her", Radical: "女"},
			{Hanzi: "地", Pinyin: "dì", Meaning: "earth, ground", Radical: "土"},
			{Hanzi: "池", Pinyin: "chí", Meaning: "pond, pool", Radical: "氵"},
			{Hanzi: "驰", Pinyin: "chí", Meaning: "gallop, speed", Radical: "马"},
		},
	},
	// ── 40. 主 ─────────────────────────────────────────────
	{
		Component: "主",
		Pinyin:    "zhǔ",
		Members: []Member{
			{Hanzi: "住", Pinyin: "zhù", Meaning: "live, reside", Radical: "亻"},
			{Hanzi: "注", Pinyin: "zhù", Meaning: "pour, annotate", Radical: "氵"},
			{Hanzi: "柱", Pinyin: "zhù", Meaning: "pillar, column", Radical: "木"},
			{Hanzi: "驻", Pinyin: "zhù", Meaning: "station, be posted", Radical: "马"},
			{Hanzi: "往", Pinyin: "wǎng", Meaning: "go toward, past", Radical: "彳"},
		},
	},
	// ── 41. 白 ─────────────────────────────────────────────
	{
		Component: "白",
		Pinyin:    "bái",
		Members: []Member{
			{Hanzi: "拍", Pinyin: "pāi", Meaning: "clap, slap, take photo", Radical: "扌"},
			{Hanzi: "伯", Pinyin: "bó", Meaning: "uncle (father's elder brother)", Radical: "亻"},
			{Hanzi: "迫", Pinyin: "pò", Meaning: "compel, press", Radical: "辶"},
			{Hanzi: "怕", Pinyin: "pà", Meaning: "fear, afraid", Radical: "忄"},
			{Hanzi: "柏", Pinyin: "bǎi", Meaning: "cypress tree", Radical: "木"},
		},
	},
	// ── 42. 舌 ─────────────────────────────────────────────
	{
		Component: "舌",
		Pinyin:    "shé",
		Members: []Member{
			{Hanzi: "话", Pinyin: "huà", Meaning: "speech, word", Radical: "讠"},
			{Hanzi: "活", Pinyin: "huó", Meaning: "live, alive", Radical: "氵"},
			{Hanzi: "括", Pinyin: "kuò", Meaning: "include, enclose", Radical: "扌"},
			{Hanzi: "刮", Pinyin: "guā", Meaning: "scrape, blow (wind)", Radical: "刂"},
		},
	},
	// ── 43. 相 ─────────────────────────────────────────────
	{
		Component: "相",
		Pinyin:    "xiāng",
		Members: []Member{
			{Hanzi: "想", Pinyin: "xiǎng", Meaning: "think, want", Radical: "心"},
			{Hanzi: "箱", Pinyin: "xiāng", Meaning: "box, chest", Radical: "竹"},
			{Hanzi: "霜", Pinyin: "shuāng", Meaning: "frost", Radical: "雨"},
			{Hanzi: "湘", Pinyin: "xiāng", Meaning: "Hunan (province)", Radical: "氵"},
		},
	},
	// ── 44. 尚 ─────────────────────────────────────────────
	{
		Component: "尚",
		Pinyin:    "shàng",
		Members: []Member{
			{Hanzi: "常", Pinyin: "cháng", Meaning: "often, normal", Radical: "巾"},
			{Hanzi: "赏", Pinyin: "shǎng", Meaning: "reward, appreciate", Radical: "贝"},
			{Hanzi: "裳", Pinyin: "shang", Meaning: "lower garment", Radical: "衤"},
			{Hanzi: "躺", Pinyin: "tǎng", Meaning: "lie down", Radical: "身"},
			{Hanzi: "淌", Pinyin: "tǎng", Meaning: "flow, drip", Radical: "氵"},
		},
	},
	// ── 45. 夹 ─────────────────────────────────────────────
	{
		Component: "夹",
		Pinyin:    "jiā",
		Members: []Member{
			{Hanzi: "峡", Pinyin: "xiá", Meaning: "gorge, strait", Radical: "山"},
			{Hanzi: "狭", Pinyin: "xiá", Meaning: "narrow", Radical: "犭"},
			{Hanzi: "颊", Pinyin: "jiá", Meaning: "cheek", Radical: "页"},
			{Hanzi: "侠", Pinyin: "xiá", Meaning: "chivalrous, hero", Radical: "亻"},
		},
	},
}
