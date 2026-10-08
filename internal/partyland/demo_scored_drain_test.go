//go:build dmoimpl1

package partyland

import (
	"os"
	"os/exec"

	"pinballfantasies/internal/presentation"

	"reflect"
	"testing"
)

func (d *demoConnected) loadScoredDrainOperands(t *testing.T) {
	t.Helper()
	d.loadHighScoreOperands(t)
	cmd := exec.Command("python3", "../../tools/check_demo_2b11_consumers.py")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("scored admission: %v %s", err, out)
	}
	b := pinnedDemoBytes(t)
	a, err := os.ReadFile("../../TABLE1.PRG")
	if err != nil {
		t.Fatal(err)
	}
	demoOK(t, d.admitScoredDrain(b, a))
}

func TestDemoScoredDrainStructural(t *testing.T) {
	for _, mode := range []string{"zero", "zero-x8", "expired", "inhibited", "inhibited-nonzero", "priority", "missing-text", "full", "nonzero-bonus", "cyclones", "happy", "mega", "match", "holdbonus"} {
		t.Run(mode, func(t *testing.T) {
			d := demoDraining(t)
			d.loadScoredDrainOperands(t)
			g := d.game
			g.Score = Number(50030)
			g.ScoreChanged = true
			g.partyFlash = false
			switch mode {
			case "zero-x8":
				g.Multiplier = 8
			case "inhibited-nonzero":
				g.inhibitEffect = true
				g.Bonus = Number(1234)
				g.Multiplier = 8
			case "expired":
				d.expired = true
			case "inhibited":
				g.inhibitEffect = true
			case "priority":
				g.Audio.Priority = 255
			case "missing-text":
				g.Display.Content.Texts["DEMO_DRAIN_TEXT"] = nil
			case "full":
				for i := 0; i < 50; i++ {
					demoOK(t, d.queue(demoTask{Site: "occupied", Delay: 65535}))
				}
			case "nonzero-bonus":
				g.Bonus = Number(1)
			case "cyclones":
				g.Cyclones = 1
			case "happy":
				g.HappyTotal = Number(1)
			case "mega":
				g.MegaTotal = Number(1)
			case "match":
				g.matchBall = true
			case "holdbonus":
				g.HoldBonus = true
			}
			before := g.Audio
			err := d.looseBall()
			early := mode == "expired" || mode == "missing-text"
			if early {
				if err == nil || g.Audio != before || g.matrix.active || !g.ScoreChanged || g.Score != Number(50030) {
					t.Fatal("pre-effect refusal", err)
				}
				assertConnectedSticky(t, d, err)
				return
			}
			if mode == "full" {
				if err == nil || !g.matrix.active || g.matrix.remaining != 5 || g.Audio.Position != 6 {
					t.Fatal("retain admitted effect before allocator refusal", err)
				}
				assertConnectedSticky(t, d, err)
				return
			}
			demoOK(t, err)
			if mode == "inhibited" || mode == "inhibited-nonzero" {
				if g.matrix.active || g.Audio.Position != before.Position || g.tasks[0] == nil {
					t.Fatal("effect guard")
				}
				if mode == "inhibited-nonzero" && g.Bonus != Number(1234) {
					t.Fatal("zero effect accounting preserves nonzero bonus under blocked matrix")
				}
				return
			}
			if !g.matrix.active || g.matrix.op != "_CLEAR4" || g.matrix.remaining != 5 || g.Audio.Position != 6 || g.Audio.Priority != 0 || g.Audio.ReturnPosition != 62 || g.Audio.JumpCount != 1 {
				t.Fatal("LOSTBALL real effect", g.Audio, g.matrix)
			}
			for i := 0; i < 91 && d.failure == nil; i++ {
				err = d.electronics(true)
			}
			negative := mode == "nonzero-bonus" || mode == "cyclones" || mode == "happy" || mode == "mega" || mode == "match" || mode == "holdbonus"
			if (err != nil) != negative || g.Score != Number(50030) || !g.ScoreChanged {
				t.Fatal("bounded bonus result", mode, err)
			}
			if negative {
				assertConnectedSticky(t, d, err)
				if d.scoredDrain() != err || d.preflightScoredTask("NEW_BALL_TASK") != err {
					t.Fatal("direct sticky")
				}
				return
			}
			if len(d.bonusNodes) != 5 || len(d.drainFires) != 1 || d.drainFires[0].Site != "SOUNDRINNER" || g.tasks[0] == nil || g.waitCounters["NEW_BALL_TASK"] != 0 || g.Session.Load().Score != Number(50030) {
				t.Fatal("zero source continuation")
			}
		})
	}
}

func TestDemoScoredTaskAdmissionAndReplacement(t *testing.T) {
	for _, mode := range []string{"missing", "panel", "party", "visa", "replace"} {
		t.Run(mode, func(t *testing.T) {
			d := demoDraining(t)
			d.loadScoredDrainOperands(t)
			g := d.game
			g.Score = Number(50030)
			g.ScoreChanged = true
			g.partyFlash = false
			g.musicOK = false
			demoOK(t, d.looseBall())
			for i := 0; i < 91; i++ {
				demoOK(t, d.electronics(true))
			}
			if mode == "replace" {
				demoOK(t, d.install(demoWait(100)))
			}
			if mode == "party" {
				g.partyFlash = true
			}
			if mode == "visa" {
				d.visaKeys = true
			}
			for i := 0; i < 30; i++ {
				demoOK(t, d.electronics(true))
			}
			if mode == "missing" {
				d.childOperands = false
			}
			if mode == "panel" {
				g.Display.Content.Texts["DEMO_BALLSTEXT"] = nil
			}
			matrix := g.matrix
			score := g.Score
			err := d.electronics(true)
			if mode == "missing" || mode == "panel" {
				if err == nil || g.waitCounters["NEW_BALL_TASK"] != 30 || g.tasks[0] == nil || !reflect.DeepEqual(g.matrix, matrix) || g.Score != score || !g.ScoreChanged {
					t.Fatal("due preflight before reset", err)
				}
				assertConnectedSticky(t, d, err)
				return
			}
			demoOK(t, err)
			if d.handoffs != 1 || g.Score != score || g.ScoreChanged || g.Physics.Ball.Lost || d.loosing || g.waitCounters["SETBALL"] != 1 || g.waitCounters["SOUNDBRICKUPP"] != 1 {
				t.Fatal("newball")
			}
			if mode == "party" || mode == "visa" {
				if g.matrix.next != matrix.next || g.matrix.op != matrix.op || d.visaKeys {
					t.Fatal("real reset guards")
				}
			} else if g.matrix.next != 1 || g.matrix.remaining != 4 || g.Display.Content.Commands[2].Args[0] != "DEMO_BALLSTEXT" {
				t.Fatal("replacement")
			}
		})
	}
}

func TestDemoScoredTailCannotDispatchCanonical(t *testing.T) {
	d := demoDraining(t)
	d.loadScoredDrainOperands(t)
	g := d.game
	g.Score = Number(50030)
	g.ScoreChanged = true
	demoOK(t, d.looseBall())
	// Structural mutation of the future same-visit tail, never a replay input.
	g.Display.Content.Commands[5] = presentation.Command{Op: "_CHANGE_PLAYER"}
	for i := 0; i < 91 && d.failure == nil; i++ {
		_ = d.electronics(true)
	}
	if d.failure == nil || d.failure.Producer != "HU_" || g.Score != Number(50030) || d.handoffs != 0 || g.BallNumber != 1 || !g.ScoreChanged || len(d.bonusNodes) != 1 {
		t.Fatal("canonical tail escaped admission", d.failure)
	}
	assertConnectedSticky(t, d, d.failure)
}

func TestDemoScoredDirectEntryAdmission(t *testing.T) {
	d := connected(newTestGame(t))
	d.loadScoredDrainOperands(t)
	before := d.game.Audio
	score := d.game.Score
	matrix := d.game.matrix
	err := d.scoredDrain()
	if err == nil || d.game.Audio != before || d.game.Score != score || !reflect.DeepEqual(d.game.matrix, matrix) {
		t.Fatal("direct scored entry bypass", err)
	}
	assertConnectedSticky(t, d, err)
	d.executeScoredTask("NEW_BALL_TASK")
	if d.game.Score != score || d.handoffs != 0 {
		t.Fatal("direct task sticky bypass")
	}
	other := connected(newTestGame(t))
	other.loadScoredDrainOperands(t)
	other.childOperands = false
	other.executeScoredTask("NEW_BALL_TASK")
	if other.failure == nil || other.handoffs != 0 || other.game.tasks[0] != nil {
		t.Fatal("direct task missing admission")
	}
}
