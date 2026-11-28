package mobile

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/therealsiege/tuilaoshi/internal/claude/claudetest"
	"github.com/therealsiege/tuilaoshi/internal/journey"
)

// openCore opens a core on a wiki holding the character 书.
func openCore(t *testing.T) *Core {
	t.Helper()
	wikiDir := t.TempDir()
	page := "---\ntype: character\nlevel: HSK1\nreview_interval: 1\nnext_review: 2020-01-02\n---\n\n# 书\n\n- **Character**: 书\n- **Pinyin**: shū\n- **Meaning**: book\n"
	if err := os.MkdirAll(filepath.Join(wikiDir, "characters"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(wikiDir, "characters", "shu1.md"), []byte(page), 0644); err != nil {
		t.Fatal(err)
	}
	c, err := Open(wikiDir, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { c.Close() }) //nolint:errcheck // test
	return c
}

// decode reads a facade call's JSON, failing the test on its error:
// decode[T](t)(c.Call()).
func decode[T any](t *testing.T) func(string, error) T {
	return func(data string, err error) T {
		t.Helper()
		var v T
		if err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal([]byte(data), &v); err != nil {
			t.Fatalf("%s: %v", data, err)
		}
		return v
	}
}

func TestWriteDiaryThroughTheServer(t *testing.T) {
	c := openCore(t)
	srv := claudetest.New(t, "SCORE: 0\nFEEDBACK: Write two sentences.", "SCORE: 2\nFEEDBACK: 很好。")
	type entry struct {
		Day   int    `json:"day"`
		Score int    `json:"score"`
		Kept  bool   `json:"kept"`
		Text  string `json:"text"`
	}
	e := decode[entry](t)(c.WriteDiary(srv.URL, "tok", "今天很好。"))
	if e.Kept || e.Score != 0 {
		t.Errorf("a failing entry %+v", e)
	}
	t0, _, _ := journey.LoadTrail(c.store, nil) //nolint:errcheck // checked below
	e = decode[entry](t)(c.WriteDiary(srv.URL, "tok", "今天我们走了很远。我很累，可是很高兴。"))
	if !e.Kept || e.Score != 2 || e.Day != 1 {
		t.Errorf("a passing entry %+v", e)
	}
	t1, _, err := journey.LoadTrail(c.store, nil)
	if err != nil || t1.Morale[journey.Master] <= t0.Morale[journey.Master] && t0.Morale[journey.Master] < 100 {
		t.Errorf("morale %d -> %d, %v", t0.Morale[journey.Master], t1.Morale[journey.Master], err)
	}
	diary := decode[[]entry](t)(c.Diary())
	if len(diary) != 1 || diary[0].Score != 2 {
		t.Errorf("diary %+v", diary)
	}
	sys, _ := srv.Requests()[1].Body["messages"].([]any)
	if !strings.Contains(strings.TrimSpace(jsonOf(sys)), "第1天") {
		t.Errorf("graded without the Mac's task: %s", jsonOf(sys))
	}
	if _, err := c.WriteDiary(srv.URL, "tok", "又写了。"); err == nil {
		t.Error("a second entry the same day was taken")
	}
}

func jsonOf(v any) string {
	data, _ := json.Marshal(v)
	return string(data)
}

func TestClaudeCallsNeedSigningIn(t *testing.T) {
	c := openCore(t)
	if _, err := c.WriteDiary("https://example.invalid", "", "今天很好。"); err == nil || !strings.Contains(err.Error(), "sign in") {
		t.Errorf("signed out: %v", err)
	}
}

func TestClaudeGazetteFallsBackToTheClerk(t *testing.T) {
	type issue struct {
		Text   string `json:"text"`
		ByHand bool   `json:"byHand"`
		Missed string `json:"missed"`
	}
	srv := claudetest.New(t, "TEXT: 本报讯：一行继续西行。\nENGLISH: Our report.")
	srv.Refuse(http.StatusTooManyRequests)
	g := decode[issue](t)(openCore(t).ClaudeGazette(srv.URL, "tok"))
	if !g.ByHand || g.Text == "" || !strings.Contains(g.Missed, "Claude limit") {
		t.Errorf("over the cap the clerk sets it: %+v", g)
	}
	srv.Refuse(0)
	c := openCore(t)
	g = decode[issue](t)(c.ClaudeGazette(srv.URL, "tok"))
	if g.ByHand || g.Text != "本报讯：一行继续西行。" || g.Missed != "" {
		t.Errorf("Claude's issue %+v", g)
	}
	if g = decode[issue](t)(c.Gazette()); g.Text != "本报讯：一行继续西行。" {
		t.Errorf("the issue wasn't kept: %+v", g)
	}
}

func TestTheStorytellerOnThePhone(t *testing.T) {
	c := openCore(t)
	srv := claudetest.New(t, "TITLE: 石猴出世\nTEXT: 花果山上有一块石头。石头里出来一只猴子。\nENGLISH: A monkey is born.\nQ: 猴子从哪里来？ | 石头 | 水 | 天")
	type shelf struct {
		Next     int  `json:"next"`
		CanTell  bool `json:"canTell"`
		Chapters []struct {
			ID    string `json:"id"`
			Title string `json:"title"`
		} `json:"chapters"`
	}
	s := decode[shelf](t)(c.Storyteller())
	if s.Next != 1 || !s.CanTell || len(s.Chapters) != 0 {
		t.Fatalf("a new shelf %+v", s)
	}
	told := decode[toldChapter](t)(c.TellChapter(srv.URL, "tok"))
	if told.ID != "story:1" || told.Title != "石猴出世" {
		t.Errorf("told %+v", told)
	}
	if words, _ := srv.Requests()[0].Body["messages"].([]any); !strings.Contains(jsonOf(words), "第1回") {
		t.Errorf("asked %s", jsonOf(words))
	}
	s = decode[shelf](t)(c.Storyteller())
	if s.Next != 2 || s.CanTell || len(s.Chapters) != 1 {
		t.Errorf("after telling %+v", s)
	}
	if _, err := c.TellChapter(srv.URL, "tok"); err == nil || len(srv.Requests()) != 1 {
		t.Error("a second chapter the same day was told")
	}
	r := decode[struct {
		Title string        `json:"title"`
		Lines []readingLine `json:"lines"`
	}](t)(c.Reading(told.ID))
	if r.Title != "说书人 第1回 石猴出世" || len(r.Lines) != 3 || r.Lines[0].English == "" {
		t.Errorf("reading %+v", r)
	}
}

func TestGradeAnswerAsksClaude(t *testing.T) {
	c := openCore(t)
	srv := claudetest.New(t, "CORRECT: true\nSCORE: 3\nFEEDBACK: 对，可是写错了一点。")
	type grade struct {
		Correct  bool   `json:"correct"`
		Score    int    `json:"score"`
		Rating   int    `json:"rating"`
		Feedback string `json:"feedback"`
	}
	g := decode[grade](t)(c.GradeAnswer(srv.URL, "tok", "characters/shu1.md", "書", "production"))
	if !g.Correct || g.Score != 3 || g.Rating != 1 || g.Feedback == "" {
		t.Errorf("graded %+v", g)
	}
	if sys, _ := srv.Requests()[0].Body["messages"].([]any); !strings.Contains(jsonOf(sys), "書") {
		t.Errorf("the answer wasn't sent: %s", jsonOf(sys))
	}
	if _, err := c.GradeAnswer(srv.URL, "tok", "characters/none.md", "x", "production"); err == nil {
		t.Error("an unknown card was graded")
	}
	if _, err := c.GradeAnswer(srv.URL, "tok", "characters/shu1.md", "x", "review"); err == nil {
		t.Error("an unknown mode was graded")
	}
	srv.Refuse(http.StatusUnauthorized)
	if _, err := c.GradeAnswer(srv.URL, "tok", "characters/shu1.md", "書", "production"); err == nil || !strings.Contains(err.Error(), "sign in again") {
		t.Errorf("a refused token: %v", err)
	}
}
