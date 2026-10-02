package partyland

import (
	"pinballfantasies/internal/physics"
	"testing"
)

func TestDownSpringChargeReleaseAndSpaceTilt(t *testing.T) {
	g := newTestGame(t)
	for i := 1; i <= 40; i++ {
		g.gameplayControl(physics.Inputs{Down: true})
		want := i
		if want > 32 {
			want = 32
		}
		if int(g.Physics.SpringPosition) != want {
			t.Fatal("source charge per sync")
		}
	}
	if g.Frame().Rect.Dy() != 350 {
		t.Fatal("physical frame height")
	}
	charged := g.Frame()
	g.gameplayControl(physics.Inputs{Release: true})
	released := g.Frame()
	if g.Physics.SpringPosition != 0 || g.Physics.Ball.VY != -5312 {
		t.Fatal("source SPRINGUP arithmetic/reset")
	}
	if charged.RGBAAt(306, 300) == released.RGBAAt(306, 300) && charged.RGBAAt(306, 307) == released.RGBAAt(306, 307) {
		t.Fatal("spring art did not move")
	}
	g.Physics.Ball.Hold = true
	g.Physics.SpringValid = false
	g.inChute = false
	for _, held := range []bool{true, false, true, false, true} {
		g.gameplayControl(physics.Inputs{Tilt: held})
	}
	if !g.Physics.Tilted || g.Physics.AllowFlip || g.matrix.op != "_EOSNURR" || g.Audio.ReturnPosition != 62 {
		t.Fatal("tilt state/matrix/audio")
	}
	for _, v := range g.Lights {
		if v {
			t.Fatal("tilt retained a logic light")
		}
	}
	for i := 0; i < 8; i++ {
		if e := g.Sync(physics.Inputs{Left: true, Right: true}); e != nil {
			t.Fatal(e)
		}
	}
	if g.Display.Dots == ([2560]bool{}) {
		t.Fatal("tilt text was not rendered")
	}
	if g.matrix.op != "_WAIT" || g.matrix.remaining > 32767 || g.Physics.Flippers[0].Frame != 0 {
		t.Fatal("tilt wait/flipper inhibition")
	}
	g.resetBall()
	if g.Physics.Tilted || !g.Physics.AllowFlip || g.Physics.TiltCounter != 0 {
		t.Fatal("next ball tilt reset")
	}
}
func TestMusicToggleKeepsJinglePriority(t *testing.T) {
	g := newTestGame(t)
	g.ToggleMusic()
	g.runTasks()
	if !g.MusicOff || g.Audio.Position != 62 {
		t.Fatal("music off")
	}
	g.music("S_TILT")
	if g.Audio.Position != g.Display.Jingle("S_TILT").Position {
		t.Fatal("tilt jingle is still audible while music is off")
	}
	g.ToggleMusic()
	if g.MusicOff || g.Audio.ReturnPosition != 0 || g.Audio.JumpCount < 1 {
		t.Fatal("music return")
	}
}

// LAMP 53 and 56 both write DAC index 111. Original LON/LOFF packets
// overwrite in call order; reconstructing the palette in numeric order loses it.
func TestSharedLampDACWriteOrder(t *testing.T) {
	g := newTestGame(t)
	g.lamp(56, false)
	palette := g.Palette()
	if got := palette[333:336]; got[0] != 73 || got[1] != 24 || got[2] != 24 {
		t.Fatal(got)
	}
	g.lamp(53, true)
	palette = g.Palette()
	if got := palette[333:336]; got[0] != 97 || got[1] != 32 || got[2] != 32 {
		t.Fatal(got)
	}
	g.lamp(56, false)
	palette = g.Palette()
	if got := palette[333:336]; got[0] != 73 || got[1] != 24 || got[2] != 24 {
		t.Fatal(got)
	}
}
