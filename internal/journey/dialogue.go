package journey

// Dialogue is a contextual line spoken by a Journey to the West companion.
type Dialogue struct {
	Speaker     string // companion name in Chinese
	English     string // companion name in English
	Line        string // the Chinese dialogue line
	Pinyin      string // pinyin for the line
	Translation string // English translation
	Trigger     string // when this shows: "streak", "rank_up", "boss_defeat", "mistake", "review_complete", "idle"
}

// RandomDialogue returns a contextual dialogue line based on the trigger event.
// Uses a simple rotation based on the count parameter to avoid randomness.
func RandomDialogue(trigger string, count int) *Dialogue {
	matches := dialogueByTrigger(trigger)
	if len(matches) == 0 {
		return nil
	}
	return &matches[count%len(matches)]
}

// dialogueByTrigger filters the dialogue pool to lines matching the given trigger.
func dialogueByTrigger(trigger string) []Dialogue {
	all := dialoguePool()
	var matches []Dialogue
	for _, d := range all {
		if d.Trigger == trigger {
			matches = append(matches, d)
		}
	}
	return matches
}

// dialoguePool returns all companion dialogue lines.
func dialoguePool() []Dialogue {
	return []Dialogue{
		// ── Streak triggers ─────────────────────────────────────────
		// Shown when the learner maintains a study streak.
		{
			Speaker:     "孙悟空",
			English:     "Sun Wukong",
			Line:        "师父，你的坚持比我的筋斗云还厉害！",
			Pinyin:      "Shīfu, nǐ de jiānchí bǐ wǒ de jīndǒuyún hái lìhai!",
			Translation: "Master, your persistence is mightier than my somersault cloud!",
			Trigger:     "streak",
		},
		{
			Speaker:     "猪八戒",
			English:     "Zhu Bajie",
			Line:        "连我都佩服你了，师兄！",
			Pinyin:      "Lián wǒ dōu pèifu nǐ le, shīxiōng!",
			Translation: "Even I admire you, senior brother!",
			Trigger:     "streak",
		},
		{
			Speaker:     "沙悟净",
			English:     "Sha Wujing",
			Line:        "师父每天都在进步，真了不起。",
			Pinyin:      "Shīfu měi tiān dōu zài jìnbù, zhēn liǎobuqǐ.",
			Translation: "Master improves every day, truly remarkable.",
			Trigger:     "streak",
		},
		{
			Speaker:     "白龙马",
			English:     "Bai Longma",
			Line:        "*稳稳地走在路上，每一步都很坚定*",
			Pinyin:      "*wěnwěn de zǒu zài lù shang, měi yī bù dōu hěn jiāndìng*",
			Translation: "*walks steadily on the road, every step firm and sure*",
			Trigger:     "streak",
		},
		{
			Speaker:     "孙悟空",
			English:     "Sun Wukong",
			Line:        "俺老孙看你这劲头，取经一定能成功！",
			Pinyin:      "Ǎn lǎo Sūn kàn nǐ zhè jìntóu, qǔjīng yīdìng néng chénggōng!",
			Translation: "Old Sun sees your drive -- fetching the sutras will surely succeed!",
			Trigger:     "streak",
		},
		{
			Speaker:     "猪八戒",
			English:     "Zhu Bajie",
			Line:        "师父，你比我勤快多了，我都不好意思偷懒了。",
			Pinyin:      "Shīfu, nǐ bǐ wǒ qínkuài duō le, wǒ dōu bù hǎoyìsi tōulǎn le.",
			Translation: "Master, you are much more diligent than me -- I'm too embarrassed to slack off now.",
			Trigger:     "streak",
		},
		{
			Speaker:     "沙悟净",
			English:     "Sha Wujing",
			Line:        "日复一日，师父的功夫越来越深了。",
			Pinyin:      "Rì fù yī rì, shīfu de gōngfu yuè lái yuè shēn le.",
			Translation: "Day after day, Master's skill grows ever deeper.",
			Trigger:     "streak",
		},
		{
			Speaker:     "孙悟空",
			English:     "Sun Wukong",
			Line:        "好样的！这份毅力连天上的神仙都佩服！",
			Pinyin:      "Hǎoyàng de! Zhè fèn yìlì lián tiānshang de shénxiān dōu pèifu!",
			Translation: "Well done! This perseverance would impress even the heavenly immortals!",
			Trigger:     "streak",
		},
		{
			Speaker:     "白龙马",
			English:     "Bai Longma",
			Line:        "*轻轻地点头，眼里满是敬意*",
			Pinyin:      "*qīngqīng de diǎntóu, yǎn lǐ mǎn shì jìngyì*",
			Translation: "*nods gently, eyes full of respect*",
			Trigger:     "streak",
		},
		{
			Speaker:     "猪八戒",
			English:     "Zhu Bajie",
			Line:        "坚持下去就有好吃的……不对，就有好结果！",
			Pinyin:      "Jiānchí xiàqù jiù yǒu hǎochī de... bùduì, jiù yǒu hǎo jiéguǒ!",
			Translation: "Keep going and there'll be good food... wait, I mean good results!",
			Trigger:     "streak",
		},
		{
			Speaker:     "沙悟净",
			English:     "Sha Wujing",
			Line:        "师父的毅力就像流沙河的水，永远不会停下来。",
			Pinyin:      "Shīfu de yìlì jiù xiàng Liúshā Hé de shuǐ, yǒngyuǎn bù huì tíng xiàlái.",
			Translation: "Master's perseverance is like the waters of the Flowing Sands River -- it never stops.",
			Trigger:     "streak",
		},

		// ── Rank up triggers ────────────────────────────────────────
		// Shown when the learner reaches a new cultivation rank.
		{
			Speaker:     "孙悟空",
			English:     "Sun Wukong",
			Line:        "好！师父又突破了一层！",
			Pinyin:      "Hǎo! Shīfu yòu tūpò le yī céng!",
			Translation: "Good! Master has broken through another level!",
			Trigger:     "rank_up",
		},
		{
			Speaker:     "沙悟净",
			English:     "Sha Wujing",
			Line:        "师父的修行又进了一步。",
			Pinyin:      "Shīfu de xiūxíng yòu jìn le yī bù.",
			Translation: "Master's cultivation has advanced another step.",
			Trigger:     "rank_up",
		},
		{
			Speaker:     "猪八戒",
			English:     "Zhu Bajie",
			Line:        "哎呀，师父升级了！我们是不是可以吃顿好的庆祝一下？",
			Pinyin:      "Āiyā, shīfu shēngjí le! Wǒmen shì bù shì kěyǐ chī dùn hǎo de qìngzhù yīxià?",
			Translation: "Oh my, Master leveled up! Can we have a nice meal to celebrate?",
			Trigger:     "rank_up",
		},
		{
			Speaker:     "白龙马",
			English:     "Bai Longma",
			Line:        "*欢快地嘶鸣，高高扬起前蹄*",
			Pinyin:      "*huānkuài de sīmíng, gāogāo yáng qǐ qián tí*",
			Translation: "*whinnies joyfully, rearing up on hind legs*",
			Trigger:     "rank_up",
		},
		{
			Speaker:     "孙悟空",
			English:     "Sun Wukong",
			Line:        "俺老孙就知道师父不简单！继续加油！",
			Pinyin:      "Ǎn lǎo Sūn jiù zhīdào shīfu bù jiǎndān! Jìxù jiāyóu!",
			Translation: "Old Sun always knew Master was extraordinary! Keep it up!",
			Trigger:     "rank_up",
		},
		{
			Speaker:     "沙悟净",
			English:     "Sha Wujing",
			Line:        "恭喜师父，修为更上一层楼。",
			Pinyin:      "Gōngxǐ shīfu, xiūwéi gèng shàng yī céng lóu.",
			Translation: "Congratulations Master, your cultivation has risen to new heights.",
			Trigger:     "rank_up",
		},
		{
			Speaker:     "猪八戒",
			English:     "Zhu Bajie",
			Line:        "师父比我强太多了，我得加把劲才行！",
			Pinyin:      "Shīfu bǐ wǒ qiáng tài duō le, wǒ děi jiā bǎ jìn cái xíng!",
			Translation: "Master is so much better than me -- I need to put in more effort!",
			Trigger:     "rank_up",
		},
		{
			Speaker:     "孙悟空",
			English:     "Sun Wukong",
			Line:        "这一关过了！下一关也不在话下！",
			Pinyin:      "Zhè yī guān guò le! Xià yī guān yě bù zài huà xià!",
			Translation: "This stage is cleared! The next one will be no problem either!",
			Trigger:     "rank_up",
		},
		{
			Speaker:     "白龙马",
			English:     "Bai Longma",
			Line:        "*用力跺蹄，表示庆贺*",
			Pinyin:      "*yònglì duò tí, biǎoshì qìnghè*",
			Translation: "*stamps hooves firmly in celebration*",
			Trigger:     "rank_up",
		},
		{
			Speaker:     "沙悟净",
			English:     "Sha Wujing",
			Line:        "师父一步一个脚印，稳扎稳打。",
			Pinyin:      "Shīfu yī bù yī gè jiǎoyìn, wěn zhā wěn dǎ.",
			Translation: "Master moves one step at a time, steady and sure.",
			Trigger:     "rank_up",
		},
		{
			Speaker:     "猪八戒",
			English:     "Zhu Bajie",
			Line:        "师父都突破了，那我今天也不偷懒了！",
			Pinyin:      "Shīfu dōu tūpò le, nà wǒ jīntiān yě bù tōulǎn le!",
			Translation: "Master has broken through, so I won't slack off today either!",
			Trigger:     "rank_up",
		},

		// ── Boss defeat triggers ────────────────────────────────────
		// Shown when the learner defeats a demon boss encounter.
		{
			Speaker:     "孙悟空",
			English:     "Sun Wukong",
			Line:        "这妖怪不过如此！吃俺老孙一棒！",
			Pinyin:      "Zhè yāoguài bùguò rúcǐ! Chī ǎn lǎo Sūn yī bàng!",
			Translation: "This demon is nothing! Take a blow from Old Sun!",
			Trigger:     "boss_defeat",
		},
		{
			Speaker:     "猪八戒",
			English:     "Zhu Bajie",
			Line:        "师兄威武！不过功劳也有我一份吧？",
			Pinyin:      "Shīxiōng wēiwǔ! Bùguò gōngláo yě yǒu wǒ yī fèn ba?",
			Translation: "Senior brother is mighty! But I deserve some credit too, right?",
			Trigger:     "boss_defeat",
		},
		{
			Speaker:     "沙悟净",
			English:     "Sha Wujing",
			Line:        "妖怪已除，前方的路安全了。",
			Pinyin:      "Yāoguài yǐ chú, qiánfāng de lù ānquán le.",
			Translation: "The demon has been vanquished -- the road ahead is safe now.",
			Trigger:     "boss_defeat",
		},
		{
			Speaker:     "白龙马",
			English:     "Bai Longma",
			Line:        "*昂首阔步，踏过妖怪的阵地*",
			Pinyin:      "*áng shǒu kuò bù, tà guò yāoguài de zhèndì*",
			Translation: "*strides proudly, stepping over the demon's territory*",
			Trigger:     "boss_defeat",
		},
		{
			Speaker:     "孙悟空",
			English:     "Sun Wukong",
			Line:        "哈哈！看你还敢不敢挡俺老孙的路！",
			Pinyin:      "Hāhā! Kàn nǐ hái gǎn bù gǎn dǎng ǎn lǎo Sūn de lù!",
			Translation: "Haha! Let's see if you dare block Old Sun's path again!",
			Trigger:     "boss_defeat",
		},
		{
			Speaker:     "猪八戒",
			English:     "Zhu Bajie",
			Line:        "打完妖怪我饿了，有没有东西吃？",
			Pinyin:      "Dǎ wán yāoguài wǒ è le, yǒu méi yǒu dōngxi chī?",
			Translation: "After beating the demon I'm hungry -- is there anything to eat?",
			Trigger:     "boss_defeat",
		},
		{
			Speaker:     "沙悟净",
			English:     "Sha Wujing",
			Line:        "师父不用担心了，这一难已经过去了。",
			Pinyin:      "Shīfu bù yòng dānxīn le, zhè yī nàn yǐjīng guòqù le.",
			Translation: "Master need not worry anymore -- this tribulation has passed.",
			Trigger:     "boss_defeat",
		},
		{
			Speaker:     "孙悟空",
			English:     "Sun Wukong",
			Line:        "师父放心，有俺老孙在，什么妖怪都不怕！",
			Pinyin:      "Shīfu fàngxīn, yǒu ǎn lǎo Sūn zài, shénme yāoguài dōu bù pà!",
			Translation: "Rest easy, Master -- with Old Sun here, no demon need be feared!",
			Trigger:     "boss_defeat",
		},
		{
			Speaker:     "白龙马",
			English:     "Bai Longma",
			Line:        "*满意地喷了一口气，甩了甩尾巴*",
			Pinyin:      "*mǎnyì de pēn le yī kǒu qì, shuǎi le shuǎi wěiba*",
			Translation: "*snorts with satisfaction and swishes tail*",
			Trigger:     "boss_defeat",
		},
		{
			Speaker:     "猪八戒",
			English:     "Zhu Bajie",
			Line:        "这个妖怪比我想的好打！嘿嘿，师父厉害！",
			Pinyin:      "Zhège yāoguài bǐ wǒ xiǎng de hǎo dǎ! Hēihēi, shīfu lìhai!",
			Translation: "This demon was easier than I expected! Hehe, Master is amazing!",
			Trigger:     "boss_defeat",
		},
		{
			Speaker:     "沙悟净",
			English:     "Sha Wujing",
			Line:        "又一个妖怪败了，师父的力量越来越强。",
			Pinyin:      "Yòu yī gè yāoguài bài le, shīfu de lìliàng yuè lái yuè qiáng.",
			Translation: "Another demon has fallen -- Master's power grows ever stronger.",
			Trigger:     "boss_defeat",
		},

		// ── Mistake triggers ────────────────────────────────────────
		// Shown when the learner gets a character wrong.
		{
			Speaker:     "孙悟空",
			English:     "Sun Wukong",
			Line:        "师父别灰心，再来一次！",
			Pinyin:      "Shīfu bié huīxīn, zài lái yī cì!",
			Translation: "Don't be discouraged, Master! Try again!",
			Trigger:     "mistake",
		},
		{
			Speaker:     "沙悟净",
			English:     "Sha Wujing",
			Line:        "没关系，慢慢来。",
			Pinyin:      "Méi guānxi, mànman lái.",
			Translation: "It's okay, take it slow.",
			Trigger:     "mistake",
		},
		{
			Speaker:     "猪八戒",
			English:     "Zhu Bajie",
			Line:        "师父，我也经常记错呢！",
			Pinyin:      "Shīfu, wǒ yě jīngcháng jì cuò ne!",
			Translation: "Master, I often get things wrong too!",
			Trigger:     "mistake",
		},
		{
			Speaker:     "白龙马",
			English:     "Bai Longma",
			Line:        "*轻轻蹭了蹭师父的手，表示安慰*",
			Pinyin:      "*qīngqīng cèng le cèng shīfu de shǒu, biǎoshì ānwèi*",
			Translation: "*gently nuzzles Master's hand in comfort*",
			Trigger:     "mistake",
		},
		{
			Speaker:     "孙悟空",
			English:     "Sun Wukong",
			Line:        "错一次怕什么？俺老孙大闹天宫也失败过！",
			Pinyin:      "Cuò yī cì pà shénme? Ǎn lǎo Sūn dànào tiāngōng yě shībài guò!",
			Translation: "What's one mistake to fear? Even Old Sun failed when wreaking havoc in heaven!",
			Trigger:     "mistake",
		},
		{
			Speaker:     "猪八戒",
			English:     "Zhu Bajie",
			Line:        "别难过了，我请师父吃桃子！哦，桃子被猴哥吃了……",
			Pinyin:      "Bié nánguò le, wǒ qǐng shīfu chī táozi! Ó, táozi bèi hóugē chī le...",
			Translation: "Don't be sad, I'll treat Master to peaches! Oh, the monkey ate them all...",
			Trigger:     "mistake",
		},
		{
			Speaker:     "沙悟净",
			English:     "Sha Wujing",
			Line:        "错误是学习的一部分，师父不必在意。",
			Pinyin:      "Cuòwù shì xuéxí de yī bùfen, shīfu bùbì zàiyì.",
			Translation: "Mistakes are part of learning -- Master need not worry about it.",
			Trigger:     "mistake",
		},
		{
			Speaker:     "孙悟空",
			English:     "Sun Wukong",
			Line:        "没事！这个字比妖怪还难，但师父一定能记住！",
			Pinyin:      "Méishì! Zhège zì bǐ yāoguài hái nán, dàn shīfu yīdìng néng jì zhù!",
			Translation: "No worries! This character is harder than a demon, but Master will surely remember it!",
			Trigger:     "mistake",
		},
		{
			Speaker:     "白龙马",
			English:     "Bai Longma",
			Line:        "*安静地站在一旁，耐心等待*",
			Pinyin:      "*ānjìng de zhàn zài yīpáng, nàixīn děngdài*",
			Translation: "*stands quietly nearby, waiting patiently*",
			Trigger:     "mistake",
		},
		{
			Speaker:     "猪八戒",
			English:     "Zhu Bajie",
			Line:        "这个字确实难，连我看了都头疼！",
			Pinyin:      "Zhège zì quèshí nán, lián wǒ kàn le dōu tóuténg!",
			Translation: "This character really is hard -- even I get a headache looking at it!",
			Trigger:     "mistake",
		},
		{
			Speaker:     "沙悟净",
			English:     "Sha Wujing",
			Line:        "师父再看一遍，我相信下次一定能记住。",
			Pinyin:      "Shīfu zài kàn yī biàn, wǒ xiāngxìn xià cì yīdìng néng jì zhù.",
			Translation: "Master, look at it once more -- I believe you will remember next time for sure.",
			Trigger:     "mistake",
		},

		// ── Review complete triggers ────────────────────────────────
		// Shown when a review session is finished.
		{
			Speaker:     "沙悟净",
			English:     "Sha Wujing",
			Line:        "今天的修行结束了，师父辛苦了。",
			Pinyin:      "Jīntiān de xiūxíng jiéshù le, shīfu xīnkǔ le.",
			Translation: "Today's practice is done -- Master has worked hard.",
			Trigger:     "review_complete",
		},
		{
			Speaker:     "白龙马",
			English:     "Bai Longma",
			Line:        "*赞许地跺了跺蹄子*",
			Pinyin:      "*zànxǔ de duò le duò tízi*",
			Translation: "*stamps hooves approvingly*",
			Trigger:     "review_complete",
		},
		{
			Speaker:     "孙悟空",
			English:     "Sun Wukong",
			Line:        "不错不错！今天的功课做完了，明天继续！",
			Pinyin:      "Búcuò búcuò! Jīntiān de gōngkè zuò wán le, míngtiān jìxù!",
			Translation: "Not bad! Today's lesson is done -- continue tomorrow!",
			Trigger:     "review_complete",
		},
		{
			Speaker:     "猪八戒",
			English:     "Zhu Bajie",
			Line:        "终于做完了！可以休息一下吃点东西了吧？",
			Pinyin:      "Zhōngyú zuò wán le! Kěyǐ xiūxi yīxià chī diǎn dōngxi le ba?",
			Translation: "Finally done! Can we rest and eat something now?",
			Trigger:     "review_complete",
		},
		{
			Speaker:     "沙悟净",
			English:     "Sha Wujing",
			Line:        "师父今天学得很认真，我来帮师父整理行李。",
			Pinyin:      "Shīfu jīntiān xué de hěn rènzhēn, wǒ lái bāng shīfu zhěnglǐ xínglǐ.",
			Translation: "Master studied very earnestly today -- let me help pack the luggage.",
			Trigger:     "review_complete",
		},
		{
			Speaker:     "孙悟空",
			English:     "Sun Wukong",
			Line:        "今天的路走完了，师父好好休息吧！",
			Pinyin:      "Jīntiān de lù zǒu wán le, shīfu hǎohǎo xiūxi ba!",
			Translation: "Today's journey is complete -- Master, please rest well!",
			Trigger:     "review_complete",
		},
		{
			Speaker:     "白龙马",
			English:     "Bai Longma",
			Line:        "*安静地卧在树下，闭上了眼睛*",
			Pinyin:      "*ānjìng de wò zài shù xià, bì shàng le yǎnjīng*",
			Translation: "*lies down quietly under a tree and closes its eyes*",
			Trigger:     "review_complete",
		},
		{
			Speaker:     "猪八戒",
			English:     "Zhu Bajie",
			Line:        "师父今天太厉害了，我都看饿了！",
			Pinyin:      "Shīfu jīntiān tài lìhai le, wǒ dōu kàn è le!",
			Translation: "Master was so amazing today that just watching made me hungry!",
			Trigger:     "review_complete",
		},
		{
			Speaker:     "沙悟净",
			English:     "Sha Wujing",
			Line:        "今天又往西天近了一步，师父早点休息吧。",
			Pinyin:      "Jīntiān yòu wǎng xītiān jìn le yī bù, shīfu zǎo diǎn xiūxi ba.",
			Translation: "Today we moved one step closer to the Western Heaven -- Master, please rest early.",
			Trigger:     "review_complete",
		},
		{
			Speaker:     "孙悟空",
			English:     "Sun Wukong",
			Line:        "师父，今天的功夫没白练！俺去给你摘些果子。",
			Pinyin:      "Shīfu, jīntiān de gōngfu méi bái liàn! Ǎn qù gěi nǐ zhāi xiē guǒzi.",
			Translation: "Master, today's practice was not in vain! Let me go pick some fruit for you.",
			Trigger:     "review_complete",
		},
		{
			Speaker:     "猪八戒",
			English:     "Zhu Bajie",
			Line:        "辛苦辛苦！下次少学一点也没关系嘛……",
			Pinyin:      "Xīnkǔ xīnkǔ! Xià cì shǎo xué yīdiǎn yě méi guānxi ma...",
			Translation: "Hard work! Maybe learn a bit less next time, that's fine too...",
			Trigger:     "review_complete",
		},

		// ── Idle triggers ───────────────────────────────────────────
		// Shown on dashboard when there is no recent study activity.
		{
			Speaker:     "猪八戒",
			English:     "Zhu Bajie",
			Line:        "师父，我们休息够了吧？该继续赶路了！",
			Pinyin:      "Shīfu, wǒmen xiūxi gòu le ba? Gāi jìxù gǎnlù le!",
			Translation: "Master, we've rested enough, right? Time to continue the journey!",
			Trigger:     "idle",
		},
		{
			Speaker:     "孙悟空",
			English:     "Sun Wukong",
			Line:        "师父，妖怪可不会等我们的！",
			Pinyin:      "Shīfu, yāoguài kě bù huì děng wǒmen de!",
			Translation: "Master, the demons won't wait for us!",
			Trigger:     "idle",
		},
		{
			Speaker:     "沙悟净",
			English:     "Sha Wujing",
			Line:        "师父，我已经把行李整理好了，随时可以出发。",
			Pinyin:      "Shīfu, wǒ yǐjīng bǎ xínglǐ zhěnglǐ hǎo le, suíshí kěyǐ chūfā.",
			Translation: "Master, I've already packed the luggage -- we can depart at any time.",
			Trigger:     "idle",
		},
		{
			Speaker:     "白龙马",
			English:     "Bai Longma",
			Line:        "*不安地踏着蹄子，想要继续赶路*",
			Pinyin:      "*bù'ān de tà zhe tízi, xiǎng yào jìxù gǎnlù*",
			Translation: "*paws the ground restlessly, eager to continue the journey*",
			Trigger:     "idle",
		},
		{
			Speaker:     "猪八戒",
			English:     "Zhu Bajie",
			Line:        "连我都坐不住了，师父快来学习吧！",
			Pinyin:      "Lián wǒ dōu zuò bù zhù le, shīfu kuài lái xuéxí ba!",
			Translation: "Even I can't sit still anymore -- Master, come study!",
			Trigger:     "idle",
		},
		{
			Speaker:     "孙悟空",
			English:     "Sun Wukong",
			Line:        "师父！前面的路还很长，不能停下来啊！",
			Pinyin:      "Shīfu! Qiánmiàn de lù hái hěn cháng, bù néng tíng xiàlái a!",
			Translation: "Master! The road ahead is still long -- we cannot stop!",
			Trigger:     "idle",
		},
		{
			Speaker:     "沙悟净",
			English:     "Sha Wujing",
			Line:        "师父，经书不会自己走过来的。",
			Pinyin:      "Shīfu, jīngshū bù huì zìjǐ zǒu guòlái de.",
			Translation: "Master, the sutras won't walk to us on their own.",
			Trigger:     "idle",
		},
		{
			Speaker:     "猪八戒",
			English:     "Zhu Bajie",
			Line:        "我都起来了，师父你还不来？那我也回去睡了！",
			Pinyin:      "Wǒ dōu qǐlái le, shīfu nǐ hái bù lái? Nà wǒ yě huíqù shuì le!",
			Translation: "Even I'm up already -- Master, aren't you coming? Then I'm going back to sleep!",
			Trigger:     "idle",
		},
		{
			Speaker:     "孙悟空",
			English:     "Sun Wukong",
			Line:        "俺在这里等了好久了！师父快来！",
			Pinyin:      "Ǎn zài zhèlǐ děng le hǎojiǔ le! Shīfu kuài lái!",
			Translation: "I've been waiting here for ages! Master, hurry up!",
			Trigger:     "idle",
		},
		{
			Speaker:     "白龙马",
			English:     "Bai Longma",
			Line:        "*朝着西方的方向看，焦急地甩着尾巴*",
			Pinyin:      "*cháo zhe xīfāng de fāngxiàng kàn, jiāojí de shuǎi zhe wěiba*",
			Translation: "*looks toward the west, swishing tail anxiously*",
			Trigger:     "idle",
		},
		{
			Speaker:     "沙悟净",
			English:     "Sha Wujing",
			Line:        "师父，不学习的日子，妖怪就会变强。",
			Pinyin:      "Shīfu, bù xuéxí de rìzi, yāoguài jiù huì biàn qiáng.",
			Translation: "Master, on days without study, the demons grow stronger.",
			Trigger:     "idle",
		},
	}
}
