//go:build dmoimpl1

package partyland

import (
	"pinballfantasies/internal/physics"
	"pinballfantasies/internal/presentation"
	"reflect"
	"testing"
)

// Structural fixtures only: real native steps detect the explicitly prepared
// ball's drain. No fresh input-only prefix or official demo witness is claimed.
func demoDraining(t *testing.T) *demoConnected {
	t.Helper()
	d := connectedQuiet(t)
	d.game.Physics.SetBall(100, 575, 0, 1024, false)
	d.game.Physics.OnEvent = func(physics.Event) { t.Fatal("canonical physics callback") }
	d.game.Physics.BeforeTargets = func() { t.Fatal("canonical BeforeTargets") }
	d.game.Physics.AfterTargets = func(physics.Inputs) { t.Fatal("canonical AfterTargets") }
	d.game.Physics.BeforeLate = func() { t.Fatal("canonical BeforeLate") }
	return d
}

func TestDemoUnscoredDrain(t *testing.T) {
	for _, expired := range []bool{false, true} {
		for _, hiRes := range []bool{false, true} {
			d := demoDraining(t)
			d.expired, d.hiRes, d.specialMode = expired, hiRes, true
			d.game.Happy, d.game.Mega, d.game.inhibitEffect = true, true, true
			d.game.Audio.Priority = 255 // direct source clear must admit priority1
			d.game.Audio.Position = 7
			beforeEvents := len(d.game.Events)
			score, bonus, ball := d.game.Score, d.game.Bonus, d.game.BallNumber
			demoOK(t, d.sync(physics.Inputs{}, true))
			if len(d.snapshots) != 4 {
				t.Fatal(d.snapshots)
			}
			entry, guards, installed, after := d.snapshots[0], d.snapshots[1], d.snapshots[2], d.snapshots[3]
			if !entry.Lost || entry.Loosing || entry.ScoreChanged || entry.Expired != expired || entry.Ball.PixelY < 576 || !entry.Special {
				t.Fatal(entry)
			}
			if !guards.Hold || !guards.Loosing || guards.AllowFlip || guards.Special || guards.Ball.PixelX != 15 || guards.Ball.PixelY != 47 || guards.Ball.VX != 0 || guards.Ball.VY != 0 || guards.Ball.High || guards.Ball.Hold {
				t.Fatal(guards)
			}
			if installed.Counter != 0 || installed.Wait != 0 || installed.Slots[0] == 0 || installed.Matrix.op != "_CLEAR4" || installed.Matrix.remaining != 5 || installed.Matrix.next != 1 || installed.Audio.Priority != 1 || installed.Audio.Position != 0 || installed.Audio.ReturnPosition != 7 || !installed.InhibitEffect {
				t.Fatal(installed)
			}
			if after.Counter != 1 || after.Wait != 1 || after.Matrix.remaining != 4 || after.Matrix.next != 1 || after.Slots != installed.Slots || after.Ball != installed.Ball || after.Expired != expired {
				t.Fatal(after)
			}
			force := int16(369)
			if hiRes {
				force = 259
			}
			if d.screenForce2 != force {
				t.Fatal(d.screenForce2)
			}
			if d.drains != 1 || d.oldExpired != expired || d.game.Random != 1 || d.game.Score != score || d.game.Bonus != bonus || d.game.BallNumber != ball || d.game.Phase != BallLost || d.game.Physics.Stopped || d.game.partyFlash || !d.game.musicOK {
				t.Fatal("drain side effects")
			}
			if len(d.game.Events) < beforeEvents+1 || d.game.Events[beforeEvents].Kind != "MatrixCommand" || d.game.Events[beforeEvents].Label != "_CLEAR4" || len(d.game.Physics.Events) != 0 || !reflect.DeepEqual(d.order, []string{"early.step", "early.step", "early.finish", "drain handoff", "ElectronicsCalculation", "task/matrix", "late.step", "complete"}) {
				t.Fatal(d.order)
			}
			for _, event := range d.game.Events[beforeEvents+1:] {
				if event.Kind != "LampChanged" {
					t.Fatal("unexpected suffix event", event.Kind)
				}
			}
			demoOK(t, d.sync(physics.Inputs{}, false))
			if d.drains != 1 || d.counter != 2 || d.game.Random != 2 || d.game.waitCounters["PARTY_ON_TASK1"] != 2 || !reflect.DeepEqual(d.game.matrix, after.Matrix) || d.game.Physics.Ball.X != after.Ball.X || d.game.Physics.Ball.Y != after.Ball.Y || d.game.Physics.Ball.VX != 0 || d.game.Physics.Ball.VY != 0 || d.game.taskIDs[0] != installed.Slots[0] || d.game.tasks[0] == nil {
				t.Fatal("pending continuation/duplicate drain")
			}
		}
	}
}

func TestDemoDrainFirstFreeAndSharedWait(t *testing.T) {
	d := demoDraining(t)
	prepareDemoReset(t, d)
	demoOK(t, d.queue(demoTask{Site: "earlier", Delay: 100}))
	demoOK(t, d.queue(demoTask{Site: "hole", Delay: 100}))
	demoOK(t, d.queue(demoTask{Site: "later", Delay: 100}))
	d.game.tasks[1] = nil
	d.game.waitCounters["PARTY_ON_TASK1"] = 7 // structural: allocator must not zero shared word
	demoOK(t, d.sync(physics.Inputs{}, false))
	installed := d.snapshots[2]
	if installed.Slots[1] == 0 || installed.Slots[0] == 0 || installed.Slots[2] == 0 || installed.Wait != 7 || d.game.waitCounters["PARTY_ON_TASK1"] != 8 || d.game.taskIDs[1] != installed.Slots[1] {
		t.Fatal(installed)
	}
	// No body dispatch while matrix budget is absent. Age30 survives until due visit.
	for d.game.waitCounters["PARTY_ON_TASK1"] < 30 {
		demoOK(t, d.sync(physics.Inputs{}, false))
	}
	beforeMatrix := d.game.matrix
	demoOK(t, d.sync(physics.Inputs{}, false))
	if d.handoffs != 1 || !reflect.DeepEqual(beforeMatrix, d.game.matrix) || !d.game.partyFlash || d.game.waitCounters["SOUNDNEWBALL"] != 0 || d.game.waitCounters["SETBALL"] != 0 || d.game.waitCounters["SOUNDBRICKUPP"] != 1 {
		t.Fatal("slot1 reset must visit only new slot2")
	}

}

func TestDemoDrainScoredAndGuardRejection(t *testing.T) {
	for _, kind := range []string{"scored", "scored-expired", "puke", "full", "nested", "matrix", "release"} {
		t.Run(kind, func(t *testing.T) {
			d := demoDraining(t)
			switch kind {
			case "scored", "scored-expired":
				d.game.ScoreChanged = true
				d.expired = kind == "scored-expired"
				d.counter = 35997
			case "puke":
				d.game.PukeForbidden = true
			case "full":
				for i := 0; i < 50; i++ {
					demoOK(t, d.queue(demoTask{Site: "occupied", Delay: 100}))
				}
			case "nested":
				demoOK(t, d.queue(demoTask{Site: "unsupported", Action: demoReplaceMatrix, Program: []presentation.Command{{Op: "_FLASHON", Args: []string{"4"}, Nums: map[int]int{0: 4}}}}))
			case "release":
				demoOK(t, d.queue(demoTask{Site: "release", Action: demoReleaseSourceHold}))
			}
			beforeEvents := append([]Event(nil), d.game.Events...)
			beforeScore, beforeBonus, beforeAudio := d.game.Score, d.game.Bonus, d.game.Audio
			err := d.sync(physics.Inputs{}, kind == "matrix")
			if kind == "matrix" {
				demoOK(t, err)
				for i := 0; i < 3; i++ {
					demoOK(t, d.sync(physics.Inputs{}, true))
				}
				demoOK(t, d.sync(physics.Inputs{}, true)) // final CLEAR4 admits FLASHON
				before := d.game.matrix
				err = d.sync(physics.Inputs{}, true)
				if err == nil || !reflect.DeepEqual(d.game.matrix, before) || before.remaining != 1 || before.next != 2 || d.game.partyFlash {
					t.Fatal("unverified PARTYONN dispatched", err)
				}
			}
			if err == nil {
				t.Fatal("missing refusal")
			}
			if d.game.Score != beforeScore || d.game.Bonus != beforeBonus || !d.game.Physics.Ball.Lost {
				t.Fatal("canonical continuation")
			}
			if kind == "scored" || kind == "scored-expired" {
				if d.failure.Consumer != "scored drain" || !d.game.ScoreChanged || d.counter != 35997 || d.game.Random != 0 || d.game.matrix.active || d.game.Audio != beforeAudio || !reflect.DeepEqual(d.game.Events, beforeEvents) || d.game.tasks[0] != nil || d.snapshots[1].Expired != d.oldExpired {
					t.Fatal("scored effects before refusal")
				}
			}
			if kind == "nested" && d.game.waitCounters["unsupported"] != 0 {
				t.Fatal("nested wait reset")
			}
			if kind == "full" && (d.counter != 0 || d.game.Random != 0) {
				t.Fatal("allocator failure proceeded to electronics")
			}
			assertConnectedSticky(t, d, err)
		})
	}
}

func TestDemoDrainEqualityOrder(t *testing.T) {
	for _, expired := range []bool{false, true} {
		d := demoDraining(t)
		d.counter = 35997
		d.expired = expired
		demoOK(t, d.sync(physics.Inputs{}, true))
		installed, after := d.snapshots[2], d.snapshots[3]
		if installed.Counter != 35997 || installed.Expired != expired || installed.Audio.Priority != 1 || installed.Matrix.remaining != 5 || installed.Wait != 0 || after.Counter != 35998 || !after.Expired || after.Audio.Priority != 255 || after.Matrix.remaining != 4 || after.Wait != 1 || d.cueRequests != 1 || d.drains != 1 {
			t.Fatal(installed, after)
		}
		if d.game.Display.Content.Commands[1].Op != "_SCROLL" {
			t.Fatal("timer equality did not replace PARTY_ONTS")
		}
	}
}

// Preserved 2B1 accounting assertion, explicitly only a structural electronics
// suffix. It makes no claim to execute LOOSE_BALL or a complete physical drain.
func TestDemoStructuralPostDrainTimerSuffix(t *testing.T) {
	for _, expired := range []bool{false, true} {
		d := newDemoCore()
		d.game.Phase = BallLost
		d.game.Physics.Ball.Lost = true
		d.counter, d.expired = 35997, expired
		demoOK(t, d.calculation(false, false))
		if d.counter != 35998 || d.game.Random != 1 || !d.expired || !d.holdStill || d.game.matrix.remaining != 5 {
			t.Fatal("structural accounting")
		}
		for _, task := range d.game.tasks {
			if task != nil {
				t.Fatal("structural suffix inserted gameplay")
			}
		}
	}
}

func TestDemoDrainMatrixCursorReplacement(t *testing.T) {
	d := demoDraining(t)
	demoOK(t, d.install(demoWait(1)))
	d.game.matrixTimeLeft = true
	demoOK(t, d.matrixVisit())
	if d.game.matrix.next != 2 {
		t.Fatal("pre-drain cursor")
	}
	demoOK(t, d.sync(physics.Inputs{}, true))
	if d.snapshots[0].Matrix.next != 2 || d.snapshots[2].Matrix.next != 1 || d.snapshots[2].Matrix.remaining != 5 || d.game.matrix.next != 1 || d.game.matrix.remaining != 4 {
		t.Fatal("direct installation did not reset cursor")
	}
	demoOK(t, d.sync(physics.Inputs{}, false))
	if d.game.matrix.next != 1 || d.game.matrix.remaining != 4 {
		t.Fatal("budgetless cursor changed")
	}
}
