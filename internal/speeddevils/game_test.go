package speeddevils

import (
	"os"
	"pinballfantasies/internal/testinputs"

	"pinballfantasies/internal/physics"

	"testing"
)

func testGame(t *testing.T) *Game {
	t.Helper()
	testinputs.Require(t, "../../TABLE2.PRG")
	b, e := os.ReadFile("../../TABLE2.PRG")
	if e != nil {
		t.Fatal(e)
	}
	table, e := physics.DecodeSpeedDevils(b)
	if e != nil {
		t.Fatal(e)
	}
	return New(table, b)
}
func quiet(g *Game) { g.Physics.Ball.Hold = true; g.Physics.SetBall(400, 0, 0, 0, false) }
func advance(t *testing.T, g *Game, n int) {
	t.Helper()
	for i := 0; i < n; i++ {
		quiet(g)
		if e := g.Sync(physics.Inputs{}); e != nil {
			t.Fatal(e)
		}
	}
}
func TestBURNINAndGear(t *testing.T) {
	g := testGame(t)
	g.touch(0)
	if g.Score.Uint64() != 7510 || g.Bonus.Uint64() != 550 || !g.Lights[16] {
		t.Fatal("bricka_b / bur_group", g.Score, g.Bonus)
	}
	g.touch(0)
	if g.Score.Uint64() != 7510 {
		t.Fatal("touch cooldown")
	}
	g.touch(1)
	g.touch(2)
	if !g.Lights[24] || g.Jackpot.Uint64() != 5100000 {
		t.Fatal("lit_A", g.Jackpot)
	}
	advance(t, g, 41)
	for i := 16; i < 19; i++ {
		if g.Lights[i] {
			t.Fatal("end_A")
		}
	}
	g.touch(3)
	g.touch(4)
	g.touch(5)
	if g.Score.Uint64() != 45090 || g.Bonus.Uint64() != 3360 || !g.Lights[25] {
		t.Fatal("nin_group", g.Score, g.Bonus)
	}
	g.Lights[22], g.Lights[23] = true, true
	if !g.checkGear() {
		t.Fatal("GEAR complete")
	}
	if g.Gear != 1 || g.PositionAvailable != 2 || g.Score.Uint64() != 545090 || g.Bonus.Uint64() != 28360 {
		t.Fatal("shift_up_man", g.Gear, g.Score, g.Bonus)
	}
	advance(t, g, 46)
	for i := 22; i <= 25; i++ {
		if g.Lamps[i] {
			t.Fatal("stop_gear")
		}
	}
	g.Gear = 5
	for n := 22; n <= 25; n++ {
		g.Lights[n] = true
	}
	g.checkGear()
	advance(t, g, 30)
	if g.Gear != 0 || !g.Lights[2] {
		t.Fatal("gearkasse hold bonus")
	}
}

func TestShiftRotationAndBallLossHighScore(t *testing.T) {
	g := testGame(t)
	// CHECK_SHIFTKEYS moves the second lamp into the first, third into
	// second and saved first into third, for BUR, NIN and PIT.
	for _, base := range []int{16, 19, 6} {
		g.light(base, true)
	}
	g.afterTargets(physics.Inputs{Left: true})
	for _, base := range []int{16, 19, 6} {
		if g.Lights[base] || g.Lights[base+1] || !g.Lights[base+2] {
			t.Fatal("CHECK_SHIFTKEYS rotation", base)
		}
	}
	g.afterTargets(physics.Inputs{Left: true})
	if !g.Lights[18] {
		t.Fatal("rotation must use key-down edge")
	}
	g.SetHighScore(number(1000))
	g.Score = number(1001)
	g.Special = true
	if g.checkHighScore() {
		t.Fatal("CHECKHIGHSCORE chute/mode guards")
	}
	// The separate ball-loss _BEATEN_MATRIX has neither guard.
	// It is reached only after a nonzero bonus passes through DO_FLORPA.
	g.Bonus = number(10)
	g.drain()
	for i := 0; i < 1200 && !g.Lights[55]; i++ {
		advance(t, g, 1)
	}
	if !g.beaten || !g.Lights[55] || g.beatHighScore() {
		t.Fatal("ball-loss high score must award exactly once")
	}
}

func TestSinglePlayerMatrixWait(t *testing.T) {
	g := testGame(t)
	for i, c := range programs.Commands {
		if c.Op != "_WAITIFMULTI" {
			continue
		}
		g.matrix = matrix{active: true, next: i}
		g.matrixDispatch()
		if g.matrix.remaining != 2 || g.matrix.next != i+1 {
			t.Fatal("FANTASIE WAITIFMULTI SISA=2", g.matrix)
		}
		g.matrixTick()
		if g.matrix.remaining != 1 || g.matrix.next != i+1 {
			t.Fatal("WAITRUT first sync")
		}
		g.matrixTick()
		if g.matrix.next == i+1 {
			t.Fatal("WAITRUT second sync did not dispatch")
		}
		return
	}
	t.Fatal("source WAITIFMULTI missing")
}
func TestPITMultiplier(t *testing.T) {
	g := testGame(t)
	g.pit(0)
	g.pit(1)
	g.pit(2)
	if g.MBStock != 1 || !g.Lights[4] || g.Score.Uint64() != 40080 || g.Bonus.Uint64() != 4040 {
		t.Fatal("enable_bonus", g.Score, g.Bonus)
	}
	g.offroadLane()
	if g.Multiplier != 2 || g.MBStock != 0 || g.Lights[4] || !g.Lights[42] || g.Score.Uint64() != 60120 || g.Bonus.Uint64() != 6130 {
		t.Fatal("do_mb", g.Score, g.Bonus)
	}
	advance(t, g, 40)
	for i := 6; i <= 8; i++ {
		if g.Lights[i] {
			t.Fatal("turn_pit_off")
		}
	}
	g.MBCollected = 8
	g.pit(0)
	g.pit(1)
	g.pit(2)
	if g.MBStock != 0 || g.Score.Uint64() != 1_100_200 {
		t.Fatal("full bonus million", g.Score)
	}
}
func TestMilesLoopsPositionAndSpeed(t *testing.T) {
	g := testGame(t)
	g.loop(true)
	if g.Miles != 2 || g.Score.Uint64() != 25000 || g.loopHigh != 390 {
		t.Fatal("first mile repeats", g.Miles, g.Score)
	}
	g.loop(false)
	if g.Miles != 3 || g.Score.Uint64() != 1050000 || !g.loopL || !g.loopH {
		t.Fatal("two_in_5", g.Score)
	}
	g.jump = true
	g.speedometer()
	if g.Speed != 1 || g.Score.Uint64() != 1300000 || !g.Lights[56] {
		t.Fatal("speedometer", g.Speed, g.Score)
	}
	g.PositionAvailable = 10
	for i := 0; i < 10; i++ {
		g.twoLoops()
	}
	if g.Position != 10 || !g.Lights[9] || !g.Lights[15] || g.jackDown != 1200 {
		t.Fatal("all_lit")
	}
	g.Miles = 19
	g.loop(true)
	if !g.Lights[1] || g.Miles != 20 {
		t.Fatal("litxball")
	}
	g.Miles = 29
	g.loop(true)
	if g.NextJump != 50 || !g.Lights[5] {
		t.Fatal("next jump")
	}
}
func TestJumpJackpotCarPartsAndExtraBall(t *testing.T) {
	g := testGame(t)
	g.Lights[5], g.Lights[15], g.Lights[13] = true, true, true
	g.lastArea = "BYGEL12"
	g.jumpLane()
	if g.Score.Uint64() != 15600000 || g.Jackpot.Uint64() != 5100000 || !g.Lights[3] || !g.Lights[53] {
		t.Fatal("jump jackpot part", g.Score, g.Jackpot)
	}
	advance(t, g, 1201)
	if g.Lights[3] {
		t.Fatal("turnoff_super")
	}
	g.Lights[1] = true
	g.lastArea = "BYGEL17"
	g.jumpLane()
	if !g.Lights[55] || g.Lights[1] {
		t.Fatal("jump under / xball")
	}
	g.Lights[2] = true
	g.pitstop()
	if !g.HoldBonus || g.Lights[2] {
		t.Fatal("holdbonus")
	}
	// SNACK_HOLE_TASK -> secondary wait -> eject, then 60-sync reenable.
	for i := 0; i < 22; i++ {
		g.runTasks()
	}
	if g.Physics.Ball.PixelX != 256 || g.Physics.Ball.PixelY != 41 {
		t.Fatal("SNACK capture")
	}
	for i := 0; i < 59; i++ {
		g.runTasks()
	}
	if g.Physics.Ball.Hold || g.Physics.Ball.VX != -2100 || g.Physics.Ball.VY != 800 {
		t.Fatal("SNACK eject", g.Physics.Ball)
	}

}
func TestModesAndBonusNewBall(t *testing.T) {
	g := testGame(t)
	g.startOffRoad()
	advance(t, g, 600)
	if !g.OffRoad || g.ModeTime == 0 {
		t.Fatal("OffRoad countdown")
	}
	g.addOffRoad()
	g.touch(0)
	if g.OffRoadTotal.Uint64() != 200000 {
		t.Fatal("addoffroad")
	}
	g.startTurbo()
	advance(t, g, 2300)
	if !g.Turbo {
		t.Fatal("Goliat deferred turbo")
	}
	g.addTurbo()
	if g.TurboTotal.Uint64() != 5000000 {
		t.Fatal("addTurboMode")
	}
	g.Bonus = number(1000)
	g.Multiplier = 2
	g.Miles = 2
	g.OffRoadTotal = number(100000)
	g.TurboTotal = number(5000000)
	g.Score = Decimal{}
	g.drain()
	for i := 0; i < 3000 && g.BallNumber == 1; i++ {
		advance(t, g, 1)
	}
	// BONUS_X applies before miles and mode totals: 2,000+200,000+100,000+5,000,000.
	if g.BallNumber != 2 || g.Score.Uint64() != 5302000 {
		t.Fatal("bonus/new ball", g.BallNumber, g.Score, g.Bonus, g.matrix.op)
	}
	for i := 0; i < 150 && g.Phase != Playing; i++ {
		advance(t, g, 1)
	}
	if g.Phase != Playing {
		t.Fatal("SETBALL task")
	}
}

func TestHighScoreExtraBallAndGameOver(t *testing.T) {
	g := testGame(t)
	g.SetHighScore(number(1000))
	g.Score = number(1001)
	quiet(g)
	advance(t, g, 2)
	if g.Lights[55] || g.beaten {
		t.Fatal("top score gated in chute")
	}
	g.lastArea = "NEDSLAPP"
	g.area("CLOSE1")
	g.matrix.active = false
	advance(t, g, 3)
	if !g.Lights[55] || !g.beaten {
		t.Fatal("_DOBEATEN must set original shoot-again lamp")
	}
	g.drain()
	for i := 0; i < 3000 && g.Phase != NewBall; i++ {
		advance(t, g, 1)
	}
	if g.BallNumber != 1 || g.Phase != NewBall {
		t.Fatal("top-score extra ball", g.BallNumber, g.Phase)
	}
	advance(t, g, 81)
	if g.Lights[55] {
		t.Fatal("shoot again consumed")
	}
	g.totalBalls = 1
	g.BallNumber = 1
	g.Score = number(12345)
	g.Bonus = Decimal{}
	g.Miles = 0
	g.OffRoadTotal = Decimal{}
	g.TurboTotal = Decimal{}
	g.drain()
	for i := 0; i < 3000 && g.Phase != GameOver; i++ {
		advance(t, g, 1)
	}
	if g.matchBall && g.Phase != GameOver {
		g.drain()
		for i := 0; i < 3000 && g.Phase != GameOver; i++ {
			advance(t, g, 1)
		}
	}
	if g.Phase != GameOver {
		t.Fatal("out of balls", g.Phase, g.matrix.op, g.matchLast)
	}
	steps := 0
	h := testGame(t)
	h.beginMatrix("OUT_OF_BALLSTS")
	for i := 0; i < 3000 && h.Phase != GameOver; i++ {
		advance(t, h, 1)
		for _, e := range h.Events {
			if e.Kind == "MatchStep" {
				steps++
			}
		}
	}
	if steps != 18 {
		t.Fatal("SDEV nof_banks=18", steps)
	}
}
