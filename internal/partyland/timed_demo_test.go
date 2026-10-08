package partyland

import (
	"os"
	"path/filepath"
	"testing"

	"pinballfantasies/internal/datalayout"
	"pinballfantasies/internal/physics"
	"pinballfantasies/internal/settings"
)

// Private: requires the legally supplied official demo, never canonical A.
func timedDemoFixture(t *testing.T) *Game {
	t.Helper()
	dir := os.Getenv("PF_10MIN_DEMO_DATA")
	if dir == "" {
		t.Skip("set PF_10MIN_DEMO_DATA to the private 10-minute demo")
	}
	intro, err := os.ReadFile(filepath.Join(dir, "INTRO.PRG"))
	if err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(filepath.Join(dir, "TABLE1.PRG"))
	if err != nil {
		t.Fatal(err)
	}
	id, err := datalayout.DetectDemoInstallation(intro, raw)
	if err != nil {
		t.Fatal(err)
	}
	data, err := datalayout.PreparePRGForProfile(id, "TABLE1.PRG", raw)
	if err != nil {
		t.Fatal(err)
	}
	rec, err := datalayout.DemoRecords(raw)
	if err != nil {
		t.Fatal(err)
	}
	table, err := physics.DecodePartyLand(data)
	if err != nil {
		t.Fatal(err)
	}
	g := NewDemo(table, data, DemoInputs{PlayersText: rec.PlayersText, ExpiryTexts: rec.ExpiryTexts})
	g.Configure(settings.Legacy())
	return g
}

func demoEvent(g *Game, kind, label string) bool {
	for _, e := range g.Events {
		if e.Kind == kind && e.Label == label {
			return true
		}
	}
	return false
}

func TestTimedDemoNoInputExpiryAndQuit(t *testing.T) {
	g := timedDemoFixture(t)
	if c, e, ok := g.Demo(); !ok || c != 0 || e {
		t.Fatal(c, e, ok)
	}
	var syncs, expiredAt, quitAt int
	for syncs = 1; syncs <= 70000 && !g.DemoFinished(); syncs++ {
		if err := g.Sync(physics.Inputs{}); err != nil {
			t.Fatal(err)
		}
		c, expired, _ := g.Demo()
		if int(c) != syncs {
			t.Fatalf("sync %d counted %d", syncs, c)
		}
		if demoEvent(g, "DemoExpired", "DEMO_TIMER") {
			if c != DemoThreshold || !expired || expiredAt != 0 || !g.Physics.Ball.Hold {
				t.Fatal(c, expired, expiredAt)
			}
			if !demoEvent(g, "Music", "S_GAMEOVER2") || !demoEvent(g, "MatrixStarted", timedDemoExpiryLabel) {
				t.Fatal(g.Events)
			}
			expiredAt = syncs
		}
		if expired != (expiredAt != 0) {
			t.Fatal("expired before equality", syncs)
		}
		if demoEvent(g, "DemoQuit", "QUIT") {
			quitAt = syncs
		}
	}
	if expiredAt != DemoThreshold || quitAt == 0 || !g.DemoFinished() {
		t.Fatal(expiredAt, quitAt)
	}
	t.Logf("no-input expiry %d, QUIT %d", expiredAt, quitAt)
	// QUIT is terminal: no further counting or matrix work.
	before, _, _ := g.Demo()
	if err := g.Sync(physics.Inputs{}); err != nil {
		t.Fatal(err)
	}
	if after, _, _ := g.Demo(); after != before {
		t.Fatal("counted after QUIT")
	}
	if g.timed.fadeLevel != 0 || g.Frame().Pix[0] != 0 || g.Frame().Pix[len(g.Frame().Pix)/2] != 0 {
		t.Fatal("expiry fade must leave the DAC black")
	}
}

func TestTimedDemoWrapRepeatsEquality(t *testing.T) {
	g := timedDemoFixture(t)
	g.timed.counter = 65535
	if err := g.Sync(physics.Inputs{}); err != nil {
		t.Fatal(err)
	}
	if c, e, _ := g.Demo(); c != 0 || e {
		t.Fatal(c, e)
	}
	g.timed.counter = DemoThreshold - 1
	g.timed.expired = true // sticky; equality is still an exact compare
	g.timed.equalities = 1
	if err := g.Sync(physics.Inputs{}); err != nil {
		t.Fatal(err)
	}
	if !demoEvent(g, "DemoExpired", "DEMO_TIMER") || g.timed.equalities != 2 || g.matrix.consumer == nil {
		t.Fatal(g.Events)
	}
	if err := g.Sync(physics.Inputs{}); err != nil {
		t.Fatal(err)
	}
	if c, e, _ := g.Demo(); c != DemoThreshold+1 || !e {
		t.Fatal("unequal counter must keep expired", c, e)
	}
}

func TestTimedDemoBallProgression(t *testing.T) {
	g := timedDemoFixture(t)
	if string(g.Display.Content.Texts["PLAYERSTEXT"]) != string(g.timed.playersText) {
		t.Fatal("duration label replaced")
	}
	digits := func() string {
		b := g.Display.Content.Texts["BALLSTEXT"]
		return string([]byte{b[4], b[5]})
	}
	if digits() != " 8" { // glyph code '8' prints as 1
		t.Fatalf("%q", digits())
	}
	var seen []string
	for i := 0; i < 40; i++ {
		if !g.demoChangePlayer() {
			t.Fatal("demo continuation rejected")
		}
		g.tasks = [50]func() bool{}
		g.waitCounters = make(map[string]uint16)
		seen = append(seen, digits())
	}
	// 2..9, 10..19, 20..29, then 10 again; ball number and player never advance.
	want := []string{" 9", " :", " ;", " <", " =", " >", " ?", " @", "87", "88"}
	for i, w := range want {
		if seen[i] != w {
			t.Fatalf("ball %d: %q want %q", i+2, seen[i], w)
		}
	}
	if seen[27] != "9@" || seen[28] != "87" || g.BallNumber != 1 || g.Session.CurrentPlayer != 1 {
		t.Fatal(seen[26:30], g.BallNumber)
	}
	g.playerText()
	if digits() != seen[39] || string(g.Display.Content.Texts["PLAYERSTEXT"]) != string(g.timed.playersText) {
		t.Fatal("canonical player text overwrote demo panel")
	}
	g.Lights[51] = true
	if g.demoChangePlayer() {
		t.Fatal("earned shoot-again must use the canonical change")
	}
}

func TestTimedDemoDrainRoutes(t *testing.T) {
	for _, scored := range []bool{false, true} {
		g := timedDemoFixture(t)
		g.timed.expired = true
		g.Phase = Playing
		g.ScoreChanged = scored
		g.Events = g.Events[:0]
		g.drain()
		if scored {
			if !demoEvent(g, "MatrixStarted", timedDemoExpiryLabel) || !demoEvent(g, "Music", "S_GAMEOVER2") || g.matrix.consumer == nil {
				t.Fatal("expired scored drain must reinstall expiry", g.Events)
			}
		} else if !demoEvent(g, "MatrixStarted", "PARTY_ONTS") || demoEvent(g, "MatrixStarted", timedDemoExpiryLabel) {
			t.Fatal("unscored drain selects PARTY_ON before the expired guard", g.Events)
		}
	}
	g := timedDemoFixture(t)
	g.ScoreChanged = true
	g.drain()
	if !demoEvent(g, "MatrixStarted", "BALL_LOSTTS") {
		t.Fatal("unexpired scored drain is the normal bonus route", g.Events)
	}
}

// A played session continues past the canonical ball limit and ends only
// through the demo timer.
func TestTimedDemoPlayedSession(t *testing.T) {
	g := timedDemoFixture(t)
	drains := 0
	for i := 0; i < 80000 && !g.DemoFinished(); i++ {
		in := physics.Inputs{Down: i%400 < 60, Release: i%400 == 60, Left: (i/7)%5 == 0, Right: (i/11)%4 == 0}
		if i%400 == 60 {
			g.Release(200, uint8(i))
		}
		if err := g.Sync(in); err != nil {
			t.Fatal(err)
		}
		if demoEvent(g, "BallLost", "LOOSE_BALL") {
			drains++
		}
		if g.Phase == GameOver {
			t.Fatal("demo reached the canonical game over at", i)
		}
	}
	if !g.DemoFinished() || drains < 4 {
		t.Fatal(g.DemoFinished(), drains)
	}
	c, _, _ := g.Demo()
	t.Logf("played session: %d drains, QUIT at timer %d, score %s", drains, c, g.Score)
}

// Attract F1-F8 set PLAYERS, but DEMOVER_CHANGE_PLAYER never rotates PLAYER.
func TestTimedDemoMultiplayerKeepsPlayerOne(t *testing.T) {
	g := timedDemoFixture(t)
	g.StartPlayers(4)
	if g.PlayerCount() != 4 {
		t.Fatal(g.PlayerCount())
	}
	for i := 0; i < 12; i++ {
		if !g.demoChangePlayer() {
			t.Fatal("demo continuation rejected")
		}
		g.tasks = [50]func() bool{}
		g.waitCounters = make(map[string]uint16)
		if g.Session.CurrentPlayer != 1 || g.BallNumber != 1 {
			t.Fatal("player rotated", g.Session.CurrentPlayer)
		}
	}
}

// The idle panel's players record shows the started count, not the loaded
// placeholder; a fresh table keeps the placeholder until the attract start.
func TestTimedDemoPlayersPanel(t *testing.T) {
	g := timedDemoFixture(t)
	placeholder := string(g.Display.Content.Texts["NO_OF_PLAYERS_TEXT"])
	for _, n := range []int{1, 3} {
		g.StartPlayers(n)
		got := g.Display.Content.Texts["NO_OF_PLAYERS_TEXT"]
		if string(got) == placeholder || got[8] != byte(n)+'7' {
			t.Fatalf("%d players: %q", n, got)
		}
		if !demoEvent(g, "MatrixStarted", "DEMO_FIRST_NO_OF_PLAYERSTS") {
			t.Fatal(g.Events)
		}
		g.Events = g.Events[:0]
		g.playerText()
		if g.Display.Content.Texts["NO_OF_PLAYERS_TEXT"][8] != byte(n)+'7' {
			t.Fatal("demo panel restored the placeholder")
		}
	}
}
