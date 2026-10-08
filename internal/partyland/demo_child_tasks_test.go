//go:build dmoimpl1

package partyland

import (
	"pinballfantasies/internal/audio"
	"pinballfantasies/internal/physics"
	"reflect"
	"testing"
)

func TestDemoChildSequence(t *testing.T) {
	for _, slot := range []int{0, 1, 3} {
		d := demoDraining(t)
		prepareDemoParty(t, d)
		for i := 0; i < slot; i++ {
			demoOK(t, d.queue(demoTask{Site: "old", Delay: 200}))
		}
		for i := 0; i < 31; i++ {
			demoOK(t, d.sync(physics.Inputs{}, true))
		}
		ids := d.game.taskIDs
		score, player := d.game.Score, d.game.SavePlayerState()
		matrix := d.game.matrix
		dots := d.game.Display.Dots
		// Exact dates depend on which new slots remain ahead of the parent scan.
		dates := map[string]int{"SOUNDNEWBALL": 82, "SETBALL": 112, "SOUNDBRICKUPP": 37}
		if slot == 0 {
			dates["SETBALL"] = 111
		}
		if slot < 2 {
			dates["SOUNDBRICKUPP"] = 36
		}
		for n := 32; n <= dates["SETBALL"]; n++ {
			before := d.game.Physics.Ball
			display := *d.game.Display
			display.Flash()
			d.order = nil
			demoOK(t, d.sync(physics.Inputs{}, true))
			if d.game.Display.Dots != dots || demoFlash(d.game.Display) != demoFlash(&display) {
				t.Fatal("dot/flash memory")
			}
			for i, site := range []string{"SOUNDNEWBALL", "SETBALL", "SOUNDBRICKUPP"} {
				initial := 0
				if slot < i {
					initial = 1
				}
				age := initial + n - 31
				if n < dates[site] {
					if d.game.waitCounters[site] != uint16(age) || d.game.tasks[i] == nil || d.game.taskIDs[i] != ids[i] {
						t.Fatal("visit/identity", slot, n, site, d.game.waitCounters)
					}
				} else if d.game.tasks[i] != nil || d.game.waitCounters[site] != 0 {
					t.Fatal("suicide/reset", slot, n, site)
				}
			}
			if n < dates["SETBALL"] && (d.game.Physics.Ball != before || !d.holdStill) {
				t.Fatal("early release")
			}
		}
		if len(d.childFires) != 3 {
			t.Fatal(d.childFires)
		}
		for _, fire := range d.childFires {
			if int(fire.Counter) != dates[fire.Site] {
				t.Fatal("firing", slot, fire)
			}
		}
		fire := d.childFires[2]
		if fire.Ball.X != 297*1024 || fire.Ball.Y != 530*1024 || fire.Ball.VX != 10 || fire.Ball.VY != 0 || fire.Ball.High || fire.Ball.Hold || fire.SourceHold || !fire.InChute {
			t.Fatal("SETBALL stores", fire)
		}
		if d.game.Physics.Ball == fire.Ball || d.counter != uint16(dates["SETBALL"]) || d.expired || d.handoffs != 1 || d.drains != 1 || d.loosing || d.game.Physics.Ball.Lost || d.game.Score != score || d.game.SavePlayerState() != player || !reflect.DeepEqual(matrix, d.game.matrix) {
			t.Fatal("release continuation")
		}
		if !reflect.DeepEqual(d.order, []string{"early.step", "early.step", "early.finish", "ElectronicsCalculation", "targets", "task/matrix", "late.step", "complete"}) {
			t.Fatal(d.order)
		}
		// Independent canonical-A shared integrator, exactly one late pass (all
		// canonical electronics removed); no expected coordinates feed candidate.
		reference := newTestGame(t).Physics
		reference.SetBall(297, 530, 10, 0, false)
		reference.BeforeTargets = func() { reference.Ball.Hold = false }
		reference.BeforeLate = nil
		reference.AfterTargets = nil
		reference.OnEvent = nil
		reference.Ball.Hold = true // suppress the canonical pair of early movement
		demoOK(t, reference.Sync(physics.Inputs{}))
		if d.game.Physics.Ball != reference.Ball {
			t.Fatal("canonical-A first late step", d.game.Physics.Ball, reference.Ball)
		}
		t.Logf("parent slot%d: child dates %v; first real late ball %+v", slot, dates, d.game.Physics.Ball)
		// Stop at the first selected ordinary area/target/event, never run gameplay.
		for n := 0; n < 500 && d.failure == nil; n++ {
			_ = d.sync(physics.Inputs{}, true)
		}
		if d.failure == nil {
			t.Fatal("missing subsequent consumer boundary")
		}
		t.Log(d.failure)
		assertConnectedSticky(t, d, d.failure)
	}
}

func TestDemoChildAdmissionBeforeReset(t *testing.T) {
	for _, site := range []string{"SOUNDBRICKUPP", "SOUNDNEWBALL", "SETBALL"} {
		missings := []string{"operands", "consumer"}
		if site != "SETBALL" {
			missings = append(missings, "playback")
		}
		for _, missing := range missings {
			t.Run(site+"/"+missing, func(t *testing.T) {
				d := demoDraining(t)
				prepareDemoReset(t, d)
				for i := 0; i < 31; i++ {
					demoOK(t, d.sync(physics.Inputs{}, false))
				}
				delay := uint16(5)
				if site == "SOUNDNEWBALL" {
					delay = 50
				}
				if site == "SETBALL" {
					delay = 80
				}
				// Declared structural due checkpoint, not a connected prefix claim.
				d.game.waitCounters[site] = delay
				if missing == "operands" {
					d.childOperands = false
				} else if missing == "playback" {
					d.game.Playback = &audio.Player{}
				} else if site == "SETBALL" {
					d.game.inChute = false
				} else {
					label := "SBRICKUPP"
					if site == "SOUNDNEWBALL" {
						label = "SNEWBALL"
					}
					old := audio.Effects[label]
					delete(audio.Effects, label)
					defer func() { audio.Effects[label] = old }()
				}
				demoOK(t, d.queue(demoTask{Site: "survivor", Delay: 200}))
				ball, matrix, ids, events := d.game.Physics.Ball, d.game.matrix, d.game.taskIDs, len(d.game.Events)
				d.game.runTasks()
				if d.failure == nil || d.failure.Producer != site || d.game.waitCounters[site] != delay || d.game.tasks[3] == nil || d.game.taskIDs != ids || d.game.Physics.Ball != ball || !reflect.DeepEqual(matrix, d.game.matrix) || len(d.game.Events) != events {
					t.Fatal("admission/reset", site, missing, d.failure)
				}
				assertConnectedSticky(t, d, d.failure)
			})
		}
	}
}

func TestDemoChildSoundIgnoresJinglePriority(t *testing.T) {
	for _, site := range []string{"SOUNDBRICKUPP", "SOUNDNEWBALL"} {
		for _, priority := range []uint8{0, 255} {
			d := demoDraining(t)
			prepareDemoReset(t, d)
			delay := uint16(5)
			label := "SBRICKUPP"
			if site == "SOUNDNEWBALL" {
				delay = 50
				label = "SNEWBALL"
			}
			d.game.Audio = silentJingle{Position: 13, Priority: priority, JumpCount: 7, ReturnPosition: 2, Active: true, entry: 13, cue: timing.Cues["13"].TrackerTicks * 71}
			before := d.game.musicClock()
			demoOK(t, d.queue(demoTask{Site: site, Delay: delay, Action: demoChildWait}))
			demoOK(t, d.queue(demoTask{Site: "survivor", Delay: 200}))
			d.game.waitCounters[site] = delay
			d.game.runTasks()
			if d.failure != nil || d.game.tasks[0] != nil || d.game.tasks[1] == nil || d.game.waitCounters[site] != 0 || d.game.musicClock() != before || len(d.childFires) != 1 || d.game.Events[len(d.game.Events)-1].Label != label {
				t.Fatal("sound/priority/suicide")
			}
			before.Sync(timing.Cues, func(string, uint64) {})
			d.game.audioTick()
			if d.game.musicClock() != before {
				t.Fatal("MusicClock continuation")
			}
		}
	}
}

func TestDemoSetBallCaptureAndPrerequisites(t *testing.T) {
	for _, kind := range []string{"capture", "table", "lost", "loosing", "phase"} {
		d := demoDraining(t)
		prepareDemoReset(t, d)
		for i := 0; i < 31; i++ {
			demoOK(t, d.sync(physics.Inputs{}, false))
		}
		d.game.waitCounters["SETBALL"] = 80
		switch kind {
		case "capture":
			d.game.Physics.Ball.Hold = true
		case "table":
			d.game.Physics.Table = nil
		case "lost":
			d.game.Physics.Ball.Lost = true
		case "loosing":
			d.loosing = true
		case "phase":
			d.game.Phase = Playing
		}
		before := d.game.Physics.Ball
		d.game.runTasks()
		if kind != "capture" {
			if d.failure == nil || d.game.waitCounters["SETBALL"] != 80 || d.game.Physics.Ball != before {
				t.Fatal("SETBALL prerequisites", kind)
			}
			assertConnectedSticky(t, d, d.failure)
			continue
		}
		if d.failure != nil || d.holdStill || !d.game.Physics.Ball.Hold {
			t.Fatal("capture cleared")
		}
		before = d.game.Physics.Ball
		demoOK(t, d.stage("late.step", physics.Inputs{}))
		if d.game.Physics.Ball != before {
			t.Fatal("false capture release")
		}
	}
}

func TestDemoSetBallFirstEquality(t *testing.T) {
	d := demoDraining(t)
	prepareDemoParty(t, d)
	d.counter = 35887 // structural unscored drain35888, not input-only prefix
	for d.counter < 35997 {
		demoOK(t, d.sync(physics.Inputs{}, true))
	}
	if d.handoffs != 1 || d.childFires[0].Counter != 35923 || d.childFires[1].Counter != 35969 || d.game.waitCounters["SETBALL"] != 80 || d.expired {
		t.Fatal("equality prefix")
	}
	before := d.game.Physics.Ball
	demoOK(t, d.sync(physics.Inputs{}, true))
	f := d.childFires[2]
	if f.Counter != 35998 || !d.expired || d.holdStill || f.Ball.X != 297*1024 || d.game.Physics.Ball == f.Ball || before.PixelX != 282 || d.game.matrix.op != "_CLEAR4" || d.game.matrix.next != 1 || d.game.matrix.remaining != 4 || d.game.Display.Content.Commands[1].Op != "_SCROLL" || d.handoffs != 1 {
		t.Fatal("expiry/release/cursor", f, d.game.matrix)
	}
	demoErr := d.sync(physics.Inputs{}, true)
	if demoErr == nil || d.counter != 35999 || d.failure.Consumer != "BYGEL12" || d.game.matrix.remaining != 4 || d.game.matrix.next != 1 || !d.expired {
		t.Fatal("next gameplay boundary before SCROLL", d.failure)
	}
	t.Log(d.failure)
	assertConnectedSticky(t, d, d.failure)
}

func TestDemoCompetingChildInstancesShareWait(t *testing.T) {
	d := demoDraining(t)
	prepareDemoReset(t, d)
	for i := 0; i < 2; i++ {
		demoOK(t, d.queue(demoTask{Site: "SOUNDBRICKUPP", Delay: 5, Action: demoChildWait}))
	}
	ids := d.game.taskIDs
	for i := 0; i < 2; i++ {
		d.game.runTasks()
		if d.game.waitCounters["SOUNDBRICKUPP"] != uint16(2*(i+1)) {
			t.Fatal("shared wait")
		}
	}
	d.game.runTasks() // slot0:4->5; slot1:compare5,reset,suicide
	if len(d.childFires) != 1 || d.game.tasks[0] == nil || d.game.tasks[1] != nil || d.game.taskIDs != ids || d.game.waitCounters["SOUNDBRICKUPP"] != 0 {
		t.Fatal("competing instance identity/reset")
	}
	for i := 0; i < 6; i++ {
		d.game.runTasks()
	}
	if len(d.childFires) != 2 || d.game.tasks[0] != nil {
		t.Fatal("surviving instance")
	}
}
