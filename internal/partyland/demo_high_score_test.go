//go:build dmoimpl1

package partyland

import (
	"fmt"
	"os"
	"os/exec"
	"reflect"
	"testing"
)

func (d *demoConnected) loadHighScoreOperands(t *testing.T) {
	t.Helper()
	d.loadToucherOperands(t)
	cmd := exec.Command("python3", "../../tools/check_demo_2b10_consumers.py")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("CHECKHIGHSCORE admission: %v %s", err, out)
	}
	// No .HI access: the admitted native factory policy starts from compiled
	// defaults, checked against native Defaults(1)'s pristine inventory identity.
	b := pinnedDemoBytes(t)
	a, err := os.ReadFile("../../TABLE1.PRG")
	if err != nil {
		t.Fatal(err)
	}
	demoOK(t, d.admitHighScore(b, a))
}

func TestDemoHighScoreComparatorStructural(t *testing.T) {
	// Independent integer oracle covers all differing decimal positions and ties.
	for _, top := range []uint64{0, 9, 10, 50030, 50000000, 999999999999} {
		for _, score := range []uint64{0, 1, 9, 10, 50029, 50030, 50031, 49999999, 50000000, 50000001, 999999999999} {
			above, valid := demoScoreAbove(Number(score), Number(top))
			if !valid || above != (score > top) {
				t.Fatal(score, top, above, valid)
			}
		}
	}
	for i := 0; i < 12; i++ {
		top := Decimal{}
		top[i] = 5
		for _, delta := range []int{-1, 0, 1} {
			score := top
			score[i] = uint8(5 + delta)
			above, valid := demoScoreAbove(score, top)
			if !valid || above != (delta > 0) {
				t.Fatal(i, delta)
			}
		}
		bad := top
		bad[i] = 10
		if _, ok := demoScoreAbove(bad, top); ok {
			t.Fatal("invalid score")
		}
		if _, ok := demoScoreAbove(top, bad); ok {
			t.Fatal("invalid top")
		}
	}
}

func TestDemoHighScoreGuardsStructural(t *testing.T) {
	for _, mode := range []string{"below", "equal", "reward", "chute", "special", "beaten", "missing", "persistent", "changed", "invalid", "invalid-score", "idle-boundary"} {
		t.Run(mode, func(t *testing.T) {
			d := connected(newTestGame(t))
			d.loadHighScoreOperands(t)
			g := d.game
			g.inChute = false
			g.Score = Number(50030)
			g.ScoreChanged = true
			g.SetHighScore(Number(0)) // poison canonical comparator: it would qualify
			switch mode {
			case "equal":
				g.Score = Number(50000000)
			case "reward":
				g.Score = Number(50000001)
			case "chute":
				g.inChute = true
				d.factoryTop = nil
			case "special":
				d.specialMode = true
				d.factoryTop = nil
			case "beaten":
				g.alreadyBeaten = true
				d.factoryTop = nil
			case "missing":
				d.factoryTop = nil
			case "persistent":
				d.factoryTopSource = "DOS-persistent"
			case "changed":
				*d.factoryTop = Number(1)
			case "invalid":
				d.factoryTop[11] = 10
			case "idle-boundary":
				d.infoCount = 720
			case "invalid-score":
				g.Score[0] = 10
			}
			before := fmt.Sprintf("%#v %#v %#v", g, g.Display, g.Physics)
			matrix := g.matrix
			info := d.infoCount
			err := d.checkHighScorePreflight()
			fail := mode == "reward" || mode == "missing" || mode == "persistent" || mode == "changed" || mode == "invalid" || mode == "invalid-score" || mode == "idle-boundary"
			if (err != nil) != fail || before != fmt.Sprintf("%#v %#v %#v", g, g.Display, g.Physics) || !reflect.DeepEqual(matrix, g.matrix) || info != d.infoCount {
				t.Fatal("guard/preflight effects", err)
			}
			if fail {
				if d.gameplayIdle() != err || d.checkHighScorePreflight() != err {
					t.Fatal("sticky")
				}
				assertConnectedSticky(t, d, err)
			} else {
				demoOK(t, d.gameplayIdle())
				if d.infoCount != info+1 || g.ScoreChanged != true || g.Bonus != Number(0) || g.Score.Uint64() == 0 {
					t.Fatal("idle continuation")
				}
			}
		})
	}
}
