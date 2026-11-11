package journey

// LoreEntry is an unlockable story passage from Journey to the West.
type LoreEntry struct {
	Tribulation int    // unlocks at this tribulation number
	Title       string // English title
	Chinese     string // Chinese title
	Passage     string // 2-3 sentences in Chinese
	Pinyin      string // pinyin for the passage
	English     string // English translation
	Unlocked    bool
}

// buildLore returns lore entries with unlock status based on current tribulation.
func buildLore(tribulation int) []LoreEntry {
	entries := loreData()
	for i := range entries {
		entries[i].Unlocked = tribulation >= entries[i].Tribulation
	}
	return entries
}

// loreData returns the raw lore entries without unlock status applied.
func loreData() []LoreEntry {
	return []LoreEntry{
		{
			Tribulation: 1,
			Title:       "Departure from Chang'an",
			Chinese:     "长安出发",
			Passage:     "唐僧接到了皇帝的命令，要去西方取回佛经。他告别了长安城，独自走上了漫长的道路。这是一段充满危险的旅程，但他的心里只有一个目标。",
			Pinyin:      "Táng Sēng jiēdào le huángdì de mìnglìng, yào qù xīfāng qǔ huí fójīng. Tā gàobié le Cháng'ān chéng, dúzì zǒu shàng le mànchǎng de dàolù. Zhè shì yī duàn chōngmǎn wéixiǎn de lǚchéng, dàn tā de xīn lǐ zhǐ yǒu yī gè mùbiāo.",
			English:     "Tang Sanzang received the emperor's command to travel west and retrieve the Buddhist sutras. He bid farewell to the city of Chang'an and set out alone on the long road. It was a journey full of danger, but in his heart there was only one goal.",
		},
		{
			Tribulation: 10,
			Title:       "Danger at Twin Fork Ridge",
			Chinese:     "双叉岭遇险",
			Passage:     "在双叉岭的山路上，一群妖怪突然出现了。唐僧非常害怕，因为他没有武器，也没有同伴。幸好有神仙暗中保护他，他才平安地度过了第一次危险。",
			Pinyin:      "Zài Shuāngchā Lǐng de shānlù shang, yī qún yāoguài tūrán chūxiàn le. Táng Sēng fēicháng hàipà, yīnwèi tā méiyǒu wǔqì, yě méiyǒu tóngbàn. Xìnghǎo yǒu shénxiān ànzhōng bǎohù tā, tā cái píng'ān de dùguò le dì yī cì wéixiǎn.",
			English:     "On the mountain road at Twin Fork Ridge, a group of demons suddenly appeared. Tang Sanzang was very afraid because he had no weapon and no companions. Fortunately, immortals were secretly protecting him, and he safely passed through his first danger.",
		},
		{
			Tribulation: 20,
			Title:       "Freeing the Monkey from Five Elements Mountain",
			Chinese:     "五行山救猴",
			Passage:     "唐僧来到五行山下，听到有人在喊他。那是孙悟空，他被佛祖压在这座山下已经五百年了。唐僧揭下了山上的符咒，孙悟空终于自由了，从此成为他的大徒弟。",
			Pinyin:      "Táng Sēng lái dào Wǔxíng Shān xià, tīng dào yǒu rén zài hǎn tā. Nà shì Sūn Wùkōng, tā bèi Fózǔ yā zài zhè zuò shān xià yǐjīng wǔbǎi nián le. Táng Sēng jiē xià le shān shang de fúzhòu, Sūn Wùkōng zhōngyú zìyóu le, cóngcǐ chéngwéi tā de dà túdì.",
			English:     "Tang Sanzang arrived at the foot of Five Elements Mountain and heard someone calling out to him. It was Sun Wukong, who had been trapped under this mountain by Buddha for five hundred years. Tang Sanzang removed the talisman from the mountain, and Sun Wukong was finally free, becoming his first disciple from that day on.",
		},
		{
			Tribulation: 30,
			Title:       "The Kasaya at Guanyin Temple",
			Chinese:     "观音院袈裟",
			Passage:     "唐僧在观音院借宿的时候，拿出了他珍贵的袈裟。寺里的老和尚看到这件美丽的袈裟，非常嫉妒，想要偷走它。这件事给唐僧带来了很大的麻烦。",
			Pinyin:      "Táng Sēng zài Guānyīn Yuàn jièsù de shíhou, ná chū le tā zhēnguì de jiāshā. Sì lǐ de lǎo héshang kàn dào zhè jiàn měilì de jiāshā, fēicháng jídù, xiǎng yào tōu zǒu tā. Zhè jiàn shì gěi Táng Sēng dài lái le hěn dà de máfan.",
			English:     "When Tang Sanzang stayed overnight at Guanyin Temple, he took out his precious kasaya robe. The old monk at the temple saw this beautiful kasaya and was very jealous, wanting to steal it. This incident brought Tang Sanzang a great deal of trouble.",
		},
		{
			Tribulation: 40,
			Title:       "Recruiting Zhu Bajie at Gao Village",
			Chinese:     "高老庄收猪",
			Passage:     "在高老庄，一个猪妖抢走了村民的女儿做妻子。孙悟空和猪妖打了一场激烈的战斗，最后把他打败了。这个猪妖就是猪八戒，他成了取经队伍的第二个徒弟。",
			Pinyin:      "Zài Gāolǎozhuāng, yī gè zhū yāo qiǎng zǒu le cūnmín de nǚ'ér zuò qīzi. Sūn Wùkōng hé zhū yāo dǎ le yī chǎng jīliè de zhàndòu, zuìhòu bǎ tā dǎ bài le. Zhège zhū yāo jiùshì Zhū Bājiè, tā chéng le qǔjīng duìwu de dì èr gè túdì.",
			English:     "At Gao Village, a pig demon had taken a villager's daughter as his wife. Sun Wukong fought a fierce battle with the pig demon and finally defeated him. This pig demon was Zhu Bajie, and he became the second disciple of the scripture-seeking party.",
		},
		{
			Tribulation: 50,
			Title:       "Sha Wujing at the Flowing Sands River",
			Chinese:     "流沙河收沙僧",
			Passage:     "取经队伍来到了流沙河，河水又深又急，没有人敢过去。河里住着一个水妖，他就是沙悟净。在观音菩萨的帮助下，沙悟净决定加入队伍，成为第三个徒弟。",
			Pinyin:      "Qǔjīng duìwu lái dào le Liúshā Hé, hé shuǐ yòu shēn yòu jí, méiyǒu rén gǎn guòqù. Hé lǐ zhù zhe yī gè shuǐ yāo, tā jiùshì Shā Wùjìng. Zài Guānyīn Púsà de bāngzhù xià, Shā Wùjìng juédìng jiārù duìwu, chéngwéi dì sān gè túdì.",
			English:     "The scripture-seeking party arrived at the Flowing Sands River. The water was deep and swift, and no one dared to cross. A water demon lived in the river -- he was Sha Wujing. With the help of Guanyin Bodhisattva, Sha Wujing decided to join the party and became the third disciple.",
		},
		{
			Tribulation: 60,
			Title:       "Borrowing the Fan at Flaming Mountains",
			Chinese:     "火焰山三借扇",
			Passage:     "前方的火焰山烧着熊熊大火，根本没办法通过。孙悟空去找铁扇公主借芭蕉扇，但是她不肯借给他。孙悟空用了很多办法，借了三次才终于拿到了扇子，扑灭了大火。",
			Pinyin:      "Qiánfāng de Huǒyàn Shān shāo zhe xióngxióng dàhuǒ, gēnběn méi bànfǎ tōngguò. Sūn Wùkōng qù zhǎo Tiěshàn Gōngzhǔ jiè bājiāoshàn, dànshì tā bù kěn jiè gěi tā. Sūn Wùkōng yòng le hěn duō bànfǎ, jiè le sān cì cái zhōngyú ná dào le shànzi, pūmiè le dàhuǒ.",
			English:     "The Flaming Mountains ahead burned with a raging fire, making it completely impossible to pass through. Sun Wukong went to ask Princess Iron Fan to lend him the Banana Leaf Fan, but she refused. Sun Wukong tried many methods, and only after borrowing three times did he finally obtain the fan and extinguish the fire.",
		},
		{
			Tribulation: 70,
			Title:       "The Fish Demon of Heaven-Reaching River",
			Chinese:     "通天河鱼精",
			Passage:     "通天河里有一条金鱼精，是从观音菩萨的莲花池里逃出来的。它每年都要村民献上一对儿童，村民们非常害怕。最后观音菩萨亲自来到河边，用竹篮把金鱼精收了回去。",
			Pinyin:      "Tōngtiān Hé lǐ yǒu yī tiáo jīnyú jīng, shì cóng Guānyīn Púsà de liánhuā chí lǐ táo chūlái de. Tā měi nián dōu yào cūnmín xiàn shàng yī duì értóng, cūnmín men fēicháng hàipà. Zuìhòu Guānyīn Púsà qīnzì lái dào hé biān, yòng zhúlán bǎ jīnyú jīng shōu le huíqù.",
			English:     "In the Heaven-Reaching River there was a goldfish spirit that had escaped from Guanyin Bodhisattva's lotus pond. Every year it demanded the villagers offer up a pair of children, and the villagers were terrified. Finally Guanyin Bodhisattva came to the riverbank in person and used a bamboo basket to recapture the goldfish spirit.",
		},
		{
			Tribulation: 80,
			Title:       "Spirit Mountain in Sight",
			Chinese:     "灵山在望",
			Passage:     "经过了很多年的旅行，唐僧和徒弟们终于看到了远方的灵山。山上金光闪闪，佛音缭绕，非常美丽。大家都很激动，因为他们的目标就在前方了。",
			Pinyin:      "Jīngguò le hěn duō nián de lǚxíng, Táng Sēng hé túdì men zhōngyú kàn dào le yuǎnfāng de Líng Shān. Shān shang jīnguāng shǎnshǎn, fóyīn liáorào, fēicháng měilì. Dàjiā dōu hěn jīdòng, yīnwèi tāmen de mùbiāo jiù zài qiánfāng le.",
			English:     "After many years of travel, Tang Sanzang and his disciples finally saw Spirit Mountain in the distance. The mountain shimmered with golden light, surrounded by the sound of Buddhist chanting, and was very beautiful. Everyone was excited because their goal was right ahead.",
		},
		{
			Tribulation: 81,
			Title:       "Obtaining the True Sutras",
			Chinese:     "取得真经",
			Passage:     "唐僧和徒弟们终于到达了大雷音寺，见到了佛祖。佛祖把真正的佛经交给了他们。经过了八十一个磨难，他们完成了使命，每个人都得到了应有的回报。",
			Pinyin:      "Táng Sēng hé túdì men zhōngyú dàodá le Dà Léiyīn Sì, jiàn dào le Fózǔ. Fózǔ bǎ zhēnzhèng de fójīng jiāo gěi le tāmen. Jīngguò le bāshíyī gè mónàn, tāmen wánchéng le shǐmìng, měi gè rén dōu dédào le yīng yǒu de huíbào.",
			English:     "Tang Sanzang and his disciples finally reached the Great Thunder Monastery and met the Buddha. The Buddha handed them the true Buddhist scriptures. Having passed through eighty-one tribulations, they completed their mission, and each person received their deserved reward.",
		},
	}
}
