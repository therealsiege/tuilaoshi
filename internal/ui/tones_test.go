package ui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

func TestSplitPinyin(t *testing.T) {
	for word, want := range map[string]string{
		"xuéxí": "xué xí", "Zhōngguó": "Zhōng guó", "shū": "shū", "nǚ'ér": "nǚ ' ér",
		"xièxie": "xiè xie", "péngyou": "péng you", "zhīdào": "zhī dào",
	} {
		if got := strings.Join(SplitPinyin(word), " "); got != want {
			t.Errorf("%s: %q, want %q", word, got, want)
		}
	}
}

func TestToneColoursKeepTheText(t *testing.T) {
	saved := Prefs
	defer func() { Prefs = saved }()
	Prefs.ToneColours = true
	if got := ansi.Strip(Pinyin("xuéxí zhōngwén")); got != "xuéxí zhōngwén" {
		t.Errorf("coloured pinyin reads %q", got)
	}
}
