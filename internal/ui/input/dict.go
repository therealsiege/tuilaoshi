package input

import (
	"sort"
	"strings"

	"github.com/cj/mandarin-wiki/internal/model"
)

// Candidate represents a Chinese character option for a given pinyin.
type Candidate struct {
	Char   string // e.g. "你"
	Pinyin string // e.g. "nǐ"
	Freq   int    // lower = more common
}

// Dict maps tone-stripped pinyin to candidate characters.
type Dict struct {
	entries map[string][]Candidate
}

// NewDict creates an empty dictionary.
func NewDict() *Dict {
	return &Dict{entries: make(map[string][]Candidate)}
}

// NewFromCards builds a Dict from flashcards extracted from wiki pages,
// supplemented with a base table of common characters.
func NewFromCards(cards []model.Card) *Dict {
	d := NewDict()

	// Add base characters first (higher freq number = lower priority)
	for i, entry := range baseChars {
		key := strings.ToLower(entry.pinyin)
		stripped, _ := stripDiacritics(key)
		if stripped == "" {
			stripped = stripToneNumber(key)
		}
		d.add(stripped, Candidate{
			Char:   entry.char,
			Pinyin: entry.pinyin,
			Freq:   500 + i, // base chars are lower priority than wiki cards
		})
	}

	// Add wiki card entries (lower freq = higher priority)
	for i, card := range cards {
		if card.Chinese == "" || card.Pinyin == "" {
			continue
		}

		// Handle multi-character words: each char maps to its syllable
		chars := []rune(card.Chinese)
		syllables := strings.Fields(card.Pinyin)

		if len(chars) == 1 && len(syllables) >= 1 {
			// Single character — map to first syllable
			py := syllables[0]
			stripped, _ := stripDiacritics(py)
			if stripped == "" {
				stripped = stripToneNumber(strings.ToLower(py))
			}
			d.add(stripped, Candidate{
				Char:   card.Chinese,
				Pinyin: py,
				Freq:   i,
			})
		}

		// Also add the full word keyed by the full pinyin (no spaces)
		fullKey, _ := stripDiacritics(strings.ReplaceAll(card.Pinyin, " ", ""))
		if fullKey == "" {
			fullKey = stripToneNumber(strings.ToLower(strings.ReplaceAll(card.Pinyin, " ", "")))
		}
		if len(chars) > 1 {
			d.add(fullKey, Candidate{
				Char:   card.Chinese,
				Pinyin: card.Pinyin,
				Freq:   i,
			})
		}
	}

	// Sort all candidate lists by frequency
	for key := range d.entries {
		sort.Slice(d.entries[key], func(i, j int) bool {
			return d.entries[key][i].Freq < d.entries[key][j].Freq
		})
		// Deduplicate by character
		d.entries[key] = dedup(d.entries[key])
	}

	return d
}

// Lookup returns candidates matching the given pinyin (tone-stripped).
func (d *Dict) Lookup(pinyin string) []Candidate {
	if d == nil {
		return nil
	}
	key := strings.ToLower(pinyin)
	return d.entries[key]
}

func (d *Dict) add(key string, c Candidate) {
	d.entries[key] = append(d.entries[key], c)
}

func dedup(cs []Candidate) []Candidate {
	seen := make(map[string]bool)
	var out []Candidate
	for _, c := range cs {
		if !seen[c.Char] {
			seen[c.Char] = true
			out = append(out, c)
		}
	}
	return out
}

type baseEntry struct {
	char   string
	pinyin string
}

// baseChars provides the ~300 most common characters so the user can type
// characters they haven't studied yet.
var baseChars = []baseEntry{
	{"的", "de"}, {"一", "yī"}, {"是", "shì"}, {"不", "bù"}, {"了", "le"},
	{"人", "rén"}, {"我", "wǒ"}, {"在", "zài"}, {"有", "yǒu"}, {"他", "tā"},
	{"这", "zhè"}, {"中", "zhōng"}, {"大", "dà"}, {"来", "lái"}, {"上", "shàng"},
	{"国", "guó"}, {"个", "gè"}, {"到", "dào"}, {"说", "shuō"}, {"们", "men"},
	{"为", "wèi"}, {"子", "zǐ"}, {"和", "hé"}, {"你", "nǐ"}, {"地", "dì"},
	{"出", "chū"}, {"会", "huì"}, {"时", "shí"}, {"要", "yào"}, {"也", "yě"},
	{"以", "yǐ"}, {"生", "shēng"}, {"能", "néng"}, {"对", "duì"}, {"里", "lǐ"},
	{"就", "jiù"}, {"学", "xué"}, {"下", "xià"}, {"自", "zì"}, {"年", "nián"},
	{"过", "guò"}, {"发", "fā"}, {"后", "hòu"}, {"作", "zuò"}, {"行", "xíng"},
	{"方", "fāng"}, {"又", "yòu"}, {"如", "rú"}, {"前", "qián"}, {"所", "suǒ"},
	{"本", "běn"}, {"见", "jiàn"}, {"经", "jīng"}, {"头", "tóu"}, {"面", "miàn"},
	{"事", "shì"}, {"好", "hǎo"}, {"小", "xiǎo"}, {"心", "xīn"}, {"她", "tā"},
	{"两", "liǎng"}, {"去", "qù"}, {"多", "duō"}, {"日", "rì"}, {"家", "jiā"},
	{"没", "méi"}, {"动", "dòng"}, {"开", "kāi"}, {"起", "qǐ"}, {"什", "shén"},
	{"么", "me"}, {"都", "dōu"}, {"把", "bǎ"}, {"那", "nà"}, {"还", "hái"},
	{"天", "tiān"}, {"想", "xiǎng"}, {"看", "kàn"}, {"话", "huà"}, {"很", "hěn"},
	{"当", "dāng"}, {"明", "míng"}, {"得", "de"}, {"走", "zǒu"}, {"问", "wèn"},
	{"回", "huí"}, {"手", "shǒu"}, {"只", "zhǐ"}, {"从", "cóng"}, {"太", "tài"},
	{"着", "zhe"}, {"进", "jìn"}, {"成", "chéng"}, {"让", "ràng"}, {"长", "cháng"},
	{"别", "bié"}, {"点", "diǎn"}, {"给", "gěi"}, {"知", "zhī"}, {"道", "dào"},
	{"吃", "chī"}, {"喝", "hē"}, {"水", "shuǐ"}, {"花", "huā"}, {"月", "yuè"},
	{"山", "shān"}, {"书", "shū"}, {"电", "diàn"}, {"话", "huà"}, {"车", "chē"},
	{"门", "mén"}, {"马", "mǎ"}, {"鱼", "yú"}, {"鸟", "niǎo"}, {"火", "huǒ"},
	{"木", "mù"}, {"口", "kǒu"}, {"目", "mù"}, {"耳", "ěr"}, {"手", "shǒu"},
	{"足", "zú"}, {"土", "tǔ"}, {"金", "jīn"}, {"白", "bái"}, {"黑", "hēi"},
	{"红", "hóng"}, {"绿", "lǜ"}, {"蓝", "lán"}, {"黄", "huáng"}, {"东", "dōng"},
	{"西", "xī"}, {"南", "nán"}, {"北", "běi"}, {"左", "zuǒ"}, {"右", "yòu"},
	{"男", "nán"}, {"女", "nǚ"}, {"老", "lǎo"}, {"少", "shǎo"}, {"新", "xīn"},
	{"高", "gāo"}, {"低", "dī"}, {"快", "kuài"}, {"慢", "màn"}, {"早", "zǎo"},
	{"晚", "wǎn"}, {"远", "yuǎn"}, {"近", "jìn"}, {"热", "rè"}, {"冷", "lěng"},
	{"买", "mǎi"}, {"卖", "mài"}, {"写", "xiě"}, {"读", "dú"}, {"听", "tīng"},
	{"说", "shuō"}, {"做", "zuò"}, {"吗", "ma"}, {"呢", "ne"}, {"了", "le"},
	{"的", "de"}, {"打", "dǎ"}, {"住", "zhù"}, {"坐", "zuò"}, {"站", "zhàn"},
	{"跑", "pǎo"}, {"飞", "fēi"}, {"笑", "xiào"}, {"哭", "kū"}, {"爱", "ài"},
	{"恨", "hèn"}, {"怕", "pà"}, {"忙", "máng"}, {"累", "lèi"}, {"饿", "è"},
	{"渴", "kě"}, {"病", "bìng"}, {"死", "sǐ"}, {"错", "cuò"}, {"对", "duì"},
	{"朋", "péng"}, {"友", "yǒu"}, {"同", "tóng"}, {"工", "gōng"}, {"钱", "qián"},
	{"店", "diàn"}, {"饭", "fàn"}, {"菜", "cài"}, {"肉", "ròu"}, {"米", "mǐ"},
	{"茶", "chá"}, {"酒", "jiǔ"}, {"果", "guǒ"}, {"衣", "yī"}, {"服", "fú"},
	{"帽", "mào"}, {"鞋", "xié"}, {"房", "fáng"}, {"间", "jiān"}, {"桌", "zhuō"},
	{"椅", "yǐ"}, {"床", "chuáng"}, {"窗", "chuāng"}, {"灯", "dēng"}, {"雨", "yǔ"},
	{"雪", "xuě"}, {"风", "fēng"}, {"云", "yún"}, {"星", "xīng"}, {"河", "hé"},
	{"海", "hǎi"}, {"路", "lù"}, {"城", "chéng"}, {"市", "shì"}, {"村", "cūn"},
	{"校", "xiào"}, {"医", "yī"}, {"院", "yuàn"}, {"猫", "māo"}, {"狗", "gǒu"},
	{"牛", "niú"}, {"羊", "yáng"}, {"猪", "zhū"}, {"鸡", "jī"}, {"龙", "lóng"},
	{"虎", "hǔ"}, {"蛇", "shé"}, {"兔", "tù"}, {"哥", "gē"}, {"姐", "jiě"},
	{"弟", "dì"}, {"妹", "mèi"}, {"爸", "bà"}, {"妈", "mā"}, {"儿", "ér"},
	{"女", "nǚ"}, {"先", "xiān"}, {"师", "shī"}, {"教", "jiào"}, {"课", "kè"},
	{"字", "zì"}, {"词", "cí"}, {"句", "jù"}, {"文", "wén"}, {"语", "yǔ"},
	{"数", "shù"}, {"百", "bǎi"}, {"千", "qiān"}, {"万", "wàn"}, {"零", "líng"},
	{"二", "èr"}, {"三", "sān"}, {"四", "sì"}, {"五", "wǔ"}, {"六", "liù"},
	{"七", "qī"}, {"八", "bā"}, {"九", "jiǔ"}, {"十", "shí"}, {"半", "bàn"},
	{"现", "xiàn"}, {"正", "zhèng"}, {"最", "zuì"}, {"已", "yǐ"}, {"才", "cái"},
	{"再", "zài"}, {"比", "bǐ"}, {"每", "měi"}, {"被", "bèi"}, {"应", "yīng"},
	{"该", "gāi"}, {"可", "kě"}, {"但", "dàn"}, {"因", "yīn"}, {"虽", "suī"},
	{"然", "rán"}, {"其", "qí"}, {"与", "yǔ"}, {"或", "huò"}, {"且", "qiě"},
	{"若", "ruò"}, {"此", "cǐ"}, {"第", "dì"}, {"等", "děng"}, {"些", "xiē"},
	{"那", "nà"}, {"谁", "shuí"}, {"哪", "nǎ"}, {"几", "jǐ"}, {"怎", "zěn"},
	{"样", "yàng"}, {"啊", "a"}, {"吧", "ba"}, {"嗯", "ǹg"}, {"喂", "wèi"},
	{"请", "qǐng"}, {"谢", "xiè"}, {"对不起", "duìbuqǐ"}, {"没关系", "méiguānxi"},
	{"再见", "zàijiàn"}, {"你好", "nǐhǎo"}, {"谢谢", "xièxie"},
}
