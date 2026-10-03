package partyland

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"pinballfantasies/internal/physics"
	"pinballfantasies/internal/presentation"
	"pinballfantasies/internal/settings"
	"pinballfantasies/internal/testinputs"
	"reflect"
	"strings"
	"testing"
)

func newTestGame(t *testing.T) *Game {
	t.Helper()
	testinputs.Require(t, "../../TABLE1.PRG")
	data, err := os.ReadFile("../../TABLE1.PRG")
	if err != nil {
		t.Fatal(err)
	}
	table, err := physics.DecodePartyLand(data)
	if err != nil {
		t.Fatal(err)
	}
	return New(table, data)
}
func ticks(t *testing.T, g *Game, n int) {
	t.Helper()
	for i := 0; i < n; i++ {
		if err := g.Sync(physics.Inputs{}); err != nil {
			t.Fatal(err)
		}
	}
}
func quiet(g *Game) { g.Physics.Ball.Hold = true; g.Physics.SetBall(100, 400, 0, 0, false) }
func check(t *testing.T, label string, got, want uint64) {
	t.Helper()
	if got != want {
		t.Fatalf("%s: got %d want %d", label, got, want)
	}
}
func fixture(t *testing.T) map[string]uint64 {
	t.Helper()
	raw, err := os.ReadFile(testinputs.Generated(t, "reference_pf4.py", "TABLE1.PRG", "reference/original-dos-source/PLAND.ASM"))
	if err != nil {
		t.Fatal(err)
	}
	var f struct{ Fixtures map[string]uint64 }
	if err = json.Unmarshal(raw, &f); err != nil {
		t.Fatal(err)
	}
	return f.Fixtures
}
func TestDecimalSourceArithmetic(t *testing.T) {
	d := Number(999999999999)
	d.Add(Number(2))
	check(t, "discard MSD carry", d.Uint64(), 1)
	d = Number(999999999999)
	d.Add(d)
	check(t, "aliased DOUBLEBONUS", d.Uint64(), 999999999998)
	d = Number(999)
	d.Add(Number(1))
	if d.String() != "000000001000" {
		t.Fatal(d)
	}
}
func TestOriginalPhysicalAwards(t *testing.T) {
	for _, c := range []struct {
		name         string
		x, y, vx, vy int16
		steps        int
		score        uint64
	}{
		{"slingshot", 56, 410, -2000, 1500, 1, 500},
		{"bumper", 211, 239, 0, 1200, 41, 1000},
	} {
		t.Run(c.name, func(t *testing.T) {
			g := newTestGame(t)
			g.Physics.SetBall(c.x, c.y, c.vx, c.vy, false)
			ticks(t, g, c.steps)
			check(t, "physical award", g.Score.Uint64(), c.score)
		})
	}
	// PF3's independently-derived flipper trajectory reaches lower duck at sync20.
	g := newTestGame(t)
	g.Physics.SetBall(115, 510, 0, 1500, false)
	for i := 0; i < 20; i++ {
		if err := g.Sync(physics.Inputs{Left: true}); err != nil {
			t.Fatal(err)
		}
	}
	check(t, "target award", g.Score.Uint64(), 7510)
	check(t, "target bonus", g.Bonus.Uint64(), 750)
	if g.Lights[54] {
		t.Fatal("DROPA3 did not lower duck")
	}
}
func TestDuckBankAndContentMasks(t *testing.T) {
	g := newTestGame(t)
	quiet(g)
	f := fixture(t)
	for i := 0; i < 3; i++ {
		g.duck(i)
	}
	check(t, "bank score", g.Score.Uint64(), f["duck_bank_score"])
	check(t, "bank bonus", g.Bonus.Uint64(), f["duck_bank_bonus"])
	if !g.snacks[0] || g.SnackNext != 1 {
		t.Fatal("ICE not qualified")
	}
	ticks(t, g, 21)
	for i := 0; i < 3; i++ {
		w := 2
		if i == 2 {
			w = 1
		}
		got := g.Physics.MaskRegion(false, 18+i, 277+18*i, w, 15)
		if !bytes.Equal(got, g.duckDown[i]) {
			t.Fatalf("duck %d down mask", i)
		}
	}
	ticks(t, g, 51)
	for i := 0; i < 3; i++ {
		if !g.Lights[52+i] || g.duckDisabled[i] {
			t.Fatalf("duck %d not reset", i)
		}
	}
}
func TestPukeFeaturesAndShift(t *testing.T) {
	g := newTestGame(t)
	quiet(g)
	f := fixture(t)
	for _, n := range []int{1, 2, 4, 5} {
		g.pukeLetter(n)
	}
	check(t, "PUKE score", g.Score.Uint64(), f["puke_first_score"])
	check(t, "PUKE bonus", g.Bonus.Uint64(), f["puke_first_bonus"])
	if !g.FiveMillion || g.Puke != 1 || !g.Lights[46] {
		t.Fatal("PUKE train/Y")
	}
	ticks(t, g, 101)
	for _, n := range []int{1, 2, 4, 5} {
		g.pukeLetter(n)
	}
	if !g.BallFeature || g.Puke != 2 {
		t.Fatal("second PUKE")
	}
	ticks(t, g, 101)
	for _, n := range []int{1, 2, 4, 5} {
		g.pukeLetter(n)
	}
	if !g.JackpotNormal || g.Puke != 3 {
		t.Fatal("third PUKE")
	}
	g = newTestGame(t)
	quiet(g)
	g.light(1, true)
	if err := g.Sync(physics.Inputs{Left: true}); err != nil {
		t.Fatal(err)
	}
	if !g.Lights[2] || g.Lights[1] {
		t.Fatal("shift rotates P to E")
	}
	if err := g.Sync(physics.Inputs{Left: true}); err != nil {
		t.Fatal(err)
	}
	if !g.Lights[2] {
		t.Fatal("held shift must not rotate again")
	}
}
func TestSkyrideReverseMultiplierAndBonusMacro(t *testing.T) {
	g := newTestGame(t)
	quiet(g)
	f := fixture(t)
	for i := 0; i < 3; i++ {
		g.skyride()
	}
	// Consecutive ramps also award PARTY_T on second completion.
	check(t, "skyride awards + T", g.Score.Uint64(), f["skyride_bank_score"]+250000)
	if !g.MB || g.Lights[47] {
		t.Fatal("MB qualification")
	}
	g.reverse()
	if g.Multiplier != 2 || g.MB || !g.Lights[47] {
		t.Fatal("MB collection")
	}
	before := g.Bonus.Uint64()
	g.pukeLetter(1)
	check(t, "ADDBONUS times two", g.Bonus.Uint64()-before, 2000)
	before = g.Bonus.Uint64()
	g.effect("BYGELSETB", 10040, 1000)
	check(t, "DOEFFECT bypasses multiplier", g.Bonus.Uint64()-before, 1000)
}
func TestOrderedReverseFeatures(t *testing.T) {
	g := newTestGame(t)
	quiet(g)
	g.Bonus = Number(1000)
	g.MB = true
	g.HB = true
	g.DB = true
	g.Events = nil
	g.reverse()
	want := []string{"RSCORE1", "MULTIBONUS", "HOLDBONUS", "DOUBLEBONUS"}
	var got []string
	for _, e := range g.Events {
		if e.Kind == "ScoreAwarded" {
			got = append(got, e.Label)
		}
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("effects %v", got)
	}
	check(t, "double after reverse and MB bonus", g.Bonus.Uint64(), 32000)
	if !g.HoldBonus || g.Multiplier != 2 || g.HB || g.DB {
		t.Fatal("reverse feature flags")
	}
}
func TestAreaDebouncePreviousCallbackAndOverlap(t *testing.T) {
	g := newTestGame(t)
	quiet(g)
	// Boundary y540 belongs to spring-invalid first, before spring-valid.
	g.Physics.SetBall(308, 532, 0, 0, false)
	g.checkAreas()
	if g.SkillTime != 300 || g.Physics.SpringValid {
		t.Fatal("area ordering")
	}
	g.SkillTime = 100
	g.checkAreas()
	if g.SkillTime != 100 {
		t.Fatal("LASTCHECK debounce")
	}
	g.Physics.SetBall(100, 400, 0, 0, false)
	g.checkAreas()
	g.Physics.SetBall(308, 532, 0, 0, false)
	g.checkAreas()
	if g.SkillTime != 300 {
		t.Fatal("reentry")
	}
	g = newTestGame(t)
	quiet(g)
	g.Physics.SetBall(102, 20, 0, 0, false)
	g.checkAreas()
	g.Physics.SetBall(210, 20, 0, 0, false)
	g.checkAreas()
	check(t, "loop from retained LASTAREA", g.Score.Uint64(), 100000)
}
func TestTunnelSkillDecayAndCycloneQuirks(t *testing.T) {
	g := newTestGame(t)
	quiet(g)
	g.SkillTime = 100
	g.tunnel()
	check(t, "skill + PARTY_P + first tunnel", g.Score.Uint64(), 2250000)
	if g.SkillTime != 0 || !g.Lights[42] || !g.Lights[14] || g.TunnelTime != 720 {
		t.Fatal("tunnel state")
	}
	// Decay is counters before targets, and turns logical 1M off.
	g.tasks = [50]func() bool{}
	quiet(g)
	ticks(t, g, 720)
	if g.Lights[14] || g.TunnelTime != 0 {
		t.Fatal("tunnel decay")
	}
	g = newTestGame(t)
	quiet(g)
	g.cyclone()
	if g.Cyclones != 2 {
		t.Fatal("first ordinary cyclone increments twice")
	}
	g.FiveX = true
	g.cyclone()
	if g.Cyclones != 7 {
		t.Fatal("5X adds five")
	}
	check(t, "cyclone awards", g.Score.Uint64(), 2600000)
	g = newTestGame(t)
	quiet(g)
	g.SkillTime = 10
	g.FiveX = true
	g.cyclone()
	if g.Cyclones != 2 || g.FiveX || !g.Lights[44] || g.SkillTime != 10 {
		t.Fatal("skill cyclone kills 5X but retains skill timer")
	}
	check(t, "skill cyclone ordering", g.Score.Uint64(), 1350000)
}
func TestModesAndDragonPriority(t *testing.T) {
	g := newTestGame(t)
	quiet(g)
	for i := 0; i < 5; i++ {
		g.party(i)
	}
	if !g.Happy || !g.JackpotTimed || g.Lights[42] {
		t.Fatal("happy start")
	}
	g.addHappy()
	g.consume(physics.Event{Kind: physics.EventSlingshotHit})
	check(t, "happy hits", g.HappyTotal.Uint64(), 2000000)
	g.BallFeature = true
	g.FiveMillion = true
	jack := g.Jackpot.Uint64()
	before := g.Score.Uint64()
	g.dragon()
	check(t, "jackpot priority", g.Score.Uint64()-before, jack)
	if !g.BallFeature || !g.FiveMillion || g.JackpotTimed || g.Jackpot.Uint64() != 10000000 {
		t.Fatal("dragon priority/reset")
	}
	g.tasks = [50]func() bool{}
	g.Physics.Ball.Hold = true
	quiet(g)
	g.ModeTime = 1
	g.modeTick()
	if g.Happy {
		t.Fatal("happy expiry")
	}
	check(t, "happy completion bonus", g.Bonus.Uint64(), 1125000)
	g = newTestGame(t)
	quiet(g)
	for i := 0; i < 5; i++ {
		g.crazy()
	}
	if !g.Mega || g.CrazyNext != 0 {
		t.Fatal("mega start")
	}
	g.addMega()
	check(t, "mega total", g.MegaTotal.Uint64(), 5000000)
}
func TestExtraBallAndBallFlow(t *testing.T) {
	g := newTestGame(t)
	g.score("BUMPER", 1000)
	g.Bonus = Number(1000)
	g.Multiplier = 2
	g.Cyclones = 2
	// Real PF3 physical drain, rather than a synthetic BallLost call.
	g.Physics.SetBall(150, 575, 0, 1024, false)
	ticks(t, g, 1)
	if g.Phase != BallLost || !g.Physics.Ball.Hold {
		t.Fatal("drain did not hold")
	}
	ticks(t, g, 800)
	if g.Phase != Playing || g.BallNumber != 2 || g.Score.Uint64() != 203000 || g.Bonus.Uint64() != 0 || g.Multiplier != 1 {
		t.Fatalf("new ball: phase=%d ball=%d score=%s bonus=%s", g.Phase, g.BallNumber, g.Score, g.Bonus)
	}
	g = newTestGame(t)
	g.BallFeature = true
	g.dragon()
	g.tasks = [50]func() bool{}
	g.drain()
	ticks(t, g, 1000)
	if g.BallNumber != 1 || g.ExtraBalls != 0 || g.Phase != Playing {
		t.Fatal("extra ball consumed numbered ball")
	}
	g = newTestGame(t)
	g.drain()
	ticks(t, g, 112)
	if g.BallNumber != 1 || g.Phase != Playing {
		t.Fatal("unscored drain should return same ball")
	}
}
func TestHeldBonusAndGameOver(t *testing.T) {
	g := newTestGame(t)
	quiet(g)
	g.Bonus = Number(1000)
	g.Multiplier = 2
	g.HoldBonus = true
	g.ScoreChanged = true
	g.drain()
	ticks(t, g, 600)
	check(t, "held multiplied bonus", g.Bonus.Uint64(), 2000)
	if g.BallNumber != 2 {
		t.Fatal("held bonus ball count")
	}
	g = newTestGame(t)
	for ball := 1; ball <= 3; ball++ {
		g.score("BUMPER", 1000)
		g.Physics.SetBall(150, 575, 0, 1024, false)
		ticks(t, g, 1)
		// A non-match is selected deterministically by the sync clock (verified below).
		ticks(t, g, 600)
	}
	if g.Phase != GameOver {
		t.Fatalf("expected game over got phase=%d score=%s ball=%d", g.Phase, g.Score, g.BallNumber)
	}
	before := g.Frame()
	ticks(t, g, 30)
	if !bytes.Equal(before.Pix, g.Frame().Pix) {
		t.Fatal("game over changed frame")
	}
}
func TestDeterministicNativeScript(t *testing.T) {
	a, b := newTestGame(t), newTestGame(t)
	// Explicit pre-PF11 oracle conditions; DOS missing-file defaults differ.
	a.Configure(settings.Legacy())
	b.Configure(settings.Legacy())
	for i := 0; i < 1200; i++ {
		if i == 100 || i == 700 {
			a.Release(32, 0)
			b.Release(32, 0)
		}
		input := physics.Inputs{Left: i%93 < 18, Right: i%71 < 14}
		if err := a.Sync(input); err != nil {
			t.Fatal(err)
		}
		if err := b.Sync(input); err != nil {
			t.Fatal(err)
		}
		if a.Score != b.Score || a.Bonus != b.Bonus || a.Lights != b.Lights || a.Phase != b.Phase || a.Physics.Ball != b.Physics.Ball || !reflect.DeepEqual(a.Events, b.Events) {
			t.Fatalf("state nondeterminism at %d", i)
		}
		// Render different numbers of intermediate frames; rule state must not change.
		if i%10 == 0 {
			a.Frame()
		}
		if i%71 == 0 && !bytes.Equal(a.Frame().Pix, b.Frame().Pix) {
			t.Fatalf("pixels at %d", i)
		}
	}
	if a.Score.Uint64() != 2_300_000 || a.BallNumber != 2 {
		t.Fatalf("accepted1200-tick oracle changed: score=%s ball=%d", a.Score, a.BallNumber)
	}
	// PLAND SHOWPLAYERSTS prints BALL at y=10. The historical aggregate
	// fixture instead synthesized PLAYERS at y=9 whenever in the chute,
	// even after selection closed. Prove that these matrix pixels account
	// for the entire old fixture difference before accepting source output.
	legacy := *a.Display
	legacy.Clear()
	legacy.InvalidateScore()
	legacy.Text(fmt.Sprintf("PLAYER %d", a.Session.CurrentPlayer), 8, 1, 5)
	legacy.Text(fmt.Sprintf("PLAYERS %d", a.Session.PlayerCount), 8, 9, 5)
	legacy.GlyphText(strings.TrimLeft(a.Score.String(), "0"), 160-8*len(strings.TrimLeft(a.Score.String(), "0")), 0, legacy.Content.ScoreFont)
	for i := 1; i <= (len(strings.TrimLeft(a.Score.String(), "0"))-1)/3; i++ {
		x := 160 - 24*i - 1
		legacy.Dots[13*160+x], legacy.Dots[14*160+x], legacy.Dots[15*160+x-1] = true, true, true
	}
	p := presentation.MatrixPaletteMode(a.Palette(), a.Physics.ReferenceMode, 242)
	old := presentation.ComposeNative(a.Physics.FramePalette(p), &legacy, p, 96, 242, a.Physics.Settings, a.Physics.ScreenOffset)
	if fmt.Sprintf("%x", sha256.Sum256(old.Pix)) != "f9b5160b3173c35f2798dd5e270e7ded40c33642605e21c0935cd083ff231a5e" {
		t.Fatal("historical fixture changed outside the proven synthetic panel")
	}
	source := *a.Display
	source.InvalidateScore()
	source.ShowPlayerBall(a.Score.String())
	if a.Display.Dots != source.Dots || !a.Display.On {
		t.Fatal("retained matrix differs from source SHOWPLAYERSTS / KILL_FLASHOR")
	}
	hash := fmt.Sprintf("%x", sha256.Sum256(a.Frame().Pix))
	t.Logf("script score=%s ball=%d frame=%s", a.Score, a.BallNumber, hash)
}
func TestContentAgainstIndependentExtraction(t *testing.T) {
	g := newTestGame(t)
	raw, err := os.ReadFile(testinputs.Generated(t, "reference_pf4.py", "TABLE1.PRG", "reference/original-dos-source/PLAND.ASM"))
	if err != nil {
		t.Fatal(err)
	}
	var f struct {
		Content map[string]struct {
			SHA string `json:"sha256"`
		}
		Lamps []struct {
			Number, Start, Count int
			SHA                  string `json:"sha256"`
		}
	}
	if err = json.Unmarshal(raw, &f); err != nil {
		t.Fatal(err)
	}
	for name, data := range map[string][]byte{"font13": g.font13, "font5": g.font5, "duck1up": g.duckUp[0], "duck1down": g.duckDown[0], "duck2up": g.duckUp[1], "duck2down": g.duckDown[1], "duck3up": g.duckUp[2], "duck3down": g.duckDown[2]} {
		if fmt.Sprintf("%x", sha256.Sum256(data)) != f.Content[name].SHA {
			t.Fatal(name)
		}
	}
	for _, c := range f.Lamps {
		n := c.Number
		if n == 40 {
			n = 39
		}
		got := g.content[c.Number]
		want := f.Lamps[n-1]
		if got.start != want.Start || len(got.rgb) != 3*want.Count || fmt.Sprintf("%x", sha256.Sum256(got.rgb)) != want.SHA {
			t.Fatalf("lamp %d", c.Number)
		}
	}
}

func TestLaneSnackAndSaucerSourceAwards(t *testing.T) {
	var source struct {
		Effects map[string]struct{ Score, Bonus uint64 }
	}
	raw, err := os.ReadFile(testinputs.Generated(t, "reference_pf4.py", "TABLE1.PRG", "reference/original-dos-source/PLAND.ASM"))
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(raw, &source); err != nil {
		t.Fatal(err)
	}
	for _, label := range []string{"BYGEL3", "BYGEL4", "BYGEL1", "BYGEL2"} {
		g := newTestGame(t)
		g.trigger(label)
		score, bonus := uint64(50030), uint64(0)
		if label == "BYGEL3" || label == "BYGEL4" {
			score = source.Effects["BYGELSETB"].Score
			bonus = source.Effects["BYGELSETB"].Bonus
		}
		check(t, label+" score", g.Score.Uint64(), score)
		check(t, label+" bonus", g.Bonus.Uint64(), bonus)
	}
	for i, name := range []string{"HSCORE", "SCORE1", "SCORE2", "SCORE3"} {
		g := newTestGame(t)
		if i > 0 {
			g.snacks[i-1] = true
		}
		g.snack()
		score := uint64(50000) + source.Effects[name].Score
		bonus := uint64(5000) + source.Effects[name].Bonus
		if i == 3 {
			score += 250000
			bonus += 25000
		}
		check(t, name+" score", g.Score.Uint64(), score)
		check(t, name+" bonus", g.Bonus.Uint64(), bonus)
		if !g.Physics.Ball.Hold || !g.Physics.Ball.High {
			t.Fatal("snack capture")
		}
		ticks(t, g, 131)
		if g.Physics.Ball.Hold || g.SnackDisabled == false {
			t.Fatal("snack release/cooldown")
		}
	}
	for _, name := range []string{"DSCORE", "MILLION5", "EXTRABALL1"} {
		g := newTestGame(t)
		g.FiveMillion = name == "MILLION5"
		g.BallFeature = name == "EXTRABALL1"
		g.dragon()
		check(t, name+" score", g.Score.Uint64(), source.Effects[name].Score)
		check(t, name+" bonus", g.Bonus.Uint64(), source.Effects[name].Bonus)
	}
}
func TestOriginalBallSettingAndTaskReset(t *testing.T) {
	g := newTestGame(t)
	g.SetOriginalBallSetting(1)
	if g.totalBalls != 5 {
		t.Fatal("S_BALLS nonzero")
	}
	g.SetOriginalBallSetting(0)
	if g.totalBalls != 3 {
		t.Fatal("S_BALLS zero")
	}
	// Reset occurs inside the executing task slot; its replacement must survive.
	g.Phase = BallLost
	g.task(func() bool { g.newBall(); return true })
	ticks(t, g, 82)
	if g.Phase != Playing || g.Physics.Ball.Hold {
		t.Fatal("new-ball SETBALL task lost during reset")
	}
}

func TestPhysicalBoundaryAreaBeforeTarget(t *testing.T) {
	g := newTestGame(t)
	g.Physics.Ball.Hold = true
	g.Physics.SetBall(22, 432, 0, 0, false) // center (30,440), BYGEL3
	g.Physics.Ball.HitX = 150
	g.Physics.Ball.HitY = 280 // collision-derived DROPA1
	ticks(t, g, 1)
	var labels []string
	for _, e := range g.Events {
		if e.Kind == "ScoreAwarded" {
			labels = append(labels, e.Label)
		}
	}
	if !reflect.DeepEqual(labels, []string{"BYGELSETB", "DROPSETA"}) {
		t.Fatalf("area/target order: %v", labels)
	}
	check(t, "combined physical boundary awards", g.Score.Uint64(), 17550)
	check(t, "combined bonus", g.Bonus.Uint64(), 1750)
}
