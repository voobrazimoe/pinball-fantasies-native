package stones

import (
	"fmt"
	"pinballfantasies/internal/physics"
	"testing"
)

func started(t *testing.T, g *Game, label string) {
	t.Helper()
	for _, e := range g.Events {
		if e.Kind == "MatrixStarted" && e.Label == label {
			return
		}
	}
	t.Fatalf("source-selected matrix %s did not start", label)
}

func TestStonesMatrixMultiplierStates(t *testing.T) {
	for i, pre := range []uint8{1, 2, 4, 6, 8} {
		t.Run(fmt.Sprint(pre), func(t *testing.T) {
			g := game(t)
			g.Multiplier = pre
			g.bonusPointer = uint8(39 + i)
			g.flash(24, 16)
			g.music.Priority = 0
			g.captureWell()
			want := []uint8{2, 4, 6, 8, 10}[i]
			if g.Multiplier != want {
				t.Fatal("transition", g.Multiplier, want)
			}
			started(t, g, fmt.Sprintf("M%dTS", want))
		})
	}
}

func TestStonesMatrixGhostStates(t *testing.T) {
	for i := uint8(0); i < 8; i++ {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			g := game(t)
			g.GhostCounter = i
			g.music.Priority = 0
			for j := 0; j < 9; j++ {
				g.touch(j)
			}
			started(t, g, fmt.Sprintf("EVENT%dTS", i+1))
			g = game(t)
			g.GhostCounter = i
			g.ghostFlashing = true
			g.music.Priority = 0
			g.captureVault()
			started(t, g, programs.Effects[ghostEffects[i]].Matrix)
			if g.GhostCounter != (i+1)%8 {
				t.Fatal("ghost transition", g.GhostCounter)
			}
		})
	}
}

func TestStonesMatrixTowerStates(t *testing.T) {
	for stage := uint8(1); stage <= 3; stage++ {
		g := game(t)
		g.TowerHunt = true
		g.TowerStage = stage
		g.music.Priority = 0
		g.captureTower()
		started(t, g, fmt.Sprintf("TOWER%dTS", stage))
		if g.TowerStage != (stage+1)%4 {
			t.Fatal("tower transition", g.TowerStage)
		}
		// Run the packed _TOWER command and verify it reaches its source row.
		for n := 0; n < 400 && g.matrix.op != "_TOWER"; n++ {
			g.matrixTick()
		}
		if g.matrix.op != "_TOWER" {
			t.Fatal("packed Tower entry absent")
		}
		row := g.matrix.nums[0]
		for n := 0; n < 200 && g.matrix.op == "_TOWER"; n++ {
			g.matrixTick()
		}
		if g.matrix.towerRow != row {
			t.Fatal("packed Tower final row", g.matrix.towerRow, row)
		}
	}
	for _, q := range []struct {
		lamp   int
		effect string
	}{{8, "EXTRABALL"}, {9, "JACKPOT"}, {10, "SUPERJACK"}, {11, "TMILLION"}, {12, "TMILLION5"}, {13, "HOLDBONUS"}, {14, "DOUBLEBONUS"}} {
		t.Run(q.effect, func(t *testing.T) {
			g := game(t)
			g.music.Priority = 0
			g.synced(q.lamp, 18, g.syncTower)
			g.captureTower()
			started(t, g, programs.Effects[q.effect].Matrix)
		})
	}
}

func TestStonesMatrixAwardsAndCaptures(t *testing.T) {
	for _, effect := range []string{"MILLION", "MILLION5", "MILLION10", "MILLION20", "BATMAN", "REDDEVIL", "MUMMYHEAD", "GHOSTHUNT", "GRIMR", "MULTIDEMONS", "HOLDBONUS", "DOUBLEBONUS"} {
		t.Run(effect, func(t *testing.T) {
			g := game(t)
			g.music.Priority = 0
			if !g.effect(effect) {
				t.Fatal("effect rejected", effect)
			}
			started(t, g, programs.Effects[effect].Matrix)
		})
	}
	for locks := 0; locks <= 2; locks++ {
		g := game(t)
		g.music.Priority = 0
		g.flash(18, 18)
		g.Lights[15] = locks > 0
		g.Lights[25] = locks > 1
		g.scream()
		started(t, g, programs.Effects[[]string{"MILLION5", "MILLION10", "MILLION20"}[locks]].Matrix)
	}
	for _, capture := range []string{"TOWER", "WELL", "VAULT"} {
		g := game(t)
		g.music.Priority = 0
		switch capture {
		case "TOWER":
			g.captureTower()
		case "WELL":
			g.captureWell()
		case "VAULT":
			g.captureVault()
		}
		started(t, g, programs.Effects["SCORE"+capture].Matrix)
	}
	for _, grim := range []bool{false, true} {
		g := game(t)
		g.captured = "TOWER"
		g.wasSpecial = true
		g.grimBackup = grim
		g.eject("TOWER")
		label := "BACK_2_OFFROADTS"
		if grim {
			label = "BACK_2_TURBOTS"
		}
		started(t, g, label)
	}
}

func TestStonesMatrixBeatenBonusMatchGameOver(t *testing.T) {
	g := game(t)
	g.Score = number(100000001)
	g.SetHighScore(number(100000000))
	found := false
	for i, c := range g.Display.Content.Commands {
		if c.Op == "_BEATEN_MATRIX" {
			g.matrix.next = i
			g.matrix.active = true
			g.matrixDispatch()
			found = true
			if !g.beaten || g.ExtraBalls != 1 || g.matrix.next != g.Display.Content.Labels["BEATEN_BH_TS"]+1 {
				t.Fatal("beaten branch", g.matrix)
			}
			break
		}
	}
	if !found {
		t.Fatal("_BEATEN_MATRIX absent")
	}
	for _, win := range []bool{false, true} {
		g = game(t)
		g.Score = number(10)
		g.matrix.matchLast = 0
		if win {
			g.matrix.matchLast = 1
		}
		// beginMatrix resets matchLast, so enter the real opcode directly.
		g.matrix.next = g.Display.Content.Labels["CHECK_XXBALLTS"]
		g.matrix.active = true
		g.matrixDispatch()
		if win && !g.matchBall {
			t.Fatal("winning match branch")
		}
		if !win {
			for i := 0; i < 100 && g.Phase != GameOver; i++ {
				g.matrixTick()
			}
			if g.Phase != GameOver {
				t.Fatal("losing match game-over branch", g.Phase)
			}
		}
	}
	for _, extra := range []uint8{0, 1} {
		g = game(t)
		g.matchBall = true
		g.ExtraBalls = extra
		g.matrix.next = g.Display.Content.Labels["NO_BONUS2TS"]
		g.matrix.active = true
		g.matrixDispatch()
		if extra > 0 && g.ExtraBalls != 0 {
			t.Fatal("_KOLLA_XXBALL extra branch")
		}
	}
	for _, mult := range []uint8{1, 2, 4, 6, 8, 10} {
		g = game(t)
		g.Multiplier = mult
		g.Bonus = number(1000)
		g.music.Priority = 0
		g.drain()
		started(t, g, "BALL_LOSTTS")
		for n := 0; n < 5000 && g.Phase == BallLost; n++ {
			if err := g.Sync(physics.Inputs{}); err != nil {
				t.Fatal(err)
			}
		}
		if g.Phase == BallLost {
			t.Fatal("bonus stalled", mult, g.matrix.op)
		}
	}
}

func TestStonesMatrixScreamAndDeferredGhostStates(t *testing.T) {
	for _, pre := range []uint16{2, 11} {
		g := game(t)
		g.Screams = pre
		g.NextJump = 30
		g.music.Priority = 0
		g.scream()
		label := "JUMP_AT_TS"
		if pre > 10 {
			label = "JUMP_AT2_TS"
		}
		started(t, g, label)
	}
	for _, idx := range []uint8{4, 7} {
		g := game(t)
		g.GhostCounter = idx
		g.ghostFlashing = true
		g.Special = true
		g.music.Priority = 0
		g.captureVault()
		g.runTasks()
		g.Special = false
		g.runTasks()
		started(t, g, programs.Effects[ghostEffects[idx]].Matrix)
	}
}
