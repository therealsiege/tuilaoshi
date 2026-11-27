package mobile

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestTheCoreReviewsACard(t *testing.T) {
	wikiDir, progress := t.TempDir(), t.TempDir()
	page := "---\ntype: character\nlevel: HSK1\nreview_interval: 1\nnext_review: 2020-01-02\n---\n\n# 书\n\n- **Character**: 书\n- **Pinyin**: shū\n- **Meaning**: book\n"
	if err := os.MkdirAll(filepath.Join(wikiDir, "characters"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(wikiDir, "characters", "shu1.md"), []byte(page), 0644); err != nil {
		t.Fatal(err)
	}
	c, err := Open(wikiDir, progress)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close() //nolint:errcheck // test
	if n, err := c.DueCount(); err != nil || n != 1 {
		t.Fatalf("due %d %v", n, err)
	}
	list, err := c.DueCards(10)
	var cards []card
	if err != nil || json.Unmarshal([]byte(list), &cards) != nil || len(cards) != 1 || cards[0].Chinese != "书" {
		t.Fatalf("due cards %s %v", list, err)
	}
	if _, err := c.Review(cards[0].ID, 2); err != nil {
		t.Fatal(err)
	}
	if n, _ := c.DueCount(); n != 0 { //nolint:errcheck // checked above
		t.Errorf("still due after a good review: %d", n)
	}
	if _, err := os.Stat(filepath.Join(progress, "progress.db")); err != nil {
		t.Errorf("progress not kept in the progress folder: %v", err)
	}
	if _, err := os.Stat(filepath.Join(wikiDir, "progress.db")); !os.IsNotExist(err) {
		t.Error("the wiki (read-only in the app bundle) got a progress file")
	}
}

func TestCampGathersTheNumbers(t *testing.T) {
	c, err := Open(t.TempDir(), t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close() //nolint:errcheck // test
	data, err := c.Camp()
	if err != nil {
		t.Fatal(err)
	}
	var got struct {
		Party []string `json:"party"`
		Term  string   `json:"term"`
		Stage *struct {
			Number int    `json:"number"`
			Region string `json:"region"`
		} `json:"stage"`
	}
	if err := json.Unmarshal([]byte(data), &got); err != nil {
		t.Fatal(err)
	}
	if got.Term == "" || got.Stage == nil || got.Stage.Number != 1 || got.Stage.Region == "" {
		t.Errorf("camp = %s; want today's term and stage 1 with its region", data)
	}
	if len(got.Party) != 1 || got.Party[0] != "唐僧" {
		t.Errorf("party = %v; a new pilgrim travels with Tang Sanzang alone", got.Party)
	}
}

// fixtureWiki writes a one-card wiki, the card due long ago.
func fixtureWiki(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	page := "---\ntype: character\nlevel: HSK1\nreview_interval: 1\nnext_review: 2020-01-02\n---\n\n# 书\n\n- **Character**: 书\n- **Pinyin**: shū\n- **Meaning**: book\n"
	if err := os.MkdirAll(filepath.Join(dir, "characters"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "characters", "shu1.md"), []byte(page), 0644); err != nil {
		t.Fatal(err)
	}
	return dir
}

// A review on the phone reaches the Mac through the shared folder, a
// review on the Mac reaches the phone, and the phone's progress survives
// closing and reopening.
func TestProgressSyncsBetweenPhoneAndMac(t *testing.T) {
	wiki, shared := fixtureWiki(t), t.TempDir()
	phoneDir, macDir := t.TempDir(), t.TempDir()
	open := func(progress string) *Core {
		c, err := Open(wiki, progress)
		if err != nil {
			t.Fatal(err)
		}
		return c
	}
	due := func(c *Core) int {
		n, err := c.DueCount()
		if err != nil {
			t.Fatal(err)
		}
		return n
	}

	phone := open(phoneDir)
	if _, err := phone.Review("characters/shu1.md", 2); err != nil {
		t.Fatal(err)
	}
	if err := phone.Close(); err != nil {
		t.Fatal(err)
	}
	phone = open(phoneDir)
	if due(phone) != 0 {
		t.Fatal("the phone's review didn't survive reopening")
	}
	if err := phone.Sync(shared, "iphone"); err != nil {
		t.Fatal(err)
	}

	mac := open(macDir)
	defer mac.Close() //nolint:errcheck // test
	if due(mac) != 1 {
		t.Fatal("the Mac should start with the card due")
	}
	if err := mac.Sync(shared, "mac"); err != nil {
		t.Fatal(err)
	}
	if due(mac) != 0 {
		t.Error("the phone's review didn't reach the Mac")
	}

	// Again on the Mac later the same day: the card comes back to the
	// phone due on the Mac's schedule, not the phone's. Review times are
	// kept to the second, so the Mac answers a second later.
	time.Sleep(1100 * time.Millisecond)
	macCard, err := mac.Review("characters/shu1.md", 0)
	if err != nil {
		t.Fatal(err)
	}
	if err := mac.Sync(shared, "mac"); err != nil {
		t.Fatal(err)
	}
	if err := phone.Sync(shared, "iphone"); err != nil {
		t.Fatal(err)
	}
	cards, err := phone.store.DrillableCards()
	if err != nil {
		t.Fatal(err)
	}
	var want struct{ Next string }
	if err := json.Unmarshal([]byte(macCard), &want); err != nil {
		t.Fatal(err)
	}
	for _, c := range cards {
		if c.ID == "characters/shu1.md" && toCard(c).Next != want.Next {
			t.Errorf("the phone has the card next due %s; the Mac's review set %s", toCard(c).Next, want.Next)
		}
	}
	if err := phone.Close(); err != nil {
		t.Fatal(err)
	}
}

// A finished session counts toward the streak; an empty one writes nothing.
func TestFinishSessionCountsTowardTheStreak(t *testing.T) {
	c, err := Open(fixtureWiki(t), t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close() //nolint:errcheck // test
	if err := c.FinishSession("review", time.Now().Add(-time.Minute).Unix(), "[]"); err != nil {
		t.Fatal(err)
	}
	if c.store.Streak() != 0 {
		t.Fatal("an empty session shouldn't count")
	}
	if err := c.FinishSession("review", time.Now().Add(-time.Minute).Unix(), `[{"id":"characters/shu1.md","correct":true,"rating":2}]`); err != nil {
		t.Fatal(err)
	}
	if c.store.Streak() != 1 {
		t.Errorf("streak = %d after a session today, want 1", c.store.Streak())
	}
	if err := c.FinishSession("battle", 0, `[{"id":"x","correct":true}]`); err == nil {
		t.Error("an unknown session kind should be refused")
	}
}

func TestShadowScoresByCharacter(t *testing.T) {
	data, err := Shadow("我爱你。", "我爱")
	if err != nil {
		t.Fatal(err)
	}
	var got struct {
		Chars []struct {
			Char  string `json:"char"`
			Heard bool   `json:"heard"`
		} `json:"chars"`
		Percent int `json:"percent"`
	}
	if err := json.Unmarshal([]byte(data), &got); err != nil {
		t.Fatal(err)
	}
	if len(got.Chars) != 3 || got.Chars[0].Char != "我" || !got.Chars[1].Heard || got.Chars[2].Heard || got.Percent != 66 {
		t.Errorf("shadow = %s; want 我 爱 heard, 你 missed, 66%%", data)
	}
}

// A reading opens glossed word by word; marking a deck word sends it back
// for review, and a sentence card joins the decks.
func TestReadingGlossesMarksAndKeepsSentences(t *testing.T) {
	dir := fixtureWiki(t)
	reading := "---\ntype: reading\nlevel: HSK1\n---\n\n# 我的书\n\n## Text\n\n这是我的书。\n\n## Vocabulary\n\n| Word | Pinyin | Meaning |\n|------|--------|---------|\n| 这是 | zhè shì | this is |\n"
	if err := os.MkdirAll(filepath.Join(dir, "readings"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "readings", "shu.md"), []byte(reading), 0644); err != nil {
		t.Fatal(err)
	}
	c, err := Open(dir, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close() //nolint:errcheck // test
	list, err := c.Readings()
	if err != nil || !strings.Contains(list, `"page:readings/shu.md"`) {
		t.Fatalf("readings = %s, %v", list, err)
	}
	data, err := c.Reading("page:readings/shu.md")
	if err != nil {
		t.Fatal(err)
	}
	var got struct {
		Title string `json:"title"`
		Lines []struct {
			Tokens []struct {
				Text    string `json:"text"`
				English string `json:"english"`
				InDeck  bool   `json:"inDeck"`
			} `json:"tokens"`
		} `json:"lines"`
	}
	if err := json.Unmarshal([]byte(data), &got); err != nil {
		t.Fatal(err)
	}
	glossed := map[string]string{}
	inDeck := map[string]bool{}
	for _, tok := range got.Lines[0].Tokens {
		glossed[tok.Text], inDeck[tok.Text] = tok.English, tok.InDeck
	}
	if glossed["这是"] != "this is" || glossed["书"] != "book" || !inDeck["书"] {
		t.Errorf("tokens = %+v; want 这是 from the reading's vocabulary and 书 from the decks", got.Lines[0].Tokens)
	}
	if ok, err := c.MarkUnknown("书"); err != nil || !ok {
		t.Errorf("MarkUnknown(书) = %v, %v", ok, err)
	}
	if ok, err := c.MarkUnknown("这是"); err != nil || ok {
		t.Errorf("MarkUnknown(这是) = %v, %v; it isn't in the decks", ok, err)
	}
	if ok, err := c.AddSentenceCard("这是", "zhè shì", "this is", "这是我的书。"); err != nil || !ok {
		t.Errorf("AddSentenceCard = %v, %v", ok, err)
	}
	cards, err := c.DueCards(50)
	if err != nil || !strings.Contains(cards, "这是") {
		t.Errorf("the sentence card should be due: %s, %v", cards, err)
	}
}
