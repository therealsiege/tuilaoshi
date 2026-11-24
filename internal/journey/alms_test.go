package journey

import "testing"

func TestAlmsScenesAreComplete(t *testing.T) {
	if len(content.Alms) < 6 {
		t.Fatalf("only %d alms scenes", len(content.Alms))
	}
	for i, a := range content.Alms {
		if a.Text == "" || a.English == "" || a.Thanks == "" || a.ThanksEnglish == "" {
			t.Errorf("alms %d is missing text, English or thanks", i)
		}
		if len(a.Choices) != 3 || len(a.EnglishChoices) != len(a.Choices) {
			t.Errorf("alms %d: want 3 choices, each with English; got %d and %d", i, len(a.Choices), len(a.EnglishChoices))
		}
	}
}
