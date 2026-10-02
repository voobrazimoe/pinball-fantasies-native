package gameshow

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"image"
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
	testinputs.Require(t, "../../TABLE3.PRG")
	b, e := os.ReadFile("../../TABLE3.PRG")
	if e != nil {
		t.Fatal(e)
	}
	v, e := physics.DecodeGameshow(b)
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
func matrixTicks(g *Game, n int) {
	for i := 0; i < n; i++ {
		g.clock += 1030
		g.audioTick()
		g.runTasks()
		g.matrixTick()
	}
}
func TestSHOWDollarAndDropGroups(t *testing.T) {
	g := game(t)
	closed := g.Physics.MaskRegion(true, 0, 96, 3, 25)
	g.touch(0)
	g.touch(1)
	if g.Score.Uint64() != 15000 || g.Bonus.Uint64() != 1020 || !g.Lights[18] {
		t.Fatal("dollar pair", g.Score, g.Bonus)
	}
	if fmt.Sprint(closed) == fmt.Sprint(g.Physics.MaskRegion(true, 0, 96, 3, 25)) {
		t.Fatal("wheel gate did not open")
	}
	tasks(g, 61)
	if g.Lights[2] || g.Lights[3] {
		t.Fatal("FLASHA23")
	}
	g.touch(2)
	g.touch(3)
	g.touch(4)
	g.touch(5)
	if g.Score.Uint64() != 45000 || g.Bonus.Uint64() != 3220 {
		t.Fatal("drop awards", g.Score, g.Bonus)
	}
	for n := 7; n <= 10; n++ {
		if g.Lights[n] {
			t.Fatal("drop still up", n)
		}
	}
	tasks(g, 60)
	if g.Lights[7] {
		t.Fatal("WAIT boundary")
	}
	tasks(g, 1)
	for n := 7; n <= 10; n++ {
		if !g.Lights[n] {
			t.Fatal("drop reset", n)
		}
	}
}
func TestSHOWPrizesJackpotsAndBillion(t *testing.T) {
	g := game(t)
	// Source path: right ramp -> skill ramp = TV; loop -> skill ramp = TRIP;
	// right ramp -> clockwise ramp = CAR. LASTAREA retains the approach contact.
	g.area("BYGEL5")
	g.lastArea = "CLOSE3"
	g.area("BYGEL10")
	if g.Skills != 2 || g.Prizes[0] != 1 {
		t.Fatal("TV/first skill", g.Skills, g.Prizes)
	}
	g.lastArea = "BYGEL11"
	g.area("BYGEL12")
	g.lastArea = "CLOSE3"
	g.area("BYGEL10")
	g.area("BYGEL5")
	g.area("BYGEL7")
	if g.Prizes != [6]uint8{1, 1, 1, 0, 0, 0} || !g.Lights[17] {
		t.Fatal("three lit", g.Prizes)
	}
	// YOU_WIN collects source order, independently of wheel's displayed sector.
	for i := 0; i < 3; i++ {
		g.winPrize()
	}
	if !g.TopThree || g.timers[1] != 1500 || g.Prizes[2] != 2 {
		t.Fatal("top three")
	}
	before := g.Score.Uint64()
	jack := g.Jackpot.Uint64()
	g.area("BYGEL5")
	if g.Score.Uint64() != before+10000+jack || g.timers[0] != 300 || g.Jackpot.Uint64() != 10000000 {
		t.Fatal("jackpot")
	}
	before = g.Score.Uint64()
	g.area("BYGEL7")
	if g.Score.Uint64() != before+10000+50000000 || g.timers[0] != 1 {
		t.Fatal("super jackpot")
	}
	// Source second group: skill TV window lights BOAT approach, TRIP window
	// lights HOUSE approach; clockwise CAR window lights PLANE approach.
	g.timers[3] = 1
	g.lastArea = "CLOSE3"
	g.area("BYGEL10")
	g.lastArea = "BYGEL12"
	g.area("BYGEL11")
	g.timers[3] = 0
	g.timers[5] = 1
	g.lastArea = "CLOSE3"
	g.area("BYGEL10")
	g.lastArea = "BYGEL12"
	g.area("BYGEL11")
	g.timers[6] = 1
	g.area("BYGEL7")
	g.lastArea = "BYGEL11"
	g.area("BYGEL12")
	if g.Prizes != [6]uint8{2, 2, 2, 1, 1, 1} || !g.Lights[17] {
		t.Fatal("second prize group", g.Prizes)
	}
	for i := 0; i < 3; i++ {
		g.winPrize()
	}
	if !g.AllSix {
		t.Fatal("all six")
	}
	g.cashCapture()
	if !g.BillionEnabled || g.Physics.Ball.PixelX != 304 || g.Physics.Ball.PixelY != 535 || g.Physics.Ball.Hold {
		t.Fatal("LOCK_THE_BALL")
	}
	before = g.Score.Uint64()
	g.wheelCapture()
	if g.Score.Uint64() != before+1000000000 || g.BillionEnabled || !g.Physics.Ball.Hold {
		t.Fatal("BILLION")
	}
	tasks(g, 250)
	if !g.Physics.Ball.Hold {
		t.Fatal("billion early eject")
	}
	tasks(g, 1)
	if g.TopThree || g.AllSix || g.Physics.Ball.VY != -3500 || g.Physics.Ball.Hold {
		t.Fatal("TURNOFF_BILLION")
	}
}
func TestSHOWMoneyManiaMultiplierAndExtraBall(t *testing.T) {
	g := game(t)
	g.Skills = 5
	g.anotherSkill()
	if !g.Special || !g.MoneyMania || g.LoopsAndTraps {
		t.Fatal("six skills mode")
	}
	g.addMoney(false)
	g.addMoney(true)
	if g.MoneyTotal.Uint64() != 500000 {
		t.Fatal("target variant")
	}
	g.ModeTime = 1
	g.matrix.op = "_COUNTDOWN"
	g.matrixTick()
	if g.MoneyMania || g.music.ReturnPosition != 3 || !g.Special {
		t.Fatal("countdown/ending presentation boundary")
	}
	g.Special = false
	g.Skills = 17
	g.anotherSkill()
	g.addMoney(false)
	g.addMoney(true)
	if !g.LoopsAndTraps || g.MoneyTotal.Uint64() != 1500000 {
		t.Fatal("loop variant")
	}
	g.Special = false
	g.timers[2] = 1
	for _, want := range []uint8{2, 3, 4, 6, 8, 10} {
		g.clockwise()
		if g.Multiplier != want {
			t.Fatal("BONUSTABLE", g.Multiplier, want)
		}
	}
	g.Skills = 11
	g.anotherSkill()
	if !g.Lights[11] {
		t.Fatal("twelve skills extra ball lit")
	}
	g.lastArea = "BYGEL11"
	g.area("BYGEL12")
	if g.Lights[11] || !g.Lights[31] {
		t.Fatal("loop collection")
	}
	g.changeBall()
	if g.BallNumber != 1 || !g.matrix.active {
		t.Fatal("shoot again")
	}
}
func TestSHOWCashCaptureWheelAndTimerOrdering(t *testing.T) {
	g := game(t)
	g.cashCapture()
	if g.Score.Uint64() != 500000 || g.Bonus.Uint64() != 510 || g.Jackpot.Uint64() != 10100000 || !g.Physics.Ball.Hold {
		t.Fatal("cash pot")
	}
	tasks(g, 200)
	if !g.Physics.Ball.Hold {
		t.Fatal("capture early eject")
	}
	tasks(g, 1)
	if g.Physics.Ball.Hold || g.Physics.Ball.VX != 100 || g.Physics.Ball.VY != 1700 {
		t.Fatal("RELEASE_IT")
	}
	g = game(t)
	g.timers[7] = 1
	g.cashCapture()
	if g.Score.Uint64() != 2500000 || g.CashPot5.Uint64() != 2500000 {
		t.Fatal("cash x5")
	}
	g = game(t)
	g.wheelCapture()
	if g.ScreenForce != 187 || g.spinCounter != 4 || g.spinLight != 6 {
		t.Fatal("hi-res spin seed")
	}
	for g.spinning {
		g.updateCounters()
	}
	want := spinScores[g.spinLight]
	if g.Physics.Ball.Hold != true {
		t.Fatal("wheel premature eject")
	}
	tasks(g, 100)
	if !g.Physics.Ball.Hold {
		t.Fatal("wheel wait")
	}
	tasks(g, 1)
	if g.Score.Uint64() != want || g.Physics.Ball.VY != -3500 {
		t.Fatal("END_OF_SPIN", g.Score, want)
	}
	g = game(t)
	g.timers[0] = 1
	g.timers[1] = 5
	g.timers[2] = 7
	g.timers[9] = 600
	g.updateCounters()
	if g.timers[1] != 5 || g.timers[2] != 7 || g.timers[9] != 600 {
		t.Fatal("expiry RETN source ordering")
	}
}
func TestSHOWBonusNewBallMatchAndHighScore(t *testing.T) {
	g := game(t)
	g.Bonus = number(1000)
	g.Multiplier = 3
	g.Skills = 2
	g.MoneyTotal = number(500000)
	g.drain()
	matrixTicks(g, 3500)
	if g.Score.Uint64() != 703000 || g.BallNumber != 2 || g.Phase != Playing || g.Physics.Ball.PixelX != 299 || g.Multiplier != 1 {
		t.Fatal("ball loss: 1000*3 + 2*100000 + 500000", g.Score, g.BallNumber, g.Phase)
	}
	g = game(t)
	g.TopThree = true
	g.AllSix = false
	g.Prizes = [6]uint8{2, 2, 2, 1, 0, 0}
	g.newBall()
	if g.Prizes != [6]uint8{2, 2, 2, 0, 0, 0} {
		t.Fatal("player state survival")
	}
	g = game(t)
	g.SetHighScore(number(100))
	g.Score = number(101)
	g.matrixTick()
	if g.beaten {
		t.Fatal("chute high guard")
	}
	g.inChute = false
	g.Special = true
	g.matrixTick()
	if g.beaten {
		t.Fatal("special guard")
	}
	g.Special = false
	g.matrixTick()
	if !g.beaten || !g.Lights[31] {
		t.Fatal("SHOW _DOBEATEN")
	}
	g = game(t)
	g.Score = number(123450)
	g.BallNumber = 3
	g.drain()
	matrixTicks(g, 3500)
	if g.Phase != GameOver {
		t.Fatal("final-ball gameover", g.Phase)
	}
	g = game(t)
	g.Score = number(123450)
	g.matchLast = 5
	g.beginMatrix("CHECK_XXBALLTS")
	if !g.matchBall {
		t.Fatal("match extra ball")
	}
	matrixTicks(g, 150)
	if g.BallNumber != 1 || g.Phase != Playing {
		t.Fatal("match newball", g.BallNumber, g.Phase)
	}
	g = game(t)
	g.beginMatrix("OUT_OF_BALLSTS")
	count := 0
	for i := 0; i < 220; i++ {
		g.clock += 1030
		g.matrixTick()
		for _, e := range g.Events {
			if e.Kind == "MatchStep" {
				count++
			}
		}
		g.Events = nil
	}
	if count != 15 {
		t.Fatal("15 banks / 14 syncs", count)
	}
}
func TestSHOWControlsTiltAndMatrixFixtures(t *testing.T) {
	g := game(t)
	g.Physics.Ball.Hold = true
	for i := 0; i < 32; i++ {
		if e := g.Sync(physics.Inputs{Down: true}); e != nil {
			t.Fatal(e)
		}
	}
	if g.Physics.SpringPosition != 32 {
		t.Fatal("charge")
	}
	g.Sync(physics.Inputs{Release: true})
	if g.Physics.SpringPosition != 0 || g.Physics.Ball.VY != -5312-int16(uint8(g.clock)) {
		t.Fatal("release source jitter")
	}
	g.inChute = false
	for i := 0; i < 3; i++ {
		g.Sync(physics.Inputs{Tilt: true})
		g.Sync(physics.Inputs{})
	}
	if !g.Physics.Tilted || g.Physics.AllowFlip {
		t.Fatal("tilt state machine")
	}
	for n := 1; n <= 38; n++ {
		if g.Lights[n] {
			t.Fatal("tilt lamps", n)
		}
	}
	g.wheelCapture()
	tasks(g, 31)
	if g.Physics.Ball.Hold || g.Physics.Ball.VY != -3500 {
		t.Fatal("tilted capture eject")
	}
	b, e := os.ReadFile("../../analysis/pf9-assets.json")
	if e != nil {
		t.Fatal(e)
	}
	var f struct {
		Animations map[string][]string `json:"animations"`
	}
	if e = json.Unmarshal(b, &f); e != nil {
		t.Fatal(e)
	}
	g = game(t)
	for name, a := range g.Display.Content.Animations {
		g.Display.Clear()
		for i, o := range a.Offsets {
			g.Display.Bitmap(o)
			raw := make([]byte, len(g.Display.Dots))
			for j, d := range g.Display.Dots {
				if d {
					raw[j] = 1
				}
			}
			got := fmt.Sprintf("%x", sha256.Sum256(raw))
			if got != f.Animations[name][i] {
				t.Fatal("independent bitmap delta", name, i)
			}
		}
	}
	// Explicit DAC packets remain ordered writes, without per-frame reconstruction.
	p := g.Palette()
	g.light(2, true)
	lit := g.Palette()
	g.light(2, false)
	c := g.lampRecord(2)
	for i, v := range c.RGB {
		want := byte((uint16(v) * 162 >> 8) >> 1)
		want = want<<2 | want>>4
		if g.palette[3*c.Start+i] != want {
			t.Fatal("LON/LOFF DAC")
		}
	}
	if p == lit {
		t.Fatal("lamp changed no palette")
	}
}
func TestSHOWAudibleAndSilentCueClocks(t *testing.T) {
	testinputs.Require(t, "../../TABLE3.MOD")
	b, e := os.ReadFile("../../TABLE3.MOD")
	if e != nil {
		t.Fatal(e)
	}
	mod, e := audio.DecodeGameshow(b)
	if e != nil {
		t.Fatal(e)
	}
	for _, label := range []string{"S_MAIN", "S_SPRING", "SJINGLE1", "SJINGLE9", "S_LOSTBALL", "S_MYSTERY", "S_TILT"} {
		a, b := game(t), game(t)
		b.AttachAudio(mod)
		a.Cue(label)
		b.Cue(label)
		for i := 0; i < 1000; i++ {
			a.audioTick()
			b.audioTick()
			if a.music != b.music {
				t.Fatal("silent/audio clock divergence", label, i)
			}
		}
		if len(b.PCM()) == 0 {
			t.Fatal("no PCM")
		}
	}
}
func TestSHOWLogicalCheckpoints(t *testing.T) {
	dir := os.Getenv("PF9_CHECKPOINT_DIR")
	if dir == "" {
		return
	}
	if e := os.MkdirAll(dir, 0755); e != nil {
		t.Fatal(e)
	}
	save := func(name string, im *image.RGBA) {
		f, e := os.Create(filepath.Join(dir, name+".png"))
		if e != nil {
			t.Fatal(e)
		}
		if e = png.Encode(f, im); e != nil {
			t.Fatal(e)
		}
		f.Close()
	}
	g := game(t)
	save("initial-chute", g.Frame())
	g.Physics.SpringPosition = 32
	save("charged-spring", g.Frame())
	g = game(t)
	g.beginMatrix("RMTS")
	matrixTicks(g, 20)
	save("matrix-text", g.Frame())
	g.beginMatrix("JACKPOTTS")
	matrixTicks(g, 180)
	save("matrix-jackpot", g.Frame())
	g = game(t)
	g.inChute = false
	g.Physics.Ball.Hold = true
	for i := 0; i < 3; i++ {
		g.Sync(physics.Inputs{Tilt: true})
		g.Sync(physics.Inputs{})
	}
	matrixTicks(g, 6)
	save("tilt", g.Frame())
	g = game(t)
	g.Bonus = number(1000)
	g.drain()
	matrixTicks(g, 250)
	save("bonus", g.Frame())
}

func TestSHOWModuleStoredPatternsAndSampleBoundaries(t *testing.T) {
	testinputs.Require(t, "../../TABLE3.MOD")
	b, e := os.ReadFile("../../TABLE3.MOD")
	if e != nil {
		t.Fatal(e)
	}
	m, e := audio.DecodeGameshow(b)
	if e != nil {
		t.Fatal(e)
	}
	fdata, e := os.ReadFile("../../analysis/pf9-assets.json")
	if e != nil {
		t.Fatal(e)
	}
	var f struct {
		Start  int      `json:"sample_start"`
		Hashes []string `json:"sample_hashes"`
	}
	if e = json.Unmarshal(fdata, &f); e != nil {
		t.Fatal(e)
	}
	if b[950] != 61 || len(m.Orders) != 63 || m.Orders[61] != 62 || m.Orders[62] != 63 || len(m.Patterns) != 64 || f.Start != 1084+64*1024 {
		t.Fatal("SHOW's referenced order extent/sample boundary")
	}
	for i, s := range m.Samples {
		raw := make([]byte, len(s.PCM))
		for j, v := range s.PCM {
			raw[j] = byte(v)
		}
		if fmt.Sprintf("%x", sha256.Sum256(raw)) != f.Hashes[i] {
			t.Fatal("independent original sample", i+1)
		}
	}
	g := game(t)
	g.AttachAudio(m)
	g.Cue("S_ENDFIG")
	g.music.ReturnPosition = 62
	seenClock, seenTracker := false, false
	for i := 0; i < 1000; i++ {
		g.audioTick()
		seenClock = seenClock || g.music.Position == 62
		seenTracker = seenTracker || g.Playback.Order == 62
	}
	if !seenClock || !seenTracker {
		t.Fatal("original match return order was never visited")
	}
	// Original order 62's empty pattern falls through to order zero; no invented Bxx.
	if g.music.Position != 0 || g.Playback.Order != 0 {
		t.Fatal("original silent pattern fall-through")
	}

}
