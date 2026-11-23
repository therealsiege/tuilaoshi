package episode

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/therealsiege/tuilaoshi/internal/journey"
	"github.com/therealsiege/tuilaoshi/internal/model"
	"github.com/therealsiege/tuilaoshi/internal/wiki"
)

// writeEpisode stores a made-up subtitle file as a private episode.
func writeEpisode(t *testing.T, dir, name string, lines ...string) {
	t.Helper()
	path := filepath.Join(dir, wiki.PrivateEpisodesDir, name+".srt")
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	var srt string
	for i, l := range lines {
		srt += fmt.Sprintf("%d\n00:00:%02d,000 --> 00:00:%02d,500\n%s\n\n", i+1, i, i, l)
	}
	if err := os.WriteFile(path, []byte(srt), 0644); err != nil {
		t.Fatal(err)
	}
}

func TestEpisodesAreNumberedAndPlacedOnTheRoad(t *testing.T) {
	dir := t.TempDir()
	writeEpisode(t, dir, "Show 98 02 Second", "第二集")
	writeEpisode(t, dir, "Show 98 01 First", "第一集")
	eps, err := Load(wiki.NewStore(dir))
	if err != nil || len(eps) != 2 {
		t.Fatalf("%d episodes, %v", len(eps), err)
	}
	if eps[0].Number != 1 || eps[1].Number != 2 || eps[0].Stage != 1 || eps[1].Stage != journey.EpisodeStage(2) {
		t.Errorf("episodes %d@%d, %d@%d", eps[0].Number, eps[0].Stage, eps[1].Number, eps[1].Stage)
	}
	if got := Opened(eps, 1); len(got) != 1 || got[0].Number != 1 {
		t.Errorf("open at tribulation 1: %v", got)
	}
	if got := Opened(eps, 0); len(got) != 2 {
		t.Error("a finished pilgrimage should open every episode")
	}
}

func TestThemeIsTheEpisodesUnlearnedWordsMostUsedFirst(t *testing.T) {
	e := Episode{Episode: model.Episode{Lines: []model.DialogueLine{
		{Chinese: "大王要兵器"}, {Chinese: "兵器在龙宫"}, {Chinese: "龙宫很远"}, {Chinese: "大王好"},
	}}}
	cards := []model.Card{
		{ID: "a", Chinese: "龙宫"},
		{ID: "b", Chinese: "兵器"},
		{ID: "c", Chinese: "大王", ReviewInterval: 5}, // learned
		{ID: "d", Chinese: "远"},                     // one character: too common to theme
		{ID: "e", Chinese: "飞机"},                    // not in the episode
	}
	words := e.Words(cards)
	if len(words) != 2 || words[0].ID != "a" && words[0].ID != "b" {
		t.Fatalf("theme %v, want 龙宫 and 兵器", words)
	}
	due := []model.Card{{ID: "x"}, cards[1], {ID: "y"}, cards[0]}
	got := Ahead(due, []Episode{e}, cards)
	if got[0].ID != "b" || got[1].ID != "a" || got[2].ID != "x" || got[3].ID != "y" {
		t.Errorf("review order %v, want the episode's words first", got)
	}
}

func TestScenesAskAboutSureLinesWithOtherLinesAsWrongAnswers(t *testing.T) {
	var lines []model.DialogueLine
	for i := 0; i < 45; i++ {
		lines = append(lines, model.DialogueLine{Chinese: fmt.Sprintf("这是第%d句很长的台词", i), English: fmt.Sprintf("This is line %d.", i)})
	}
	lines[3].English += " (listen with p)"
	e := Episode{ID: "ep", Episode: model.Episode{Lines: lines}}
	scenes := e.Scenes()
	if len(scenes) != 3 {
		t.Fatalf("%d scenes from 45 lines, want 3", len(scenes))
	}
	for _, s := range scenes {
		asked := s.Lines[len(s.Lines)-1]
		if s.Choices[s.Right] != asked.English || strings.Contains(asked.English, "listen") {
			t.Errorf("scene %d asks %q with right choice %q", s.Index, asked.English, s.Choices[s.Right])
		}
		if len(s.Choices) != choiceCount || len(s.Lines) > contextSize+1 {
			t.Errorf("scene %d: %d choices, %d lines of context", s.Index, len(s.Choices), len(s.Lines))
		}
		for i, c := range s.Choices {
			for _, l := range lines[s.Index*sceneSize : min((s.Index+1)*sceneSize, len(lines))] {
				if i != s.Right && c == l.English {
					t.Errorf("scene %d offers its own line %q as a wrong answer", s.Index, c)
				}
			}
		}
	}
	if again := e.Scenes(); again[1].Asked != scenes[1].Asked || again[1].Right != scenes[1].Right {
		t.Error("scenes differ between calls")
	}
}
