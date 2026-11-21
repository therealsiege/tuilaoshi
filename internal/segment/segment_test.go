package segment

import (
	"reflect"
	"testing"
)

func TestSegmentPrefersLongestKnownWord(t *testing.T) {
	known := map[string]bool{"服务员": true, "服务": true, "菜单": true}
	got := Split("服务员：这是菜单。", known)
	want := []Token{
		{"服务员", true}, {"：", false}, {"这", true}, {"是", true}, {"菜单", true}, {"。", false},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("segment = %v, want %v", got, want)
	}
}
