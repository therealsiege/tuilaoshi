package journey

import (
	"fmt"
	"testing"
)

func TestCalculateRank(t *testing.T) {
	tests := []struct {
		name        string
		mastered    int
		wantName    string
		wantEnglish string
		wantLevel   int
		wantCurrent int
		wantReq     int
		wantNext    string
		wantNextEn  string
	}{
		// Rank boundaries — exact thresholds
		{
			name:        "zero mastered is Mortal",
			mastered:    0,
			wantName:    "凡人",
			wantEnglish: "Mortal",
			wantLevel:   0,
			wantCurrent: 0,
			wantReq:     100,
			wantNext:    "练气",
			wantNextEn:  "Qi Refining",
		},
		{
			name:        "99 mastered still Mortal",
			mastered:    99,
			wantName:    "凡人",
			wantEnglish: "Mortal",
			wantLevel:   0,
			wantCurrent: 99,
			wantReq:     100,
			wantNext:    "练气",
			wantNextEn:  "Qi Refining",
		},
		{
			name:        "exactly 100 reaches Qi Refining",
			mastered:    100,
			wantName:    "练气",
			wantEnglish: "Qi Refining",
			wantLevel:   1,
			wantCurrent: 0,
			wantReq:     300,
			wantNext:    "筑基",
			wantNextEn:  "Foundation Building",
		},
		{
			name:        "150 is mid-Qi Refining",
			mastered:    150,
			wantName:    "练气",
			wantEnglish: "Qi Refining",
			wantLevel:   1,
			wantCurrent: 50,
			wantReq:     300,
			wantNext:    "筑基",
			wantNextEn:  "Foundation Building",
		},
		{
			name:        "exactly 300 reaches Foundation Building",
			mastered:    300,
			wantName:    "筑基",
			wantEnglish: "Foundation Building",
			wantLevel:   2,
			wantCurrent: 0,
			wantReq:     600,
			wantNext:    "金丹",
			wantNextEn:  "Golden Core",
		},
		{
			name:        "exactly 600 reaches Golden Core",
			mastered:    600,
			wantName:    "金丹",
			wantEnglish: "Golden Core",
			wantLevel:   3,
			wantCurrent: 0,
			wantReq:     1200,
			wantNext:    "元婴",
			wantNextEn:  "Nascent Soul",
		},
		{
			name:        "exactly 1200 reaches Nascent Soul",
			mastered:    1200,
			wantName:    "元婴",
			wantEnglish: "Nascent Soul",
			wantLevel:   4,
			wantCurrent: 0,
			wantReq:     2000,
			wantNext:    "化神",
			wantNextEn:  "Spirit Severing",
		},
		{
			name:        "exactly 2000 reaches Spirit Severing",
			mastered:    2000,
			wantName:    "化神",
			wantEnglish: "Spirit Severing",
			wantLevel:   5,
			wantCurrent: 0,
			wantReq:     3000,
			wantNext:    "大乘",
			wantNextEn:  "Mahayana",
		},
		{
			name:        "exactly 3000 reaches Mahayana",
			mastered:    3000,
			wantName:    "大乘",
			wantEnglish: "Mahayana",
			wantLevel:   6,
			wantCurrent: 0,
			wantReq:     5000,
			wantNext:    "斗战胜佛",
			wantNextEn:  "Victorious Fighting Buddha",
		},
		{
			name:        "exactly 5000 reaches max rank",
			mastered:    5000,
			wantName:    "斗战胜佛",
			wantEnglish: "Victorious Fighting Buddha",
			wantLevel:   7,
			wantCurrent: 5000,
			wantReq:     5000,
			wantNext:    "",
			wantNextEn:  "",
		},
		{
			name:        "above max rank threshold",
			mastered:    9999,
			wantName:    "斗战胜佛",
			wantEnglish: "Victorious Fighting Buddha",
			wantLevel:   7,
			wantCurrent: 9999,
			wantReq:     5000,
			wantNext:    "",
			wantNextEn:  "",
		},
		// One-below boundary tests
		{
			name:        "299 still Qi Refining",
			mastered:    299,
			wantName:    "练气",
			wantEnglish: "Qi Refining",
			wantLevel:   1,
			wantCurrent: 199,
			wantReq:     300,
			wantNext:    "筑基",
			wantNextEn:  "Foundation Building",
		},
		{
			name:        "4999 still Mahayana",
			mastered:    4999,
			wantName:    "大乘",
			wantEnglish: "Mahayana",
			wantLevel:   6,
			wantCurrent: 1999,
			wantReq:     5000,
			wantNext:    "斗战胜佛",
			wantNextEn:  "Victorious Fighting Buddha",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calculateRank(tt.mastered)
			if got.Name != tt.wantName {
				t.Errorf("Name = %q, want %q", got.Name, tt.wantName)
			}
			if got.English != tt.wantEnglish {
				t.Errorf("English = %q, want %q", got.English, tt.wantEnglish)
			}
			if got.Level != tt.wantLevel {
				t.Errorf("Level = %d, want %d", got.Level, tt.wantLevel)
			}
			if got.Current != tt.wantCurrent {
				t.Errorf("Current = %d, want %d", got.Current, tt.wantCurrent)
			}
			if got.Required != tt.wantReq {
				t.Errorf("Required = %d, want %d", got.Required, tt.wantReq)
			}
			if got.NextName != tt.wantNext {
				t.Errorf("NextName = %q, want %q", got.NextName, tt.wantNext)
			}
			if got.NextEnglish != tt.wantNextEn {
				t.Errorf("NextEnglish = %q, want %q", got.NextEnglish, tt.wantNextEn)
			}
		})
	}
}

func TestCalculateProgress(t *testing.T) {
	tests := []struct {
		trib, cleared int
		wantLoc       string
		wantPct       float64
	}{
		// Each tribulation's own place (content/tribulations.yaml).
		{1, 0, "灵山", 0},
		{10, 9, "观音禅院", 9.0 / 81 * 100},
		{50, 49, "火焰山", 49.0 / 81 * 100},
		{81, 81, "长安", 100},
	}
	for _, tt := range tests {
		got := calculateProgress(tt.trib, tt.cleared, 0, 0)
		if got.Tribulation != tt.trib || got.Location != tt.wantLoc || got.Percent != tt.wantPct {
			t.Errorf("progress(%d, %d cleared) = %+v, want %s at %.1f%%", tt.trib, tt.cleared, got, tt.wantLoc, tt.wantPct)
		}
	}
}

// clearedThrough marks stages 1..n cleared.
func clearedThrough(n int) map[int]string {
	c := map[int]string{}
	for i := 1; i <= n; i++ {
		c[i] = "2026-11-21"
	}
	return c
}

func TestStagesOpenByMasteryAndClearByWinning(t *testing.T) {
	stages := Stages(Counts{}, nil)
	if len(stages) != 81 || stages[0].Status != StageOpen || stages[1].Status != StageLocked {
		t.Fatalf("new learner: stage 1 %v, stage 2 %v; want stage 1 open, the rest locked", stages[0].Status, stages[1].Status)
	}
	// Mastery alone doesn't skip ahead: stage 2 waits for stage 1's win.
	if s := Stages(Counts{Learned: 9999, Mastered: 9999}, nil); s[1].Status != StageLocked {
		t.Error("stage 2 opened before stage 1 was cleared")
	}
	s := Stages(Counts{Learned: stages[1].Unlock}, clearedThrough(1))
	if s[0].Status != StageCleared || s[1].Status != StageOpen || CurrentStage(s).Number != 2 {
		t.Errorf("after clearing 1 with stage 2's mastery: %v %v, current %d", s[0].Status, s[1].Status, CurrentStage(s).Number)
	}
	if s := Stages(Counts{Learned: stages[1].Unlock - 1, Mastered: 9999}, clearedThrough(1)); s[1].Status != StageLocked {
		t.Error("stage 2 opened without its mastery")
	}
	if CurrentStage(Stages(Counts{Learned: 9999, Mastered: 9999}, clearedThrough(81))) != nil {
		t.Error("a finished pilgrimage still has a current stage")
	}
}

func TestEarlyStagesCountLearnedLaterMastered(t *testing.T) {
	last := content.Progression.LearningUntil
	early, late := Stages(Counts{}, nil)[1], Stages(Counts{}, nil)[last]
	learner := Counts{Learned: late.Unlock, Mastered: 0}
	if s := Stages(learner, clearedThrough(1)); s[1].Status != StageOpen {
		t.Errorf("stage 2 with %d learned characters is %v, want open", early.Unlock, s[1].Status)
	}
	if s := Stages(learner, clearedThrough(last)); s[last].Status != StageLocked {
		t.Errorf("stage %d opened on learned characters; it counts mastered ones", last+1)
	}
	if got := NeedText(late, Counts{Mastered: late.Unlock - 3}); got != "master 3 more characters" {
		t.Errorf("NeedText = %q", got)
	}
	if got := NeedText(early, Counts{}); got != fmt.Sprintf("learn %d more characters", early.Unlock) {
		t.Errorf("NeedText = %q", got)
	}
}

func TestStageOpponents(t *testing.T) {
	stages := Stages(Counts{}, nil)
	for _, s := range stages {
		if s.Opponent.Name == "" {
			t.Errorf("stage %d has no opponent", s.Number)
		}
		if s.Boss != nil && s.Opponent.Name != BossEntry(*s.Boss).Name {
			t.Errorf("boss stage %d fights %s, not its boss %s", s.Number, s.Opponent.Name, s.Boss.Name)
		}
	}
	if stages[19].Opponent.Name != "白骨精" || stages[20].Opponent.Name != "白骨精" {
		t.Errorf("stage 20 %s, stage 21 %s; want White Bone Demon carried forward", stages[19].Opponent.Name, stages[20].Opponent.Name)
	}
}

func TestCompanionsJoinAtTheirStage(t *testing.T) {
	joined := func(c []Companion) (names []string) {
		for _, x := range c {
			if x.Unlocked {
				names = append(names, x.Name)
			}
		}
		return names
	}
	if got := joined(buildCompanions(nil)); len(got) != 1 || got[0] != "唐僧" {
		t.Errorf("at the start: %v, want only Tang Sanzang", got)
	}
	if got := joined(buildCompanions(clearedThrough(12))); len(got) != 4 {
		t.Errorf("after stage 12: %v, want Tang, Wukong, the horse and Bajie", got)
	}
}

func TestBosses(t *testing.T) {
	for _, b := range Bosses(Counts{}, clearedThrough(8)) {
		if b.Required != TribulationAt(b.Stage).Unlock {
			t.Errorf("%s needs %d characters, not its stage's %d", b.English, b.Required, TribulationAt(b.Stage).Unlock)
		}
		if want := b.Stage <= 8; b.Defeated != want {
			t.Errorf("%s (stage %d) Defeated = %v after clearing stages 1-8", b.English, b.Stage, b.Defeated)
		}
	}
	if next := NextBoss(Bosses(Counts{}, clearedThrough(8))); next == nil || next.Name != "白骨精" {
		t.Errorf("next boss after stage 8 = %v, want White Bone Demon", next)
	}
	if NextBoss(Bosses(Counts{}, clearedThrough(81))) != nil {
		t.Error("a boss is left after every stage is cleared")
	}
}

func TestBuildArtifacts(t *testing.T) {
	tests := []struct {
		name     string
		mastered int
		streak   int
		want     []bool // earned state for each of the 6 artifacts
	}{
		{
			name:     "nothing earned",
			mastered: 0,
			streak:   0,
			want:     []bool{false, false, false, false, false, false},
		},
		{
			name:     "only streak artifact earned at 7 days",
			mastered: 0,
			streak:   7,
			want:     []bool{false, true, false, false, false, false},
		},
		{
			name:     "streak 6 not enough for Headband Spell",
			mastered: 0,
			streak:   6,
			want:     []bool{false, false, false, false, false, false},
		},
		{
			name:     "100 mastered earns Ruyi Jingu Bang",
			mastered: 100,
			streak:   0,
			want:     []bool{true, false, false, false, false, false},
		},
		{
			name:     "99 mastered not enough for Ruyi Jingu Bang",
			mastered: 99,
			streak:   0,
			want:     []bool{false, false, false, false, false, false},
		},
		{
			name:     "500 mastered earns Banana Leaf Fan",
			mastered: 500,
			streak:   0,
			want:     []bool{true, false, true, false, false, false},
		},
		{
			name:     "1000 mastered earns Ginseng Fruit",
			mastered: 1000,
			streak:   0,
			want:     []bool{true, false, true, true, false, false},
		},
		{
			name:     "2000 mastered earns Fiery Eyes",
			mastered: 2000,
			streak:   0,
			want:     []bool{true, false, true, true, true, false},
		},
		{
			name:     "3000 mastered earns all mastery artifacts",
			mastered: 3000,
			streak:   0,
			want:     []bool{true, false, true, true, true, true},
		},
		{
			name:     "all artifacts earned",
			mastered: 3000,
			streak:   7,
			want:     []bool{true, true, true, true, true, true},
		},
		{
			name:     "high mastery with streak earns everything",
			mastered: 5000,
			streak:   30,
			want:     []bool{true, true, true, true, true, true},
		},
	}

	artifactNames := []string{
		"Ruyi Jingu Bang", "Headband Spell", "Banana Leaf Fan",
		"Ginseng Fruit", "Fiery Eyes", "Sea-Calming Needle",
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildArtifacts(tt.mastered, tt.streak)
			if len(got) != 6 {
				t.Fatalf("got %d artifacts, want 6", len(got))
			}
			for i, a := range got {
				if a.Earned != tt.want[i] {
					t.Errorf("artifact %q Earned = %v, want %v (mastered=%d, streak=%d)",
						a.Name, a.Earned, tt.want[i], tt.mastered, tt.streak)
				}
				if a.Name != artifactNames[i] {
					t.Errorf("artifact[%d].Name = %q, want %q", i, a.Name, artifactNames[i])
				}
			}
		})
	}
}

func TestBuildArtifactsFields(t *testing.T) {
	artifacts := buildArtifacts(0, 0)

	expected := []struct {
		name    string
		chinese string
		earnAt  int
		kind    string
	}{
		{"Ruyi Jingu Bang", "金箍棒", 100, "mastery"},
		{"Headband Spell", "紧箍咒", 7, "streak"},
		{"Banana Leaf Fan", "芭蕉扇", 500, "mastery"},
		{"Ginseng Fruit", "人参果", 1000, "mastery"},
		{"Fiery Eyes", "火眼金睛", 2000, "mastery"},
		{"Sea-Calming Needle", "定海神针", 3000, "mastery"},
	}

	for i, e := range expected {
		a := artifacts[i]
		if a.Chinese != e.chinese {
			t.Errorf("artifact[%d].Chinese = %q, want %q", i, a.Chinese, e.chinese)
		}
		if a.EarnAt != e.earnAt {
			t.Errorf("artifact[%d].EarnAt = %d, want %d", i, a.EarnAt, e.earnAt)
		}
		if a.Kind != e.kind {
			t.Errorf("artifact[%d].Kind = %q, want %q", i, a.Kind, e.kind)
		}
		if a.Desc == "" {
			t.Errorf("artifact[%d].Desc is empty", i)
		}
	}
}
