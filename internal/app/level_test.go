package app

import (
	"strings"
	"testing"

	"github.com/therealsiege/tuilaoshi/internal/journey"
)

func TestLevelGainsSayWhatTheLevelBrought(t *testing.T) {
	st := journey.State{Level: journey.LevelFor(225)} // level 10
	gains := strings.Join(levelGains(st, 8, 0), "\n")
	if !strings.Contains(gains, "Max HP +8 (now 136)") || !strings.Contains(gains, "master 50 more") {
		t.Errorf("gains for level 8 -> 10:\n%s", gains)
	}
	stage := journey.Stage{Status: journey.StageOpen}
	stage.Number = 5
	st.Stage = &stage
	if gains := strings.Join(levelGains(st, 9, 0), "\n"); !strings.Contains(gains, "Tribulation 5 is open") {
		t.Errorf("a level that opened a stage should say so:\n%s", gains)
	}
	if gains := strings.Join(levelGains(st, 9, 5), "\n"); strings.Contains(gains, "Tribulation") {
		t.Errorf("a stage already open is not news:\n%s", gains)
	}
}
