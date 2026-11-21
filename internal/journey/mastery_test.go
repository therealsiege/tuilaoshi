package journey

import (
	"testing"

	"github.com/cj/mandarin-wiki/internal/model"
)

func TestIsMastered(t *testing.T) {
	tests := []struct {
		name string
		page model.Page
		want bool
	}{
		{"character at threshold", model.Page{Type: "character", ReviewInterval: masteredInterval}, true},
		{"character below threshold", model.Page{Type: "character", ReviewInterval: masteredInterval - 1}, false},
		{"vocabulary never counts", model.Page{Type: "vocabulary", ReviewInterval: 90}, false},
	}
	for _, tt := range tests {
		if got := IsMastered(tt.page); got != tt.want {
			t.Errorf("%s: IsMastered() = %v, want %v", tt.name, got, tt.want)
		}
	}
}
