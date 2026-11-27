package mobile

import (
	"encoding/json"
	"os"
	"path/filepath"
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
