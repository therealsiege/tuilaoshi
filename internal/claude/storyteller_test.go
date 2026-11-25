package claude

import (
	"strings"
	"testing"
)

func TestParseChapter(t *testing.T) {
	reply := "TITLE: 灵根孕育 石猴出世\nTEXT: 很久以前，有一座花果山。\n山上有一块石头。\n\n石头里出来一只猴子。欲知后事如何，且听下回分解。\nENGLISH: A stone monkey is born\non the Mountain of Flowers and Fruit."
	ch, err := ParseChapter(reply)
	if err != nil {
		t.Fatal(err)
	}
	if ch.Title != "灵根孕育 石猴出世" {
		t.Errorf("title %q", ch.Title)
	}
	if want := "很久以前，有一座花果山。\n山上有一块石头。\n石头里出来一只猴子。欲知后事如何，且听下回分解。"; ch.Text != want {
		t.Errorf("text %q, want %q", ch.Text, want)
	}
	if ch.English != "A stone monkey is born on the Mountain of Flowers and Fruit." {
		t.Errorf("english %q", ch.English)
	}
}

func TestParseChapterAddsTheStorytellersEnding(t *testing.T) {
	ch, err := ParseChapter("TITLE: 第二回\nTEXT: 猴王去学本事。\nENGLISH: The Monkey King goes to learn.")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(ch.Text, "\n"+StoryEnding) {
		t.Errorf("text %q doesn't end with the storyteller's words", ch.Text)
	}
	if _, err := ParseChapter("TITLE: 空\nENGLISH: nothing"); err == nil {
		t.Error("a reply without text should be an error")
	}
}
