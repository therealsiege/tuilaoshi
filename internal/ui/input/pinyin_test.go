package input

import "testing"

func TestIsValidSyllable(t *testing.T) {
	tests := []struct {
		in   string
		want bool
	}{
		{"ni3", true},
		{"zhong1", true}, // two-letter initial
		{"shi4", true},
		{"lv4", true},
		{"a", true}, // no initial
		{"ang4", true},
		{"er2", true},
		{"zh", false}, // initial only
		{"x", false},
		{"qq", false},
		{"3", false},
		{"", false},
	}
	for _, tt := range tests {
		if got := IsValidSyllable(tt.in); got != tt.want {
			t.Errorf("IsValidSyllable(%q) = %v, want %v", tt.in, got, tt.want)
		}
	}
}
