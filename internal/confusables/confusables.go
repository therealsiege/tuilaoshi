package confusables

// Pair is a set of commonly confused characters.
type Pair struct {
	Chars    []Confused // the confusable characters (2-4)
	Tip      string     // how to distinguish them
	Category string     // "visual", "tonal", "semantic"
}

// Confused describes one character within a confusion group.
type Confused struct {
	Hanzi   string
	Pinyin  string
	Meaning string
	Key     string // the distinguishing feature
}

// Pairs returns all confusion pair data.
func Pairs() []Pair {
	return pairs
}

var pairs = []Pair{
	// ----------------------------------------------------------------
	// VISUALLY SIMILAR
	// ----------------------------------------------------------------
	{
		Category: "visual",
		Tip:      "己 is open at top-right; 已 has the middle stroke touching the right side, closing the gap halfway; 巳 is fully enclosed — the middle stroke meets the top.",
		Chars: []Confused{
			{Hanzi: "已", Pinyin: "yǐ", Meaning: "already", Key: "middle stroke touches right wall halfway"},
			{Hanzi: "己", Pinyin: "jǐ", Meaning: "self", Key: "open at top-right, middle stroke does not touch"},
			{Hanzi: "巳", Pinyin: "sì", Meaning: "sixth earthly branch", Key: "fully enclosed, middle stroke meets the top"},
		},
	},
	{
		Category: "visual",
		Tip:      "大 has no dot; 太 adds a dot at the lower-right; 犬 adds a dot at the upper-right.",
		Chars: []Confused{
			{Hanzi: "大", Pinyin: "dà", Meaning: "big", Key: "no dot"},
			{Hanzi: "太", Pinyin: "tài", Meaning: "too, extremely", Key: "dot at bottom-right below the stroke"},
			{Hanzi: "犬", Pinyin: "quǎn", Meaning: "dog", Key: "dot at upper-right above the stroke"},
		},
	},
	{
		Category: "visual",
		Tip:      "土 has a longer bottom stroke (earth is wide); 士 has a longer top stroke (a scholar holds his head high).",
		Chars: []Confused{
			{Hanzi: "土", Pinyin: "tǔ", Meaning: "earth, soil", Key: "bottom horizontal stroke is longer"},
			{Hanzi: "士", Pinyin: "shì", Meaning: "scholar, warrior", Key: "top horizontal stroke is longer"},
		},
	},
	{
		Category: "visual",
		Tip:      "末 has a longer top stroke (the end caps everything); 未 has a longer bottom stroke (not yet — still growing upward).",
		Chars: []Confused{
			{Hanzi: "末", Pinyin: "mò", Meaning: "end, tip", Key: "top horizontal stroke is longer"},
			{Hanzi: "未", Pinyin: "wèi", Meaning: "not yet", Key: "bottom horizontal stroke is longer"},
		},
	},
	{
		Category: "visual",
		Tip:      "买 (buy) has no top stroke — you give money away; 卖 (sell) has 十 on top — you put goods on display.",
		Chars: []Confused{
			{Hanzi: "买", Pinyin: "mǎi", Meaning: "buy", Key: "no extra stroke on top"},
			{Hanzi: "卖", Pinyin: "mài", Meaning: "sell", Key: "十 (cross) on top"},
		},
	},
	{
		Category: "visual",
		Tip:      "干 has two horizontal strokes; 千 tilts the top stroke into a left-falling piě; 于 hooks the bottom stroke leftward.",
		Chars: []Confused{
			{Hanzi: "干", Pinyin: "gān", Meaning: "dry, to do", Key: "two flat horizontal strokes"},
			{Hanzi: "千", Pinyin: "qiān", Meaning: "thousand", Key: "top stroke is a left-falling piě (撇)"},
			{Hanzi: "于", Pinyin: "yú", Meaning: "in, at, to", Key: "bottom stroke hooks to the left"},
		},
	},
	{
		Category: "visual",
		Tip:      "日 is tall and narrow (sun); 目 is taller with two inner strokes (eye has pupils); 曰 is wide and flat (a mouth speaking).",
		Chars: []Confused{
			{Hanzi: "日", Pinyin: "rì", Meaning: "sun, day", Key: "narrow rectangle, one inner horizontal"},
			{Hanzi: "目", Pinyin: "mù", Meaning: "eye", Key: "taller, two inner horizontals"},
			{Hanzi: "曰", Pinyin: "yuē", Meaning: "to say (classical)", Key: "wide and flat, wider than 日"},
		},
	},
	{
		Category: "visual",
		Tip:      "人 spreads its strokes apart symmetrically; 入 tucks the left stroke under the right, with the right piě extending higher.",
		Chars: []Confused{
			{Hanzi: "人", Pinyin: "rén", Meaning: "person", Key: "strokes spread evenly, piě starts higher"},
			{Hanzi: "入", Pinyin: "rù", Meaning: "enter", Key: "right nà stroke extends higher than left piě"},
		},
	},
	{
		Category: "visual",
		Tip:      "刀 has a curved inner stroke (a blade); 力 has a straight inner stroke going down-left (exerting force).",
		Chars: []Confused{
			{Hanzi: "刀", Pinyin: "dāo", Meaning: "knife", Key: "inner stroke curves (blade shape)"},
			{Hanzi: "力", Pinyin: "lì", Meaning: "strength, force", Key: "inner stroke goes straight down-left"},
		},
	},
	{
		Category: "visual",
		Tip:      "天 starts with two horizontal strokes then 人; 夫 starts with the horizontals but the top horizontal is crossed by a vertical that extends above it.",
		Chars: []Confused{
			{Hanzi: "天", Pinyin: "tiān", Meaning: "sky, day", Key: "top stroke does not cross above"},
			{Hanzi: "夫", Pinyin: "fū", Meaning: "husband, man", Key: "top horizontal has a stroke extending above it"},
		},
	},
	{
		Category: "visual",
		Tip:      "午 has a flat top and vertical going straight down; 牛 has the vertical stroke extending above the top horizontal.",
		Chars: []Confused{
			{Hanzi: "午", Pinyin: "wǔ", Meaning: "noon", Key: "vertical stops at the top horizontal"},
			{Hanzi: "牛", Pinyin: "niú", Meaning: "cow, ox", Key: "vertical extends above the top horizontal"},
		},
	},
	{
		Category: "visual",
		Tip:      "由 is open at the top (growing out); 甲 is open at the bottom (armor protects from above); 申 extends both up and down (stretching through).",
		Chars: []Confused{
			{Hanzi: "由", Pinyin: "yóu", Meaning: "from, by", Key: "vertical extends below the box"},
			{Hanzi: "甲", Pinyin: "jiǎ", Meaning: "first, armor", Key: "vertical extends above the box"},
			{Hanzi: "申", Pinyin: "shēn", Meaning: "to state, to extend", Key: "vertical extends both above and below"},
		},
	},
	{
		Category: "visual",
		Tip:      "折 has 扌+ 斤 (a hand breaking/folding); 拆 has 扌+ a different right side with a dot on top (a hand tearing apart).",
		Chars: []Confused{
			{Hanzi: "折", Pinyin: "zhé", Meaning: "to fold, to break", Key: "right side is 斤 (axe)"},
			{Hanzi: "拆", Pinyin: "chāi", Meaning: "to tear down, to dismantle", Key: "right side differs — dot on top"},
		},
	},
	{
		Category: "visual",
		Tip:      "拔 has 扌+ 犮 (the right component has a nà stroke); 拨 has 扌+ 发 minus the top (right component ends with a dot/piě).",
		Chars: []Confused{
			{Hanzi: "拔", Pinyin: "bá", Meaning: "to pull out", Key: "right side: 犮 with a nà (捺) at bottom-right"},
			{Hanzi: "拨", Pinyin: "bō", Meaning: "to poke, to stir", Key: "right side: like 发 — ends differently"},
		},
	},
	{
		Category: "visual",
		Tip:      "Same pronunciation fēng but different radicals reveal meaning: 山 (mountain) for peak, 钅(metal) for sharp edge, 虫 (insect) for bee.",
		Chars: []Confused{
			{Hanzi: "峰", Pinyin: "fēng", Meaning: "mountain peak", Key: "radical 山 (mountain)"},
			{Hanzi: "锋", Pinyin: "fēng", Meaning: "sharp edge, vanguard", Key: "radical 钅 (metal)"},
			{Hanzi: "蜂", Pinyin: "fēng", Meaning: "bee, wasp", Key: "radical 虫 (insect)"},
		},
	},
	{
		Category: "visual",
		Tip:      "戊 is empty inside (wù); 戌 has a single dot inside (xū); 戍 has a dot at the top (shù); 戎 has 十 inside (róng).",
		Chars: []Confused{
			{Hanzi: "戊", Pinyin: "wù", Meaning: "fifth heavenly stem", Key: "empty inside — nothing within the frame"},
			{Hanzi: "戌", Pinyin: "xū", Meaning: "eleventh earthly branch", Key: "one horizontal stroke inside"},
			{Hanzi: "戍", Pinyin: "shù", Meaning: "to garrison", Key: "a dot at the top inside"},
			{Hanzi: "戎", Pinyin: "róng", Meaning: "military, weapons", Key: "十 (cross) inside the frame"},
		},
	},
	{
		Category: "visual",
		Tip:      "冫 (two dots, ice radical) vs 氵 (three dots, water radical). Count the dots on the left side.",
		Chars: []Confused{
			{Hanzi: "冫", Pinyin: "bīng", Meaning: "ice radical", Key: "two dots — fewer strokes"},
			{Hanzi: "氵", Pinyin: "shuǐ", Meaning: "water radical", Key: "three dots — one more than ice"},
		},
	},
	{
		Category: "visual",
		Tip:      "壁 has 土 at the bottom (earthen wall); 璧 has 玉 at the bottom (jade disc). Check the bottom component.",
		Chars: []Confused{
			{Hanzi: "壁", Pinyin: "bì", Meaning: "wall", Key: "bottom is 土 (earth)"},
			{Hanzi: "璧", Pinyin: "bì", Meaning: "jade disc", Key: "bottom is 玉 (jade)"},
		},
	},
	{
		Category: "visual",
		Tip:      "衤(clothing radical, left side) has a dot; 礻(spirit/altar radical, left side) has no dot. Clothes need a button (dot).",
		Chars: []Confused{
			{Hanzi: "衤", Pinyin: "yī", Meaning: "clothing radical", Key: "has an extra dot (button on clothes)"},
			{Hanzi: "礻", Pinyin: "shì", Meaning: "spirit/altar radical", Key: "no extra dot"},
		},
	},
	{
		Category: "visual",
		Tip:      "睛 has 目 radical (eye) — eyeball; 晴 has 日 radical (sun) — sunny; 清 has 氵radical (water) — clear; 情 has 忄 radical (heart) — emotion; 请 has 讠 radical (speech) — to invite/please.",
		Chars: []Confused{
			{Hanzi: "睛", Pinyin: "jīng", Meaning: "eyeball", Key: "目 radical (eye)"},
			{Hanzi: "晴", Pinyin: "qíng", Meaning: "sunny, clear sky", Key: "日 radical (sun)"},
			{Hanzi: "清", Pinyin: "qīng", Meaning: "clear, clean", Key: "氵 radical (water)"},
			{Hanzi: "情", Pinyin: "qíng", Meaning: "emotion, feeling", Key: "忄 radical (heart)"},
		},
	},
	{
		Category: "visual",
		Tip:      "辨 has a 刂 (knife) in the middle — to differentiate (cut apart); 辩 has 言 in the middle — to debate (words); 辫 has 纟 in the middle — a braid (silk/thread).",
		Chars: []Confused{
			{Hanzi: "辨", Pinyin: "biàn", Meaning: "to differentiate", Key: "middle is 刂 (knife — cutting apart ideas)"},
			{Hanzi: "辩", Pinyin: "biàn", Meaning: "to debate, to argue", Key: "middle is 言 (speech — arguing with words)"},
			{Hanzi: "辫", Pinyin: "biàn", Meaning: "braid, pigtail", Key: "middle is 纟 (silk — threading hair)"},
		},
	},
	{
		Category: "visual",
		Tip:      "贝 is the simplified shell/money radical; 见 is simplified see/perceive. Bottom differs: 贝 has two inner dots, 见 has 儿 (legs) at the bottom.",
		Chars: []Confused{
			{Hanzi: "贝", Pinyin: "bèi", Meaning: "shell, money radical", Key: "two dots inside, closed bottom"},
			{Hanzi: "见", Pinyin: "jiàn", Meaning: "to see", Key: "儿 (legs) at bottom — person looking"},
		},
	},
	{
		Category: "visual",
		Tip:      "候 has a middle vertical stroke (the small 丨 between 亻 and the right side); 侯 does not. Think: you wait (候) for the extra stroke.",
		Chars: []Confused{
			{Hanzi: "候", Pinyin: "hòu", Meaning: "to wait; time/season", Key: "has a short vertical stroke in the middle"},
			{Hanzi: "侯", Pinyin: "hóu", Meaning: "marquis, nobleman", Key: "no middle vertical stroke"},
		},
	},
	{
		Category: "visual",
		Tip:      "采 has a claw (爫) on top and a tree (木) below — to pick/pluck; 彩 adds 彡 (decoration strokes) — color is decorated; 踩 adds 足 (foot) — to step on.",
		Chars: []Confused{
			{Hanzi: "采", Pinyin: "cǎi", Meaning: "to pick, to pluck", Key: "just 爫 + 木, no extra radical"},
			{Hanzi: "彩", Pinyin: "cǎi", Meaning: "color, splendor", Key: "adds 彡 (decorative strokes) on the right"},
			{Hanzi: "踩", Pinyin: "cǎi", Meaning: "to step on", Key: "adds 足 (foot) on the left"},
		},
	},
	{
		Category: "visual",
		Tip:      "Both share a similar right side but differ on the left: 冒 has 目 below 日-like cap (risking your eyes); 帽 adds 巾 radical (cloth) on the left — a hat is cloth.",
		Chars: []Confused{
			{Hanzi: "冒", Pinyin: "mào", Meaning: "to risk; to emit", Key: "no radical on the left — just the core character"},
			{Hanzi: "帽", Pinyin: "mào", Meaning: "hat, cap", Key: "巾 (cloth) radical on the left"},
		},
	},

	// ----------------------------------------------------------------
	// TONALLY CONFUSING
	// ----------------------------------------------------------------
	{
		Category: "tonal",
		Tip:      "All share the sound qīng/qíng/qǐng. The radical reveals the meaning: 讠 = speech (please/invite), 氵= water (clear), 忄 = heart (emotion), 日 = sun (sunny), 青 is the base character (green/young).",
		Chars: []Confused{
			{Hanzi: "请", Pinyin: "qǐng", Meaning: "please, to invite", Key: "tone 3; 讠 radical (speech) — asking politely"},
			{Hanzi: "清", Pinyin: "qīng", Meaning: "clear, clean", Key: "tone 1; 氵 radical (water) — clear water"},
			{Hanzi: "情", Pinyin: "qíng", Meaning: "emotion, feeling", Key: "tone 2; 忄 radical (heart) — feelings of the heart"},
			{Hanzi: "晴", Pinyin: "qíng", Meaning: "sunny, clear weather", Key: "tone 2; 日 radical (sun) — sun is out"},
			{Hanzi: "青", Pinyin: "qīng", Meaning: "green, young", Key: "tone 1; standalone — the phonetic component itself"},
		},
	},
	{
		Category: "tonal",
		Tip:      "买 (tone 3) = buy — you pick up goods; 卖 (tone 4) = sell — you push goods out. Falling tone (4) is more forceful, like a seller calling out.",
		Chars: []Confused{
			{Hanzi: "买", Pinyin: "mǎi", Meaning: "to buy", Key: "tone 3 (dipping) — receiving"},
			{Hanzi: "卖", Pinyin: "mài", Meaning: "to sell", Key: "tone 4 (falling) — pushing out"},
		},
	},
	{
		Category: "tonal",
		Tip:      "All are pronounced shì (tone 4), but meanings and radicals differ. 是 (to be) is fundamental; 事 (matter) has no clear radical hint — just memorize; 市 has 巾 (cloth/market stalls); 试 has 讠 (speech — testing with words).",
		Chars: []Confused{
			{Hanzi: "是", Pinyin: "shì", Meaning: "to be, yes", Key: "most common character — the copula verb"},
			{Hanzi: "事", Pinyin: "shì", Meaning: "matter, affair", Key: "complex structure — things/events"},
			{Hanzi: "市", Pinyin: "shì", Meaning: "city, market", Key: "巾 (cloth) at bottom — market stalls"},
			{Hanzi: "试", Pinyin: "shì", Meaning: "to try, to test", Key: "讠 (speech) radical — testing with questions"},
		},
	},
	{
		Category: "tonal",
		Tip:      "All pronounced zuò (tone 4). 做 means to do/make (general action); 作 means to work/create (more formal); 坐 means to sit (two people 人人 on earth 土); 座 means seat/place (广 shelter + 坐).",
		Chars: []Confused{
			{Hanzi: "做", Pinyin: "zuò", Meaning: "to do, to make", Key: "亻 + 故 — a person acting"},
			{Hanzi: "作", Pinyin: "zuò", Meaning: "to work, to create", Key: "亻 + 乍 — formal/literary creation"},
			{Hanzi: "坐", Pinyin: "zuò", Meaning: "to sit", Key: "two 人 on 土 — people sitting on earth"},
			{Hanzi: "座", Pinyin: "zuò", Meaning: "seat, place", Key: "广 (shelter) over 坐 — a designated seat"},
		},
	},
	{
		Category: "tonal",
		Tip:      "All share the jī/jǐ/jì sound. 记 (tone 4) = to remember (讠 speech/words); 几 (tone 3) = how many/几 table; 机 (tone 1) = machine (木 wood — originally wooden loom); 鸡 (tone 1) = chicken (鸟 bird radical).",
		Chars: []Confused{
			{Hanzi: "记", Pinyin: "jì", Meaning: "to remember, to record", Key: "tone 4; 讠 radical (words you remember)"},
			{Hanzi: "几", Pinyin: "jǐ", Meaning: "how many; small table", Key: "tone 3; simple character — a small table shape"},
			{Hanzi: "机", Pinyin: "jī", Meaning: "machine, opportunity", Key: "tone 1; 木 radical (wooden mechanism)"},
			{Hanzi: "鸡", Pinyin: "jī", Meaning: "chicken", Key: "tone 1; 鸟 radical (it is a bird)"},
		},
	},
	{
		Category: "tonal",
		Tip:      "四/死 are both sǐ/sì — four vs. death. 四 (tone 4) is the number, enclosed box with two inner strokes; 死 (tone 3) has 歹 (bad) + 匕 (person) — a person gone bad.",
		Chars: []Confused{
			{Hanzi: "四", Pinyin: "sì", Meaning: "four", Key: "tone 4; enclosed box shape — just a number"},
			{Hanzi: "死", Pinyin: "sǐ", Meaning: "to die, death", Key: "tone 3; 歹 (bad/bone) radical — decay"},
		},
	},
	{
		Category: "tonal",
		Tip:      "All pronounced with the base sound mǎ/má/mā/mà. This is the classic tone drill. 妈 (tone 1) = mother; 麻 (tone 2) = hemp/numb; 马 (tone 3) = horse; 骂 (tone 4) = to scold.",
		Chars: []Confused{
			{Hanzi: "妈", Pinyin: "mā", Meaning: "mother", Key: "tone 1 (flat); 女 radical — female parent"},
			{Hanzi: "麻", Pinyin: "má", Meaning: "hemp, numb", Key: "tone 2 (rising); 广 + 林 — plants under a shelter"},
			{Hanzi: "马", Pinyin: "mǎ", Meaning: "horse", Key: "tone 3 (dipping); standalone pictograph"},
			{Hanzi: "骂", Pinyin: "mà", Meaning: "to scold, to curse", Key: "tone 4 (falling); two 口 (mouths) on top — yelling"},
		},
	},
	{
		Category: "tonal",
		Tip:      "All yī/yì/yí/yǐ. Context determines meaning: 一 = one (number); 以 = by/with (preposition); 已 = already (adverb); 义 = justice/meaning.",
		Chars: []Confused{
			{Hanzi: "一", Pinyin: "yī", Meaning: "one", Key: "tone 1; single horizontal stroke"},
			{Hanzi: "以", Pinyin: "yǐ", Meaning: "by means of, with", Key: "tone 3; function word — no obvious radical hint"},
			{Hanzi: "已", Pinyin: "yǐ", Meaning: "already", Key: "tone 3; resembles 己 but middle stroke touches right"},
			{Hanzi: "义", Pinyin: "yì", Meaning: "justice, meaning", Key: "tone 4; simplified from 義 — moral righteousness"},
		},
	},
	{
		Category: "tonal",
		Tip:      "All shī/shí/shǐ/shì: 师 (tone 1) = teacher; 十 (tone 2) = ten; 使 (tone 3) = to cause/send; 世 (tone 4) = world/era.",
		Chars: []Confused{
			{Hanzi: "师", Pinyin: "shī", Meaning: "teacher, master", Key: "tone 1; originally a military unit — leader"},
			{Hanzi: "十", Pinyin: "shí", Meaning: "ten", Key: "tone 2; cross shape — just a number"},
			{Hanzi: "使", Pinyin: "shǐ", Meaning: "to cause, to send", Key: "tone 3; 亻 radical — a person sent as envoy"},
			{Hanzi: "世", Pinyin: "shì", Meaning: "world, era, generation", Key: "tone 4; three connected 十 — spanning generations"},
		},
	},
	{
		Category: "tonal",
		Tip:      "Both zhǐ but different tones and meanings: 只 (tone 3) = only/measure word; 纸 (tone 3) = paper (纟 silk radical — made from fibers); 指 (tone 3) = finger/to point (扌 hand radical).",
		Chars: []Confused{
			{Hanzi: "只", Pinyin: "zhǐ", Meaning: "only; measure word for animals", Key: "simple character, very common — 'only'"},
			{Hanzi: "纸", Pinyin: "zhǐ", Meaning: "paper", Key: "纟 (silk/fiber) radical — paper is made from fibers"},
			{Hanzi: "指", Pinyin: "zhǐ", Meaning: "finger; to point at", Key: "扌 (hand) radical — pointing with a hand"},
		},
	},
	{
		Category: "tonal",
		Tip:      "All share the base sound chéng/chēng/chèng: 成 = to become (tone 2); 城 = city (tone 2, 土 radical — earthen walls); 称 = to weigh/to call (tone 1 or tone 4 depending on meaning).",
		Chars: []Confused{
			{Hanzi: "成", Pinyin: "chéng", Meaning: "to become; to succeed", Key: "tone 2; base character — achievement"},
			{Hanzi: "城", Pinyin: "chéng", Meaning: "city, city wall", Key: "tone 2; 土 radical — earthen fortification"},
			{Hanzi: "称", Pinyin: "chēng", Meaning: "to weigh; to call/name", Key: "tone 1; 禾 radical — weighing grain"},
		},
	},
	{
		Category: "tonal",
		Tip:      "Both share the sound gōng/gōng but differ: 工 = work (tone 1, simple shape — a beam); 公 = public (tone 1, 八 + 厶 — dividing from private); 功 = merit/achievement (tone 1, 工 + 力 — work plus effort).",
		Chars: []Confused{
			{Hanzi: "工", Pinyin: "gōng", Meaning: "work, labor", Key: "simple I-beam shape — a tool or beam"},
			{Hanzi: "公", Pinyin: "gōng", Meaning: "public, fair", Key: "八 (divide) + 厶 (private) — opening up what is private"},
			{Hanzi: "功", Pinyin: "gōng", Meaning: "merit, achievement", Key: "工 (work) + 力 (strength) — effort applied"},
		},
	},

	// ----------------------------------------------------------------
	// SEMANTICALLY CONFUSING
	// ----------------------------------------------------------------
	{
		Category: "semantic",
		Tip:      "看 = to look at (active gaze, hand over eyes 手+目); 见 = to see/perceive (result of looking); 观 = to observe (formal/careful examination); 望 = to gaze into the distance (王+月+亡 — looking far at the moon).",
		Chars: []Confused{
			{Hanzi: "看", Pinyin: "kàn", Meaning: "to look at, to watch", Key: "active looking — hand (手) shading eyes (目)"},
			{Hanzi: "见", Pinyin: "jiàn", Meaning: "to see, to perceive", Key: "result of looking — you perceived it"},
			{Hanzi: "观", Pinyin: "guān", Meaning: "to observe, to view", Key: "careful, formal observation"},
			{Hanzi: "望", Pinyin: "wàng", Meaning: "to gaze, to hope", Key: "looking far away — distance and longing"},
		},
	},
	{
		Category: "semantic",
		Tip:      "说 = to say/speak (everyday); 讲 = to explain/lecture (more structured); 谈 = to discuss/chat (informal two-way); 告 = to tell/inform (one-way notification, often formal or legal).",
		Chars: []Confused{
			{Hanzi: "说", Pinyin: "shuō", Meaning: "to say, to speak", Key: "most general — everyday speech"},
			{Hanzi: "讲", Pinyin: "jiǎng", Meaning: "to explain, to lecture", Key: "structured speech — giving a talk"},
			{Hanzi: "谈", Pinyin: "tán", Meaning: "to discuss, to chat", Key: "two-way conversation — discussion"},
			{Hanzi: "告", Pinyin: "gào", Meaning: "to tell, to inform", Key: "one-way — announcing or notifying"},
		},
	},
	{
		Category: "semantic",
		Tip:      "走 = to walk/go (general movement on foot); 跑 = to run (足 foot radical + 包); 行 = to travel/go (more abstract, also 'okay/feasible').",
		Chars: []Confused{
			{Hanzi: "走", Pinyin: "zǒu", Meaning: "to walk, to go", Key: "general walking — basic locomotion"},
			{Hanzi: "跑", Pinyin: "pǎo", Meaning: "to run", Key: "足 radical — running with feet"},
			{Hanzi: "行", Pinyin: "xíng", Meaning: "to go, to travel; okay", Key: "abstract movement — also means 'feasible'"},
		},
	},
	{
		Category: "semantic",
		Tip:      "想 = to think/want/miss (心 heart + 相 — thinking with heart and mind); 思 = to ponder/contemplate (田 + 心 — field of the heart, deeper reflection); 念 = to miss/recite (今 + 心 — present in the heart).",
		Chars: []Confused{
			{Hanzi: "想", Pinyin: "xiǎng", Meaning: "to think, to want, to miss", Key: "general thinking/wanting — everyday use"},
			{Hanzi: "思", Pinyin: "sī", Meaning: "to ponder, to contemplate", Key: "deeper reflection — literary/formal thinking"},
			{Hanzi: "念", Pinyin: "niàn", Meaning: "to miss, to read aloud", Key: "held in the heart now — also 'to recite'"},
		},
	},
	{
		Category: "semantic",
		Tip:      "学 = to learn/study (acquiring knowledge); 习 = to practice/review (repeated drilling); 教 = to teach (孝 + 攵 — passing on filial knowledge).",
		Chars: []Confused{
			{Hanzi: "学", Pinyin: "xué", Meaning: "to learn, to study", Key: "acquiring new knowledge"},
			{Hanzi: "习", Pinyin: "xí", Meaning: "to practice, to review", Key: "repeating and drilling — making it habitual"},
			{Hanzi: "教", Pinyin: "jiāo", Meaning: "to teach, to instruct", Key: "transferring knowledge to others"},
		},
	},
	{
		Category: "semantic",
		Tip:      "听 = to listen (耳 ear implied); 闻 = to hear/smell (门 + 耳 — ear at the door, news reaching you); the difference is active (听) vs. passive (闻) perception.",
		Chars: []Confused{
			{Hanzi: "听", Pinyin: "tīng", Meaning: "to listen", Key: "active — deliberately paying attention"},
			{Hanzi: "闻", Pinyin: "wén", Meaning: "to hear; to smell; news", Key: "passive — something reaches your senses"},
		},
	},
	{
		Category: "semantic",
		Tip:      "穿 = to wear (for items you put on through an opening: shirts, pants, shoes); 戴 = to wear (for accessories placed on top: hats, glasses, gloves, jewelry).",
		Chars: []Confused{
			{Hanzi: "穿", Pinyin: "chuān", Meaning: "to wear (put through)", Key: "threading through — clothes you step/pull into"},
			{Hanzi: "戴", Pinyin: "dài", Meaning: "to wear (put on top)", Key: "placing on — accessories you put onto body"},
		},
	},
	{
		Category: "semantic",
		Tip:      "把 = to hold/grasp (扌 hand — holding an object); 拿 = to take/carry (合 + 手 — joining hand to object and bringing it); 拉 = to pull (扌 + 立 — hand pulling upright).",
		Chars: []Confused{
			{Hanzi: "把", Pinyin: "bǎ", Meaning: "to hold, to grasp; handle", Key: "gripping something — also the 把 grammar particle"},
			{Hanzi: "拿", Pinyin: "ná", Meaning: "to take, to carry", Key: "picking up and bringing along"},
			{Hanzi: "拉", Pinyin: "lā", Meaning: "to pull, to drag", Key: "pulling something toward you"},
		},
	},
	{
		Category: "semantic",
		Tip:      "关 = to close/turn off (general closing); 闭 = to close/shut (more formal — 门 + 才, closing a door firmly); 合 = to close/combine (bringing two halves together).",
		Chars: []Confused{
			{Hanzi: "关", Pinyin: "guān", Meaning: "to close, to turn off", Key: "everyday closing — 关门 (close door), 关灯 (turn off light)"},
			{Hanzi: "闭", Pinyin: "bì", Meaning: "to shut, to close", Key: "formal/literary — 闭嘴 (shut mouth), 闭幕 (closing ceremony)"},
			{Hanzi: "合", Pinyin: "hé", Meaning: "to close, to combine, to fit", Key: "bringing together — 合上 (close a book), 合作 (cooperate)"},
		},
	},
	{
		Category: "semantic",
		Tip:      "快 = fast (speed); 早 = early (time); 急 = urgent/anxious (emotional urgency). 快 is about velocity, 早 about timing, 急 about pressure.",
		Chars: []Confused{
			{Hanzi: "快", Pinyin: "kuài", Meaning: "fast, quick", Key: "speed of action — how fast something moves"},
			{Hanzi: "早", Pinyin: "zǎo", Meaning: "early", Key: "timing — before the expected time"},
			{Hanzi: "急", Pinyin: "jí", Meaning: "urgent, anxious", Key: "emotional pressure — needing to act now"},
		},
	},
	{
		Category: "semantic",
		Tip:      "常 = often/always (frequency — it happens regularly); 经常 is the full word; 总 = always/overall (inevitably, summing up); 往往 = tend to (pattern observed in the past).",
		Chars: []Confused{
			{Hanzi: "常", Pinyin: "cháng", Meaning: "often, always", Key: "regular frequency — 'it often happens'"},
			{Hanzi: "总", Pinyin: "zǒng", Meaning: "always, overall, general", Key: "inevitability — 'it always turns out this way'"},
			{Hanzi: "往", Pinyin: "wǎng", Meaning: "toward; past (往往 = tend to)", Key: "observed pattern — 'it tends to happen'"},
		},
	},
	{
		Category: "semantic",
		Tip:      "用 = to use (general utility — using a tool); 穿 is for wearing; 打 is a wildcard verb (hit, make a phone call, play); context determines 打's meaning.",
		Chars: []Confused{
			{Hanzi: "用", Pinyin: "yòng", Meaning: "to use", Key: "general purpose — applying something as a tool"},
			{Hanzi: "打", Pinyin: "dǎ", Meaning: "to hit; to make (a call); to play", Key: "wildcard verb — meaning depends entirely on context"},
		},
	},
	{
		Category: "semantic",
		Tip:      "对 = correct/toward/pair; 正 = straight/upright/just; 确 = certain/accurate. 对 is about matching (right answer), 正 about alignment (proper), 确 about precision (confirmed).",
		Chars: []Confused{
			{Hanzi: "对", Pinyin: "duì", Meaning: "correct, toward, pair", Key: "matching/corresponding — 'that is right'"},
			{Hanzi: "正", Pinyin: "zhèng", Meaning: "straight, upright, just", Key: "proper alignment — 'in the right position'"},
			{Hanzi: "确", Pinyin: "què", Meaning: "certain, accurate", Key: "verified precision — 'confirmed correct'"},
		},
	},
	{
		Category: "semantic",
		Tip:      "因 = because/cause (reason behind something); 为 = for/because of (purpose or cause); 由 = from/due to (origin or source). 因为 is the common compound; 由于 is more formal.",
		Chars: []Confused{
			{Hanzi: "因", Pinyin: "yīn", Meaning: "because, cause", Key: "the reason — 'because of this reason'"},
			{Hanzi: "为", Pinyin: "wèi", Meaning: "for, because of", Key: "purpose or motivation — 'for this purpose'"},
			{Hanzi: "由", Pinyin: "yóu", Meaning: "from, due to, by", Key: "origin or source — 'stemming from this'"},
		},
	},
	{
		Category: "semantic",
		Tip:      "带 = to bring/carry (带 also means belt — carrying on your person); 拿 = to hold/take (in hand); 送 = to deliver/give (taking it to someone else).",
		Chars: []Confused{
			{Hanzi: "带", Pinyin: "dài", Meaning: "to bring, to carry; belt", Key: "carrying along with you — 带来 (bring here)"},
			{Hanzi: "拿", Pinyin: "ná", Meaning: "to hold, to take", Key: "holding in hand — 拿起来 (pick up)"},
			{Hanzi: "送", Pinyin: "sòng", Meaning: "to deliver, to give, to see off", Key: "transferring to someone — 送给 (give to)"},
		},
	},
	{
		Category: "semantic",
		Tip:      "到 = to arrive/to (destination reached); 去 = to go (away from here); 来 = to come (toward here). Direction matters: 来 is toward speaker, 去 is away, 到 emphasizes arrival.",
		Chars: []Confused{
			{Hanzi: "到", Pinyin: "dào", Meaning: "to arrive, to reach", Key: "destination reached — 到了 (arrived)"},
			{Hanzi: "去", Pinyin: "qù", Meaning: "to go (away)", Key: "movement away from speaker"},
			{Hanzi: "来", Pinyin: "lái", Meaning: "to come (toward)", Key: "movement toward speaker"},
		},
	},
	{
		Category: "semantic",
		Tip:      "会 = can (learned skill, also 'will' for future); 能 = can (physical ability or permission); 可以 = may (permission or possibility). 会 = I learned how; 能 = I am able to; 可以 = I am allowed to.",
		Chars: []Confused{
			{Hanzi: "会", Pinyin: "huì", Meaning: "can (learned); will", Key: "acquired ability — 'I learned to swim'"},
			{Hanzi: "能", Pinyin: "néng", Meaning: "can (able); may", Key: "inherent ability — 'I am physically able'"},
			{Hanzi: "可", Pinyin: "kě", Meaning: "may, can (permission)", Key: "permission/possibility — 'you may do this'"},
		},
	},

	// ----------------------------------------------------------------
	// MORE VISUAL PAIRS
	// ----------------------------------------------------------------
	{
		Category: "visual",
		Tip:      "问 has 口 (mouth) inside 门 (door) — asking at the door; 间 has 日 (sun) inside 门 — sunlight through a door crack (space/gap).",
		Chars: []Confused{
			{Hanzi: "问", Pinyin: "wèn", Meaning: "to ask", Key: "口 (mouth) inside 门 — speaking at the door"},
			{Hanzi: "间", Pinyin: "jiān", Meaning: "between, space, room", Key: "日 (sun) inside 门 — light through the gap"},
		},
	},
	{
		Category: "visual",
		Tip:      "万 has no dot (ten thousand); 方 has a dot on the upper left (direction/square). Think: 方 points the way (has a pointer/dot).",
		Chars: []Confused{
			{Hanzi: "万", Pinyin: "wàn", Meaning: "ten thousand", Key: "no dot — clean top"},
			{Hanzi: "方", Pinyin: "fāng", Meaning: "direction, square, method", Key: "dot/piě at upper left — pointing the way"},
		},
	},
	{
		Category: "visual",
		Tip:      "厂 is an open cliff/factory (tone 3); 广 has a dot on top, representing a shelter/roof (tone 3 or 4). The dot is the roof.",
		Chars: []Confused{
			{Hanzi: "厂", Pinyin: "chǎng", Meaning: "factory, cliff", Key: "no dot — open cliff face"},
			{Hanzi: "广", Pinyin: "guǎng", Meaning: "wide, broad; shelter radical", Key: "dot on top — a covered shelter"},
		},
	},
	{
		Category: "visual",
		Tip:      "失 has a short top; 矢 has the top stroke extending more prominently. 失 = to lose (losing something small); 矢 = arrow (the top is the arrowhead).",
		Chars: []Confused{
			{Hanzi: "失", Pinyin: "shī", Meaning: "to lose, to miss", Key: "top stroke is short — something slipping away"},
			{Hanzi: "矢", Pinyin: "shǐ", Meaning: "arrow", Key: "top stroke is prominent — the arrowhead"},
		},
	},
	{
		Category: "visual",
		Tip:      "左 has 工 (work) on the bottom — the left hand works; 右 has 口 (mouth) on the bottom — the right hand eats.",
		Chars: []Confused{
			{Hanzi: "左", Pinyin: "zuǒ", Meaning: "left", Key: "bottom component is 工 (work)"},
			{Hanzi: "右", Pinyin: "yòu", Meaning: "right", Key: "bottom component is 口 (mouth)"},
		},
	},
}
