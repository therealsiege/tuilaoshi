package journey

import (
	"math"
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
		name       string
		mastered   int
		totalChars int
		wantTrib   int
		wantLoc    string
		wantLocEn  string
		wantPct    float64
	}{
		// Zero-total edge case
		{
			name:       "zero total chars defaults to tribulation 1 at Chang'an",
			mastered:   0,
			totalChars: 0,
			wantTrib:   1,
			wantLoc:    "长安",
			wantLocEn:  "Chang'an (Capital)",
			wantPct:    0.0,
		},
		// Starting out
		{
			name:       "zero mastered with chars starts at Chang'an",
			mastered:   0,
			totalChars: 1000,
			wantTrib:   1,
			wantLoc:    "长安",
			wantLocEn:  "Chang'an (Capital)",
			wantPct:    0.0,
		},
		// Low progress
		{
			name:       "small progress still at Chang'an",
			mastered:   1,
			totalChars: 1000,
			wantTrib:   1,
			wantLoc:    "长安",
			wantLocEn:  "Chang'an (Capital)",
			wantPct:    0.1,
		},
		// Tribulation exactly at location boundaries
		{
			name:       "tribulation 10 reaches Twin Fork Ridge",
			mastered:   1000,
			totalChars: 8100,
			wantTrib:   10,
			wantLoc:    "双叉岭",
			wantLocEn:  "Twin Fork Ridge",
			wantPct:    (1000.0 / 8100.0) * 100.0,
		},
		{
			name:       "tribulation 20 reaches Five Elements Mountain",
			mastered:   2000,
			totalChars: 8100,
			wantTrib:   20,
			wantLoc:    "五行山",
			wantLocEn:  "Five Elements Mountain",
			wantPct:    (2000.0 / 8100.0) * 100.0,
		},
		{
			name:       "tribulation 30 reaches Guanyin Temple",
			mastered:   3000,
			totalChars: 8100,
			wantTrib:   30,
			wantLoc:    "观音院",
			wantLocEn:  "Guanyin Temple",
			wantPct:    (3000.0 / 8100.0) * 100.0,
		},
		{
			name:       "tribulation 40 reaches Gao Village",
			mastered:   4000,
			totalChars: 8100,
			wantTrib:   40,
			wantLoc:    "高老庄",
			wantLocEn:  "Gao Village",
			wantPct:    (4000.0 / 8100.0) * 100.0,
		},
		{
			name:       "tribulation 50 reaches Flowing Sands River",
			mastered:   5000,
			totalChars: 8100,
			wantTrib:   50,
			wantLoc:    "流沙河",
			wantLocEn:  "Flowing Sands River",
			wantPct:    (5000.0 / 8100.0) * 100.0,
		},
		{
			name:       "tribulation 60 reaches Flaming Mountains",
			mastered:   6000,
			totalChars: 8100,
			wantTrib:   60,
			wantLoc:    "火焰山",
			wantLocEn:  "Flaming Mountains",
			wantPct:    (6000.0 / 8100.0) * 100.0,
		},
		{
			name:       "tribulation 70 reaches Heaven-Reaching River",
			mastered:   7000,
			totalChars: 8100,
			wantTrib:   70,
			wantLoc:    "通天河",
			wantLocEn:  "Heaven-Reaching River",
			wantPct:    (7000.0 / 8100.0) * 100.0,
		},
		{
			name:       "tribulation 80 reaches Spirit Mountain",
			mastered:   8000,
			totalChars: 8100,
			wantTrib:   80,
			wantLoc:    "灵山",
			wantLocEn:  "Spirit Mountain",
			wantPct:    (8000.0 / 8100.0) * 100.0,
		},
		// Full completion
		{
			name:       "complete journey reaches Great Thunder Monastery",
			mastered:   8100,
			totalChars: 8100,
			wantTrib:   81,
			wantLoc:    "大雷音寺",
			wantLocEn:  "Great Thunder Monastery",
			wantPct:    100.0,
		},
		// Tribulation clamped at 81
		{
			name:       "mastered exceeding total clamps tribulation at 81",
			mastered:   9000,
			totalChars: 8100,
			wantTrib:   81,
			wantLoc:    "大雷音寺",
			wantLocEn:  "Great Thunder Monastery",
			wantPct:    100.0,
		},
		// Mid-journey
		{
			name:       "half mastered gives midpoint tribulation",
			mastered:   500,
			totalChars: 1000,
			wantTrib:   40,
			wantLoc:    "高老庄",
			wantLocEn:  "Gao Village",
			wantPct:    50.0,
		},
		// Percent capped at 100 when mastered > total
		{
			name:       "percent capped at 100",
			mastered:   1200,
			totalChars: 1000,
			wantTrib:   81,
			wantLoc:    "大雷音寺",
			wantLocEn:  "Great Thunder Monastery",
			wantPct:    100.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calculateProgress(tt.mastered, tt.totalChars)
			if got.Tribulation != tt.wantTrib {
				t.Errorf("Tribulation = %d, want %d", got.Tribulation, tt.wantTrib)
			}
			if got.Location != tt.wantLoc {
				t.Errorf("Location = %q, want %q", got.Location, tt.wantLoc)
			}
			if got.LocationEn != tt.wantLocEn {
				t.Errorf("LocationEn = %q, want %q", got.LocationEn, tt.wantLocEn)
			}
			if got.Mastered != tt.mastered {
				t.Errorf("Mastered = %d, want %d", got.Mastered, tt.mastered)
			}
			if got.TotalChars != tt.totalChars {
				t.Errorf("TotalChars = %d, want %d", got.TotalChars, tt.totalChars)
			}
			if math.Abs(got.Percent-tt.wantPct) > 0.01 {
				t.Errorf("Percent = %f, want %f", got.Percent, tt.wantPct)
			}
		})
	}
}

func TestBuildCompanions(t *testing.T) {
	tests := []struct {
		name     string
		mastered int
		want     []bool // unlocked state for each of the 5 companions
	}{
		{
			name:     "zero mastered unlocks only Tang Sanzang",
			mastered: 0,
			want:     []bool{true, false, false, false, false},
		},
		{
			name:     "49 mastered still only Tang Sanzang",
			mastered: 49,
			want:     []bool{true, false, false, false, false},
		},
		{
			name:     "50 mastered unlocks Sun Wukong",
			mastered: 50,
			want:     []bool{true, true, false, false, false},
		},
		{
			name:     "199 mastered still two companions",
			mastered: 199,
			want:     []bool{true, true, false, false, false},
		},
		{
			name:     "200 mastered unlocks Zhu Bajie",
			mastered: 200,
			want:     []bool{true, true, true, false, false},
		},
		{
			name:     "499 mastered still three companions",
			mastered: 499,
			want:     []bool{true, true, true, false, false},
		},
		{
			name:     "500 mastered unlocks Sha Wujing",
			mastered: 500,
			want:     []bool{true, true, true, true, false},
		},
		{
			name:     "999 mastered still four companions",
			mastered: 999,
			want:     []bool{true, true, true, true, false},
		},
		{
			name:     "1000 mastered unlocks all companions",
			mastered: 1000,
			want:     []bool{true, true, true, true, true},
		},
		{
			name:     "large number keeps all unlocked",
			mastered: 5000,
			want:     []bool{true, true, true, true, true},
		},
	}

	companionNames := []string{"唐僧", "孙悟空", "猪八戒", "沙悟净", "白龙马"}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildCompanions(tt.mastered)
			if len(got) != 5 {
				t.Fatalf("got %d companions, want 5", len(got))
			}
			for i, c := range got {
				if c.Unlocked != tt.want[i] {
					t.Errorf("companion %s (%s) Unlocked = %v, want %v",
						c.Name, c.English, c.Unlocked, tt.want[i])
				}
				if c.Name != companionNames[i] {
					t.Errorf("companion[%d].Name = %q, want %q", i, c.Name, companionNames[i])
				}
			}
		})
	}
}

func TestBuildCompanionsFields(t *testing.T) {
	companions := buildCompanions(0)

	expected := []struct {
		name     string
		english  string
		emoji    string
		unlockAt int
	}{
		{"唐僧", "Tang Sanzang", "@", 0},
		{"孙悟空", "Sun Wukong", "M", 50},
		{"猪八戒", "Zhu Bajie", "P", 200},
		{"沙悟净", "Sha Wujing", "S", 500},
		{"白龙马", "Bai Longma", "H", 1000},
	}

	for i, e := range expected {
		c := companions[i]
		if c.English != e.english {
			t.Errorf("companion[%d].English = %q, want %q", i, c.English, e.english)
		}
		if c.Emoji != e.emoji {
			t.Errorf("companion[%d].Emoji = %q, want %q", i, c.Emoji, e.emoji)
		}
		if c.UnlockAt != e.unlockAt {
			t.Errorf("companion[%d].UnlockAt = %d, want %d", i, c.UnlockAt, e.unlockAt)
		}
		if c.Role == "" {
			t.Errorf("companion[%d].Role is empty", i)
		}
	}
}

func TestBuildBosses(t *testing.T) {
	thresholds := []int{100, 250, 500, 800, 1200, 1800, 2500, 4000}

	tests := []struct {
		name     string
		mastered int
		want     []bool // defeated state for each of the 8 bosses
	}{
		{
			name:     "zero mastered no bosses defeated",
			mastered: 0,
			want:     []bool{false, false, false, false, false, false, false, false},
		},
		{
			name:     "99 mastered no bosses defeated",
			mastered: 99,
			want:     []bool{false, false, false, false, false, false, false, false},
		},
		{
			name:     "100 mastered defeats first boss",
			mastered: 100,
			want:     []bool{true, false, false, false, false, false, false, false},
		},
		{
			name:     "249 mastered still only first boss",
			mastered: 249,
			want:     []bool{true, false, false, false, false, false, false, false},
		},
		{
			name:     "250 mastered defeats second boss",
			mastered: 250,
			want:     []bool{true, true, false, false, false, false, false, false},
		},
		{
			name:     "500 mastered defeats three bosses",
			mastered: 500,
			want:     []bool{true, true, true, false, false, false, false, false},
		},
		{
			name:     "800 mastered defeats four bosses",
			mastered: 800,
			want:     []bool{true, true, true, true, false, false, false, false},
		},
		{
			name:     "1200 mastered defeats five bosses",
			mastered: 1200,
			want:     []bool{true, true, true, true, true, false, false, false},
		},
		{
			name:     "1800 mastered defeats six bosses",
			mastered: 1800,
			want:     []bool{true, true, true, true, true, true, false, false},
		},
		{
			name:     "2500 mastered defeats seven bosses",
			mastered: 2500,
			want:     []bool{true, true, true, true, true, true, true, false},
		},
		{
			name:     "4000 mastered defeats all bosses",
			mastered: 4000,
			want:     []bool{true, true, true, true, true, true, true, true},
		},
		{
			name:     "large number defeats all bosses",
			mastered: 10000,
			want:     []bool{true, true, true, true, true, true, true, true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildBosses(tt.mastered)
			if len(got) != 8 {
				t.Fatalf("got %d bosses, want 8", len(got))
			}
			for i, b := range got {
				if b.Defeated != tt.want[i] {
					t.Errorf("boss %s (%s) Defeated = %v, want %v (required=%d, mastered=%d)",
						b.Name, b.English, b.Defeated, tt.want[i], thresholds[i], tt.mastered)
				}
				if b.Required != thresholds[i] {
					t.Errorf("boss[%d].Required = %d, want %d", i, b.Required, thresholds[i])
				}
			}
		})
	}
}

func TestFindNextBoss(t *testing.T) {
	tests := []struct {
		name        string
		mastered    int
		wantNil     bool
		wantEnglish string
		wantReq     int
	}{
		{
			name:        "zero mastered next boss is Havoc Monster",
			mastered:    0,
			wantEnglish: "Havoc Monster",
			wantReq:     100,
		},
		{
			name:        "99 mastered still Havoc Monster",
			mastered:    99,
			wantEnglish: "Havoc Monster",
			wantReq:     100,
		},
		{
			name:        "100 mastered next boss is White Bone Demon",
			mastered:    100,
			wantEnglish: "White Bone Demon",
			wantReq:     250,
		},
		{
			name:        "250 mastered next boss is Red Boy",
			mastered:    250,
			wantEnglish: "Red Boy",
			wantReq:     500,
		},
		{
			name:        "500 mastered next boss is Spider Demon",
			mastered:    500,
			wantEnglish: "Spider Demon",
			wantReq:     800,
		},
		{
			name:        "800 mastered next boss is Gold Horn King",
			mastered:    800,
			wantEnglish: "Gold Horn King",
			wantReq:     1200,
		},
		{
			name:        "1200 mastered next boss is Bull Demon King",
			mastered:    1200,
			wantEnglish: "Bull Demon King",
			wantReq:     1800,
		},
		{
			name:        "1800 mastered next boss is Golden-Winged Roc",
			mastered:    1800,
			wantEnglish: "Golden-Winged Roc",
			wantReq:     2500,
		},
		{
			name:        "2500 mastered next boss is Buddha",
			mastered:    2500,
			wantEnglish: "Buddha (Final Trial)",
			wantReq:     4000,
		},
		{
			name:     "4000 mastered all defeated returns nil",
			mastered: 4000,
			wantNil:  true,
		},
		{
			name:     "huge mastered returns nil",
			mastered: 99999,
			wantNil:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := findNextBoss(tt.mastered)
			if tt.wantNil {
				if got != nil {
					t.Errorf("expected nil, got boss %q", got.English)
				}
				return
			}
			if got == nil {
				t.Fatal("expected a boss, got nil")
			}
			if got.English != tt.wantEnglish {
				t.Errorf("English = %q, want %q", got.English, tt.wantEnglish)
			}
			if got.Required != tt.wantReq {
				t.Errorf("Required = %d, want %d", got.Required, tt.wantReq)
			}
			if got.Defeated {
				t.Error("next boss should not be defeated")
			}
		})
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
