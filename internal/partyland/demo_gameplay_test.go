//go:build dmoimpl1

package partyland

import (
	"fmt"
	"os"
	"os/exec"

	"pinballfantasies/internal/physics"
	"pinballfantasies/internal/presentation"
	"pinballfantasies/internal/settings"

	"reflect"
	"testing"
)

// The saved deterministic BYGEL/drain script, without a trajectory search.
func demoFixedInput(n int) physics.Inputs {
	in := physics.Inputs{Down: n >= 35438 && n < 35460, Release: n == 35460}
	if n >= 35460 {
		k := n - 35460
		in.Left = (k+46)%52 < 8
		in.Right = (k+22)%30 < 21
	}
	return in
}

func TestDemoFreshGameplayBoundary(t *testing.T) {
	g := newTestGame(t)
	g.Configure(settings.Legacy())
	reference := newTestGame(t)
	reference.Configure(settings.Legacy())
	d := connected(g)
	d.loadGameplayOperands(t)
	g.Physics.OnEvent = func(physics.Event) { t.Fatal("canonical OnEvent fallback") }
	g.Physics.BeforeTargets = func() { t.Fatal("canonical BeforeTargets fallback") }
	g.Physics.AfterTargets = func(physics.Inputs) { t.Fatal("canonical AfterTargets fallback") }
	g.Physics.BeforeLate = func() { t.Fatal("canonical BeforeLate fallback") }
	if d.counter != 0 || g.clock != 0 || g.ScoreChanged || g.Score.Uint64() != 0 || g.Bonus.Uint64() != 0 || g.Session.CurrentPlayer != 1 || g.BallNumber != 1 || g.matrix.active || d.expired || d.loosing || d.holdStill || g.Physics.Ball.Lost || g.Physics.Ball.Hold || len(g.waitCounters) != 0 {
		t.Fatal("nonfresh entry")
	}
	for _, task := range g.tasks {
		if task != nil {
			t.Fatal("nonfresh task")
		}
	}
	counts := map[string]int{}
	for n := 1; n <= 35877; n++ {
		if err := reference.Sync(demoFixedInput(n)); err != nil {
			t.Fatal(err)
		}
		if err := d.sync(demoFixedInput(n), true); err != nil {
			t.Logf("boundary %v ball=%+v score=%s bonus=%s tasks=%v waits=%v matrix=%+v last=%s/%s", err, g.Physics.Ball, g.Score.String(), g.Bonus.String(), g.taskIDs, g.waitCounters, g.matrix, g.lastCheck, g.lastArea)
			t.Logf("boundary counters skill=%d reverse=%d clock=%d random=%d springValid=%t touchDisabled=%t arcade=%t audio=%+v lamps=%v flashes=%v", g.SkillTime, g.InhibitReverseTime, g.clock, g.Random, g.Physics.SpringValid, g.touchDisabled, g.Arcade, g.Audio, g.Lamps, g.flashes)
			if g.touchDisabled || g.Arcade || g.Physics.Ball.HitX != 132 || g.Physics.Ball.HitY != 199 || g.ScoreChanged {
				t.Fatal("TOUCHER effects before refusal")
			}
			for _, task := range g.tasks {
				if task != nil {
					t.Fatal("target task before refusal")
				}
			}
			if n != 35712 || d.failure.Consumer != "target/OnEvent" || d.failure.Producer != "checkSpringAndTargets target=0" || d.counter != 35712 || g.Physics.Syncs != 35711 {
				t.Fatal("fixed stopping boundary drift", err)
			}
			if !reflect.DeepEqual(counts, map[string]int{"BYGEL12": 2, "BYGEL28": 1, "CLOSE1": 1, "BYGEL9": 1, "BYGEL11": 1}) {
				t.Fatal("fixed callback counts", counts)
			}
			assertConnectedSticky(t, d, err)
			return
		}
		if g.Physics.Ball != reference.Physics.Ball || !reflect.DeepEqual(g.Physics.Flippers, reference.Physics.Flippers) || g.Physics.SpringPosition != reference.Physics.SpringPosition || g.Physics.ScreenOffset != reference.Physics.ScreenOffset || g.Physics.Raster != reference.Physics.Raster || g.Physics.ScreenSpeed != reference.Physics.ScreenSpeed || g.Physics.ScreenPosition != reference.Physics.ScreenPosition {
			t.Fatalf("reference trajectory mismatch at %d candidate=%+v reference=%+v", n, g.Physics.Ball, reference.Physics.Ball)
		}
		if d.counter != uint16(n) || g.Random != reference.Random || g.clock != reference.clock || g.Score != reference.Score || g.Bonus != reference.Bonus || g.ScoreChanged != reference.ScoreChanged || g.SkillTime != reference.SkillTime || g.InhibitReverseTime != reference.InhibitReverseTime || g.lastArea != reference.lastArea || g.lastCheck != reference.lastCheck || g.Lights != reference.Lights {
			t.Fatalf("electronics mismatch at %d", n)
		}
		if n == 35460 {
			t.Logf("release clock=%d low8=%d ball=%+v spring=%d", g.clock, uint8(g.clock), g.Physics.Ball, g.Physics.SpringPosition)
		}
		for _, e := range g.Events {
			if e.Kind == "Switch" {
				counts[e.Label]++
				t.Logf("callback calculation=%d %s ball=%+v timers=%d/%d score=%s", n, e.Label, g.Physics.Ball, g.SkillTime, g.InhibitReverseTime, g.Score.String())
			}
		}
	}
	t.Fatal("missing bounded stopping boundary")
}

func (d *demoConnected) loadGameplayOperands(t *testing.T) {
	t.Helper()
	_ = pinnedDemoBytes(t)
	cmd := exec.Command("python3", "../../tools/check_demo_2b8_consumers.py")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("consumer admission: %v %s", err, out)
	}
	b := pinnedDemoBytes(t)
	a, err := os.ReadFile("../../TABLE1.PRG")
	if err != nil {
		t.Fatal(err)
	}
	demoOK(t, d.admitGameplay(b, a))
}

// Structural state injection below is deliberately separate from the fresh test.
func TestDemoBygelStructural(t *testing.T) {
	for _, label := range []string{"BYGEL1", "BYGEL2", "BYGEL9", "BYGEL11", "CLOSE1"} {
		t.Run(label, func(t *testing.T) {
			d := connected(newTestGame(t))
			d.loadGameplayOperands(t)
			g := d.game
			positions := map[string][2]int16{"BYGEL1": {0, 450}, "BYGEL2": {280, 450}, "BYGEL9": {200, 15}, "BYGEL11": {90, 15}, "CLOSE1": {280, 300}}
			p := positions[label]
			g.Physics.SetBall(p[0], p[1], 0, 0, false)
			g.Audio.Priority = 255
			g.Bonus = Number(1234)
			if label == "BYGEL11" {
				g.lastArea = "BYGEL9"
				g.InhibitReverseTime = 7
			}
			if label == "CLOSE1" {
				g.lastArea = "BYGEL12"
				d.visaKeys = true
			}
			g.Events = nil
			demoOK(t, d.gameplayArea())
			if g.lastArea != label || g.lastCheck != label || g.Bonus.Uint64() != 1234 {
				t.Fatal("bookkeeping or bonus")
			}
			if label == "BYGEL1" || label == "BYGEL2" {
				if g.Score.Uint64() != 50030 || !g.ScoreChanged || len(g.Events) != 3 || g.Events[1].Label != "SBYGEL1" || g.Events[2].Kind != "ScoreAwarded" || g.Audio.Priority != 255 {
					t.Fatal("score/sound independent of jingle priority", g.Events)
				}
			} else if g.ScoreChanged || g.Score.Uint64() != 0 {
				t.Fatal("spurious award")
			}
			if label == "BYGEL11" && g.InhibitReverseTime != 0 {
				t.Fatal("inhibited reverse exit")
			}
			if label == "CLOSE1" && (g.inChute || d.visaKeys || g.partyFlash || g.Audio.Priority != 255 || g.Audio.JumpCount != 0 || g.Audio.ReturnPosition != 1 || !g.matrix.active) {
				t.Fatal("priority rejection must still replace matrix")
			}
			before := len(g.Events)
			demoOK(t, d.gameplayArea())
			if len(g.Events) != before {
				t.Fatal("duplicate callback")
			}
		})
	}
}
func TestDemoBygelRejectBeforeNestedEffects(t *testing.T) {
	for _, kind := range []string{"lit-lane", "loop", "reverse", "unknown", "missing-operands", "score-panel"} {
		t.Run(kind, func(t *testing.T) {
			d := connected(newTestGame(t))
			d.loadGameplayOperands(t)
			g := d.game
			switch kind {
			case "lit-lane":
				g.Physics.SetBall(0, 450, 0, 0, false)
				g.Lights[39] = true
			case "loop":
				g.Physics.SetBall(200, 15, 0, 0, false)
				g.lastArea = "BYGEL11"
			case "reverse":
				g.Physics.SetBall(90, 15, 0, 0, false)
				g.lastArea = "BYGEL9"
			case "unknown":
				g.Physics.SetBall(120, 165, 0, 0, false)
			case "missing-operands":
				g.Physics.SetBall(297, 530, 0, 0, false)
				d.gameplayOperands = false
			case "score-panel":
				g.Physics.SetBall(0, 450, 0, 0, false)
				d.fjantText = true
				g.Display.Content.Texts["PLAYERSTEXT"] = nil
			}
			before := fmt.Sprintf("%#v %#v", g, g.Physics)
			err := d.gameplayArea()
			if err == nil || before != fmt.Sprintf("%#v %#v", g, g.Physics) {
				t.Fatal("effects before admission", err)
			}
			if d.gameplayArea() != err || d.gameplayIdle() != err {
				t.Fatal("direct sticky consumer")
			}
			assertConnectedSticky(t, d, err)
		})
	}
}
func TestDemoFreshInitialRejection(t *testing.T) {
	_ = pinnedDemoBytes(t)
	d := connected(newTestGame(t))
	d.game.Configure(settings.Legacy())
	err := d.sync(demoFixedInput(1), true)
	if err == nil || d.failure.Consumer != "BYGEL12" || d.calls != 1 || d.counter != 1 || d.game.Physics.Ball.PixelX != 297 || d.game.Physics.Ball.PixelY != 530 || d.game.lastCheck != "" || d.game.SkillTime != 0 {
		t.Fatal("initial fixed boundary", err)
	}
	assertConnectedSticky(t, d, err)
}

func TestDemoBygelTaskMatrixOrdering(t *testing.T) {
	d := connected(newTestGame(t))
	d.loadGameplayOperands(t)
	g := d.game
	// Structural CLOSE1 checkpoint, not an input-only witness.
	g.Physics.SetBall(280, 300, 0, 0, false)
	g.lastArea = "BYGEL12"
	demoOK(t, d.gameplayArea())
	demoOK(t, d.queue(demoTask{Site: "preserve", Action: demoPreserve}))
	demoOK(t, d.sync(physics.Inputs{}, false))
	if g.tasks[0] != nil || g.matrix.remaining != 5 || d.counter != 1 {
		t.Fatal("task scan/budget order")
	}
	demoOK(t, d.sync(physics.Inputs{}, true))
	if g.matrix.remaining != 4 || d.counter != 2 || g.lastCheck != "CLOSE1" {
		t.Fatal("matrix visit or callback repeated")
	}
}
func TestDemoBygelScorePanelAdmission(t *testing.T) {
	d := connected(newTestGame(t))
	d.loadGameplayOperands(t)
	g := d.game
	g.Physics.SetBall(0, 450, 0, 0, false)
	d.fjantText = true
	d.infoCount = 37
	demoOK(t, d.gameplayArea())
	if d.fjantText || d.infoCount != 0 || g.Score.Uint64() != 50030 || g.Bonus.Uint64() != 0 || !g.ScoreChanged || g.matrix.op != "_CLEAR4" || !d.ownedMatrix {
		t.Fatal("ADDSCORE/UPDAT_INFOBAR")
	}
	for _, bad := range []presentation.Command{
		{Op: "_PRINT5", Args: []string{"PLAYERSTEXT", "999"}, Nums: map[int]int{1: 340}},
		{Op: "_PRINT5", Args: []string{"PLAYERSTEXT", "340"}, Nums: map[int]int{0: 340}},
		{Op: "_PARTYOFF", Args: []string{"2"}, Nums: map[int]int{0: 2}},
	} {
		if d.command(bad) {
			t.Fatal("malformed source operand")
		}
	}
}

func TestDemoBygelMatrixFreeEffect(t *testing.T) {
	for _, priority := range []uint8{0, 1, 255} {
		for _, inhibit := range []bool{false, true} {
			d := connected(newTestGame(t))
			d.loadGameplayOperands(t)
			g := d.game
			g.Physics.SetBall(25, 435, 0, 0, false)
			g.Audio.Priority = priority
			g.inhibitEffect = inhibit
			g.Multiplier = 8
			audio, matrix := g.Audio, g.matrix
			demoOK(t, d.gameplayArea())
			if g.Score.Uint64() != 10040 || g.Bonus.Uint64() != 1000 || !g.ScoreChanged || g.effectAccepted || !g.effectEnded || g.Audio != audio || !reflect.DeepEqual(matrix, g.matrix) {
				t.Fatal("matrix-free DOEFFECT priority/arithmetic")
			}
			if g.Events[len(g.Events)-1].Label != "SBYGEL2" {
				t.Fatal("effect sound order")
			}
		}
	}
}
