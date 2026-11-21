package wiki

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/therealsiege/tuilaoshi/internal/model"
)

func TestParseSubtitlesSRT(t *testing.T) {
	srt := "\ufeff1\r\n00:00:01,000 --> 00:00:03,000\r\n<i>孙悟空：师父，快走！</i>\r\nMaster, hurry!\r\n\r\n" +
		"2\r\n00:00:04,000 --> 00:00:06,000\r\n{\\an8}这里有妖怪。\r\n\r\n3\r\n00:00:07,000 --> 00:00:08,000\r\n(music)\r\n"
	ep := ParseSubtitles("ep01.srt", []byte(srt))
	want := []model.DialogueLine{
		{Speaker: "孙悟空", Chinese: "师父，快走！", English: "Master, hurry!", Start: time.Second, End: 3 * time.Second},
		{Chinese: "这里有妖怪。", Start: 4 * time.Second, End: 6 * time.Second},
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

func TestEpisodeFolderMatchesEnglishByTimeAndFindsAudio(t *testing.T) {
	store := NewStore(t.TempDir())
	dir := filepath.Join(store.Dir, PrivateEpisodesDir, "ep01")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	zh := "1\n00:00:00,970 --> 00:00:04,490\n(Transcribed by TurboScribe. Go Unlimited to remove this message.) 猴王出世\n\n" +
		"2\n00:00:09,670 --> 00:00:12,510\n在很久很久以前\n\n3\n00:00:13,710 --> 00:00:16,610\n古称神州的中国大地上\n"
	en := "1\n00:00:02,130 --> 00:00:05,770\n(Transcribed by TurboScribe. Go Unlimited to remove this message.) The Monkey King appears.\n\n" +
		"2\n00:00:10,150 --> 00:00:12,410\nA long, long time ago,\n\n3\n00:00:13,670 --> 00:00:15,190\nin the ancient China known as Shenzhou,\n\n" +
		"4\n00:09:00,000 --> 00:09:01,000\nstray line far from any Chinese\n"
	for name, body := range map[string]string{"ep01.srt": en, "ep01.srt copy": zh, "ep01.mp3": "ID3", "notes.md": "x"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0644); err != nil {
			t.Fatal(err)
		}
	}
	eps, err := store.PrivateEpisodes()
	if err != nil || len(eps) != 1 {
		t.Fatalf("episodes %+v, err %v", eps, err)
	}
	ep := eps[0]
	if ep.Title != "ep01" || filepath.Base(ep.Audio) != "ep01.mp3" || len(ep.Lines) != 3 {
		t.Fatalf("episode %q audio %q, %d lines", ep.Title, ep.Audio, len(ep.Lines))
	}
	want := []string{"The Monkey King appears.", "A long, long time ago,", "in the ancient China known as Shenzhou,"}
	for i, l := range ep.Lines {
		if l.English != want[i] {
			t.Errorf("line %d %q: English %q, want %q", i, l.Chinese, l.English, want[i])
		}
	}
	if ep.Lines[0].Chinese != "猴王出世" || ep.Lines[1].Start != 9670*time.Millisecond {
		t.Errorf("watermark or timing wrong: %+v", ep.Lines[:2])
	}
}
