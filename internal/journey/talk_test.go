package journey

import (
	"strings"
	"testing"

	"github.com/therealsiege/tuilaoshi/internal/model"
)

func TestTalkPromptCarriesVoiceAndWords(t *testing.T) {
	b := *BondFor("猪八戒")
	p := TalkPrompt(Companion{Name: "猪八戒", English: "Zhu Bajie"}, b, []model.Card{{Chinese: "馒头", Pinyin: "mántou", English: "steamed bun"}})
	for _, want := range []string{"猪八戒", "hungry", "馒头 (mántou, steamed bun)", "BOND: n"} {
		if !strings.Contains(p, want) {
			t.Errorf("prompt lacks %q", want)
		}
	}
}

func TestParseTalkReply(t *testing.T) {
	text, grade := ParseTalkReply("好兄弟，你也饿了吗？\nBOND: 2\n")
	if text != "好兄弟，你也饿了吗？" || grade != 2 {
		t.Errorf("got %q, %d", text, grade)
	}
	if text, grade := ParseTalkReply("没有分数。"); text != "没有分数。" || grade != 0 {
		t.Errorf("no grade line: %q, %d", text, grade)
	}
}
