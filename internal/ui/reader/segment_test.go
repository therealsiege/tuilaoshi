package reader

import (
	"reflect"
	"testing"
)

func TestTextSection(t *testing.T) {
	body := "# Title\n\n## Text\n\n第一行。\n\n第二行。\n\n## Pinyin\n\nDì yī háng.\n"
	if got := textSection(body); !reflect.DeepEqual(got, []string{"第一行。", "第二行。"}) {
		t.Errorf("textSection = %v", got)
	}
}
