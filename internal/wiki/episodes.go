package wiki

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"unicode"

	"github.com/therealsiege/tuilaoshi/internal/model"
)

// PrivateEpisodesDir holds subtitle files the learner adds from shows they
// own. It is git-ignored: the files stay on the learner's machine.
const PrivateEpisodesDir = "private/episodes"

// PrivateEpisodes reads the learner's .srt and .txt subtitle files as
// episodes, sorted by file name. A missing folder means none.
func (s *Store) PrivateEpisodes() ([]model.Episode, error) {
	dir := filepath.Join(s.Dir, PrivateEpisodesDir)
	entries, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read private episodes: %w", err)
	}
	var episodes []model.Episode
	for _, e := range entries {
		ext := strings.ToLower(filepath.Ext(e.Name()))
		if e.IsDir() || (ext != ".srt" && ext != ".txt") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", e.Name(), err)
		}
		ep := ParseSubtitles(e.Name(), data)
		if len(ep.Lines) > 0 {
			episodes = append(episodes, ep)
		}
	}
	sort.Slice(episodes, func(i, j int) bool { return episodes[i].Source < episodes[j].Source })
	return episodes, nil
}

var (
	cueNumber  = regexp.MustCompile(`^\d+$`)
	cueTiming  = regexp.MustCompile(`-->`)
	markupTags = regexp.MustCompile(`<[^>]*>|\{[^}]*\}`)
)

// ParseSubtitles reads an .srt file (or a .txt with one line per line) into
// an episode. Lines with Chinese become spoken lines; a line without Chinese
// right after one, in the same subtitle, is taken as its English (bilingual
// subtitles). A "Speaker：line" prefix names the speaker.
func ParseSubtitles(name string, data []byte) model.Episode {
	ep := model.Episode{Title: strings.TrimSuffix(name, filepath.Ext(name)), Source: name}
	text := strings.ReplaceAll(strings.TrimPrefix(string(data), "\ufeff"), "\r\n", "\n")
	translatable := false // the previous line was Chinese in the same subtitle
	for _, raw := range strings.Split(text, "\n") {
		line := strings.TrimSpace(markupTags.ReplaceAllString(raw, ""))
		if line == "" || cueNumber.MatchString(line) || cueTiming.MatchString(line) {
			translatable = false
			continue
		}
		if !hasHan(line) {
			if translatable {
				ep.Lines[len(ep.Lines)-1].English = line
			}
			translatable = false
			continue
		}
		translatable = true
		dl := model.DialogueLine{Chinese: line}
		if who, said, ok := strings.Cut(line, "："); ok && len([]rune(who)) <= 6 && hasHan(said) {
			dl.Speaker, dl.Chinese = strings.TrimSpace(who), strings.TrimSpace(said)
		}
		ep.Lines = append(ep.Lines, dl)
	}
	return ep
}

func hasHan(s string) bool {
	for _, r := range s {
		if unicode.Is(unicode.Han, r) {
			return true
		}
	}
	return false
}
