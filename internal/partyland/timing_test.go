package partyland

import (
	"encoding/json"
	"os"
	"pinballfantasies/internal/physics"
	"pinballfantasies/internal/testinputs"
	"reflect"
	"testing"
)

type traceEntry struct {
	Tick uint64 `json:"tick"`
	Op   string `json:"op"`
}

func timingFixtures(t *testing.T) struct {
	Traces map[string][]traceEntry
	Cues   map[string]struct {
		Syncs        uint64
		TrackerTicks uint32 `json:"tracker_ticks"`
	}
} {
	t.Helper()
	var v struct {
		Traces map[string][]traceEntry
		Cues   map[string]struct {
			Syncs        uint64
			TrackerTicks uint32 `json:"tracker_ticks"`
		}
	}
	b, err := os.ReadFile(testinputs.Generated(t, "reference_pf45.py", "TABLE1.MOD", "reference/original-dos-source/PLAND.ASM"))
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(b, &v); err != nil {
		t.Fatal(err)
	}
	return v
}
func matrixEvents(g *Game) []traceEntry {
	var a []traceEntry
	for _, e := range g.Events {
		if e.Kind == "MatrixCommand" {
			a = append(a, traceEntry{e.Tick, e.Label})
		}
	}
	return a
}
func TestPF45IndependentMatrixTraces(t *testing.T) {
	fixtures := timingFixtures(t)
	for _, name := range []string{"HIDDENTS", "MYSTERYTS", "HAPPYHOURTS", "MEGALAUGHTS"} {
		t.Run(name, func(t *testing.T) {
			g := newTestGame(t)
			quiet(g)
			g.Events = nil
			switch name {
			case "HIDDENTS":
				g.hidden()
				quiet(g)
			case "MYSTERYTS":
				g.Arcade = true
				g.arcade()
				quiet(g)
			case "HAPPYHOURTS":
				g.startHappy()
			case "MEGALAUGHTS":
				g.startMega()
			}
			got := matrixEvents(g)
			want := fixtures.Traces[name]
			end := want[len(want)-1].Tick
			// Mystery finishes then its task preempts on tick320; test the complete
			// matrix independently here, and real arcade continuation separately below.
			if name == "MYSTERYTS" {
				g.tasks = [50]func() bool{}
			}
			for g.Tick < end {
				ticks(t, g, 1)
				got = append(got, matrixEvents(g)...)
			}
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("command trace\ngot  %v\nwant %v", got, want)
			}
			if name == "HAPPYHOURTS" || name == "MEGALAUGHTS" {
				if g.ModeTime != 1776 {
					t.Fatalf("countdown starts after intro: %d", g.ModeTime)
				}
				ticks(t, g, 1)
				if g.ModeTime != 1775 {
					t.Fatal("first countdown call must decrement 26 to25")
				}
			}
		})
	}
}
func TestPF45TaskSlotAndSharedWaitOrdering(t *testing.T) {
	g := newTestGame(t)
	quiet(g)
	var trace []string
	g.task(func() bool {
		trace = append(trace, "A")
		g.task(func() bool { trace = append(trace, "C"); return true }) // later slot executes now.
		return true
	})
	g.task(func() bool {
		trace = append(trace, "B")
		g.task(func() bool { trace = append(trace, "D"); return true }) // vacated earlier slot waits.
		return true
	})
	ticks(t, g, 1)
	if !reflect.DeepEqual(trace, []string{"A", "B", "C"}) {
		t.Fatal(trace)
	}
	ticks(t, g, 1)
	if !reflect.DeepEqual(trace, []string{"A", "B", "C", "D"}) {
		t.Fatal(trace)
	}
	// Two instances share ONE WAITLIST word: first invocation increments0->1,
	// second compares1 and completes. Instance one completes on tick5.
	var ready []uint64
	g.waitAt("same original macro", 1, func() { ready = append(ready, g.Tick) })
	g.waitAt("same original macro", 1, func() { ready = append(ready, g.Tick) })
	ticks(t, g, 3)
	if !reflect.DeepEqual(ready, []uint64{3, 5}) {
		t.Fatal(ready)
	}

}
func TestPF45PriorityAndCompletionOrdering(t *testing.T) {
	g := newTestGame(t)
	quiet(g)
	g.Events = nil
	g.effect("SPIN500K", 500000, 25000)
	if g.Events[0].Kind != "Music" || g.Events[1].Kind != "ScoreAwarded" {
		t.Fatal(g.Events)
	}
	// Completion at ceil(8rows*5tracker ticks*71/50)=57 native syncs.
	fixture := timingFixtures(t)
	if fixture.Cues["41"].Syncs != 57 {
		t.Fatal("module oracle changed")
	}
	g.waitForAudioDrop(45, 70)
	var at57 []string
	for i := 1; i <= 57; i++ {
		ticks(t, g, 1)
		if i < 57 && g.Audio.ReadyLogic {
			t.Fatalf("early completion at%d", i)
		}
		if i == 57 {
			for _, e := range g.Events {
				if e.Kind == "AudioComplete" || e.Kind == "TaskReady" {
					at57 = append(at57, e.Kind+":"+e.Label)
				}
			}
		}
	}
	if !reflect.DeepEqual(at57, []string{"AudioComplete:JINGLEREADY", "TaskReady:START_DROP_WHEN_READY"}) {
		t.Fatal(at57)
	}
	if g.Audio.ReadyLogic || g.dropMinimum != 65535 {
		t.Fatal("release consumes logic flag")
	}
	// Equal priority replaces; lower priority still adds score and sets EOTS.
	g = newTestGame(t)
	quiet(g)
	g.effect("TSCORE3", 5000000, 500000)
	g.effectEnded = false
	before := g.Score.Uint64()
	g.effect("DSCORE", 250000, 10030)
	if !g.effectEnded || g.Score.Uint64() != before+250000 || g.matrix.op != "_CLEAR4" || g.Audio.Position != 47 {
		t.Fatal("priority rejection changed arithmetic or preempted matrix")
	}
	g.effect("MILLION5", 5000000, 10000)
	if g.Audio.Position != 47 || g.Audio.elapsed != 0 {
		t.Fatal("equal priority must restart")
	}

}

func TestPF45ReadyLatchAndUpperBound(t *testing.T) {
	for _, tc := range []struct {
		name  string
		ready uint64
		want  uint64
	}{
		{"completion before minimum", 10, 45}, {"completion after minimum", 57, 57}, {"no completion", 0, 70},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newTestGame(t)
			quiet(g)
			g.Audio.ReadyLogic = false
			g.waitForAudioDrop(45, 70)
			var release uint64
			for i := uint64(1); i <= 70; i++ {
				if i == tc.ready {
					g.Audio.ReadyLogic = true
				}
				if tc.ready > 0 && i == tc.ready+1 {
					g.Audio.ReadyLogic = false
				} // later jingle clears flag.
				ticks(t, g, 1)
				for _, e := range g.Events {
					if e.Kind == "TaskReady" && e.Label == "START_DROP_WHEN_READY" {
						release = e.Tick
					}
				}
			}
			if release != tc.want {
				t.Fatalf("release %d want %d", release, tc.want)
			}
		})
	}
}
func TestPF45MatrixPreemptionWakesSharedEOTS(t *testing.T) {
	g := newTestGame(t)
	quiet(g)
	g.hidden()
	quiet(g)
	ticks(t, g, 10)
	g.effect("TSCORE3", 5000000, 500000) // interrupts hidden scroll; EOTS remains false.
	ticks(t, g, 320)
	if g.waitCounters["HIDDENTASK1"] != 0 {
		t.Fatal("interruption must not complete hidden waiter")
	}
	g.effect("DSCORE", 250000, 10030) // lower priority or post-jingle timing can allow;
	// Explicit source priority255 proves suppression sets the shared EOTS latch.
	g.Audio.Priority = 255
	g.effect("BYGELSETB", 10040, 1000)
	if !g.effectEnded {
		t.Fatal("DOEFFECT suppression must set EOTS")
	}
	ticks(t, g, 1)
	found := false
	for _, e := range g.Events {
		if e.Label == "HIDDENTASK0" {
			found = true
		}
	}
	if !found {
		t.Fatal("EOTS waiter did not resume on next task scan")
	}
}
func TestPF45ArcadeAudioContinuation(t *testing.T) {
	g := newTestGame(t)
	quiet(g)
	g.Arcade = true
	g.arcade()
	quiet(g)
	var completion, resume, award uint64
	for i := 0; i < 321; i++ {
		ticks(t, g, 1)
		for _, e := range g.Events {
			if e.Kind == "AudioComplete" && e.Value == 30 {
				completion = e.Tick
			}
			if e.Kind == "TaskReady" && e.Label == "WAIT_FOR_SPIN_TASK" {
				resume = e.Tick
			}
			if e.Kind == "ScoreAwarded" && e.Label != "MYSTERY" {
				award = e.Tick
			}
		}
	}
	if completion != 319 || resume != 320 || award != 320 {
		t.Fatalf("completion=%d task=%d prize=%d", completion, resume, award)
	}
}
func TestPF45JingleJumpCountAcrossPatterns(t *testing.T) {
	g := newTestGame(t)
	quiet(g)
	g.music("SJINGLE4")
	var cues []uint64
	var positions []uint64
	for i := 0; i < 1600; i++ {
		ticks(t, g, 1)
		for _, e := range g.Events {
			if e.Kind == "AudioCue" && len(cues) < 3 {
				cues = append(cues, e.Tick)
				positions = append(positions, e.Value)
			}
		}
		if len(cues) == 3 {
			break
		}
	}
	// Derived from MOD orders24 ->32 ->33 and original repeat byte3.
	f := timingFixtures(t)
	a := f.Cues["24"].TrackerTicks
	b := f.Cues["32"].TrackerTicks
	c := f.Cues["33"].TrackerTicks
	want := []uint64{uint64(a*71+49) / 50, uint64((a+b)*71+49) / 50, uint64((a+b+c)*71+49) / 50}
	if !reflect.DeepEqual(cues, want) || !reflect.DeepEqual(positions, []uint64{24, 32, 33}) || !g.Audio.ReadyAnim {
		t.Fatalf("cues=%v positions=%v want=%v", cues, positions, want)
	}
}
func TestPF45TimedLampsAndRestart(t *testing.T) {
	g := newTestGame(t)
	quiet(g)
	g.flash(21, 7, 0, false)
	var on, off []uint64
	for i := 0; i < 15; i++ {
		ticks(t, g, 1)
		for _, e := range g.Events {
			if e.Kind == "LampChanged" && e.Label == "21" {
				if e.Value == 1 {
					on = append(on, e.Tick)
				} else {
					off = append(off, e.Tick)
				}
			}
		}
	}
	if !reflect.DeepEqual(on, []uint64{1, 15}) || !reflect.DeepEqual(off, []uint64{8}) {
		t.Fatalf("on=%v off=%v", on, off)
	}
	g.waitAt("reset-site", 9, func() { t.Fatal("terminated task ran") })
	ticks(t, g, 3)
	g.newBall()
	if g.waitCounters["reset-site"] != 0 {
		t.Fatal("new ball did not reset WAITLIST")
	}
	ticks(t, g, 81)
	if g.Phase != Playing {
		t.Fatal("SETBALL completion missing")
	}
}
func TestPF45LongScriptStateAndFrames(t *testing.T) {
	a, b := newTestGame(t), newTestGame(t)
	for i := 0; i < 8000; i++ {
		if i%600 == 100 {
			a.Release(32, 0)
			b.Release(32, 0)
		}
		in := physics.Inputs{Left: i%93 < 18, Right: i%71 < 14}
		if err := a.Sync(in); err != nil {
			t.Fatal(err)
		}
		if err := b.Sync(in); err != nil {
			t.Fatal(err)
		}
		if a.Audio != b.Audio || a.Tick != b.Tick || a.ModeTime != b.ModeTime || a.Score != b.Score || a.Bonus != b.Bonus || a.Phase != b.Phase || a.Lamps != b.Lamps || !reflect.DeepEqual(a.waitCounters, b.waitCounters) || !reflect.DeepEqual(a.Events, b.Events) || !reflect.DeepEqual(a.matrix, b.matrix) {
			t.Fatalf("state diverged at%d", i)
		}
		if i%137 == 0 {
			if !reflect.DeepEqual(a.Frame().Pix, b.Frame().Pix) {
				t.Fatalf("frame diverged at%d", i)
			}
		}
	}
}

func TestPF45IndependentDrainToNextBallTrace(t *testing.T) {
	var f struct {
		BallTrace []struct {
			Tick        uint64
			Kind, Label string
		} `json:"ball_trace"`
	}
	b, err := os.ReadFile(testinputs.Generated(t, "reference_pf45.py", "TABLE1.MOD", "reference/original-dos-source/PLAND.ASM"))
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(b, &f); err != nil {
		t.Fatal(err)
	}
	g := newTestGame(t)
	quiet(g)
	g.Score = Number(1000)
	g.ScoreChanged = true
	g.Bonus = Number(1000)
	g.Multiplier = 2
	g.Cyclones = 2
	g.Events = nil
	g.drain()
	var got []struct {
		Tick        uint64
		Kind, Label string
	}
	appendEvents := func() {
		for _, e := range g.Events {
			include := e.Kind == "BallLost" || e.Kind == "NewBall" || e.Kind == "BonusMultiplied" || e.Kind == "BonusAdded" || (e.Kind == "ScoreAwarded" && e.Label == "DO_FLORPA") || (e.Kind == "Sound" && (e.Label == "SBRICKUPP" || e.Label == "SNEWBALL")) || (e.Kind == "TaskReady" && e.Label == "SETBALL")
			if include {
				got = append(got, struct {
					Tick        uint64
					Kind, Label string
				}{e.Tick, e.Kind, e.Label})
			}
		}
	}
	appendEvents()
	for i := 0; i < 600; i++ {
		ticks(t, g, 1)
		appendEvents()
	}
	if !reflect.DeepEqual(got, f.BallTrace) {
		t.Fatalf("drain checkpoints\ngot  %v\nwant %v", got, f.BallTrace)
	}
	if g.Phase != Playing || g.BallNumber != 2 || g.Score.Uint64() != 203000 {
		t.Fatal("next-ball state")
	}
}
func TestPF45EveryGameplayEffectMatrix(t *testing.T) {
	// Exercise every reachable presentation program, not only common trajectories.
	// No missing command/animation/scroll may silently fall back to an estimate.
	for label := range timing.Effects {
		t.Run(label, func(t *testing.T) {
			g := newTestGame(t)
			quiet(g)
			if label == "HAPPYHOUR" {
				g.startHappy()
			} else if label == "CRAZYSCORE" {
				g.startMega()
			} else {
				g.effect(label, 0, 0)
			}
			for i := 0; i < 2400 && g.matrix.active; i++ {
				ticks(t, g, 1)
			}
			if g.matrix.active && label != "LOSTBALL" {
				t.Fatal("effect did not complete")
			}
		})
	}
}

func TestPF45MissedMatrixBudget(t *testing.T) {
	g := newTestGame(t)
	quiet(g)
	g.startMega()
	var ready uint64
	g.waitAt("unaffected-task", 1, func() { ready = g.Tick })
	for i := 0; i < 10; i++ {
		if err := g.SyncWithMatrixBudget(physics.Inputs{}, false); err != nil {
			t.Fatal(err)
		}
	}
	if ready != 2 || g.ModeTime != 0 || g.matrix.op != "_EOSNURR" {
		t.Fatal("budget skip affected tasks or advanced matrix")
	}
	ticks(t, g, 171)
	if g.ModeTime != 1776 || g.Tick != 181 {
		t.Fatal("matrix time must count successful scans")
	}
}
func TestPF45MatrixBeforeLatePhysics(t *testing.T) {
	g := newTestGame(t)
	quiet(g)
	g.startHappy()
	// Source matrix expires during VBLANK, before the third physics pass.
	g.matrix.op = "_COUNTDOWN"
	g.ModeTime = 1
	g.Physics.BeforeLate = func() {
		g.matrixTick()
		if g.Happy {
			t.Fatal("mode still active at late-raster boundary")
		}
	}
	ticks(t, g, 1)
}

func TestPF45RejectedJingleStillSwapsRepeat(t *testing.T) {
	g := newTestGame(t)
	quiet(g)
	g.music("SJINGLE4")
	if g.Audio.JumpCount != 3 {
		t.Fatal("jackpot repeat")
	}
	if g.playJingle("S_MYSTERY") {
		t.Fatal("lower priority accepted")
	}
	if g.Audio.Position != 24 || g.Audio.JumpCount != 1 {
		t.Fatal("ASM repeat-swap quirk was cleaned up")
	}
	ticks(t, g, 392)
	if !g.Audio.ReadyAnim || g.Audio.Priority != 1 {
		t.Fatal("callback must restore GAME_MUSIC priority")
	}
}
func TestPF45PendingModeStartsAfterNextTaskScan(t *testing.T) {
	g := newTestGame(t)
	quiet(g)
	g.startHappy()
	g.MegaPending = true
	g.ModeTime = 1
	g.matrix.op = "_COUNTDOWN"
	ticks(t, g, 1)
	if g.Happy || g.Mega || g.MegaPending {
		t.Fatal("expiry did not schedule pending mode")
	}
	ticks(t, g, 400)
	if g.Mega {
		t.Fatal("WAITSYNCS400 completed before its401st call")
	}
	ticks(t, g, 1)
	if !g.Mega || g.ModeTime != 0 {
		t.Fatal("pending mode did not start intro at402")
	}
}
func TestPF45JackpotPausesAndResumesCountdown(t *testing.T) {
	g := newTestGame(t)
	quiet(g)
	g.startHappy()
	ticks(t, g, 350)
	remaining := g.ModeTime
	g.JackpotTimed = true
	g.dragon()
	quiet(g)
	ticks(t, g, 100)
	if g.ModeTime != remaining {
		t.Fatal("jackpot matrix must pause mode timer")
	}
	// Linked jackpot animation329 calls, clears5+5 and print1: COUNTDOWN2
	// installed340 scans after direct entry; resume decrements displayed seconds.
	ticks(t, g, 240)
	want := ((remaining+70)/71-1)*71 + 1
	if g.ModeTime != want || g.matrix.op != "_COUNTDOWN2" {
		t.Fatalf("resume %d want%d op%s", g.ModeTime, want, g.matrix.op)
	}
}

func TestPF45SkillMatrixInhibitsLaterEffectsOnly(t *testing.T) {
	for _, label := range []string{"tunnel", "cyclone"} {
		t.Run(label, func(t *testing.T) {
			g := newTestGame(t)
			quiet(g)
			g.SkillTime = 300
			if label == "tunnel" {
				g.tunnel()
			} else {
				g.cyclone()
			}
			if g.inhibitEffect || !g.effectEnded || g.Audio.Position != 56 || (label == "tunnel" && g.Score.Uint64() != 2250000) || (label == "cyclone" && g.Score.Uint64() != 1350000) {
				t.Fatalf("inhibition/award: flag%v eots%v pos%d score%s", g.inhibitEffect, g.effectEnded, g.Audio.Position, g.Score)
			}
			var starts []string
			for _, e := range g.Events {
				if e.Kind == "MatrixStarted" {
					starts = append(starts, e.Label)
				}
			}
			expected := "SKILLTUNNELTS"
			if label == "cyclone" {
				expected = "SKILLCYCLONETS"
			}
			for _, start := range starts {
				if start != expected {
					t.Fatalf("later award replaced skill matrix: %v", starts)
				}
			}
		})
	}
}

func TestPF45SilentOrderCursorAndReturnPosition(t *testing.T) {
	g := newTestGame(t)
	quiet(g)
	g.music("S_MAIN")
	// Original main orders1..5 each span64 rows*6 tracker ticks. Order2 starts
	// at ceil(384*71/50)=546 syncs, before the eventual B01 callback.
	ticks(t, g, 545)
	if g.Audio.Position != 1 {
		t.Fatal("early pattern transition")
	}
	ticks(t, g, 1)
	if g.Audio.Position != 2 {
		t.Fatal("module order cursor did not advance")
	}
	g.music("SJINGLE22")
	if g.Audio.ReturnPosition != 2 {
		t.Fatal("force position lost old current order")
	}
	ticks(t, g, 57)
	if g.Audio.Position != 2 || !g.Audio.ReadyLogic {
		t.Fatal("jingle did not return to saved order")
	}
}
