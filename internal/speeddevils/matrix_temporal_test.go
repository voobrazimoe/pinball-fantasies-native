package speeddevils

import (
	"fmt"
	"pinballfantasies/internal/presentation"
	"reflect"
	"testing"
)

// Temporal expectations come from FANTASIE WAITRUT, rclear1/2/3/4,
// WAITJINGLE and linked ANIM/SCROLLE, not prior native fixtures.
func TestAllSourceMatrixRoutineBoundaries(t *testing.T) {
	seed := testGame(t)
	allowed := map[int]bool{}
	for label, index := range programs.Labels {
		if label == "AFTERDEMOMODETS" || label == "ONCE_MORETS" || label == "GAMEOVERTS" || label == "URBANOVERTS" || label == "SHOWHIGHSTS" || label == "AFTERDEMOMODETS2" || label == "SHOWINFOTS" {
			continue
		}
		for j := index; j < len(programs.Commands) && programs.Commands[j].Op != "0"; j++ {
			if programs.Commands[j].Op == "_SETLOOP" || programs.Commands[j].Op == "_INIT_SCORE" {
				break
			}
			allowed[j] = true
		}
	}
	for index, c := range programs.Commands {
		if !allowed[index] {
			continue
		}
		if index+1 < len(programs.Commands) {
			switch programs.Commands[index+1].Op {
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
			g := testGame(t)
			g.matrix = matrix{active: true, next: index}
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
		for index, c := range programs.Commands {
			if c.Op != op {
				continue
			}
			g := testGame(t)
			g.music.ReadyAnim = false
			g.matrix = matrix{active: true, next: index}
			g.matrixDispatch()
			for sync := 0; sync < 7; sync++ {
				g.matrixTick()
				if g.matrix.next != index+1 {
					t.Fatal("jingle wait bypassed ready flag", op)
				}
			}
			g.music.ReadyAnim = true
			g.matrixTick()
			if g.matrix.next == index+1 {
				t.Fatal("jingle ready did not dispatch on polling sync", op)
			}
			break
		}
	}
	g := testGame(t)
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

// FANTASIE DOEFFECT gates DO_MATRIX on jingle/effect priority. Rejection
// still awards arithmetic, but never executes KILL_FLASHOR or writes dots.
func TestEveryRejectedSourceEffectPreservesDisplay(t *testing.T) {
	for label, spec := range programs.Effects {
		g := testGame(t)
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
		g.music.Priority = 255
		g.effect(label)
		if !reflect.DeepEqual(before, *g.Display) || !reflect.DeepEqual(matrixBefore, g.matrix) {
			t.Fatal(label, "rejected effect mutated display ownership")
		}
	}
}

func TestNewBallRunsSourcePlayerPanel(t *testing.T) {
	// FANTASIE WHEN_NEW_BALL_RESET/NONEWPL calls DO_MATRIX SHOWPLAYERSTS.
	// The table program's clear and two PRINT5 routines consume real syncs.
	g := testGame(t)
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
	g := testGame(t)
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
	seed := testGame(t)
	for index, c := range seed.Display.Content.Commands {
		if c.Op != "_CHANGE_PLAYER" {
			continue
		}
		g := testGame(t)
		g.matrix = matrix{active: true, next: index}
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
	seed := testGame(t)
	for index, c := range seed.Display.Content.Commands {
		if c.Op != "_WAITJINGLE2" {
			continue
		}
		g := testGame(t)
		n := 25
		before := seed.Display.Content.Commands[index-1]
		if before.Op == "_SETDECCOR" {
			n = before.Num(0)
		}
		g.Display.NoSound = true
		g.Display.SetJingleCountdown(uint16(n))
		g.matrix = matrix{active: true, next: index}
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

// SDEV _SHOOT_AGAIN_ONN writes only the player digit and WAITRUT1.
// Award code owns lamp55; a matrix print must not invent an extra-ball lamp.
func TestShootAgainDigitDoesNotAwardLamp(t *testing.T) {
	for _, lit := range []bool{false, true} {
		g := testGame(t)
		g.Session.CurrentPlayer = 2
		g.light(55, lit)
		found := false
		for i, c := range programs.Commands {
			if c.Op != "_SHOOT_AGAIN_ONN" {
				continue
			}
			found = true
			g.matrix = matrix{active: true, next: i}
			g.matrixDispatch()
			if g.Lights[55] != lit || g.Lamps[55] != lit {
				t.Fatal("digit handler mutated award lamp")
			}
			if g.Display.Content.Texts["SHOOT_AGAIN_TEXT"][19] != '9' {
				t.Fatal("source player digit")
			}
			if g.matrix.remaining != 1 {
				t.Fatal("source WAITRUT1")
			}
			break
		}
		if !found {
			t.Fatal("source handler missing")
		}
	}
}

// SDEV _turnonturbo always JMPs TURBOTS in the same dispatch, including
// GetTheGoalYouFool's branch which postpones gameplay via Goliat.
func TestTurnOnTurboTailCallWhileAnotherModeActive(t *testing.T) {
	g := testGame(t)
	g.Special = true
	for i, c := range programs.Commands {
		if c.Op != "_TURNONTURBO" {
			continue
		}
		g.matrix = matrix{active: true, next: i}
		g.matrixDispatch()
		first := programs.Commands[programs.Labels["TURBOTS"]]
		if g.matrix.op != first.Op || g.matrix.next != programs.Labels["TURBOTS"]+1 {
			t.Fatal("tail call did not install TURBOTS first command", g.matrix)
		}
		return
	}
	t.Fatal("source handler missing")
}

func TestSilentPCMIsNotDOSNoSoundCapability(t *testing.T) {
	for _, musicOff := range []bool{false, true} {
		g := testGame(t)
		g.Playback = nil // No PCM device; virtual tracker still runs.
		g.MusicOff = musicOff
		g.Cue("S_LOSTBALL")
		g.Display.SetJingleCountdown(1)
		if g.Display.NoSound || g.Display.JingleDone(false, true) {
			t.Fatal("silence or music setting selected INT66 bit3")
		}
		for tick := 0; tick < 10000 && !g.music.ReadyAnim; tick++ {
			g.audioTick()
		}
		if !g.music.ReadyAnim || g.Display.NoSound || !g.Display.JingleDone(g.music.ReadyAnim, true) {
			t.Fatal("virtual jingle did not finish with silent PCM", musicOff)
		}
	}
}

func TestMatchSourceFirstAndLastVisits(t *testing.T) {
	g := testGame(t)
	g.beginMatrix("OUT_OF_BALLSTS")
	for g.matrix.op != "_KNACKET" {
		g.matrixTick()
	}
	total := 1 + 18*13

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
			t.Fatal("late match completion", tick, g.matrix.matchRemaining)
		}
	}
}

func TestSupermodeReturnFalsePreservesIdleRoutine(t *testing.T) {
	g := testGame(t)
	for i, c := range programs.Commands {
		if c.Op != "_RETURN_OF_THE_EVIL_SUPERMODE" {
			continue
		}
		g.matrix = matrix{active: true, next: i}
		g.inhibitCountdown = false
		g.matrixDispatch()
		if g.matrix.active {
			t.Fatal("INGA_KONSTIGHETER installed an invented WAITRUT visit")
		}
		g.matrixTick()
		if g.matrix.op == "_RETURN_OF_THE_EVIL_SUPERMODE" && g.matrix.active {
			t.Fatal("return handler did not preserve NODOT")
		}
		return
	}
	t.Fatal("missing source return handler")
}

func TestLostMatchRunsSourceScoreEntryClear(t *testing.T) {
	g := testGame(t)
	g.Score = number(0)
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
	g := testGame(t)
	g.SetHighScore(number(0))
	g.Score = number(1)
	g.Phase = Playing
	g.inChute = false
	g.Physics.SpringValid = false
	g.Display.TakeIdlePanel(false)
	g.matrixTick()
	i := programs.Labels["BEATENTS"]
	want := programs.Commands[i+1].Op
	if !g.matrix.active || g.matrix.op != want {
		t.Fatal("CHECKHIGHSCORE/NODOT did not reenter NEXT_A", g.matrix.op, want)
	}
}

func TestMatchEarnedExtraBallOwnsImmediateReplacement(t *testing.T) {
	g := testGame(t)
	g.matchBall = true
	g.Lights[55] = true
	g.HoldBonus = true
	g.beginMatrix("NO_BONUS2TS")
	if g.matrix.op == "_WAITIFMULTI" || g.ScoreEntryPending() {
		t.Fatal("earned extra ball did not tail-jump shoot-again", g.matrix.op)
	}
	if g.Bonus.Uint64() != 0 {
		t.Fatal("KOLLA_XXBALL restored held bonus owned by CHANGE_PLAYER")
	}
}

func TestChangePlayerTailJumpRetainsFlashPalette(t *testing.T) {
	for _, extra := range []bool{false, true} {
		g := testGame(t)
		g.Display.BeginResolved("_FLASHON", []string{"3"}, map[int]int{0: 3})
		g.Display.Visit(0, 0, g.matrixNumber)
		for i := 0; i < 3; i++ {
			g.Display.Flash()
		}
		if g.Display.On {
			t.Fatal("flash precondition")
		}
		if extra {
			g.Lights[55] = true
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
		g := testGame(t)
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
