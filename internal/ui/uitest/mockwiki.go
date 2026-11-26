package uitest

import (
	"fmt"
	"strings"
)

// mockWords are HSK4 vocabulary rows: word, pinyin, meaning, example.
var mockWords = [][4]string{
	{"开车", "kāichē", "to drive", "你会开车吗？"},
	{"驾照", "jiàzhào", "driver's license", "我去年拿到了驾照。"},
	{"停车", "tíngchē", "to park", "这附近很难找到停车的地方。"},
	{"堵车", "dǔchē", "traffic jam", "早高峰堵车特别严重。"},
	{"交通", "jiāotōng", "traffic", "北京的交通非常拥挤。"},
	{"罚款", "fákuǎn", "fine", "闯红灯会被罚款两百块。"},
	{"保险", "bǎoxiǎn", "insurance", "买车以后一定要买保险。"},
	{"事故", "shìgù", "accident", "高速公路上发生了一起事故。"},
	{"推荐", "tuījiàn", "to recommend", "我推荐你去看这部电影。"},
	{"展览", "zhǎnlǎn", "exhibition", "博物馆的展览很精彩。"},
	{"演出", "yǎnchū", "performance", "今天晚上的演出八点开始。"},
	{"观众", "guānzhòng", "audience", "观众都站起来鼓掌。"},
	{"比赛", "bǐsài", "match", "这场比赛非常激烈。"},
	{"报名", "bàomíng", "to sign up", "我已经报名参加了。"},
	{"放松", "fàngsōng", "to relax", "周末我喜欢在家放松。"},
	{"感动", "gǎndòng", "to move", "这个故事让我很感动。"},
	{"打扰", "dǎrǎo", "to disturb", "对不起打扰你一下。"},
	{"提醒", "tíxǐng", "to remind", "请提醒我明天带护照。"},
	{"失望", "shīwàng", "disappointed", "考试结果让他很失望。"},
	{"顺利", "shùnlì", "smoothly", "祝你一路顺利。"},
}

// mockLines are example lines with English, five to a page.
var mockLines = [][]string{
	{"我们明天早上出发。 — We set off tomorrow morning.", "这条路晚上很安全。 — This road is safe at night.",
		"他把行李放在门口了。 — He left the luggage by the door.", "火车比飞机便宜多了。 — The train is much cheaper than the plane.",
		"我们在车站等了很久。 — We waited at the station for a long time."},
	{"我下个月要换工作。 — I'm changing jobs next month.", "老板对我的报告很满意。 — The boss was pleased with my report.",
		"这个会议推迟到周五了。 — The meeting was put off until Friday.", "她每天加班到很晚。 — She works late every day.",
		"公司打算招聘新员工。 — The company plans to hire new staff."},
	{"你应该多喝水少喝咖啡。 — You should drink more water and less coffee.", "他感冒了所以没来上课。 — He had a cold, so he missed class.",
		"医生让我好好休息。 — The doctor told me to rest well.", "每天跑步对身体有好处。 — Running every day is good for you.",
		"我最近睡得不太好。 — I haven't been sleeping well lately."},
	{"今天的天气特别热。 — It's especially hot today.", "下雨了别忘了带伞。 — It's raining, don't forget an umbrella.",
		"北方的冬天非常冷。 — Winters in the north are very cold.", "明天可能会下雪。 — It may snow tomorrow.",
		"春天是最舒服的季节。 — Spring is the most comfortable season."},
}

// MockTestPages is a small HSK4 wiki for mock test screens: two
// vocabulary pages with example columns, four pages of example lines, and
// an HSK3 page of common words for splitting sentences into tiles.
func MockTestPages() map[string]string {
	pages := map[string]string{}
	for p := 0; p < 2; p++ {
		var b strings.Builder
		fmt.Fprintf(&b, "---\ntype: vocabulary\nlevel: HSK4\n---\n\n# Words %d\n\n| Word | Pinyin | Meaning | Example |\n|---|---|---|---|\n", p+1)
		for _, w := range mockWords[p*10 : p*10+10] {
			fmt.Fprintf(&b, "| %s | %s | %s | %s |\n", w[0], w[1], w[2], w[3])
		}
		pages[fmt.Sprintf("vocabulary/words-%d.md", p+1)] = b.String()
	}
	pages["vocabulary/common.md"] = "---\ntype: vocabulary\nlevel: HSK3\n---\n\n# Common\n\n| Word | Pinyin | Meaning |\n|---|---|---|\n" +
		"| 我们 | wǒmen | we |\n| 明天 | míngtiān | tomorrow |\n| 早上 | zǎoshang | morning |\n| 出发 | chūfā | to set off |\n" +
		"| 今天 | jīntiān | today |\n| 天气 | tiānqì | weather |\n| 特别 | tèbié | especially |\n| 冬天 | dōngtiān | winter |\n" +
		"| 非常 | fēicháng | very |\n| 北方 | běifāng | the north |\n| 可能 | kěnéng | maybe |\n| 下雪 | xiàxuě | to snow |\n"
	for i, lines := range mockLines {
		body := "---\ntype: vocabulary\nlevel: HSK4\n---\n\n# Lines\n\n## Example Sentences\n\n- " + strings.Join(lines, "\n- ") + "\n"
		pages[fmt.Sprintf("vocabulary/lines-%d.md", i+1)] = body
	}
	return pages
}
