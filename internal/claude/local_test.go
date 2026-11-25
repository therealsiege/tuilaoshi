package claude

import (
	"testing"

	"github.com/therealsiege/tuilaoshi/internal/model"
)

func TestLocalGrade(t *testing.T) {
	book := model.Card{Chinese: "书", Pinyin: "shū", English: "book; letter (to write)"}
	eat := model.Card{Chinese: "吃饭", Pinyin: "chī fàn", English: "to eat a meal"}
	tests := []struct {
		card   model.Card
		answer string
		mode   model.DrillMode
		want   bool
	}{
		{book, "Book", model.ModeRecognition, true},
		{book, "letter", model.ModeRecognition, true},
		{book, "books and more", model.ModeRecognition, false},
		{eat, "eat a meal", model.ModeRecognition, true},
		{eat, "吃饭", model.ModeProduction, true},
		{eat, "chi fan", model.ModeListening, true},
		{eat, "吃", model.ModeProduction, false},
		{book, "", model.ModeProduction, false},
	}
	for _, tt := range tests {
		if got := LocalGrade(tt.card, tt.answer, tt.mode); got.Correct != tt.want {
			t.Errorf("%s %q (%s): correct %v, want %v", tt.card.Chinese, tt.answer, tt.mode, got.Correct, tt.want)
		}
	}
}

func TestWithoutAKeyDrillsAreGradedLocally(t *testing.T) {
	t.Setenv("ANTHROPIC_API_KEY", "")
	c := New()
	if c.Ready() {
		t.Fatal("ready without a key")
	}
	res, err := c.EvaluateDrill(t.Context(), model.Card{Chinese: "书", Pinyin: "shū", English: "book"}, "book", model.ModeRecognition)
	if err != nil || !res.Correct {
		t.Errorf("local grade %+v %v", res, err)
	}
}
