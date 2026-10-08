package frontend

import (
	"image"
	"image/png"
	"os"
	"path/filepath"
	"pinballfantasies/internal/stones"
	"pinballfantasies/internal/tablelogic"
	"testing"
)

func TestNativeStonesLifecycleAndPersistence(t *testing.T) {
	dir := t.TempDir()
	data := stageInstallationSettings(t)
	r, e := Load(data, FileStore{Directory: dir, SeedDirectory: data})
	if e != nil {
		t.Fatal(e)
	}
	m := r.Model
	save := func(name string, im *image.RGBA) {
		dir := os.Getenv("PF10_CHECKPOINT_DIR")
		if dir == "" {
			return
		}
		os.MkdirAll(dir, 0755)
		f, e := os.Create(filepath.Join(dir, name+".png"))
		if e != nil {
			t.Fatal(e)
		}
		if e = png.Encode(f, im); e != nil {
			t.Fatal(e)
		}
		f.Close()
	}
	key(t, m, Space)
	key(t, m, F4)
	if m.Selected != 4 || m.Mode != TableAttract {
		t.Fatal("F4 native attract")
	}
	g, ok := m.Session.(*stones.Game)
	if !ok {
		t.Fatal("wrong table implementation")
	}
	if g.Frame().Rect.Dy() != 350 || g.AttractCue() != "S_NOHIGH" {
		t.Fatal("native content")
	}
	title := *r.View.StonesMatrix
	title.Clear()
	title.Begin("_PRINT13", []string{"PL_TEXT", "SW*2/4*2+3*4"})
	title.Visit(0, 0, func(string) string { return "0" })
	foundTitle := false
	for i := 0; i < 450; i++ {
		d := r.View.StonesMatrix.Attract(i, [4]string{}, [4]string{})
		if d.On && d.Dots == title.Dots {
			m.Counter = i
			foundTitle = true
			break
		}
	}
	if !foundTitle {
		t.Fatal("original Stones attract title")
	}
	save("attract", r.Frame())
	key(t, m, Enter)
	g = m.Session.(*stones.Game)
	key(t, m, P)
	tick := g.Tick
	for i := 0; i < 30; i++ {
		update(t, m, Input{Left: true, Release: true})
	}
	if g.Tick != tick {
		t.Fatal("pause advanced Stones")
	}
	key(t, m, Key(30))
	for i := 0; i < 31; i++ {
		update(t, m, Input{})
	}
	key(t, m, P)
	key(t, m, Escape)
	key(t, m, Key(21))
	if m.Mode != Selector || m.Selected != 4 {
		t.Fatal("abort selected identity")
	}
	key(t, m, F4)
	key(t, m, Enter)
	key(t, m, Escape)
	if m.Mode != TableAttract || m.Selected != 4 {
		t.Fatal("abort in chute")
	}
	key(t, m, Enter)
	g = m.Session.(*stones.Game)
	// Fixture starts at the final ball with a qualifying source score: three
	// SuperJack awards are 3*50,000,000. Rule awards themselves are tested locally.
	g.SetOriginalBallSetting(0)
	g.BallNumber = 3
	g.Score = tablelogic.Number(150000000)
	// Isolate the original final-ball/high-score lifecycle from the separately
	// covered top-score shoot-again branch.
	g.SetHighScore(tablelogic.Number(999999999999))
	for i := 0; i < 4000 && m.Mode == Playing; i++ {
		if g.Phase == stones.Playing {
			g.Physics.SetBall(20, 576, 0, 0, false)
		}
		update(t, m, Input{})
	}
	for i := 0; i < 5 && m.Mode == GameEnd; i++ {
		update(t, m, Input{})
	}
	if m.Mode != Initials || m.Selected != 4 {
		t.Fatal("Stones high score", m.Mode, g.Phase)
	}
	save("high-score-entry", r.Frame())
	update(t, m, Input{}) // GET_IT_FROM_KEYBOARD clears any queued make.
	key(t, m, Key(30))
	key(t, m, Key(48))
	key(t, m, Key(46))
	if m.Mode != EntryWait {
		t.Fatal("initials")
	}
	b, e := os.ReadFile(filepath.Join(dir, "TABLE4.HI"))
	if e != nil || len(b) != 64 {
		t.Fatal("64-byte TABLE4.HI persistence", e)
	}
	scores, e := DecodeScores(b)
	if e != nil || scores[0].Name != [3]byte{'A', 'B', 'C'} || scores[0].Digits.Uint64() != 150000000 {
		t.Fatal(scores, e)
	}
	if _, e = os.Stat(filepath.Join(dir, "TABLE3.HI")); !os.IsNotExist(e) {
		t.Fatal("wrote Gameshow scores")
	}
	if _, e = os.Stat(filepath.Join(dir, "TABLE2.HI")); !os.IsNotExist(e) {
		t.Fatal("wrote Speed Devils scores")
	}
	if _, e = os.Stat(filepath.Join(dir, "TABLE1.HI")); !os.IsNotExist(e) {
		t.Fatal("table identity crossed")
	}
	for i := 0; i < 62; i++ {
		update(t, m, Input{})
	}
	if m.Mode != TableAttract || m.Selected != 4 {
		t.Fatal("return to Stones attract")
	}
	key(t, m, Escape)
	key(t, m, Key(21))
	if m.Mode != Selector || m.Selected != 4 {
		t.Fatal("return to frontend")
	}
}
