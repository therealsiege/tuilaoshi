package journey

import "strings"

// The travel rescript (通关文牒) Tang Sanzang carried was stamped by every
// kingdom on the road. Here each kingdom's gate officer asks four questions
// in official Chinese; a typed answer passes when it names the fact asked
// for, so the stamp needs no Claude. Stamps are collected in the passport.

// Gate is a kingdom's gate: its town, officer, greeting and questions.
type Gate struct {
	Town            string         `yaml:"town"`
	Officer         string         `yaml:"officer"`
	OfficerEnglish  string         `yaml:"officer_english"`
	Greeting        string         `yaml:"greeting"`
	GreetingEnglish string         `yaml:"greeting_english"`
	Questions       []GateQuestion `yaml:"questions"`
}

// GateQuestion is one question: its kind (name, from, to, purpose) and how
// this officer asks it.
type GateQuestion struct {
	Kind    string `yaml:"kind"`
	Ask     string `yaml:"ask"`
	English string `yaml:"english"`
}

// facts are the words an answer of each kind must contain one of.
var facts = map[string][]string{
	"name":    {"三藏", "玄奘", "唐僧", "陈玄奘"},
	"from":    {"大唐", "东土", "长安", "唐朝"},
	"to":      {"西天", "天竺", "灵山", "雷音寺"},
	"purpose": {"取经", "求经", "真经", "拜佛"},
}

// sample is an answer of each kind, shown after a second miss.
var sample = map[string]string{
	"name":    "贫僧法名玄奘，人们叫我唐三藏。",
	"from":    "贫僧从东土大唐来。",
	"to":      "贫僧要去西天灵山。",
	"purpose": "贫僧去西天拜佛取经。",
}

// StampPay is the copper a kingdom's stamp brings (the court's gift).
const StampPay = 30

// Gates returns every kingdom's gate, west to east as in the content.
func Gates() []Gate { return content.Gates }

// GateFor returns a town's gate, if the town is a kingdom with one.
func GateFor(town string) (Gate, bool) {
	for _, g := range content.Gates {
		if g.Town == town {
			return g, true
		}
	}
	return Gate{}, false
}

// Accepts reports whether an answer gives the fact the question asks for.
func (q GateQuestion) Accepts(answer string) bool {
	for _, f := range facts[q.Kind] {
		if strings.Contains(answer, f) {
			return true
		}
	}
	return false
}

// Sample is an answer to the question that would pass.
func (q GateQuestion) Sample() string { return sample[q.Kind] }
