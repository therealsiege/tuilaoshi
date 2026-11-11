package journey

// BestiaryEntry describes a demon from Journey to the West.
type BestiaryEntry struct {
	Name     string // Chinese name
	English  string // English name
	Title    string // demon's title/epithet in Chinese
	TitleEng string // English title
	Origin   string // 1-sentence origin story
	Powers   string // what the demon can do
	Weakness string // how they were defeated
	Chapter  string // approximate chapter reference
	Art      string // 3-5 line ASCII art of the demon
}

// Bestiary returns all demon entries in the catalog.
func Bestiary() []BestiaryEntry { return bestiary }

// BestiaryLookup finds a demon by its Chinese name, or returns nil.
func BestiaryLookup(name string) *BestiaryEntry {
	for i := range bestiary {
		if bestiary[i].Name == name {
			return &bestiary[i]
		}
	}
	return nil
}

var bestiary = []BestiaryEntry{
	{
		Name:     "混世魔王",
		English:  "Havoc Monster",
		Title:    "混世魔王",
		TitleEng: "Demon King of Confusion",
		Origin:   "A lesser demon who seized the Water Curtain Cave while Sun Wukong was away training under Master Subhuti.",
		Powers:   "Commands a horde of minor demons and wields a massive iron blade",
		Weakness: "Sun Wukong reclaimed his cave and slew him with a single stroke of a stolen sword, his very first boss battle",
		Chapter:  "Chapter 2-3",
		Art: `    _/\_
   (o  o)
  --\  /--
    |\/|
    /  \`,
	},
	{
		Name:     "白骨精",
		English:  "White Bone Demon",
		Title:    "尸魔",
		TitleEng: "Corpse Demon",
		Origin:   "A skeleton demon who achieved sentience and shape-shifting powers after centuries of absorbing yin energy on White Tiger Ridge.",
		Powers:   "Transforms into three disguises — a young woman, an old woman, and an old man — to trick Tang Sanzang",
		Weakness: "Sun Wukong saw through all three disguises with his fiery golden eyes and struck her down, though Tang Sanzang banished him for it",
		Chapter:  "Chapters 27-28",
		Art: `    .-.
   (x x)
    |=|
   /|+|\
    d  b`,
	},
	{
		Name:     "黄风怪",
		English:  "Yellow Wind Demon",
		Title:    "黄风大王",
		TitleEng: "Great King of Yellow Wind",
		Origin:   "Originally a marten who lived beneath the lamp of Lingji Bodhisattva's temple and stole its lamp oil, then fled to Yellow Wind Ridge.",
		Powers:   "Commands devastating samadhi wind storms that blind and disorient opponents, even Sun Wukong",
		Weakness: "Lingji Bodhisattva subdued him with the Stabilizing Wind Pill and a golden dragon claw lance",
		Chapter:  "Chapters 20-21",
		Art: `   ~~//~~
  ( @.@ )
  /|~~~|\
   | W |
   ^   ^`,
	},
	{
		Name:     "黄袍怪",
		English:  "Yellow Robe Demon",
		Title:    "奎木狼",
		TitleEng: "Star Spirit Kui Mu Lang",
		Origin:   "The Wood Wolf star spirit of the Kui constellation who fell to earth to live with his celestial lover Baihua, the princess of Baoxiang Kingdom.",
		Powers:   "Transforms Tang Sanzang into a tiger with dark magic and possesses tremendous combat skill as a celestial warrior",
		Weakness: "Sun Wukong was recalled from exile, freed Sanzang, and fought him until the celestial authorities summoned the star spirit back to heaven",
		Chapter:  "Chapters 28-31",
		Art: `   /|  |\
  ( o--o )
   |/~~\|
   | ][ |
   _/  \_`,
	},
	{
		Name:     "金角大王",
		English:  "Gold Horn King",
		Title:    "金角大王",
		TitleEng: "Great King Gold Horn",
		Origin:   "Originally a boy attendant who tended Laozi's golden brazier, sent to earth by Guanyin to test the pilgrims.",
		Powers:   "Possesses five magical treasures including the Purple Gold Red Gourd, the Jade Vase, and the Plantain Fan",
		Weakness: "Sun Wukong outsmarted them through cunning and deception, swapping the treasures and trapping them in their own gourds",
		Chapter:  "Chapters 32-35",
		Art: `   ,/|\.
  / o o \
  | _L_ |
  |/|||\ |
   // \\`,
	},
	{
		Name:     "银角大王",
		English:  "Silver Horn King",
		Title:    "银角大王",
		TitleEng: "Great King Silver Horn",
		Origin:   "The other boy attendant from Laozi's palace, partner to Gold Horn, equally armed with stolen celestial treasures.",
		Powers:   "Wields the Seven Star Sword and commands the magical rope that binds immortals, works in tandem with his brother",
		Weakness: "Tricked by Sun Wukong's shape-shifting and trapped in the Purple Gold Red Gourd; Laozi reclaimed both demons",
		Chapter:  "Chapters 32-35",
		Art: `   ,/|\.
  / - - \
  | _v_ |
  |/|||\|
   // \\`,
	},
	{
		Name:     "红孩儿",
		English:  "Red Boy",
		Title:    "圣婴大王",
		TitleEng: "Holy Infant King",
		Origin:   "Son of the Bull Demon King and Princess Iron Fan, raised in the volcanic Fire Cloud Cave where he mastered samadhi true fire.",
		Powers:   "Breathes samadhi true fire from his mouth and nostrils, which cannot be extinguished by ordinary water",
		Weakness: "Guanyin trapped him with celestial blades on a lotus throne and made him her disciple, the Boy of Wealth",
		Chapter:  "Chapters 40-42",
		Art: `   *\~/*
   (>.< )
  ~|/^\|~
   | V |
   ^   ^`,
	},
	{
		Name:     "黑熊精",
		English:  "Black Bear Demon",
		Title:    "黑风山主",
		TitleEng: "Lord of Black Wind Mountain",
		Origin:   "A black bear who cultivated the Dao for centuries on Black Wind Mountain near the Guanyin Temple.",
		Powers:   "Immense physical strength, skilled in martial combat, and cunning enough to befriend corrupt monks",
		Weakness: "Guanyin disguised herself as a Daoist offering a pill of immortality, which was actually a mind-controlling pellet that subdued him",
		Chapter:  "Chapters 16-17",
		Art: `   (n n)
  [| . |]
  [|\_/|]
  [| | |]
   _/ \_`,
	},
	{
		Name:     "青牛精",
		English:  "Green Bull Demon",
		Title:    "独角兕大王",
		TitleEng: "Great King Single-Horn Rhinoceros",
		Origin:   "Laozi's own blue-green bull mount that escaped the celestial stables and descended to earth as a demon king.",
		Powers:   "Possesses the Diamond Jade Bracelet (金刚琢) that absorbs and neutralizes all weapons, including Sun Wukong's staff",
		Weakness: "Only Laozi himself could reclaim the bracelet and subdue his own mount, returning it to the celestial realm",
		Chapter:  "Chapters 50-52",
		Art: `  (__/)__
  (o   o )>
  /| U |/
 / |===| \
    || ||`,
	},
	{
		Name:     "蜘蛛精",
		English:  "Spider Demons",
		Title:    "七情迷",
		TitleEng: "Seven Passions",
		Origin:   "Seven spider sisters who established themselves at Silk Cave (盘丝洞), each embodying one of the seven emotions.",
		Powers:   "Weave poisonous silk webs from their navels, transform into beautiful women, and ensnare victims with illusion and desire",
		Weakness: "Sun Wukong destroyed their webs while Zhu Bajie fought them at the bathhouse; their senior, the Hundred-Eye Demon, was subdued by Vairocana Buddha's mother",
		Chapter:  "Chapters 72-73",
		Art: `  /\()/\
 //    \\
((  ><  ))
 \\    //
  \/()\/`,
	},
	{
		Name:     "蝎子精",
		English:  "Scorpion Demon",
		Title:    "琵琶精",
		TitleEng: "Pipa Spirit",
		Origin:   "A scorpion spirit who once stung the Buddha himself when he was lecturing at the Thunder Monastery, then fled to Pipa Cave.",
		Powers:   "Her tail stinger is lethally venomous, powerful enough to wound even Buddha; she can transform into a beautiful woman",
		Weakness: "Defeated by the Star Lord of the Pleiades (Mao Ri Xing), a celestial rooster whose crowing is her bane",
		Chapter:  "Chapter 55-56",
		Art: `    )>
   /o o\
  | === |
  \\ ~ /
   \\|/=>`,
	},
	{
		Name:     "六耳猕猴",
		English:  "Six-Eared Macaque",
		Title:    "假悟空",
		TitleEng: "False Wukong",
		Origin:   "One of the four primates of chaos, a creature born from heaven and earth who can hear and know all things across the universe.",
		Powers:   "A perfect copy of Sun Wukong in appearance, power, and even his staff, indistinguishable by any spell or mirror",
		Weakness: "Only Buddha could tell the two apart; he revealed the impostor under his golden bowl and Sun Wukong struck him down",
		Chapter:  "Chapters 56-58",
		Art: `   ,.-.
  ( o.o)
   |=|?
  /|M|\
   d  b`,
	},
	{
		Name:     "牛魔王",
		English:  "Bull Demon King",
		Title:    "大力牛魔王",
		TitleEng: "Great Strength Bull Demon King",
		Origin:   "Once a sworn brother of Sun Wukong from their demon days, now the most powerful independent demon king, lord of Cuiyun Mountain.",
		Powers:   "Can transform into any shape, possesses enormous strength, rides iron chariots, and rivals Sun Wukong in raw combat power",
		Weakness: "Defeated by the combined forces of Sun Wukong, Zhu Bajie, Nezha, and the Heavenly King Li Jing's demon-revealing mirror",
		Chapter:  "Chapters 59-61",
		Art: `  (\\||//)
   (o  o)
    |UU|
   /|  |\
  /_|  |_\`,
	},
	{
		Name:     "铁扇公主",
		English:  "Princess Iron Fan",
		Title:    "罗刹女",
		TitleEng: "Rakshasi",
		Origin:   "Wife of the Bull Demon King and mother of Red Boy, she guards the palm-leaf fan at Palm Leaf Cave near the Flaming Mountains.",
		Powers:   "Wields the Plantain Fan (芭蕉扇) which can extinguish fire or create hurricanes with a single wave",
		Weakness: "Sun Wukong shrank himself and entered her stomach, causing unbearable pain until she surrendered the fan",
		Chapter:  "Chapters 59-61",
		Art: `    ___
   / ~ \
  ( -.- )
   \\_//
    )|(`,
	},
	{
		Name:     "白鹿精",
		English:  "White Deer Demon",
		Title:    "比丘国国丈",
		TitleEng: "State Preceptor of Bhikku Kingdom",
		Origin:   "The mount of the Star of Longevity (寿星) who escaped and posed as a Daoist sage in the kingdom of Bhikku.",
		Powers:   "Convinced the king to harvest the hearts of 1,111 children to brew an immortality elixir",
		Weakness: "Sun Wukong saved the children and chased the demon until the Star of Longevity arrived to reclaim his mount",
		Chapter:  "Chapters 78-79",
		Art: `   /\\  /\\
  (  oo  )
   | -- |
   | || |
    \\||/`,
	},
	{
		Name:     "玉兔精",
		English:  "Jade Rabbit Demon",
		Title:    "天竺假公主",
		TitleEng: "False Princess of Tianzhu",
		Origin:   "Chang'e's jade rabbit who descended to earth to settle a karmic debt and impersonated the princess of the Tianzhu Kingdom.",
		Powers:   "Skilled in martial arts using a pestle weapon, shape-shifts into a perfect copy of the real princess",
		Weakness: "Chang'e herself descended from the moon to reclaim the rabbit after Sun Wukong exposed the deception",
		Chapter:  "Chapters 93-95",
		Art: `   (\\()
   (o.o )
    (\")(")
     || ||
     *  *`,
	},
	{
		Name:     "金翅大鹏",
		English:  "Golden-Winged Great Roc",
		Title:    "大鹏金翅",
		TitleEng: "Great Peng of Golden Wings",
		Origin:   "Born from a cosmic phoenix egg, he is the nephew of the Buddha and the fastest creature in all of creation.",
		Powers:   "Flies at inconceivable speed, possesses strength that overwhelms even Sun Wukong, and commands the demons of Lion Camel Ridge",
		Weakness: "Only Buddha himself could subdue him; the Roc was made to serve as a dharma protector atop Buddha's throne",
		Chapter:  "Chapters 74-77",
		Art: `   __/V\__
  / /   \ \
 / / o o \ \
 \=\  ^  /=/
   \\===//`,
	},
	{
		Name:     "金鱼精",
		English:  "Goldfish Demon",
		Title:    "灵感大王",
		TitleEng: "Great King of Spiritual Touch",
		Origin:   "A goldfish that lived in the lotus pond of Guanyin's temple and achieved demon form by absorbing the light of moonlit scriptures.",
		Powers:   "Controls the waters of the Heaven-Reaching River, creates floods, and demands child sacrifices",
		Weakness: "Guanyin arrived with a bamboo basket and recaptured her escaped fish with a simple scoop",
		Chapter:  "Chapters 47-49",
		Art: `    ,<><,
   / o  o\
  (  ><  )
   \~~~~/>
    `+"``"+`--'`,
	},
	{
		Name:     "黄眉怪",
		English:  "Yellow Brow Demon",
		Title:    "黄眉大王",
		TitleEng: "Great King Yellow Brow",
		Origin:   "A page boy who served under Maitreya Buddha and stole his master's golden cymbals and human-skin bag before fleeing to earth.",
		Powers:   "Traps enemies in the human-skin bag and crushes them between the golden cymbals; created a false Buddhist paradise",
		Weakness: "Maitreya Buddha arrived in disguise as a melon farmer and lured the demon back into captivity",
		Chapter:  "Chapters 65-66",
		Art: `    _O_
   /o o\\
  | === |
  |\\_//|
   \\|_|/`,
	},
	{
		Name:     "白毛鼠精",
		English:  "White-Haired Rat Demon",
		Title:    "半截观音",
		TitleEng: "Half-Length Guanyin",
		Origin:   "A rat spirit who once stole food from the Buddha's lamp and was spared execution by Li Jing, then posed as a kidnapped maiden in Pitfall Cave.",
		Powers:   "Transforms into a beautiful woman to seduce Tang Sanzang, and impersonates a half-bodied Guanyin shrine",
		Weakness: "Heavenly King Li Jing and Prince Nezha arrived to arrest their wayward adopted daughter and returned her to heaven",
		Chapter:  "Chapters 80-83",
		Art: `    ()()
   (o.o )
    |>|
   /| |\
    ~  ~`,
	},
	{
		Name:     "九头虫",
		English:  "Nine-Headed Bug",
		Title:    "九头驸马",
		TitleEng: "Nine-Headed Prince Consort",
		Origin:   "A nine-headed bird demon who married a dragon princess of Biyou Palace and stole the Buddha's relic from the Jinguang Temple.",
		Powers:   "Possesses nine heads that are nearly impossible to sever permanently, each capable of biting and spitting poison",
		Weakness: "Erlang Shen's celestial hound bit off one head; the injured demon fled and was never fully captured, one of the few to escape",
		Chapter:  "Chapters 62-63",
		Art: `  ooo/|\ooo
   \o-o-o/
    |===|
    /| |\
   / | | \`,
	},
	{
		Name:     "如来佛祖",
		English:  "Buddha",
		Title:    "如来佛祖",
		TitleEng: "Tathagata Buddha",
		Origin:   "The supreme Buddha who trapped Sun Wukong under Five Elements Mountain and ordained the Journey to the West as a test of faith.",
		Powers:   "Infinite wisdom and cosmic power; his palm can become a mountain, his perception encompasses all universes",
		Weakness: "He is not a demon to be defeated but the final judge; the pilgrims earn his scriptures through 81 tribulations of devotion",
		Chapter:  "Chapters 7, 98-100",
		Art: `     _
    /.\
   / . \
  ( -o- )
   \___/`,
	},
}
