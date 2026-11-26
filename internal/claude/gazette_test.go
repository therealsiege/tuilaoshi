package claude

import "testing"

func TestParseGazette(t *testing.T) {
	zh, en, err := ParseGazette("TEXT: 本报讯：圣僧一行抵达宝象国。\n\n据悉，国王亲自盖印。\nENGLISH: Our report: the holy monk reached Baoxiang.\nWe hear the king stamped the papers himself.\n")
	if err != nil || zh != "本报讯：圣僧一行抵达宝象国。\n据悉，国王亲自盖印。" || en != "Our report: the holy monk reached Baoxiang.\nWe hear the king stamped the papers himself." {
		t.Errorf("%q %q %v", zh, en, err)
	}
	if _, _, err := ParseGazette("ENGLISH: nothing"); err == nil {
		t.Error("an issue without TEXT should fail")
	}
}
