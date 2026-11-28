package mobile

import (
	"encoding/json"
	"errors"
	"math/rand"
	"slices"
	"strings"

	"github.com/therealsiege/tuilaoshi/internal/road"
)

// The grammar gate (语法关) before a region's lair, by the Mac's rules
// (internal/road): five rounds of laying a sentence's word tiles in
// order. On the Mac with Claude the last round is a sentence of the
// learner's own, which Claude grades; the phone has no grader yet, so it
// plays the gate as the Mac does without one, all five rounds tiles.
// CheckSentence keeps a typed sentence to the pattern's simple rules, the
// half of the typed round that needs no grader.

// gateRand seeds a gate's choice of sentences and its shuffles; tests fix it.
var gateRand = func() *rand.Rand { return rand.New(rand.NewSource(rand.Int63())) }

type gateExample struct {
	Chinese string `json:"chinese"`
	English string `json:"english"`
}

type gateTile struct {
	Word string `json:"word"`
	Laid bool   `json:"laid,omitempty"`
}

type roadGate struct {
	Region   string        `json:"region"`
	Pattern  string        `json:"pattern"`
	Pinyin   string        `json:"pinyin"`
	Explain  string        `json:"explain"`
	Examples []gateExample `json:"examples"`
	Round    int           `json:"round"`  // the round under way, from 0; rounds once passed
	Rounds   int           `json:"rounds"` // how many pass the gate
	Passed   bool          `json:"passed,omitempty"`
	Typed    bool          `json:"typed,omitempty"`   // this round is a sentence of the learner's own (not yet on the phone)
	English  string        `json:"english,omitempty"` // the tile round's sentence, in English
	Tiles    []gateTile    `json:"tiles"`             // in the order shown; lay them by position
	Built    string        `json:"built"`             // the tiles laid so far, in order
	Note     string        `json:"note,omitempty"`    // what came of the last try
	Laid     string        `json:"laid,omitempty"`    // the last tile laid: more, right or wrong
}

var laidWords = map[road.Laid]string{road.LaidMore: "more", road.LaidRight: "right", road.LaidWrong: "wrong"}

func (c *Core) toGate(laid road.Laid) *roadGate {
	g := c.gate
	if g == nil {
		return nil
	}
	out := &roadGate{Region: g.Region, Pattern: g.Pattern, Pinyin: g.Pinyin, Explain: g.Explain, Round: g.Round, Rounds: road.GateRounds,
		Passed: g.Passed(), Typed: g.Typed(), Note: g.Note, Laid: laidWords[laid], Examples: []gateExample{}, Tiles: []gateTile{}}
	for _, e := range g.Examples {
		out.Examples = append(out.Examples, gateExample{e.Chinese, e.English})
	}
	if out.Passed || out.Typed {
		return out
	}
	s := g.Sentence()
	out.English, out.Built = s.English, strings.Join(g.Chosen(g.Picked), "")
	for p, w := range g.Tiles {
		out.Tiles = append(out.Tiles, gateTile{s.Words[w], slices.Contains(g.Picked, p)})
	}
	return out
}

// openGate starts the region's gate, tiles only. The caller holds c.mu.
func (c *Core) openGate() *roadGate {
	g, ok := road.OpenGate(c.walk.Region.Region.ID, false, gateRand())
	if !ok {
		return nil
	}
	c.gate = &g
	return c.toGate(road.LaidNot)
}

// Gate is the grammar gate open, as JSON: its pattern, what it means and
// model sentences, the round under way with its tiles as shown, and what
// came of the last try.
func (c *Core) Gate() (string, error) {
	return c.onGate(func(*road.Gate) (road.Laid, error) { return road.LaidNot, nil })
}

// LayTile lays the shown tile at position p (from 0) next and returns the
// gate. When every tile is laid the sentence is checked: in order, the
// next round; out of order, the tiles are shuffled again. The fifth
// sentence laid passes the gate, for good.
func (c *Core) LayTile(p int) (string, error) {
	return c.onGate(func(g *road.Gate) (road.Laid, error) {
		laid := g.Lay(p)
		if laid == road.LaidRight && g.Passed() {
			return laid, c.walk.PassGate(g)
		}
		return laid, nil
	})
}

// UndoTile takes back the last tile laid and returns the gate.
func (c *Core) UndoTile() (string, error) {
	return c.onGate(func(g *road.Gate) (road.Laid, error) {
		g.Undo()
		return road.LaidNot, nil
	})
}

func (c *Core) onGate(do func(*road.Gate) (road.Laid, error)) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.gate == nil {
		return "", errors.New("no grammar gate is open")
	}
	laid, err := do(c.gate)
	if err != nil {
		return "", err
	}
	data, err := json.Marshal(c.toGate(laid))
	return string(data), err
}

type sentenceCheck struct {
	Uses     bool   `json:"uses"`
	Note     string `json:"note,omitempty"` // why it doesn't use the pattern yet
	Task     string `json:"task"`           // for the grader
	Criteria string `json:"criteria"`
}

// CheckSentence holds a typed sentence to the open gate's pattern by its
// simple rules, as the Mac does before Claude grades it, and returns the
// task and criteria a grader would judge it by. It passes nothing: the
// phone's gates are passed on tiles until grading comes through the
// game's server.
func (c *Core) CheckSentence(text string) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.gate == nil {
		return "", errors.New("no grammar gate is open")
	}
	note := c.gate.Unused(strings.TrimSpace(text))
	data, err := json.Marshal(sentenceCheck{note == "", note, c.gate.Task(), c.gate.Criteria()})
	return string(data), err
}
