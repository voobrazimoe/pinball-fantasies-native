package stones

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"image/png"
	"os"
	"path/filepath"
	"pinballfantasies/internal/audio"
	"pinballfantasies/internal/physics"
	"pinballfantasies/internal/testinputs"
	"testing"
)

func game(t *testing.T) *Game {
	t.Helper()
	testinputs.Require(t, "../../TABLE4.PRG")
	b, e := os.ReadFile("../../TABLE4.PRG")
	if e != nil {
		t.Fatal(e)
	}
	v, e := physics.DecodeStones(b)
	if e != nil {
		t.Fatal(e)
	}
	return New(v, b)
}
func tasks(g *Game, n int) {
	for i := 0; i < n; i++ {
		g.runTasks()
	}
}
func checkpoint(t *testing.T, g *Game, name string) {
	t.Helper()
	dir := os.Getenv("PF10_CHECKPOINT_DIR")
	if dir == "" {
		return
	}
	os.MkdirAll(dir, 0755)
	f, e := os.Create(filepath.Join(dir, name+".png"))
	if e != nil {
		t.Fatal(e)
	}
	if e = png.Encode(f, g.Frame()); e != nil {
		t.Fatal(e)
	}
	f.Close()
}
func TestStonesTargetsAndSourceWaits(t *testing.T) {
	g := game(t)
	for i := 0; i < 9; i++ {
		g.touch(i)
	}
	// TOUCHSETA four*27530; TOUCHSETB five*17520; EVENT_LIT1=100000.
	if g.Score.Uint64() != 297720 || g.Bonus.Uint64() != 5790 || !g.sbDisabled || !g.ghostFlashing || len(g.tasks) != 20 || len(g.flashes) != 64 {
		t.Fatal("STONE/BONE completion", g.Score, g.Bonus)
	}
	tasks(g, 2)
	if !g.sbDisabled {
		t.Fatal("WAIT2 early")
	}
	tasks(g, 1)
	tasks(g, 70)
	if g.sbDisabled {
		t.Fatal("WAIT70 completion")
	}
	for _, n := range stoneBone {
		if g.Lights[n] {
			t.Fatal("bank cleanup")
		}
	}
	g = game(t)
	g.touch(0)
	g.touch(0)
	if g.Score.Uint64() != 27530 {
		t.Fatal("contact guard")
	}
	tasks(g, 10)
	if !g.sbGuards[1] {
		t.Fatal("WAIT10")
	}
	tasks(g, 1)
	g.touch(0)
	if g.Score.Uint64() != 55060 {
		t.Fatal("guard rearm")
	}
}
func TestStonesKeyRIPSkillAndTowerProgression(t *testing.T) {
	g := game(t)
	g.SkillKey = 1
	g.key(1)
	g.key(2)
	g.key(3)
	if g.SkillScore.Uint64() != 1000000 || g.Score.Uint64() != 1030180 || g.Bonus.Uint64() != 3030 || !g.enabled(11) || !g.towerOpen {
		t.Fatal("KEY and skill", g.Score, g.Bonus)
	}
	if g.VaultValue.Uint64() != 664300 || g.TowerValue.Uint64() != 1223470 || g.WellValue.Uint64() != 164190 || g.Jackpot.Uint64() != 10200000 {
		t.Fatal("source skill increments")
	}
	tasks(g, 71)
	g.SkillKey = 0
	g.key(1)
	g.key(2)
	g.key(3)
	if !g.enabled(12) {
		t.Fatal("second KEY award")
	}
	g.rip(4)
	g.rip(5)
	g.rip(6)
	if !g.enabled(44) || !g.ripDisabled {
		t.Fatal("RIP kickback")
	}
	tasks(g, 70)
	if !g.ripDisabled {
		t.Fatal("RIP WAIT70 early")
	}
	tasks(g, 1)
	if g.ripDisabled {
		t.Fatal("RIP cleanup")
	}
}
func TestStonesWellMultiplierAndCaptures(t *testing.T) {
	g := game(t)
	g.gridLeft()
	g.captureWell()
	if g.Multiplier != 2 || g.Score.Uint64() != 210030 || g.Bonus.Uint64() != 26040 || g.captured != "WELL" {
		t.Fatal("well multiplier", g.Score, g.Bonus, g.Multiplier)
	}
	g.ejectTask("WELL")
	tasks(g, 10)
	if !g.Physics.Ball.Hold {
		t.Fatal("10-WAIT early")
	}
	tasks(g, 1)
	if g.Physics.Ball.Hold || g.Physics.Ball.VX != -800 || g.Physics.Ball.VY != 2000 || g.Physics.Ball.High {
		t.Fatal("well eject")
	}
	for _, q := range []struct {
		name     string
		x, y, vy int16
		high     bool
	}{{"TOWER", 141, 143, -4000, true}, {"VAULT", 2, 532, -2880, false}} {
		g = game(t)
		g.Physics.Tilted = true
		if q.name == "TOWER" {
			g.captureTower()
		} else {
			g.captureVault()
		}
		g.ejectTask(q.name)
		tasks(g, 11)
		if g.Physics.Ball.PixelX != q.x || g.Physics.Ball.PixelY != q.y || g.Physics.Ball.VY != q.vy || g.Physics.Ball.High != q.high {
			t.Fatal(q.name, g.Physics.Ball)
		}
	}
}
func TestStonesEightGhostStateMachines(t *testing.T) {
	for idx, l := range ghostEffects {
		g := game(t)
		g.GhostCounter = uint8(idx)
		g.ghostFlashing = true
		g.captureVault()
		if idx == 4 {
			g.matrixTick()
		}
		if g.GhostCounter != uint8((idx+1)%8) {
			t.Fatal("ghost order", l)
		}
		switch idx {
		case 0:
			if g.Score.Uint64() != 5500000 {
				t.Fatal("BATMAN", g.Score)
			}
		case 1:
			if !g.TowerHunt || g.TowerStage != 1 || g.timers[5] != 2400 {
				t.Fatal("TOWERHUNT")
			}
		case 2:
			if !g.enabled(8) {
				t.Fatal("SMILER")
			}
		case 3:
			if g.Score.Uint64() != 10500000 {
				t.Fatal("REDDEVIL", g.Score)
			}
		case 4:
			if !g.Special || !g.GhostHunt || !g.enabled(9) {
				t.Fatal("GHOSTHUNT")
			}
		case 5:
			if !g.MultiDemon || g.timers[1] != 2100 || !g.enabled(15) || !g.enabled(25) || !g.enabled(18) {
				t.Fatal("MULTIDEMONS")
			}
		case 6:
			if g.Score.Uint64() != 15500000 {
				t.Fatal("MUMMYHEAD")
			}
		case 7:
			if !g.ghostInhibit {
				t.Fatal("eight ghost wrap")
			}
		}
	}
}
func TestStonesJackpotTowerAwardsAndLocks(t *testing.T) {
	g := game(t)
	g.flash(9, 18)
	g.captureTower()
	if g.Score.Uint64() != 12100000 || g.Jackpot.Uint64() != 10000000 || !g.enabled(10) {
		t.Fatal("jackpot plus tower", g.Score)
	}
	g = game(t)
	g.flash(8, 18)
	g.flash(14, 18)
	g.flash(13, 18)
	g.Bonus = number(1000)
	g.captureTower()
	if g.ExtraBalls != 1 || !g.holdBonus || g.Bonus.Uint64() != 62000 {
		t.Fatal("tower awards", g.Bonus)
	}
	for locks := 0; locks < 3; locks++ {
		g = game(t)
		g.flash(18, 18)
		if locks > 0 {
			g.light(15, true)
		}
		if locks > 1 {
			g.light(25, true)
		}
		g.scream()
		want := []uint64{5000000, 10000000, 20000000}[locks] + 30060
		if g.Score.Uint64() != want {
			t.Fatal("Multi Demon", locks, g.Score, want)
		}
	}
	g = game(t)
	g.flash(25, 18)
	g.captureWell()
	if !g.Lights[25] || g.Physics.Ball.PixelX != 302 || g.Physics.Ball.PixelY != 535 || g.Physics.Ball.Hold || !g.holdMulti {
		t.Fatal("well lock")
	}
}
func TestStonesModesScreamsAndTimers(t *testing.T) {
	g := game(t)
	g.GhostHunt = true
	g.addGhost()
	g.Grim = true
	g.addGrim()
	if g.GhostTotal.Uint64() != 1000000 || g.GrimTotal.Uint64() != 5000000 {
		t.Fatal("mode additions")
	}
	g = game(t)
	g.gridLeft()
	g.scream()
	g.loop()
	if g.ComboStage != 0 || g.Score.Uint64() != 5050120 || g.Screams != 2 {
		t.Fatal("grid scream loop combo", g.Score, g.Screams)
	}
	g = game(t)
	g.Screams = 9
	g.scream()
	if g.Screams != 10 || g.NextJump != 20 || !g.enabled(8) {
		t.Fatal("Scream ten extra ball")
	}
	g.Screams = 19
	g.scream()
	if !g.enabled(12) || g.NextJump != 30 {
		t.Fatal("later tens 5 million")
	}
	g = game(t)
	g.gridLeft()
	for i := 0; i < 359; i++ {
		g.updateCounters()
	}
	if g.timers[2] != 91 {
		t.Fatal("SULP timing")
	}
	g.updateCounters()
	for i := 0; i < 90; i++ {
		g.updateCounters()
	}
	if g.enabled(19) || g.enabled(17) {
		t.Fatal("SULP expiry")
	}
}
func TestStonesBonusNewBallHighAndTilt(t *testing.T) {
	g := game(t)
	g.Bonus = number(1000)
	g.Multiplier = 4
	g.Screams = 2
	g.GhostTotal = number(1000000)
	g.GrimTotal = number(5000000)
	g.drain()
	for i := 0; i < 5000 && g.Phase == BallLost; i++ {
		if e := g.Sync(physics.Inputs{}); e != nil {
			t.Fatal(e)
		}
	}
	if g.Score.Uint64() != 6204000 || g.BallNumber != 2 || g.Phase != NewBall || g.Physics.Ball.PixelX != 282 {
		t.Fatal("source bonus flow", g.Score, g.Phase, g.BallNumber, g.Physics.Ball)
	}
	tasks(g, int(80-g.waits["SETBALL"]))
	if !g.Physics.Ball.Hold {
		t.Fatal("SETBALL80 early")
	}
	tasks(g, 1)
	if g.Physics.Ball.PixelX != 297 || g.Physics.Ball.Hold {
		t.Fatal("SETBALL81")
	}
	g = game(t)
	g.Score = number(100000001)
	g.SetHighScore(number(100000000))
	g.matrixTick()
	if g.beaten {
		t.Fatal("chute high guard")
	}
	g.inChute = false
	g.matrixTick()
	if !g.beaten || g.ExtraBalls != 1 {
		t.Fatal("top score extra ball")
	}
	g = game(t)
	g.inChute = false
	for i := 0; i < 3; i++ {
		g.afterTargets(physics.Inputs{Tilt: true})
		g.afterTargets(physics.Inputs{})
	}
	if !g.Physics.Tilted || g.Physics.AllowFlip {
		t.Fatal("source tilt edges")
	}
	checkpoint(t, g, "tilt")
}
func TestStonesIndependentBitmapAndAudioFixtures(t *testing.T) {
	g := game(t)
	b, e := os.ReadFile("../../analysis/pf10-assets.json")
	if e != nil {
		t.Fatal(e)
	}
	var f struct {
		Animations  map[string][]string
		Tower       map[string]string `json:"tower_windows"`
		Samples     []string          `json:"sample_hashes"`
		SampleStart int               `json:"sample_start"`
	}
	if e = json.Unmarshal(b, &f); e != nil {
		t.Fatal(e)
	}
	for name, hashes := range f.Animations {
		g.Display.Clear()
		g.Display.Begin("_ANIMATION", []string{name})
		for i, want := range hashes {
			g.Display.Bitmap(g.Display.Content.Animations[name].Offsets[i])
			raw := make([]byte, 2560)
			for j, on := range g.Display.Dots {
				if on {
					raw[j] = 1
				}
			}
			if fmt.Sprintf("%x", sha256.Sum256(raw)) != want {
				t.Fatal("independent animation", name, i)
			}
		}
	}
	for row, want := range f.Tower {
		var y int
		fmt.Sscan(row, &y)
		g.Display.TowerWindow(0x4a1f0, y)
		raw := make([]byte, 2560)
		for j, on := range g.Display.Dots {
			if on {
				raw[j] = 1
			}
		}
		if fmt.Sprintf("%x", sha256.Sum256(raw)) != want {
			t.Fatal("tower packed window", row)
		}
	}
	testinputs.Require(t, "../../TABLE4.MOD")
	b, e = os.ReadFile("../../TABLE4.MOD")
	if e != nil {
		t.Fatal(e)
	}
	m, e := audio.DecodeStones(b)
	if e != nil {
		t.Fatal(e)
	}
	if len(m.Orders) != 66 || len(m.Patterns) != 64 || f.SampleStart != 66620 {
		t.Fatal("module physical extent")
	}
	for i, s := range m.Samples {
		raw := make([]byte, len(s.PCM))
		for j, v := range s.PCM {
			raw[j] = byte(v)
		}
		if fmt.Sprintf("%x", sha256.Sum256(raw)) != f.Samples[i] {
			t.Fatal("sample offset", i)
		}
	}
	g.AttachAudio(m)
	for i := 0; i < 120; i++ {
		g.audioTick()
	}
	if len(g.PCM()) != (audio.Rate/71)*4 && len(g.PCM()) != (audio.Rate/71+1)*4 {
		t.Fatal("stereo PCM")
	}
}
func TestStonesSourceVisualCheckpoints(t *testing.T) {
	g := game(t)
	checkpoint(t, g, "initial-chute")
	g.Physics.SpringPosition = 32
	checkpoint(t, g, "charged-spring")
	g.beginMatrix("RIPEFFTS")
	for i := 0; i < 20; i++ {
		g.matrixTick()
	}
	checkpoint(t, g, "rule-text")
	g.beginMatrix("JACKPOTTS")
	for i := 0; i < 20; i++ {
		g.matrixTick()
	}
	checkpoint(t, g, "bitmap-animation")
	g = game(t)
	g.captureVault()
	checkpoint(t, g, "capture-vault")
	g = game(t)
	g.Bonus = number(1000)
	g.drain()
	for i := 0; i < 150; i++ {
		g.Sync(physics.Inputs{})
	}
	checkpoint(t, g, "bonus")
	g.beginMatrix("URBANOVERTS")
	for i := 0; i < 8; i++ {
		g.matrixTick()
	}
	checkpoint(t, g, "game-over")
}

func TestStonesSourcePaletteChronology(t *testing.T) {
	g := game(t)
	if string(g.Display.Content.Texts["PL_TEXT"]) != "STONES N BONES\x00" {
		t.Fatal("original double-quoted title")
	}
	if len(programs.LampOrder) != 44 || len(g.Display.Content.LampFlash) != 19 {
		t.Fatal("source lamp declarations")
	}
	// LON38 first color17,23,30: percentage conversion then OFF half DAC.
	if got := g.palette[112*3 : 112*3+3]; fmt.Sprint(got) != "[20 28 36]" {
		t.Fatal("SLACK_LIGHTS", got)
	}
	g.light(33, true)
	before := g.palette[80*3]
	g.Display.On = false
	g.matrixPalette(false)
	if g.palette[80*3] != 48 || before == 48 {
		t.Fatal("MATRIXOFF overlapping MUMMY", before, g.palette[80*3])
	}
	g.light(33, true)
	if g.palette[80*3] != before {
		t.Fatal("later LON wins")
	}
	g.Display.On = true
	g.matrixPalette(false)
	if g.palette[80*3] != before {
		t.Fatal("MATRIXON writes only79")
	}
	if g.palette[79*3] != 243 {
		t.Fatal("source matrix RGB", g.palette[79*3])
	}
}

func TestStonesSourceNewBallDelay(t *testing.T) {
	g := game(t)
	g.Phase = BallLost
	g.changeBall()
	tasks(g, 30)
	if g.Phase != BallLost {
		t.Fatal("NEW_BALL_TASK WAITSYNCS30 early")
	}
	tasks(g, 1)
	if g.Phase != NewBall || g.Physics.Ball.PixelX != 282 {
		t.Fatal("NEW_BALL_TASK completion")
	}
	g = game(t)
	g.Physics.SpringValid = true
	for i := 0; i < 32; i++ {
		g.afterTargets(physics.Inputs{Down: true})
	}
	if g.Physics.SpringPosition != 32 {
		t.Fatal("source spring saturation")
	}
	g.Release(32, 0)
	if g.Physics.Ball.VY != -5312 {
		t.Fatal("source charge/release")
	}
}
