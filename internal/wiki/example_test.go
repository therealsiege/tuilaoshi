package wiki

import "testing"

func TestParseExample(t *testing.T) {
	zh, en, ok := ParseExample("- 请把门关上。(Qǐng bǎ mén guān shàng.) — Please close the door.")
	if !ok || zh != "请把门关上。" || en != "Please close the door." {
		t.Errorf("got %q %q %v", zh, en, ok)
	}
	if _, _, ok := ParseExample("- 没有翻译"); ok {
		t.Error("parsed a line with no translation")
	}
}
