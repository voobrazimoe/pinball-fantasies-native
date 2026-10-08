package frontend

import (
	"os"
	"path/filepath"
	"pinballfantasies/internal/partyland"
	"pinballfantasies/internal/speeddevils"
	"testing"
)

func TestNativeSpeedDevilsLifecycleAndPersistence(t *testing.T) {
	dir := t.TempDir()
	data := stageInstallationSettings(t)
	r, e := Load(data, FileStore{Directory: dir, SeedDirectory: data})
	if e != nil {
		t.Fatal(e)
	}
	m := r.Model
	key(t, m, Space)
	key(t, m, F4)
	if m.Session == nil || m.Mode != TableAttract || m.Selected != 4 {
		t.Fatal("Stones attract integration")
	}
	key(t, m, Escape)
	key(t, m, Key(21))
	key(t, m, F2)
	if m.Selected != 2 || m.Mode != TableAttract {
		t.Fatal("F2 native attract")
	}
	g, ok := m.Session.(*speeddevils.Game)
	if !ok {
		t.Fatal("wrong table implementation")
	}
	if g.Frame().Rect.Dy() != 350 || g.AttractCue() != "S_NOHIGH" {
		t.Fatal("native content")
	}
	key(t, m, Enter)
	g = m.Session.(*speeddevils.Game)
	key(t, m, P)
	tick := g.Tick
	for i := 0; i < 30; i++ {
		update(t, m, Input{Left: true, Release: true})
	}
	if g.Tick != tick {
		t.Fatal("pause advanced Speed Devils")
	}
	key(t, m, Key(30))
	for i := 0; i < 31; i++ {
		update(t, m, Input{})
	}
	key(t, m, P)
	key(t, m, Escape)
	key(t, m, Key(21))
	if m.Mode != Selector || m.Selected != 2 {
		t.Fatal("abort selected identity")
	}
	key(t, m, F2)
	key(t, m, Enter)
	key(t, m, Escape)
	if m.Mode != TableAttract || m.Selected != 2 {
		t.Fatal("abort in chute")
	}
	key(t, m, Enter)
	g = m.Session.(*speeddevils.Game)
	// Fixture starts at the final ball with a qualifying source score: three
	// SuperJack awards are 3*50,000,000. Rule awards themselves are tested locally.
	g.SetOriginalBallSetting(0)
	g.BallNumber = 3
	g.Score = partyland.Number(150000000)
	// Isolate the original final-ball/high-score lifecycle from the separately
	// covered top-score shoot-again branch.
	g.SetHighScore(partyland.Number(999999999999))
	for i := 0; i < 4000 && m.Mode == Playing; i++ {
		if g.Phase == speeddevils.Playing {
			g.Physics.SetBall(20, 576, 0, 0, false)
		}
		update(t, m, Input{})
	}
	for i := 0; i < 5 && m.Mode == GameEnd; i++ {
		update(t, m, Input{})
	}
	if m.Mode != Initials || m.Selected != 2 {
		t.Fatal("Speed Devils high score", m.Mode, g.Phase)
	}
	update(t, m, Input{}) // GET_IT_FROM_KEYBOARD clears any queued make.
	key(t, m, Key(30))
	key(t, m, Key(48))
	key(t, m, Key(46))
	if m.Mode != EntryWait {
		t.Fatal("initials")
	}
	b, e := os.ReadFile(filepath.Join(dir, "TABLE2.HI"))
	if e != nil || len(b) != 64 {
		t.Fatal("64-byte TABLE2.HI persistence", e)
	}
	scores, e := DecodeScores(b)
	if e != nil || scores[0].Name != [3]byte{'A', 'B', 'C'} || scores[0].Digits.Uint64() != 150000000 {
		t.Fatal(scores, e)
	}
	if _, e = os.Stat(filepath.Join(dir, "TABLE1.HI")); !os.IsNotExist(e) {
		t.Fatal("table identity crossed")
	}
	for i := 0; i < 62; i++ {
		update(t, m, Input{})
	}
	if m.Mode != TableAttract || m.Selected != 2 {
		t.Fatal("return to Speed Devils attract")
	}
	key(t, m, Escape)
	key(t, m, Key(21))
	if m.Mode != Selector || m.Selected != 2 {
		t.Fatal("return to frontend")
	}
}

func TestSpeedDevilsAttractIdentity(t *testing.T) {
	// The first title sequence has the same source command durations, but
	// PL_TEXT is table-local. Never display Party Land over Speed Devils.
	label, _, _ := attractPanelForTable(268, Defaults(2), 2)
	if label != "SPEED DEVILS" {
		t.Fatal("SDEV PL_TEXT", label)
	}
	label, _, _ = attractPanel(268, Defaults(1))
	if label != "PARTY LAND" {
		t.Fatal("PLAND presentation changed", label)
	}
}
