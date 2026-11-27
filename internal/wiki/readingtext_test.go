package wiki

import (
	"reflect"
	"testing"
)

func TestReadingText(t *testing.T) {
	body := "# Title\n\n## Text\n\n第一行。\n\n第二行。\n\n## Pinyin\n\nDì yī háng.\n"
	if got := ReadingText(body); !reflect.DeepEqual(got, []string{"第一行。", "第二行。"}) {
		t.Errorf("ReadingText = %v", got)
	}
}
