package mnemonic

// Decomposition breaks a character into its components with a mnemonic story.
type Decomposition struct {
	Character  string      // the full character: 好
	Components []Component // its parts
	Story      string      // mnemonic story connecting parts to meaning
}

// Component describes one radical or sub-part of a character.
type Component struct {
	Part     string // 女, 子
	Meaning  string // "woman", "child"
	Position string // "left", "right", "top", "bottom", "enclosure", "inside"
}

// Lookup returns the decomposition for a character, if known.
func Lookup(char string) (Decomposition, bool) {
	d, ok := decompositions[char]
	return d, ok
}

// All returns every known decomposition keyed by character.
func All() map[string]Decomposition {
	return decompositions
}

// decompositions maps single characters to their component breakdowns and
// mnemonic stories. Entries are ordered roughly by frequency / HSK level.
var decompositions = map[string]Decomposition{

	// ── Semantic combinations ────────────────────────────────────────

	"好": {
		Character: "好",
		Components: []Component{
			{Part: "女", Meaning: "woman", Position: "left"},
			{Part: "子", Meaning: "child", Position: "right"},
		},
		Story: "A woman holding her child — that is the definition of GOOD.",
	},
	"明": {
		Character: "明",
		Components: []Component{
			{Part: "日", Meaning: "sun", Position: "left"},
			{Part: "月", Meaning: "moon", Position: "right"},
		},
		Story: "The sun and moon shining together — everything is BRIGHT and clear.",
	},
	"休": {
		Character: "休",
		Components: []Component{
			{Part: "亻", Meaning: "person", Position: "left"},
			{Part: "木", Meaning: "tree", Position: "right"},
		},
		Story: "A tired person leans against a tree and finally RESTS.",
	},
	"林": {
		Character: "林",
		Components: []Component{
			{Part: "木", Meaning: "tree", Position: "left"},
			{Part: "木", Meaning: "tree", Position: "right"},
		},
		Story: "Two trees side by side — a FOREST is forming.",
	},
	"森": {
		Character: "森",
		Components: []Component{
			{Part: "木", Meaning: "tree", Position: "top"},
			{Part: "林", Meaning: "forest", Position: "bottom"},
		},
		Story: "A tree on top of a forest — trees as far as you can see, DENSE FOREST.",
	},
	"从": {
		Character: "从",
		Components: []Component{
			{Part: "人", Meaning: "person", Position: "left"},
			{Part: "人", Meaning: "person", Position: "right"},
		},
		Story: "One person walking behind another — to FOLLOW.",
	},
	"众": {
		Character: "众",
		Components: []Component{
			{Part: "人", Meaning: "person", Position: "top"},
			{Part: "从", Meaning: "follow", Position: "bottom"},
		},
		Story: "One person standing on two others — a whole CROWD of people.",
	},
	"男": {
		Character: "男",
		Components: []Component{
			{Part: "田", Meaning: "field", Position: "top"},
			{Part: "力", Meaning: "strength", Position: "bottom"},
		},
		Story: "Strength working the field — that was a MAN's role in ancient China.",
	},
	"安": {
		Character: "安",
		Components: []Component{
			{Part: "宀", Meaning: "roof", Position: "top"},
			{Part: "女", Meaning: "woman", Position: "bottom"},
		},
		Story: "A woman safe under a roof — the feeling of PEACE and safety.",
	},
	"家": {
		Character: "家",
		Components: []Component{
			{Part: "宀", Meaning: "roof", Position: "top"},
			{Part: "豕", Meaning: "pig", Position: "bottom"},
		},
		Story: "A pig under the roof — ancient families kept pigs indoors; where the pig lives is HOME.",
	},
	"灾": {
		Character: "灾",
		Components: []Component{
			{Part: "宀", Meaning: "roof", Position: "top"},
			{Part: "火", Meaning: "fire", Position: "bottom"},
		},
		Story: "Fire under the roof — your house is burning, a DISASTER.",
	},
	"尖": {
		Character: "尖",
		Components: []Component{
			{Part: "小", Meaning: "small", Position: "top"},
			{Part: "大", Meaning: "big", Position: "bottom"},
		},
		Story: "Small on top tapering from big on the bottom — something POINTY, like a pyramid.",
	},
	"看": {
		Character: "看",
		Components: []Component{
			{Part: "手", Meaning: "hand", Position: "top"},
			{Part: "目", Meaning: "eye", Position: "bottom"},
		},
		Story: "A hand shading the eyes to peer into the distance — to LOOK / SEE.",
	},
	"拿": {
		Character: "拿",
		Components: []Component{
			{Part: "合", Meaning: "together", Position: "top"},
			{Part: "手", Meaning: "hand", Position: "bottom"},
		},
		Story: "Hands coming together around an object — to TAKE / HOLD.",
	},
	"打": {
		Character: "打",
		Components: []Component{
			{Part: "扌", Meaning: "hand", Position: "left"},
			{Part: "丁", Meaning: "nail", Position: "right"},
		},
		Story: "A hand hammering a nail — to HIT / STRIKE.",
	},
	"闻": {
		Character: "闻",
		Components: []Component{
			{Part: "门", Meaning: "gate", Position: "enclosure"},
			{Part: "耳", Meaning: "ear", Position: "inside"},
		},
		Story: "An ear pressed against the gate, eavesdropping — to HEAR (also: to SMELL news).",
	},
	"问": {
		Character: "问",
		Components: []Component{
			{Part: "门", Meaning: "gate", Position: "enclosure"},
			{Part: "口", Meaning: "mouth", Position: "inside"},
		},
		Story: "A mouth calling through the gate — to ASK a question.",
	},
	"间": {
		Character: "间",
		Components: []Component{
			{Part: "门", Meaning: "gate", Position: "enclosure"},
			{Part: "日", Meaning: "sun", Position: "inside"},
		},
		Story: "Sunlight slicing through the crack in the gate — the SPACE BETWEEN.",
	},
	"闪": {
		Character: "闪",
		Components: []Component{
			{Part: "门", Meaning: "gate", Position: "enclosure"},
			{Part: "人", Meaning: "person", Position: "inside"},
		},
		Story: "A person darting through the gate — to FLASH past, to dodge.",
	},

	// ── Phonetic-hint characters ─────────────────────────────────────

	"清": {
		Character: "清",
		Components: []Component{
			{Part: "氵", Meaning: "water", Position: "left"},
			{Part: "青", Meaning: "blue-green / qing", Position: "right"},
		},
		Story: "Water so pure you can see the blue-green bottom — CLEAR (qing). The 青 gives the sound.",
	},
	"请": {
		Character: "请",
		Components: []Component{
			{Part: "讠", Meaning: "speech", Position: "left"},
			{Part: "青", Meaning: "blue-green / qing", Position: "right"},
		},
		Story: "Polite speech — PLEASE (qing). Same 青 sound component, different radical.",
	},
	"情": {
		Character: "情",
		Components: []Component{
			{Part: "忄", Meaning: "heart", Position: "left"},
			{Part: "青", Meaning: "blue-green / qing", Position: "right"},
		},
		Story: "What the heart feels — EMOTION, feelings (qing). The heart radical tells you it is about feelings.",
	},
	"晴": {
		Character: "晴",
		Components: []Component{
			{Part: "日", Meaning: "sun", Position: "left"},
			{Part: "青", Meaning: "blue-green / qing", Position: "right"},
		},
		Story: "The sun in a blue-green sky — CLEAR WEATHER, sunny (qing).",
	},
	"睛": {
		Character: "睛",
		Components: []Component{
			{Part: "目", Meaning: "eye", Position: "left"},
			{Part: "青", Meaning: "blue-green / qing", Position: "right"},
		},
		Story: "The eye's pupil, bright and clear — EYEBALL (jing). Used in 眼睛.",
	},
	"妈": {
		Character: "妈",
		Components: []Component{
			{Part: "女", Meaning: "woman", Position: "left"},
			{Part: "马", Meaning: "horse / ma", Position: "right"},
		},
		Story: "The woman who is your MAMA — 马 gives the 'ma' sound.",
	},
	"吗": {
		Character: "吗",
		Components: []Component{
			{Part: "口", Meaning: "mouth", Position: "left"},
			{Part: "马", Meaning: "horse / ma", Position: "right"},
		},
		Story: "A mouth sound at the end of a sentence — the question particle MA. Horse gives the sound.",
	},
	"骂": {
		Character: "骂",
		Components: []Component{
			{Part: "马", Meaning: "horse / ma", Position: "top"},
			{Part: "马", Meaning: "horse / ma", Position: "top"},
			{Part: "口", Meaning: "mouth", Position: "bottom"},
		},
		Story: "Two horses above a mouth — imagine two horses screaming at each other: to SCOLD (ma).",
	},

	// ── Body / nature / common objects ────────────────────────────────

	"想": {
		Character: "想",
		Components: []Component{
			{Part: "相", Meaning: "mutual / appearance", Position: "top"},
			{Part: "心", Meaning: "heart", Position: "bottom"},
		},
		Story: "An image (相) held in the heart (心) — to THINK, to miss someone.",
	},
	"忘": {
		Character: "忘",
		Components: []Component{
			{Part: "亡", Meaning: "dead / gone", Position: "top"},
			{Part: "心", Meaning: "heart", Position: "bottom"},
		},
		Story: "Something has died in your heart — you FORGOT it.",
	},
	"忙": {
		Character: "忙",
		Components: []Component{
			{Part: "忄", Meaning: "heart", Position: "left"},
			{Part: "亡", Meaning: "dead / gone", Position: "right"},
		},
		Story: "Your heart is dying from overwork — you are BUSY.",
	},
	"怕": {
		Character: "怕",
		Components: []Component{
			{Part: "忄", Meaning: "heart", Position: "left"},
			{Part: "白", Meaning: "white", Position: "right"},
		},
		Story: "Your heart turns white — you are AFRAID. Face goes pale with fear.",
	},
	"快": {
		Character: "快",
		Components: []Component{
			{Part: "忄", Meaning: "heart", Position: "left"},
			{Part: "夬", Meaning: "decisive", Position: "right"},
		},
		Story: "A decisive heart — acts FAST, feels HAPPY. Quick and cheerful.",
	},
	"慢": {
		Character: "慢",
		Components: []Component{
			{Part: "忄", Meaning: "heart", Position: "left"},
			{Part: "曼", Meaning: "prolonged", Position: "right"},
		},
		Story: "A heart that takes its time, prolonged — SLOW.",
	},
	"吃": {
		Character: "吃",
		Components: []Component{
			{Part: "口", Meaning: "mouth", Position: "left"},
			{Part: "乞", Meaning: "beg", Position: "right"},
		},
		Story: "A mouth begging for food — to EAT.",
	},
	"喝": {
		Character: "喝",
		Components: []Component{
			{Part: "口", Meaning: "mouth", Position: "left"},
			{Part: "曷", Meaning: "why / he", Position: "right"},
		},
		Story: "Your mouth is so dry it asks 'why?' — to DRINK (he).",
	},
	"说": {
		Character: "说",
		Components: []Component{
			{Part: "讠", Meaning: "speech", Position: "left"},
			{Part: "兑", Meaning: "exchange", Position: "right"},
		},
		Story: "Exchanging speech — to SAY, to speak.",
	},
	"话": {
		Character: "话",
		Components: []Component{
			{Part: "讠", Meaning: "speech", Position: "left"},
			{Part: "舌", Meaning: "tongue", Position: "right"},
		},
		Story: "What the tongue speaks — WORDS, speech, conversation.",
	},
	"谢": {
		Character: "谢",
		Components: []Component{
			{Part: "讠", Meaning: "speech", Position: "left"},
			{Part: "身", Meaning: "body", Position: "right"},
			{Part: "寸", Meaning: "inch", Position: "right"},
		},
		Story: "Words from the body, measured carefully — to THANK someone sincerely.",
	},
	"读": {
		Character: "读",
		Components: []Component{
			{Part: "讠", Meaning: "speech", Position: "left"},
			{Part: "卖", Meaning: "sell", Position: "right"},
		},
		Story: "Speech you sell to an audience — to READ aloud.",
	},
	"花": {
		Character: "花",
		Components: []Component{
			{Part: "艹", Meaning: "grass/plant", Position: "top"},
			{Part: "化", Meaning: "change / hua", Position: "bottom"},
		},
		Story: "A plant that transforms into something beautiful — a FLOWER (hua).",
	},
	"草": {
		Character: "草",
		Components: []Component{
			{Part: "艹", Meaning: "grass/plant", Position: "top"},
			{Part: "早", Meaning: "early", Position: "bottom"},
		},
		Story: "The first plant that sprouts early in spring — GRASS.",
	},
	"茶": {
		Character: "茶",
		Components: []Component{
			{Part: "艹", Meaning: "grass/plant", Position: "top"},
			{Part: "人", Meaning: "person", Position: "middle"},
			{Part: "木", Meaning: "tree", Position: "bottom"},
		},
		Story: "A plant that a person picks from a tree — TEA.",
	},
	"药": {
		Character: "药",
		Components: []Component{
			{Part: "艹", Meaning: "grass/plant", Position: "top"},
			{Part: "约", Meaning: "approximately / yue", Position: "bottom"},
		},
		Story: "A healing plant — MEDICINE. Ancient medicine was all herbs.",
	},
	"菜": {
		Character: "菜",
		Components: []Component{
			{Part: "艹", Meaning: "grass/plant", Position: "top"},
			{Part: "采", Meaning: "pick / gather", Position: "bottom"},
		},
		Story: "Plants you pick and gather — VEGETABLES, dishes.",
	},
	"字": {
		Character: "字",
		Components: []Component{
			{Part: "宀", Meaning: "roof", Position: "top"},
			{Part: "子", Meaning: "child", Position: "bottom"},
		},
		Story: "A child learning under a roof — the school where you learn CHARACTERS.",
	},
	"学": {
		Character: "学",
		Components: []Component{
			{Part: "⺍", Meaning: "hands holding", Position: "top"},
			{Part: "冖", Meaning: "cover/roof", Position: "middle"},
			{Part: "子", Meaning: "child", Position: "bottom"},
		},
		Story: "Hands guiding a child under a roof — to STUDY, to learn.",
	},
	"教": {
		Character: "教",
		Components: []Component{
			{Part: "孝", Meaning: "filial piety", Position: "left"},
			{Part: "攵", Meaning: "strike/tap", Position: "right"},
		},
		Story: "A hand with a rod guiding through filial piety — to TEACH.",
	},

	// ── Water radical family ─────────────────────────────────────────

	"河": {
		Character: "河",
		Components: []Component{
			{Part: "氵", Meaning: "water", Position: "left"},
			{Part: "可", Meaning: "can / he", Position: "right"},
		},
		Story: "A body of water you can navigate — a RIVER (he).",
	},
	"海": {
		Character: "海",
		Components: []Component{
			{Part: "氵", Meaning: "water", Position: "left"},
			{Part: "每", Meaning: "every / mei", Position: "right"},
		},
		Story: "Water everywhere, in every direction — the SEA (hai).",
	},
	"湖": {
		Character: "湖",
		Components: []Component{
			{Part: "氵", Meaning: "water", Position: "left"},
			{Part: "胡", Meaning: "reckless / hu", Position: "right"},
		},
		Story: "A body of water just sitting there — a LAKE (hu). 胡 gives the sound.",
	},
	"洗": {
		Character: "洗",
		Components: []Component{
			{Part: "氵", Meaning: "water", Position: "left"},
			{Part: "先", Meaning: "first / xian", Position: "right"},
		},
		Story: "Water first! — you WASH (xi) before doing anything else.",
	},
	"流": {
		Character: "流",
		Components: []Component{
			{Part: "氵", Meaning: "water", Position: "left"},
			{Part: "㐬", Meaning: "flow", Position: "right"},
		},
		Story: "Water that moves without stopping — to FLOW (liu).",
	},
	"汉": {
		Character: "汉",
		Components: []Component{
			{Part: "氵", Meaning: "water", Position: "left"},
			{Part: "又", Meaning: "again", Position: "right"},
		},
		Story: "The great river people returned to again and again — HAN (the Han River, the Han people).",
	},
	"没": {
		Character: "没",
		Components: []Component{
			{Part: "氵", Meaning: "water", Position: "left"},
			{Part: "殳", Meaning: "weapon/action", Position: "right"},
		},
		Story: "Something sinks into water and vanishes — NOT HAVE, did not (mei).",
	},
	"活": {
		Character: "活",
		Components: []Component{
			{Part: "氵", Meaning: "water", Position: "left"},
			{Part: "舌", Meaning: "tongue", Position: "right"},
		},
		Story: "Water on the tongue — as long as you can drink, you are ALIVE (huo).",
	},

	// ── Person radical family ────────────────────────────────────────

	"他": {
		Character: "他",
		Components: []Component{
			{Part: "亻", Meaning: "person", Position: "left"},
			{Part: "也", Meaning: "also / ye", Position: "right"},
		},
		Story: "That other person — HE / HIM (ta). Also a person, but not you.",
	},
	"她": {
		Character: "她",
		Components: []Component{
			{Part: "女", Meaning: "woman", Position: "left"},
			{Part: "也", Meaning: "also / ye", Position: "right"},
		},
		Story: "The female 'other person' — SHE / HER (ta). Same sound as 他, woman radical makes it feminine.",
	},
	"们": {
		Character: "们",
		Components: []Component{
			{Part: "亻", Meaning: "person", Position: "left"},
			{Part: "门", Meaning: "gate / men", Position: "right"},
		},
		Story: "People at the gate — the PLURAL marker (men). 门 gives the sound.",
	},
	"你": {
		Character: "你",
		Components: []Component{
			{Part: "亻", Meaning: "person", Position: "left"},
			{Part: "尔", Meaning: "you (classical)", Position: "right"},
		},
		Story: "A person who is the classical 'you' — YOU (ni).",
	},
	"住": {
		Character: "住",
		Components: []Component{
			{Part: "亻", Meaning: "person", Position: "left"},
			{Part: "主", Meaning: "master / zhu", Position: "right"},
		},
		Story: "A person who is the master of a place — to LIVE / RESIDE (zhu).",
	},
	"作": {
		Character: "作",
		Components: []Component{
			{Part: "亻", Meaning: "person", Position: "left"},
			{Part: "乍", Meaning: "suddenly / zha", Position: "right"},
		},
		Story: "A person who suddenly starts doing things — to DO / MAKE / WORK (zuo).",
	},
	"体": {
		Character: "体",
		Components: []Component{
			{Part: "亻", Meaning: "person", Position: "left"},
			{Part: "本", Meaning: "root / origin", Position: "right"},
		},
		Story: "The root of a person — the BODY (ti).",
	},
	"信": {
		Character: "信",
		Components: []Component{
			{Part: "亻", Meaning: "person", Position: "left"},
			{Part: "言", Meaning: "speech", Position: "right"},
		},
		Story: "A person standing by their word — TRUST, a LETTER, to BELIEVE (xin).",
	},

	// ── Fire / light ─────────────────────────────────────────────────

	"灯": {
		Character: "灯",
		Components: []Component{
			{Part: "火", Meaning: "fire", Position: "left"},
			{Part: "丁", Meaning: "nail / ding", Position: "right"},
		},
		Story: "Fire on a nail — a LAMP nailed to the wall (deng).",
	},
	"烧": {
		Character: "烧",
		Components: []Component{
			{Part: "火", Meaning: "fire", Position: "left"},
			{Part: "尧", Meaning: "high / yao", Position: "right"},
		},
		Story: "Fire rising high — to BURN (shao).",
	},

	// ── Earth radical family ─────────────────────────────────────────

	"地": {
		Character: "地",
		Components: []Component{
			{Part: "土", Meaning: "earth", Position: "left"},
			{Part: "也", Meaning: "also / ye", Position: "right"},
		},
		Story: "Earth is also everywhere — the GROUND, earth, land (di/de).",
	},
	"场": {
		Character: "场",
		Components: []Component{
			{Part: "土", Meaning: "earth", Position: "left"},
			{Part: "昜", Meaning: "expand / yang", Position: "right"},
		},
		Story: "Earth expanded into a flat space — a FIELD, a venue, a place (chang).",
	},
	"城": {
		Character: "城",
		Components: []Component{
			{Part: "土", Meaning: "earth", Position: "left"},
			{Part: "成", Meaning: "become / cheng", Position: "right"},
		},
		Story: "Earth built up to become something — a CITY wall, a city (cheng).",
	},

	// ── Metal radical family ─────────────────────────────────────────

	"钱": {
		Character: "钱",
		Components: []Component{
			{Part: "钅", Meaning: "metal", Position: "left"},
			{Part: "戋", Meaning: "small / jian", Position: "right"},
		},
		Story: "Small pieces of metal — MONEY (qian). Coins were small metal discs.",
	},
	"铁": {
		Character: "铁",
		Components: []Component{
			{Part: "钅", Meaning: "metal", Position: "left"},
			{Part: "失", Meaning: "lose / shi", Position: "right"},
		},
		Story: "Metal that never loses its strength — IRON (tie).",
	},

	// ── Hand radical family ──────────────────────────────────────────

	"找": {
		Character: "找",
		Components: []Component{
			{Part: "扌", Meaning: "hand", Position: "left"},
			{Part: "戈", Meaning: "spear / weapon", Position: "right"},
		},
		Story: "A hand reaching for a weapon — to SEARCH FOR, to look for (zhao).",
	},
	"把": {
		Character: "把",
		Components: []Component{
			{Part: "扌", Meaning: "hand", Position: "left"},
			{Part: "巴", Meaning: "cling / ba", Position: "right"},
		},
		Story: "A hand clinging to something — to GRASP, a handle (ba).",
	},
	"拉": {
		Character: "拉",
		Components: []Component{
			{Part: "扌", Meaning: "hand", Position: "left"},
			{Part: "立", Meaning: "stand / la", Position: "right"},
		},
		Story: "A hand pulling someone to stand — to PULL (la).",
	},
	"推": {
		Character: "推",
		Components: []Component{
			{Part: "扌", Meaning: "hand", Position: "left"},
			{Part: "隹", Meaning: "short-tailed bird", Position: "right"},
		},
		Story: "A hand shoving a stubborn bird — to PUSH (tui).",
	},

	// ── Roof/shelter family ──────────────────────────────────────────

	"宝": {
		Character: "宝",
		Components: []Component{
			{Part: "宀", Meaning: "roof", Position: "top"},
			{Part: "玉", Meaning: "jade", Position: "bottom"},
		},
		Story: "Jade kept safe under the roof — a TREASURE (bao).",
	},
	"完": {
		Character: "完",
		Components: []Component{
			{Part: "宀", Meaning: "roof", Position: "top"},
			{Part: "元", Meaning: "origin / yuan", Position: "bottom"},
		},
		Story: "Everything under the roof is in its original place — COMPLETE, finished (wan).",
	},
	"室": {
		Character: "室",
		Components: []Component{
			{Part: "宀", Meaning: "roof", Position: "top"},
			{Part: "至", Meaning: "arrive", Position: "bottom"},
		},
		Story: "Arriving under the roof — you have reached your ROOM (shi).",
	},
	"客": {
		Character: "客",
		Components: []Component{
			{Part: "宀", Meaning: "roof", Position: "top"},
			{Part: "各", Meaning: "each / ge", Position: "bottom"},
		},
		Story: "Each person under your roof — a GUEST, a customer (ke).",
	},
	"定": {
		Character: "定",
		Components: []Component{
			{Part: "宀", Meaning: "roof", Position: "top"},
			{Part: "正", Meaning: "correct / upright", Position: "bottom"},
		},
		Story: "Correct and settled under a roof — FIXED, decided, definite (ding).",
	},
	"实": {
		Character: "实",
		Components: []Component{
			{Part: "宀", Meaning: "roof", Position: "top"},
			{Part: "头", Meaning: "head", Position: "bottom"},
		},
		Story: "A solid thing under the roof — REAL, solid, truth (shi).",
	},

	// ── Eye/seeing family ────────────────────────────────────────────

	"眼": {
		Character: "眼",
		Components: []Component{
			{Part: "目", Meaning: "eye", Position: "left"},
			{Part: "艮", Meaning: "tough / gen", Position: "right"},
		},
		Story: "The tough part of the face that sees — the EYE (yan).",
	},
	"睡": {
		Character: "睡",
		Components: []Component{
			{Part: "目", Meaning: "eye", Position: "left"},
			{Part: "垂", Meaning: "droop / hang down", Position: "right"},
		},
		Story: "Eyes drooping shut — to SLEEP (shui).",
	},

	// ── Wood / tree family ───────────────────────────────────────────

	"树": {
		Character: "树",
		Components: []Component{
			{Part: "木", Meaning: "tree", Position: "left"},
			{Part: "对", Meaning: "correct / pair", Position: "right"},
		},
		Story: "Wood standing upright, correct and tall — a TREE (shu).",
	},
	"村": {
		Character: "村",
		Components: []Component{
			{Part: "木", Meaning: "tree", Position: "left"},
			{Part: "寸", Meaning: "inch / measure", Position: "right"},
		},
		Story: "Trees measured out for settlement — a VILLAGE (cun).",
	},
	"桥": {
		Character: "桥",
		Components: []Component{
			{Part: "木", Meaning: "tree/wood", Position: "left"},
			{Part: "乔", Meaning: "tall / qiao", Position: "right"},
		},
		Story: "Tall wood spanning a gap — a BRIDGE (qiao).",
	},
	"机": {
		Character: "机",
		Components: []Component{
			{Part: "木", Meaning: "tree/wood", Position: "left"},
			{Part: "几", Meaning: "small table / ji", Position: "right"},
		},
		Story: "A wooden mechanism — MACHINE, opportunity (ji). Originally a weaving loom.",
	},

	// ── Mouth family ─────────────────────────────────────────────────

	"叫": {
		Character: "叫",
		Components: []Component{
			{Part: "口", Meaning: "mouth", Position: "left"},
			{Part: "丩", Meaning: "twist / cry", Position: "right"},
		},
		Story: "A mouth twisted in a cry — to CALL OUT, to shout (jiao).",
	},
	"听": {
		Character: "听",
		Components: []Component{
			{Part: "口", Meaning: "mouth", Position: "left"},
			{Part: "斤", Meaning: "axe / jin", Position: "right"},
		},
		Story: "Mouth speaks, words hit like an axe — to LISTEN (ting).",
	},
	"回": {
		Character: "回",
		Components: []Component{
			{Part: "囗", Meaning: "enclosure", Position: "enclosure"},
			{Part: "口", Meaning: "mouth", Position: "inside"},
		},
		Story: "A mouth inside an enclosure, turning back on itself — to RETURN (hui).",
	},

	// ── Food / eating family ─────────────────────────────────────────

	"饭": {
		Character: "饭",
		Components: []Component{
			{Part: "饣", Meaning: "food", Position: "left"},
			{Part: "反", Meaning: "opposite / fan", Position: "right"},
		},
		Story: "The food you always come back to — cooked RICE, a meal (fan). 反 gives the sound.",
	},
	"饿": {
		Character: "饿",
		Components: []Component{
			{Part: "饣", Meaning: "food", Position: "left"},
			{Part: "我", Meaning: "I / me / wo", Position: "right"},
		},
		Story: "Me wanting food — HUNGRY (e). When I think about food, I must be starving.",
	},

	// ── Silk / thread family ─────────────────────────────────────────

	"红": {
		Character: "红",
		Components: []Component{
			{Part: "纟", Meaning: "silk/thread", Position: "left"},
			{Part: "工", Meaning: "work / gong", Position: "right"},
		},
		Story: "Thread that workers dye first — RED (hong), the most auspicious color.",
	},
	"绿": {
		Character: "绿",
		Components: []Component{
			{Part: "纟", Meaning: "silk/thread", Position: "left"},
			{Part: "录", Meaning: "record / lu", Position: "right"},
		},
		Story: "Thread the color of nature's record — GREEN (lv). 录 hints at the sound.",
	},
	"给": {
		Character: "给",
		Components: []Component{
			{Part: "纟", Meaning: "silk/thread", Position: "left"},
			{Part: "合", Meaning: "together / he", Position: "right"},
		},
		Story: "Thread that brings people together — to GIVE (gei).",
	},
	"经": {
		Character: "经",
		Components: []Component{
			{Part: "纟", Meaning: "silk/thread", Position: "left"},
			{Part: "圣", Meaning: "holy", Position: "right"},
		},
		Story: "The holy threads running through a loom — to PASS THROUGH, scripture, already (jing).",
	},

	// ── Movement / walking ───────────────────────────────────────────

	"道": {
		Character: "道",
		Components: []Component{
			{Part: "辶", Meaning: "walk/movement", Position: "enclosure"},
			{Part: "首", Meaning: "head / chief", Position: "inside"},
		},
		Story: "Where the head leads, the feet follow — the WAY, a road, the Dao (dao).",
	},
	"远": {
		Character: "远",
		Components: []Component{
			{Part: "辶", Meaning: "walk/movement", Position: "enclosure"},
			{Part: "元", Meaning: "origin / yuan", Position: "inside"},
		},
		Story: "Walking back to the origin — it is FAR away (yuan).",
	},
	"近": {
		Character: "近",
		Components: []Component{
			{Part: "辶", Meaning: "walk/movement", Position: "enclosure"},
			{Part: "斤", Meaning: "axe / jin", Position: "inside"},
		},
		Story: "Walking just an axe-throw away — NEAR, close (jin).",
	},
	"还": {
		Character: "还",
		Components: []Component{
			{Part: "辶", Meaning: "walk/movement", Position: "enclosure"},
			{Part: "不", Meaning: "not", Position: "inside"},
		},
		Story: "Not done walking yet — STILL, also, to return (hai/huan).",
	},
	"进": {
		Character: "进",
		Components: []Component{
			{Part: "辶", Meaning: "walk/movement", Position: "enclosure"},
			{Part: "井", Meaning: "well / jing", Position: "inside"},
		},
		Story: "Walking toward the well at the center of town — to ENTER, advance (jin).",
	},
	"过": {
		Character: "过",
		Components: []Component{
			{Part: "辶", Meaning: "walk/movement", Position: "enclosure"},
			{Part: "寸", Meaning: "inch", Position: "inside"},
		},
		Story: "Walking inch by inch — to PASS, to cross over (guo).",
	},
	"送": {
		Character: "送",
		Components: []Component{
			{Part: "辶", Meaning: "walk/movement", Position: "enclosure"},
			{Part: "关", Meaning: "close / barrier", Position: "inside"},
		},
		Story: "Walking someone past the barrier — to SEND OFF, deliver (song).",
	},

	// ── Strength / power ─────────────────────────────────────────────

	"动": {
		Character: "动",
		Components: []Component{
			{Part: "云", Meaning: "cloud", Position: "left"},
			{Part: "力", Meaning: "strength", Position: "right"},
		},
		Story: "Clouds move with great force — to MOVE, motion (dong).",
	},
	"功": {
		Character: "功",
		Components: []Component{
			{Part: "工", Meaning: "work", Position: "left"},
			{Part: "力", Meaning: "strength", Position: "right"},
		},
		Story: "Strength applied to work — MERIT, achievement, kung fu (gong).",
	},
	"加": {
		Character: "加",
		Components: []Component{
			{Part: "力", Meaning: "strength", Position: "left"},
			{Part: "口", Meaning: "mouth", Position: "right"},
		},
		Story: "Adding your voice to strength — to ADD (jia).",
	},
	"助": {
		Character: "助",
		Components: []Component{
			{Part: "且", Meaning: "moreover", Position: "left"},
			{Part: "力", Meaning: "strength", Position: "right"},
		},
		Story: "Moreover, I lend my strength — to HELP (zhu).",
	},
	"办": {
		Character: "办",
		Components: []Component{
			{Part: "力", Meaning: "strength", Position: "middle"},
			{Part: "丶", Meaning: "dot", Position: "left"},
			{Part: "丶", Meaning: "dot", Position: "right"},
		},
		Story: "Strength with effort marks on both sides — to MANAGE, to handle affairs (ban).",
	},

	// ── Sun / time family ────────────────────────────────────────────

	"时": {
		Character: "时",
		Components: []Component{
			{Part: "日", Meaning: "sun", Position: "left"},
			{Part: "寸", Meaning: "inch", Position: "right"},
		},
		Story: "The sun measured in inches on a sundial — TIME (shi).",
	},
	"早": {
		Character: "早",
		Components: []Component{
			{Part: "日", Meaning: "sun", Position: "top"},
			{Part: "十", Meaning: "ten", Position: "bottom"},
		},
		Story: "The sun has only risen ten degrees — it is EARLY (zao).",
	},
	"晚": {
		Character: "晚",
		Components: []Component{
			{Part: "日", Meaning: "sun", Position: "left"},
			{Part: "免", Meaning: "exempt / avoid", Position: "right"},
		},
		Story: "The sun has been exempted from the sky — EVENING, late (wan).",
	},

	// ── Rain / weather family ────────────────────────────────────────

	"雪": {
		Character: "雪",
		Components: []Component{
			{Part: "雨", Meaning: "rain", Position: "top"},
			{Part: "彐", Meaning: "hand / broom", Position: "bottom"},
		},
		Story: "Rain you can sweep with a broom — SNOW (xue).",
	},
	"雷": {
		Character: "雷",
		Components: []Component{
			{Part: "雨", Meaning: "rain", Position: "top"},
			{Part: "田", Meaning: "field", Position: "bottom"},
		},
		Story: "Rain crashing down onto the fields — THUNDER (lei).",
	},

	// ── Knife / cutting family ───────────────────────────────────────

	"到": {
		Character: "到",
		Components: []Component{
			{Part: "至", Meaning: "arrive", Position: "left"},
			{Part: "刂", Meaning: "knife", Position: "right"},
		},
		Story: "Arriving sharp as a knife — to ARRIVE, to reach (dao).",
	},
	"别": {
		Character: "别",
		Components: []Component{
			{Part: "另", Meaning: "another", Position: "left"},
			{Part: "刂", Meaning: "knife", Position: "right"},
		},
		Story: "A knife cutting one from another — DON'T, to separate, other (bie).",
	},
	"利": {
		Character: "利",
		Components: []Component{
			{Part: "禾", Meaning: "grain", Position: "left"},
			{Part: "刂", Meaning: "knife", Position: "right"},
		},
		Story: "A knife harvesting grain — PROFIT, sharp, beneficial (li).",
	},

	// ── Enclosure / box family ───────────────────────────────────────

	"国": {
		Character: "国",
		Components: []Component{
			{Part: "囗", Meaning: "enclosure/border", Position: "enclosure"},
			{Part: "玉", Meaning: "jade", Position: "inside"},
		},
		Story: "Jade protected within borders — a COUNTRY, a nation (guo). The most precious thing within walls.",
	},
	"园": {
		Character: "园",
		Components: []Component{
			{Part: "囗", Meaning: "enclosure/border", Position: "enclosure"},
			{Part: "元", Meaning: "origin / yuan", Position: "inside"},
		},
		Story: "An enclosed space at the origin of nature — a GARDEN, park (yuan).",
	},
	"图": {
		Character: "图",
		Components: []Component{
			{Part: "囗", Meaning: "enclosure/border", Position: "enclosure"},
			{Part: "冬", Meaning: "winter / dong", Position: "inside"},
		},
		Story: "Winter mapped out within a frame — a MAP, diagram, picture (tu).",
	},

	// ── Heart family (full form) ─────────────────────────────────────

	"思": {
		Character: "思",
		Components: []Component{
			{Part: "田", Meaning: "field", Position: "top"},
			{Part: "心", Meaning: "heart", Position: "bottom"},
		},
		Story: "The field of the heart — to THINK (si). Thoughts are cultivated like crops.",
	},
	"意": {
		Character: "意",
		Components: []Component{
			{Part: "音", Meaning: "sound", Position: "top"},
			{Part: "心", Meaning: "heart", Position: "bottom"},
		},
		Story: "The sound of the heart — MEANING, intention, idea (yi).",
	},
	"感": {
		Character: "感",
		Components: []Component{
			{Part: "咸", Meaning: "all / salty", Position: "top"},
			{Part: "心", Meaning: "heart", Position: "bottom"},
		},
		Story: "All the heart's responses — to FEEL, sensation, emotion (gan).",
	},
	"愿": {
		Character: "愿",
		Components: []Component{
			{Part: "原", Meaning: "origin/plain", Position: "top"},
			{Part: "心", Meaning: "heart", Position: "bottom"},
		},
		Story: "The original desire of the heart — to WISH, to be willing (yuan).",
	},

	// ── Miscellaneous high-frequency characters ──────────────────────

	"笔": {
		Character: "笔",
		Components: []Component{
			{Part: "⺮", Meaning: "bamboo", Position: "top"},
			{Part: "毛", Meaning: "fur/hair", Position: "bottom"},
		},
		Story: "Bamboo shaft with animal hair at the tip — a writing BRUSH / PEN (bi).",
	},
	"等": {
		Character: "等",
		Components: []Component{
			{Part: "⺮", Meaning: "bamboo", Position: "top"},
			{Part: "寺", Meaning: "temple / si", Position: "bottom"},
		},
		Story: "Sitting on a bamboo mat at the temple — to WAIT, rank, class (deng).",
	},
	"答": {
		Character: "答",
		Components: []Component{
			{Part: "⺮", Meaning: "bamboo", Position: "top"},
			{Part: "合", Meaning: "together / he", Position: "bottom"},
		},
		Story: "Bamboo slips brought together — an ANSWER, to reply (da).",
	},
	"简": {
		Character: "简",
		Components: []Component{
			{Part: "⺮", Meaning: "bamboo", Position: "top"},
			{Part: "间", Meaning: "between / jian", Position: "bottom"},
		},
		Story: "Bamboo strips with space between — SIMPLE, brief (jian). Simple messages on bamboo.",
	},
	"骑": {
		Character: "骑",
		Components: []Component{
			{Part: "马", Meaning: "horse", Position: "left"},
			{Part: "奇", Meaning: "strange / qi", Position: "right"},
		},
		Story: "Strange things happen on horseback — to RIDE (qi).",
	},
	"鲜": {
		Character: "鲜",
		Components: []Component{
			{Part: "鱼", Meaning: "fish", Position: "left"},
			{Part: "羊", Meaning: "sheep", Position: "right"},
		},
		Story: "Fish and mutton together — the FRESHEST, most delicious flavors (xian).",
	},
	"鸡": {
		Character: "鸡",
		Components: []Component{
			{Part: "又", Meaning: "hand/again", Position: "left"},
			{Part: "鸟", Meaning: "bird", Position: "right"},
		},
		Story: "A bird you can grab with your hand — a CHICKEN (ji). Domesticated and easy to catch.",
	},
	"破": {
		Character: "破",
		Components: []Component{
			{Part: "石", Meaning: "stone", Position: "left"},
			{Part: "皮", Meaning: "skin / pi", Position: "right"},
		},
		Story: "Stone tearing through skin — BROKEN, to break (po).",
	},
	"确": {
		Character: "确",
		Components: []Component{
			{Part: "石", Meaning: "stone", Position: "left"},
			{Part: "角", Meaning: "horn / corner", Position: "right"},
		},
		Story: "Solid as stone, sharp as a horn — CERTAIN, definite, exact (que).",
	},
	"岁": {
		Character: "岁",
		Components: []Component{
			{Part: "山", Meaning: "mountain", Position: "top"},
			{Part: "夕", Meaning: "evening", Position: "bottom"},
		},
		Story: "Mountains watch the evenings pass — counting YEARS of age (sui).",
	},
	"岛": {
		Character: "岛",
		Components: []Component{
			{Part: "鸟", Meaning: "bird", Position: "top"},
			{Part: "山", Meaning: "mountain", Position: "bottom"},
		},
		Story: "A mountain where only birds can reach — an ISLAND (dao).",
	},
}
