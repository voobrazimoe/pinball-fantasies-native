package partyland

import (
	"bytes"
	"fmt"
	"pinballfantasies/internal/presentation"
	"reflect"
	"testing"
)

// Temporal expectations come from FANTASIE WAITRUT, rclear1/2/3/4,
// WAITJINGLE and linked ANIM/SCROLLE, not prior native fixtures.
func TestAllSourceMatrixRoutineBoundaries(t *testing.T) {
	seed := newTestGame(t)
	allowed := map[int]bool{}
	for label, index := range timing.Labels {
		if label == "AFTERDEMOMODETS" || label == "ONCE_MORETS" || label == "GAMEOVERTS" || label == "URBANOVERTS" || label == "SHOWHIGHSTS" || label == "AFTERDEMOMODETS2" || label == "SHOWINFOTS" {
			continue
		}
		for j := index; j < len(timing.Commands) && timing.Commands[j].Op != "0"; j++ {
			if timing.Commands[j].Op == "_SETLOOP" || timing.Commands[j].Op == "_INIT_SCORE" {
				break
			}
			allowed[j] = true
		}
	}
	for index, c := range timing.Commands {
		if !allowed[index] {
			continue
		}
		if index+1 < len(timing.Commands) {
			switch timing.Commands[index+1].Op {
			case "_SETLOOP", "_INIT_SCORE", "_CHECK_HIGH":
				continue
			}
		}
		duration := 0
		switch c.Op {
		case "_WAIT":
			duration = c.Num(0)
			if duration == 0 {
				duration = 65536
			}
		case "_CLEAR1":
			duration = 1
		case "_CLEAR2":
			duration = 17
		case "_CLEAR3":
			duration = 81
		case "_CLEAR4":
			duration = 5
		case "_SCROLL":
			duration = (len(seed.Display.Content.Texts[c.Arg(0)])-21)*4 + 1
		case "_RULLGARDIN_UPP":
			duration = 16 - c.Num(1)
		case "_RULLGARDIN_NED":
			duration = 13 + c.Num(1)
		case "_ANIMATION":
			a := seed.Display.Content.Animations[c.Arg(0)]
			offset, loops := 0, int(a.Header[1])
			duration = 1
			for {
				if offset == int(a.Header[2]) {
					loops--
					if loops == 0 {
						break
					}
					next := int(a.Header[0]) + 4
					duration += int(a.Durations[offset/4])
					offset = next
				} else {
					duration += int(a.Durations[offset/4])
					offset += 4
				}
			}
		default:
			continue
		}
		t.Run(fmt.Sprintf("%d/%s", index, c.Op), func(t *testing.T) {
			g := newTestGame(t)
			g.matrix = matrixState{active: true, next: index, sourceProgram: false}
			g.matrixDispatch()
			for tick := 1; tick <= duration; tick++ {
				g.matrixTick()
				if tick < duration && (!g.matrix.active || g.matrix.next != index+1 || g.matrix.op != c.Op) {
					t.Fatalf("sync %d ended %s before source boundary %d", tick, c.Op, duration)
				}
				if tick == duration && g.matrix.next == index+1 {
					t.Fatalf("sync %d did not dispatch following source command", tick)
				}
			}
		})
	}
}
func TestDOSJinglePollAndReplacement(t *testing.T) {
	for _, op := range []string{"_WAITJINGLE", "_WAITJINGLE2"} {
		for index, c := range timing.Commands {
			if c.Op != op {
				continue
			}
			g := newTestGame(t)
			g.Audio.ReadyAnim = false
			g.matrix = matrixState{active: true, next: index, sourceProgram: false}
			g.matrixDispatch()
			for sync := 0; sync < 7; sync++ {
				g.matrixTick()
				if g.matrix.next != index+1 {
					t.Fatal("jingle wait bypassed ready flag", op)
				}
			}
			g.Audio.ReadyAnim = true
			g.matrixTick()
			if g.matrix.next == index+1 {
				t.Fatal("jingle ready did not dispatch on polling sync", op)
			}
			break
		}
	}
	g := newTestGame(t)
	g.Display.BeginCommand(presentation.Command{Op: "_FLASHON", Args: []string{"1"}, Nums: map[int]int{0: 1}})
	g.Display.Flash()
	if g.Display.On {
		t.Fatal("source flash setup")
	}
	g.beginMatrix("SHOWPLAYERSTS")
	if !g.Display.On {
		t.Fatal("accepted DO_MATRIX failed KILL_FLASHOR")
	}
	for tick := 0; tick < 10; tick++ {
		g.Display.Flash()
		g.matrixTick()
		if !g.Display.On {
			t.Fatal("replacement flash leaked")
		}
	}
}

func TestPartyLettersRelinquishFlashAtSourceEnd(t *testing.T) {
	// PLAND PARTY_* headers are loop=20,length=8; frame durations 1,1,2.
	// Old-BX loop behavior takes 60 calls: frames0/1 once, frame2 and frame1
	// on the subsequent 19 loops. CLEAR4 before/after adds 5 calls each.
	for _, label := range []string{"PARTY_PTS", "PARTY_ATS", "PARTY_RTS", "PARTY_TTS", "PARTY_YTS"} {
		g := newTestGame(t)
		g.Display.BeginCommand(presentation.Command{Op: "_FLASHON", Args: []string{"1"}, Nums: map[int]int{0: 1}})
		g.Display.Flash()
		g.beginMatrix(label)
		for sync := 1; sync <= 70; sync++ {
			g.Display.Flash()
			g.matrixTick()
			if !g.Display.On {
				t.Fatal(label, "inherited flash")
			}
			if sync < 70 && !g.matrix.active {
				t.Fatal(label, "early end", sync)
			}
		}
		if g.matrix.active {
			t.Fatal(label, "late end", g.matrix.op)
		}
		// The next NODOT source visit paints ordinary score; Frame itself
		// must preserve the final CLEAR4 frame until that visit occurs.
		g.matrixTick()
		first := g.Frame()
		for sync := 0; sync < 710; sync++ {
			g.Display.Flash()
			g.matrixTick()
			if !g.Display.On {
				t.Fatal(label, "idle score flash")
			}
		}
		if !bytes.Equal(first.Pix, g.Frame().Pix) {
			t.Fatal(label, "idle matrix changed")
		}
	}
}
func TestSideLaneScrollCompletesBeforeDropEjection(t *testing.T) {
	// PLAND SPINXB_RUT: XBALL_TIME=160; DROPTASK1 WAITSYNCS30;
	// DROPTASK2 WAITSYNCS(7*4-1). SCROLLE tests SI+20: 70-byte
	// SIDELANETEXT requires (70-21)*4+1=197 successful matrix calls.
	g := newTestGame(t)
	g.Physics.Ball.Hold = true
	g.spinReward(0)
	for sync := 1; sync <= 197; sync++ {
		g.runTasks()
		g.matrixTick()
		if sync < 197 && (!g.matrix.active || g.matrix.op != "_SCROLL") {
			t.Fatal("side lane replaced at sync", sync, g.matrix.op)
		}
		if !g.Physics.Ball.Hold {
			t.Fatal("drop ejected before source scroll end", sync)
		}
	}
	if g.matrix.active {
		t.Fatal("side lane missed source terminator")
	}
}

func TestSideLaneExtraBallDrainReplacesAwardWithLostBall(t *testing.T) {
	// PLAND BYGEL1/2 awards EXTRABALL2 and lights 51. LOOSE_BALL
	// explicitly drops priority before EFFECT LOSTBALL, rather than waiting
	// for EXTRABALL2TS. BALL_LOSTTS then runs CLEARIT (5 calls), PRINT13
	// (1 call), WAIT80: even with no bonus, no new ball before sync 86.
	for _, lane := range []string{"BYGEL1", "BYGEL2"} {
		t.Run(lane, func(t *testing.T) {
			g := newTestGame(t)
			g.light(39, true)
			g.trigger(lane)
			if g.ExtraBalls != 1 || !g.Lights[51] || !g.ScoreChanged {
				t.Fatal("source extra-ball award not recorded")
			}
			for tick := 0; tick < 6; tick++ {
				g.matrixTick()
			}
			if g.matrix.op != "_ANIMATION" {
				t.Fatal("extra-ball heart did not start")
			}
			g.drain()
			if g.Phase != BallLost || g.matrix.op != "_CLEAR4" {
				t.Fatal("LOOSE_BALL did not replace the award with BALL_LOSTTS")
			}
			for tick := 1; tick <= 86; tick++ {
				g.runTasks()
				g.matrixTick()
				if g.Phase != BallLost {
					t.Fatal("new round bypassed source WAIT80", tick)
				}
				if tick == 5 && g.matrix.op != "_PRINT13" || tick >= 6 && tick < 86 && g.matrix.op != "_WAIT" {
					t.Fatal("BALL_LOSTTS dispatch boundary", tick, g.matrix.op)
				}
			}
		})
	}
}
func TestPartyFlashSurvivesNewBallUntilGameOn(t *testing.T) {
	g := newTestGame(t)
	g.beginMatrix("PARTY_ONTS")
	for sync := 0; sync < 8; sync++ {
		g.Display.Flash()
		g.matrixTick()
	}
	if g.matrix.op != "_PARTYON" || !g.partyFlash {
		t.Fatal("PARTYRUT not installed")
	}
	g.newBall()
	for sync := 0; sync < 100; sync++ {
		g.Display.Flash()
		g.matrixTick()
		if g.matrix.op != "_PARTYON" || !g.matrix.active {
			t.Fatal("WHEN_NEW_BALL_RESET killed PARTYFLASH")
		}
	}
	g.lastArea = "BYGEL12"
	g.trigger("CLOSE1")
	for sync := 0; sync < 8; sync++ {
		g.Display.Flash()
		g.matrixTick()
	}
	if g.partyFlash || g.matrix.op == "_PARTYON" || !g.Display.On {
		t.Fatal("CLOSE1 did not relinquish party flash")
	}
}

// FANTASIE DOEFFECT gates DO_MATRIX on jingle/effect priority. Rejection
// still awards arithmetic, but never executes KILL_FLASHOR or writes dots.
func TestEveryRejectedSourceEffectPreservesDisplay(t *testing.T) {
	for label, spec := range timing.Effects {
		g := newTestGame(t)
		// Priority 255 source jingles are accepted even at the maximum
		// current priority; they are not rejection cases.
		if spec.Jingle != "0" && g.Display.Jingle(spec.Jingle).Priority == 255 {
			continue
		}
		if spec.Jingle == "0" && spec.Priority == 255 {
			continue
		}
		g.beginMatrix("SHOWPLAYERSTS")
		g.Display.BeginCommand(presentation.Command{Op: "_FLASHON", Args: []string{"3"}, Nums: map[int]int{0: 3}})
		g.Display.Flash()
		g.Display.Flash()
		g.Display.Flash()
		before := *g.Display
		matrixBefore := g.matrix
		g.Audio.Priority = 255
		g.effect(label, 0, 0)
		if !reflect.DeepEqual(before, *g.Display) || !reflect.DeepEqual(matrixBefore, g.matrix) {
			t.Fatal(label, "rejected effect mutated display ownership")
		}
	}
}

func TestNewBallRunsSourcePlayerPanel(t *testing.T) {
	// FANTASIE WHEN_NEW_BALL_RESET/NONEWPL calls DO_MATRIX SHOWPLAYERSTS.
	// The table program's clear and two PRINT5 routines consume real syncs.
	g := newTestGame(t)
	g.newBall()
	start := g.Display.Content.Labels["SHOWPLAYERSTS"]
	commands := g.Display.Content.Commands
	first := commands[start].Op
	clearCalls := 1
	if first == "_CLEAR4" {
		clearCalls = 5
	}
	if first == "_ANIMATION" {
		a := g.Display.Content.Animations[commands[start].Arg(0)]
		clearCalls = 1
		for i := 0; i < int(a.Header[2])/4; i++ {
			clearCalls += int(a.Durations[i])
		}
	}
	if !g.matrix.active || g.matrix.op != first {
		t.Fatal("new ball bypassed source clear", g.matrix.op)
	}
	for tick := 1; tick <= clearCalls+2; tick++ {
		g.matrixTick()
		if tick < clearCalls && g.matrix.op != first {
			t.Fatal("early clear completion", tick)
		}
		if tick >= clearCalls && tick < clearCalls+2 && g.matrix.op != "_PRINT5" {
			t.Fatal("source player/ball print boundary", tick, g.matrix.op)
		}
		if tick == clearCalls+2 && g.matrix.active {
			t.Fatal("late player panel completion", tick)
		}
	}
}

func TestIdleScoreIsDrivenBySourceSync(t *testing.T) {
	// DO_MATRIX writes BEHOVS_PROVAD; DO_THE_ANIMATIONS/ts_slut
	// retains the final dots, then the following NODOT visit installs
	// SHOWPLAYERSTS via DO_SPEC_MATRIX and paints ONLY_SCORE.
	g := newTestGame(t)
	g.inChute = false
	g.beginMatrix("SHOWPLAYERSTS")
	for g.matrix.active {
		g.matrixTick()
	}
	final := g.Display.Dots
	for i := 0; i < 3; i++ {
		g.Frame()
	}
	if final != g.Display.Dots {
		t.Fatal("host rendering advanced matrix memory")
	}
	g.matrixTick()
	first := g.Display.Content.Commands[g.Display.Content.Labels["SHOWPLAYERSTS"]+1].Op
	if !g.matrix.active || g.matrix.op != first {
		t.Fatal("NODOT did not tail-call source panel", g.matrix.op)
	}
	if final == g.Display.Dots {
		t.Fatal("ONLY_SCORE did not run on first idle sync")
	}
	for g.matrix.active {
		g.matrixTick()
	}
	g.matrixTick()
	stable := g.Display.Dots
	for i := 0; i < 710; i++ {
		g.Display.Flash()
		g.matrixTick()
		if !g.Display.On || stable != g.Display.Dots {
			t.Fatal("ordinary score changed without source event", i)
		}
	}
}

func TestChangePlayerTailCallsFollowingSourceCommand(t *testing.T) {
	// Each table's _CHANGE_PLAYER ends its ordinary player path at HU_:
	// ADD BX,2 followed by JMP [BX]. NEW_BALL_TASK runs later; there is
	// no early NODOT handoff and no extra wait visit at the branch itself.
	seed := newTestGame(t)
	for index, c := range seed.Display.Content.Commands {
		if c.Op != "_CHANGE_PLAYER" {
			continue
		}
		g := newTestGame(t)
		g.matrix = matrixState{active: true, next: index, sourceProgram: true}
		g.matrixDispatch()
		want := seed.Display.Content.Commands[index+1].Op
		if !g.matrix.active || g.matrix.op != want {
			t.Fatal("HU_ did not tail-call", index, g.matrix.op, want)
		}
		if g.BallNumber != 2 {
			t.Fatal("source player identity did not advance before NEW_BALL_TASK")
		}
	}
}

func TestSourceNoSoundJingleWaitBoundary(t *testing.T) {
	// FANTASIE WAITJINGLE2's INT66 function21 bit3 selects DEC DECCOR,
	// even when JINGLE_READY_ANIM is true. SETDECCOR supplies each table's
	// data operand; PLAND GROPB supplies 25 before EFFECT MYSTERY.
	seed := newTestGame(t)
	for index, c := range seed.Display.Content.Commands {
		if c.Op != "_WAITJINGLE2" {
			continue
		}
		g := newTestGame(t)
		n := 25
		before := seed.Display.Content.Commands[index-1]
		if before.Op == "_SETDECCOR" {
			n = before.Num(0)
		}
		g.Display.NoSound = true
		g.Display.SetJingleCountdown(uint16(n))
		g.matrix = matrixState{active: true, next: index, sourceProgram: true}
		g.matrixDispatch()
		for tick := 1; tick <= n; tick++ {
			g.matrixTick()
			if tick < n && (g.matrix.op != "_WAITJINGLE2" || g.matrix.next != index+1) {
				t.Fatal("NOSOUND ended early", index, tick, n)
			}
			if tick == n && g.matrix.next == index+1 {
				t.Fatal("NOSOUND ended late", index, tick)
			}
		}
	}
}

func TestSilentPCMIsNotDOSNoSoundCapability(t *testing.T) {
	for _, musicOff := range []bool{false, true} {
		g := newTestGame(t)
		g.Playback = nil // No PCM device; virtual tracker still runs.
		g.MusicOff = musicOff
		g.music("S_LOSTBALL")
		g.Display.SetJingleCountdown(1)
		if g.Display.NoSound || g.Display.JingleDone(false, true) {
			t.Fatal("silence or music setting selected INT66 bit3")
		}
		for tick := 0; tick < 10000 && !g.Audio.ReadyAnim; tick++ {
			g.audioTick()
		}
		if !g.Audio.ReadyAnim || g.Display.NoSound || !g.Display.JingleDone(g.Audio.ReadyAnim, true) {
			t.Fatal("virtual jingle did not finish with silent PCM", musicOff)
		}
	}
}

func TestMatchSourceFirstAndLastVisits(t *testing.T) {
	g := newTestGame(t)
	g.startMatrix("OUT_OF_BALLSTS", true)
	for g.matrix.op != "_KNACKET" {
		g.matrixTick()
	}
	total := 1 + 22*11

	retained := g.Display.Dots
	for tick := 1; tick <= total; tick++ {
		g.matrixTick()
		if tick == 1 {
			// KNACKRUT1 prints LAST_TEXT without erasing the lower matrix.
			for y := 5; y < 16; y++ {
				for x := 0; x < 160; x++ {
					if g.Display.Dots[y*160+x] != retained[y*160+x] {
						t.Fatal("KNACKRUT1 erased lower retained memory")
					}
				}
			}
		}
		if tick < total && g.matrix.op != "_KNACKET" {
			t.Fatal("early match completion", tick, total, g.matrix.op)
		}
		if tick == total && g.matrix.op == "_KNACKET" {
			t.Fatal("late match completion", tick, g.matchRemaining)
		}
	}
}

func TestLostMatchRunsSourceScoreEntryClear(t *testing.T) {
	g := newTestGame(t)
	g.Score = Number(0)
	g.matchLast = 1
	g.beginMatrix("CHECK_XXBALLTS")

	if g.ScoreEntryPending() || g.Phase == GameOver {
		t.Fatal("lost match bypassed source clear")
	}
	first := g.matrix.op
	if first != "_CLEAR1" && first != "_CLEAR4" {
		t.Fatal("missing AFTER_XXBALL clear", first)
	}
	duration := 1
	if first == "_CLEAR4" {
		duration = 5
	}
	for tick := 1; tick < duration; tick++ {
		g.matrixTick()
		if g.ScoreEntryPending() {
			t.Fatal("early qualification", tick)
		}
	}
	g.matrixTick()
	if !g.ScoreEntryPending() || g.Phase != GameOver {
		t.Fatal("CHECK_HIGH handoff missing", g.matrix.op)
	}
	g.FinishScoreEntry()
	if g.matrix.op != "_MATRIXLGT" || g.Phase != BallLost {
		t.Fatal("SPINTSEL completion did not dispatch illumination")
	}
	g.matrixTick()
	if g.Phase == GameOver {
		t.Fatal("TO_DEMO_FROM_GAME executed before its task scan")
	}
	g.runTasks()
	if g.Phase != GameOver {
		t.Fatal("source demo task missing")
	}
}

func TestIdleHighScoreReturnsZeroAndReenters(t *testing.T) {
	g := newTestGame(t)
	g.SetHighScore(Number(0))
	g.Score = Number(1)
	g.Phase = Playing
	g.inChute = false
	g.Physics.SpringValid = false
	g.Display.TakeIdlePanel(false)
	g.matrixTick()
	i := g.Display.Content.Labels["BEATENTS"]
	want := g.Display.Content.Commands[i+1].Op
	if !g.matrix.active || g.matrix.op != want {
		t.Fatal("CHECKHIGHSCORE/NODOT did not reenter NEXT_A", g.matrix.op, want)
	}
}

func TestMatchEarnedExtraBallOwnsImmediateReplacement(t *testing.T) {
	g := newTestGame(t)
	g.matchBall = true
	g.Lights[51] = true
	g.ExtraBalls = 1
	g.HoldBonus = true
	g.beginMatrix("NO_BONUS2TS")
	if g.matrix.op == "_WAITIFMULTI" || g.ScoreEntryPending() {
		t.Fatal("earned extra ball did not tail-jump shoot-again", g.matrix.op)
	}
	if g.ExtraBalls != 0 {
		t.Fatal("source extra-ball decrement missing")
	}
	if g.Bonus.Uint64() != 0 {
		t.Fatal("KOLLA_XXBALL restored held bonus owned by CHANGE_PLAYER")
	}
}

func TestChangePlayerTailJumpRetainsFlashPalette(t *testing.T) {
	for _, extra := range []bool{false, true} {
		g := newTestGame(t)
		g.Display.BeginResolved("_FLASHON", []string{"3"}, map[int]int{0: 3})
		g.Display.Visit(0, 0, g.matrixNumber)
		for i := 0; i < 3; i++ {
			g.Display.Flash()
		}
		if g.Display.On {
			t.Fatal("flash precondition")
		}
		if extra {
			g.Lights[51] = true
			g.ExtraBalls = 1
		} else {
			g.BallNumber = g.totalBalls
		}
		if g.changeBall() {
			t.Fatal("expected source program replacement")
		}
		if g.Display.On {
			t.Fatal("tail jump incorrectly invoked DO_MATRIX/KILL_FLASHOR", extra)
		}
	}
}

// FANTASIE DOADDTASK returns BX=TASKLIST[slot]. _2_DEMO_MODE
// does not save BX; HU_ therefore calls the next task, not its WAIT data.
func TestDemoCommandUsesTaskListTailAndPreservesMatrix(t *testing.T) {
	for _, occupied := range []bool{false, true} {
		g := newTestGame(t)
		calls := 0
		for i := range g.tasks {
			g.tasks[i] = nil
		}
		if occupied {
			g.tasks[1] = func() bool { calls++; return false }
		}
		for i := range g.Display.Dots {
			g.Display.Dots[i] = i%3 == 0
		}
		before := g.Display.Dots
		g.Phase = BallLost
		commands := g.Display.Content.Commands
		for i, c := range commands {
			if c.Op != "_2_DEMO_MODE" {
				continue
			}
			g.matrix.sourceProgram = true
			g.matrix.active = true
			g.matrix.next = i
			g.matrixDispatch()
			if g.matrix.active || g.matrix.op != "_2_DEMO_MODE" {
				t.Fatal("ADDTASK BX was incorrectly preserved")
			}
			if g.Display.Dots != before || g.Phase != BallLost {
				t.Fatal("dispatch cleared pixels or ran demo task")
			}
			if (calls == 1) != occupied {
				t.Fatal("HU_ did not call following task slot", calls)
			}
			g.runTasks()
			if g.Phase != GameOver {
				t.Fatal("demo task missing")
			}
			break
		}
	}
}
