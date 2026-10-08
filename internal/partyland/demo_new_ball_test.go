//go:build dmoimpl1

package partyland

import (
	"bytes"

	"pinballfantasies/internal/physics"
	"reflect"
	"testing"
)

// Connected structural checkpoints, never fresh input-only demo witnesses.
func TestDemoNewBallConnectedHandoff(t *testing.T) {
	for _, slot := range []int{0, 1, 3} {
		for _, expiry := range []bool{false, true} {
			d := demoDraining(t)
			prepareDemoReset(t, d)
			d.visaKeys = true // PARTYFLASH must skip this guard and retain it.
			for i := 0; i < slot; i++ {
				demoOK(t, d.queue(demoTask{Site: "old", Delay: 200}))
			}
			if expiry {
				d.counter = 35967
			}
			saved := PlayerState{Score: Decimal{}, Bonus: Decimal{}, Cyclones: 7}
			saved.Score.AddNumber(123)
			saved.Bonus.AddNumber(456)
			saved.SkillTunnel.AddNumber(789)
			saved.Progress[0], saved.Progress[16] = true, true
			d.game.Session.Save(saved)
			d.game.HappyTotal.AddNumber(5)
			d.game.MegaTotal.AddNumber(6)
			for i := 0; i < 30; i++ {
				demoOK(t, d.sync(physics.Inputs{}, false))
				if d.game.waitCounters["PARTY_ON_TASK1"] != uint16(i+1) || d.handoffs != 0 {
					t.Fatal("compare-before-increment")
				}
			}
			if d.counter == 35998 {
				t.Fatal("expiry too early")
			}
			// Structural reset operands, after connected age30, before due visit.
			d.game.inhibitEffect, d.game.PukeForbidden = true, true
			d.game.Happy, d.game.Mega, d.game.ScoreChanged = true, true, true
			d.game.TunnelTime, d.game.LoopTime, d.game.ReverseTime = 100, 100, 100
			d.game.SkillTime = 100 // not a RESET_VARS store; UPDATE_COUNTERS alone decrements
			d.shiftPressed, d.inhibitCountdown = true, true
			d.game.Physics.Tilted, d.game.Physics.AllowFlip = true, false
			d.game.Physics.TiltCounter = 17
			d.game.Physics.Ball.High = true
			d.game.ExtraBalls = 1
			for i := 0; i < 3; i++ {
				d.game.duckMask(i, true)
			}
			oldID := d.game.taskIDs[slot]
			matrix := d.game.matrix
			d.game.waitCounters["stale"] = 99
			demoOK(t, d.sync(physics.Inputs{}, false))
			g, b := d.game, d.game.Physics.Ball
			if d.handoffs != 1 || d.loosing || b.Lost || !d.holdStill || b.Hold || b.High || b.PixelX != 282 || b.PixelY != 530 || b.X != 282*1024 || b.Y != 530*1024 || b.VX != 0 || b.VY != 0 || !g.inChute || !g.Physics.SpringValid || !g.Physics.AllowFlip || g.Physics.Tilted || g.Phase != NewBall || !g.partyFlash || !d.visaKeys {
				t.Fatal("new-ball state", b)
			}
			if g.SavePlayerState() != saved || g.HappyTotal != (Decimal{}) || g.MegaTotal != (Decimal{}) || g.ScoreChanged || g.inhibitEffect || d.specialMode || !d.keyTaskEmpty || !d.dotReady || d.bonusX != 1 || d.screenForce2 != -1 || g.Physics.TargetRaster != -1 || g.Display.Content.Texts["BONUS_TEXT"][11] != '8' {
				t.Fatal("reset/player state")
			}
			if d.shiftPressed || d.inhibitCountdown || g.PukeForbidden || g.Happy || g.Mega || g.TunnelTime != 0 || g.LoopTime != 0 || g.ReverseTime != 0 || g.SkillTime != 99 || g.Physics.TiltCounter != 0 || !g.Lights[51] || !g.Lights[52] || !g.Lights[53] || !g.Lights[54] || g.flashes[0] != (flash{14, 1, 8}) || g.flashes[1] != (flash{26, 1, 9}) {
				t.Fatal("transient/lamp reset")
			}
			for i, y := range []int{277, 295, 313} {
				w := 2
				if i == 2 {
					w = 1
				}
				if !bytes.Equal(g.Physics.MaskRegion(false, []int{18, 19, 20}[i], y, w, 15), g.duckUp[i]) {
					t.Fatal("duck restore", i)
				}
			}

			if expiry {
				if d.counter != 35998 || !d.expired || d.cueRequests != 1 || g.Display.Content.Commands[1].Op != "_SCROLL" || g.matrix.remaining != 5 {
					t.Fatal("expiry lost")
				}
			} else if d.counter != 31 || d.expired || !reflect.DeepEqual(matrix, g.matrix) {
				t.Fatal("matrix/timer changed")
			}
			for i, f := range g.tasks {
				if (i < 3) != (f != nil) {
					t.Fatal("TASKLIST", i)
				}
			}
			for i := 0; i < 3; i++ {
				if g.taskIDs[i] <= oldID {
					t.Fatal("old identity survived")
				}
			}
			want := map[string]uint16{}
			if slot < 1 {
				want["SETBALL"] = 1
			}
			if slot < 2 {
				want["SOUNDBRICKUPP"] = 1
			}
			if !reflect.DeepEqual(g.waitCounters, want) {
				t.Fatal("WAITLIST remaining scan", slot, g.waitCounters, want)
			}
			demoOK(t, d.sync(physics.Inputs{}, false))
			if d.handoffs != 1 || g.waitCounters["SOUNDNEWBALL"] != 1 || g.waitCounters["SETBALL"] != want["SETBALL"]+1 || g.waitCounters["SOUNDBRICKUPP"] != want["SOUNDBRICKUPP"]+1 || g.Physics.Ball != b {
				t.Fatal("next scan or immediate release")
			}
		}
	}
}

func TestDemoNewBallPreflightBeforeWaitReset(t *testing.T) {
	for _, missing := range []string{"duck", "player", "music", "addplayers", "reset-inputs", "reset-text"} {
		d := demoDraining(t)
		prepareDemoReset(t, d)
		for i := 0; i < 30; i++ {
			demoOK(t, d.sync(physics.Inputs{}, false))
		}
		switch missing {
		case "reset-inputs":
			d.resetOperands = false
		case "reset-text":
			d.game.Display.Content.Texts["BONUS_TEXT"] = nil
		case "duck":
			d.game.duckUp[2] = nil
		case "player":
			d.game.Session.CurrentPlayer = 0
		case "music":
			d.game.musicOK = false
		case "addplayers":
			d.addPlayers = true
		}
		ball, matrix, ids := d.game.Physics.Ball, d.game.matrix, d.game.taskIDs
		waits := map[string]uint16{}
		for k, v := range d.game.waitCounters {
			waits[k] = v
		}
		err := d.sync(physics.Inputs{}, false)
		if err == nil || d.game.partyFlash || d.handoffs != 0 || d.game.Physics.Ball != ball || !reflect.DeepEqual(matrix, d.game.matrix) || ids != d.game.taskIDs || !reflect.DeepEqual(waits, d.game.waitCounters) {
			t.Fatal("partial reset", missing, err)
		}
		assertConnectedSticky(t, d, err)
	}
}

func TestDemoNewBallChildBoundaries(t *testing.T) {
	for _, child := range []demoTask{{Site: "SOUNDBRICKUPP", Delay: 5}, {Site: "SOUNDNEWBALL", Delay: 50}, {Site: "SETBALL", Delay: 80}} {
		d := demoDraining(t)
		prepareDemoReset(t, d)
		for i := 0; i < 31; i++ {
			demoOK(t, d.sync(physics.Inputs{}, false))
		}
		d.childOperands = false // deliberately remove the newly admitted consumer
		// For the first child use actual connected waiting. Other child boundaries
		// are explicit structural due checkpoints; no reachability claim is made.
		if child.Site == "SOUNDBRICKUPP" {
			for d.game.waitCounters[child.Site] < child.Delay {
				demoOK(t, d.sync(physics.Inputs{}, false))
			}
		} else {
			d.game.waitCounters[child.Site] = child.Delay
		}
		d.game.Audio.Active = false // freeze prior clock for this isolated refusal snapshot
		b, m, a, ids, events := d.game.Physics.Ball, d.game.matrix, d.game.Audio, d.game.taskIDs, len(d.game.Events)
		err := d.sync(physics.Inputs{}, false)
		if err == nil || d.failure.Producer != child.Site || d.game.waitCounters[child.Site] != child.Delay || d.game.Physics.Ball != b || !reflect.DeepEqual(m, d.game.matrix) || a != d.game.Audio || ids != d.game.taskIDs || len(d.game.Events) != events || d.handoffs != 1 {
			t.Fatal("child effects before rejection", child.Site, err)
		}
		assertConnectedSticky(t, d, err)
	}
}

func TestDemoNewBallEqualityBudgetAndStickyExpiry(t *testing.T) {
	for _, budget := range []bool{false, true} {
		d := demoDraining(t)
		prepareDemoParty(t, d)
		d.counter = 35967
		for i := 0; i < 30; i++ {
			demoOK(t, d.sync(physics.Inputs{}, budget))
		}
		if d.counter != 35997 || d.expired {
			t.Fatal("pre-equality")
		}
		demoOK(t, d.sync(physics.Inputs{}, budget))
		want := uint16(5)
		if budget {
			want = 4
		}
		if d.counter != 35998 || !d.expired || !d.holdStill || !d.game.partyFlash || d.handoffs != 1 || d.game.matrix.op != "_CLEAR4" || d.game.matrix.remaining != want || d.game.matrix.next != 1 || d.game.Display.Content.Commands[1].Op != "_SCROLL" || demoFlash(d.game.Display).Enabled {
			t.Fatal("expiry replaced/restored")
		}
	}
	d := demoDraining(t)
	prepareDemoReset(t, d)
	d.expired = true
	for i := 0; i < 31; i++ {
		demoOK(t, d.sync(physics.Inputs{}, false))
	}
	if d.counter != 31 || !d.expired || d.cueRequests != 0 || d.handoffs != 1 {
		t.Fatal("reset cleared sticky expired or timer")
	}
}

func TestDemoResetOperandRejectionSticky(t *testing.T) {
	d := newDemoCore()
	err := d.loadResetOperands(nil)
	if err == nil || d.resetOperands || d.loadResetOperands(nil) != err || d.calculation(false, false) != err || d.counter != 0 {
		t.Fatal("reset operand gate")
	}
}
