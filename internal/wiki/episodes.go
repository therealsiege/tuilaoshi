package wiki

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/therealsiege/tuilaoshi/internal/model"
)

// PrivateEpisodesDir holds subtitle files the learner adds from shows they
// own. It is git-ignored: the files stay on the learner's machine. A file
// is an episode; so is a folder holding Chinese subtitles, optionally
// English subtitles and the episode's audio.
const PrivateEpisodesDir = "private/episodes"

var audioExts = map[string]bool{".mp3": true, ".m4a": true, ".wav": true, ".aac": true, ".ogg": true}

// PrivateEpisodes reads the learner's episodes, sorted by name. A missing
// folder means none.
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
		path := filepath.Join(dir, e.Name())
		var ep model.Episode
		switch ext := strings.ToLower(filepath.Ext(e.Name())); {
		case e.IsDir():
			ep, err = readEpisodeDir(path)
		case ext == ".srt" || ext == ".txt":
			var data []byte
			data, err = os.ReadFile(path)
			ep = ParseSubtitles(e.Name(), data)
		default:
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("read episode %s: %w", e.Name(), err)
		}
		if len(ep.Lines) > 0 {
			episodes = append(episodes, ep)
		}
	}
	sort.Slice(episodes, func(i, j int) bool { return episodes[i].Source < episodes[j].Source })
	return episodes, nil
}

// readEpisodeDir builds one episode from a folder: the subtitle file with
// the most Chinese is the dialogue, another subtitle file supplies English
// by timing, and an audio file plays each line.
func readEpisodeDir(dir string) (model.Episode, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return model.Episode{}, err
	}
	type subs struct {
		name string
		data []byte
		han  int
	}
	var files []subs
	audio := ""
	for _, e := range entries {
		if e.IsDir() || strings.HasPrefix(e.Name(), ".") {
			continue
		}
		path := filepath.Join(dir, e.Name())
		if audioExts[strings.ToLower(filepath.Ext(e.Name()))] {
			audio = path
			continue
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return model.Episode{}, err
		}
		if cueTiming.Match(data) { // any subtitle file, whatever its extension ("x.srt copy")
			files = append(files, subs{e.Name(), data, hanCount(string(data))})
		}
	}
	if len(files) == 0 {
		return model.Episode{}, nil
	}
	sort.Slice(files, func(i, j int) bool { return files[i].han > files[j].han })
	ep := ParseSubtitles(filepath.Base(dir), files[0].data)
	ep.Source, ep.Audio = filepath.Base(dir), audio
	for _, f := range files[1:] {
		if f.han == 0 {
			matchEnglish(ep.Lines, parseCues(f.data))
		}
	}
	return ep, nil
}

var (
	cueTiming  = regexp.MustCompile(`-->`)
	timing     = regexp.MustCompile(`(\d+):(\d+):(\d+)[,.](\d+)\s*-->\s*(\d+):(\d+):(\d+)[,.](\d+)`)
	markupTags = regexp.MustCompile(`<[^>]*>|\{[^}]*\}|\(Transcribed by [^)]*\)`)
)

// cue is one subtitle: when it shows and its text lines.
type cue struct {
	start, end time.Duration
	lines      []string
}

// parseCues reads the cues of an .srt file.
func parseCues(data []byte) []cue {
	var cues []cue
	var cur *cue
	for _, raw := range strings.Split(strings.ReplaceAll(strings.TrimPrefix(string(data), "\ufeff"), "\r\n", "\n"), "\n") {
		line := strings.TrimSpace(markupTags.ReplaceAllString(raw, ""))
		if m := timing.FindStringSubmatch(line); m != nil {
			cues = append(cues, cue{start: clockTime(m[1:5]), end: clockTime(m[5:9])})
			cur = &cues[len(cues)-1]
			continue
		}
		if line == "" {
			cur = nil
			continue
		}
		if cur != nil { // text after a timing line (cue numbers come before it)
			cur.lines = append(cur.lines, line)
		}
	}
	return cues
}

// clockTime reads [hours minutes seconds milliseconds] from a subtitle
// timing; the timing pattern guarantees they are digits.
func clockTime(hmsMs []string) time.Duration {
	units := []time.Duration{time.Hour, time.Minute, time.Second, time.Millisecond}
	var d time.Duration
	for i, s := range hmsMs {
		if n, err := strconv.Atoi(s); err == nil {
			d += time.Duration(n) * units[i]
		}
	}
	return d
}

// ParseSubtitles reads an .srt file (or a .txt with one line per line) into
// an episode. Lines with Chinese become spoken lines, timed by their
// subtitle; a line without Chinese right after one, in the same subtitle,
// is taken as its English (bilingual subtitles). A "Speaker：line" prefix
// names the speaker.
func ParseSubtitles(name string, data []byte) model.Episode {
	ep := model.Episode{Title: strings.TrimSuffix(name, filepath.Ext(name)), Source: name}
	cues := parseCues(data)
	if len(cues) == 0 { // plain text: one cue per paragraph
		for _, para := range strings.Split(strings.ReplaceAll(strings.TrimPrefix(string(data), "\ufeff"), "\r\n", "\n"), "\n\n") {
			var c cue
			for _, l := range strings.Split(para, "\n") {
				if l = strings.TrimSpace(markupTags.ReplaceAllString(l, "")); l != "" {
					c.lines = append(c.lines, l)
				}
			}
			cues = append(cues, c)
		}
	}
	for _, c := range cues {
		translatable := false // the previous line was Chinese in this subtitle
		for _, line := range c.lines {
			if !hasHan(line) {
				if translatable {
					ep.Lines[len(ep.Lines)-1].English = line
				}
				translatable = false
				continue
			}
			translatable = true
			dl := model.DialogueLine{Chinese: line, Start: c.start, End: c.end}
			if who, said, ok := strings.Cut(line, "："); ok && len([]rune(who)) <= 6 && hasHan(said) {
				dl.Speaker, dl.Chinese = strings.TrimSpace(who), strings.TrimSpace(said)
			}
			ep.Lines = append(ep.Lines, dl)
		}
	}
	return ep
}

// matchEnglish gives each English subtitle to the Chinese line it overlaps
// most in time (the two files are timed separately), or the nearest line.
func matchEnglish(lines []model.DialogueLine, english []cue) {
	if len(lines) == 0 {
		return
	}
	for _, en := range english {
		text := strings.Join(en.lines, " ")
		if text == "" {
			continue
		}
		best, bestScore := 0, time.Duration(-1<<62)
		for i, l := range lines {
			score := min(l.End, en.end) - max(l.Start, en.start) // overlap; negative is the gap
			if score > bestScore {
				best, bestScore = i, score
			}
		}
		if bestScore < -2*time.Second {
			continue // nothing near it
		}
		if lines[best].English == "" {
			lines[best].English = text
		} else {
			lines[best].English += " " + text
		}
	}
}

func hasHan(s string) bool { return hanCount(s) > 0 }

func hanCount(s string) int {
	n := 0
	for _, r := range s {
		if unicode.Is(unicode.Han, r) {
			n++
		}
	}
	return n
}
