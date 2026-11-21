package reader

import (
	"reflect"
	"testing"
)

func TestSegmentPrefersLongestKnownWord(t *testing.T) {
	known := map[string]bool{"服务员": true, "服务": true, "菜单": true}
	got := segment("服务员：这是菜单。", known)
	want := []token{
		{"服务员", true}, {"：", false}, {"这", true}, {"是", true}, {"菜单", true}, {"。", false},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("segment = %v, want %v", got, want)
	}
}

func TestTextSection(t *testing.T) {
	body := "# Title\n\n## Text\n\n第一行。\n\n第二行。\n\n## Pinyin\n\nDì yī háng.\n"
	if got := textSection(body); !reflect.DeepEqual(got, []string{"第一行。", "第二行。"}) {
		t.Errorf("textSection = %v", got)
	}
}
