package episode

import "testing"

func TestCompareShadow(t *testing.T) {
	for _, tt := range []struct {
		line, heard string
		pct         int
		missed      string
	}{
		{"师父，前面有一座大山，我们小心一点儿。", "师父,前面有一座大山,我们小心一点。", 100, ""},
		{"我们去西天取经。", "我们去西天取经", 100, ""},
		{"我们去西天取经。", "我们去天取经", 85, "西"},
		{"妖怪来了！", "", 0, "妖怪来了"},
		{"妖怪来了！", "咬怪来了", 75, "妖"},
		{"第1集第3句", "第一集第三句", 100, ""},
	} {
		s := CompareShadow(tt.line, tt.heard)
		missed := ""
		for _, c := range s.Chars {
			if !c.Heard {
				missed += string(c.Char)
			}
		}
		if s.Percent != tt.pct || missed != tt.missed {
			t.Errorf("%q heard as %q: %d%% missing %q, want %d%% missing %q", tt.line, tt.heard, s.Percent, missed, tt.pct, tt.missed)
		}
	}
}
