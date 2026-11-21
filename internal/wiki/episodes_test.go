package wiki

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/therealsiege/tuilaoshi/internal/model"
)

func TestParseSubtitlesSRT(t *testing.T) {
	srt := "\ufeff1\r\n00:00:01,000 --> 00:00:03,000\r\n<i>孙悟空：师父，快走！</i>\r\nMaster, hurry!\r\n\r\n" +
		"2\r\n00:00:04,000 --> 00:00:06,000\r\n{\\an8}这里有妖怪。\r\n\r\n3\r\n00:00:07,000 --> 00:00:08,000\r\n(music)\r\n"
	ep := ParseSubtitles("ep01.srt", []byte(srt))
	want := []model.DialogueLine{
		{Speaker: "孙悟空", Chinese: "师父，快走！", English: "Master, hurry!"},
		{Chinese: "这里有妖怪。"},
	}
	if ep.Title != "ep01" || len(ep.Lines) != 2 || ep.Lines[0] != want[0] || ep.Lines[1] != want[1] {
		t.Errorf("episode %+v", ep)
	}
}

func TestPrivateEpisodesReadsOnlySubtitles(t *testing.T) {
	store := NewStore(t.TempDir())
	if eps, err := store.PrivateEpisodes(); err != nil || eps != nil {
		t.Fatalf("no folder: %v %v", eps, err)
	}
	dir := filepath.Join(store.Dir, PrivateEpisodesDir)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	for name, body := range map[string]string{"b.txt": "八戒：我饿了！\nI'm hungry!\n", "a.srt": "1\n00:00:01,000 --> 00:00:02,000\n你好\n", "notes.md": "# 不是字幕"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0644); err != nil {
			t.Fatal(err)
		}
	}
	eps, err := store.PrivateEpisodes()
	if err != nil || len(eps) != 2 || eps[0].Source != "a.srt" || eps[1].Lines[0].Speaker != "八戒" {
		t.Errorf("episodes %+v, err %v", eps, err)
	}
}
