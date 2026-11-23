package episode

import (
	"fmt"
	"hash/fnv"
	"math/rand"
	"strings"
	"unicode"

	"github.com/therealsiege/tuilaoshi/internal/model"
)

// A scene is a stretch of an episode's dialogue with one comprehension
// question: which English says what one of its lines says. The English is
// shown only once answered; the wrong choices are other lines of the same
// episode, so the question needs the line understood, not the topic.

const (
	sceneSize   = 20 // lines a scene spans
	contextSize = 2  // lines shown before the asked one
	choiceCount = 4
	minHan      = 5 // characters a line needs to be asked about
)

// Pay for scenes, once each, and for finishing all of an episode's.
const (
	ScenePay     = 15
	FinishCopper = 50
	FinishIngots = 1
	unsureMark   = "(listen with p)" // English for a line the transcript isn't sure of
)

// Scene is one stretch of an episode and its question.
type Scene struct {
	Index   int
	Lines   []model.DialogueLine // the context shown, the asked line last
	Asked   int                  // the asked line's index in the episode
	Choices []string             // English meanings, shuffled
	Right   int                  // the index of the right choice
}

// QuestID names a scene's question for the quest log.
func (e Episode) QuestID(scene int) string { return fmt.Sprintf("episode:%s:scene:%d", e.ID, scene) }

// FinishID names the reward for answering all the episode's scenes.
func (e Episode) FinishID() string { return "episode:" + e.ID }

// Scenes returns the episode's scenes that can carry a question.
func (e Episode) Scenes() []Scene {
	var scenes []Scene
	for start := 0; start < len(e.Lines); start += sceneSize {
		if s, ok := e.scene(len(scenes), start, min(start+sceneSize, len(e.Lines))); ok {
			scenes = append(scenes, s)
		}
	}
	return scenes
}

// scene builds the question for lines [start, end), the same every time.
func (e Episode) scene(index, start, end int) (Scene, bool) {
	h := fnv.New64a()
	fmt.Fprintf(h, "%s:%d", e.ID, start) //nolint:errcheck // hash writes never fail
	rng := rand.New(rand.NewSource(int64(h.Sum64())))
	var candidates []int
	for i := start; i < end; i++ {
		if askable(e.Lines[i]) {
			candidates = append(candidates, i)
		}
	}
	if len(candidates) == 0 {
		return Scene{}, false
	}
	asked := candidates[rng.Intn(len(candidates))]
	right := cleanEnglish(e.Lines[asked].English)
	choices := []string{right}
	seen := map[string]bool{strings.ToLower(right): true}
	for _, i := range rng.Perm(len(e.Lines)) {
		if len(choices) == choiceCount {
			break
		}
		other := cleanEnglish(e.Lines[i].English)
		if (i >= start && i < end) || !askable(e.Lines[i]) || seen[strings.ToLower(other)] {
			continue
		}
		seen[strings.ToLower(other)] = true
		choices = append(choices, other)
	}
	if len(choices) < 3 {
		return Scene{}, false
	}
	rng.Shuffle(len(choices), func(i, j int) { choices[i], choices[j] = choices[j], choices[i] })
	s := Scene{Index: index, Asked: asked, Choices: choices}
	for i, c := range choices {
		if c == right {
			s.Right = i
		}
	}
	s.Lines = e.Lines[max(asked-contextSize, start) : asked+1]
	return s, true
}

// askable reports whether a line is worth a question: enough Chinese, a
// sure English meaning.
func askable(l model.DialogueLine) bool {
	han := 0
	for _, r := range l.Chinese {
		if unicode.Is(unicode.Han, r) {
			han++
		}
	}
	en := strings.TrimSpace(l.English)
	return han >= minHan && en != "" && !strings.Contains(en, unsureMark)
}

func cleanEnglish(s string) string { return strings.TrimSpace(strings.ReplaceAll(s, unsureMark, "")) }
