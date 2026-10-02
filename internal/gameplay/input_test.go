package gameplay_test

import (
	"pinballfantasies/internal/gameplay"
	"pinballfantasies/internal/physics"
	"testing"
)

func TestKeyboardSpringDOSOracle(t *testing.T) {
	var pos uint8
	g := &physics.Game{SpringValid: true}
	release := func(c uint8) { g.Release(c, 255) }
	for tick := 0; tick < 40; tick++ {
		gameplay.Spring(&pos, true, gameplay.Controls{Down: true}, release)
		want := tick + 1
		if want > 32 {
			want = 32
		}
		if int(pos) != want {
			t.Fatalf("SPRINGIT tick %d: %d", tick, pos)
		}
	}
	gameplay.Spring(&pos, true, gameplay.Controls{Release: true, Down: true}, release)
	if pos != 0 || g.Ball.VY != -166*32-255 || g.Ball.Rotation != 15 {
		t.Fatalf("SPRINGUP: pos=%d ball=%+v", pos, g.Ball)
	}
}

func TestMouseSpringDOSOracle(t *testing.T) {
	var pos uint8
	launches := 0
	g := &physics.Game{SpringValid: true}
	release := func(c uint8) { launches++; g.Release(c, 7) }
	step := func(in gameplay.Controls) { gameplay.Spring(&pos, true, in, release) }
	step(gameplay.Controls{MouseY: -9})
	if pos != 0 {
		t.Fatal("lower clamp")
	}
	for i := 0; i < 40; i++ {
		step(gameplay.Controls{MouseY: 99})
	}
	if pos != 32 {
		t.Fatal("upper clamp", pos)
	}
	step(gameplay.Controls{MouseY: -99})
	if pos != 31 {
		t.Fatal("SPRINGSTEEN must decrement once per task", pos)
	}
	step(gameplay.Controls{MouseFire: true})
	if launches != 1 || g.Ball.VY != -166*31-7 || pos != 0 {
		t.Fatal("mouse SPRINGUP arithmetic/reset", launches, pos, g.Ball.VY)
	}
	step(gameplay.Controls{})
	step(gameplay.Controls{MouseFire: true})
	if launches != 1 {
		t.Fatal("zero charge invented a launch")
	}
}

func TestMixedDevicesOneSpringAndValidity(t *testing.T) {
	var pos uint8
	var charges []uint8
	release := func(c uint8) { charges = append(charges, c) }
	for i := 0; i < 4; i++ {
		gameplay.Spring(&pos, true, gameplay.Controls{Down: true}, release)
	}
	gameplay.Spring(&pos, true, gameplay.Controls{MouseY: 1}, release)
	gameplay.Spring(&pos, true, gameplay.Controls{MouseY: -1}, release)
	gameplay.Spring(&pos, true, gameplay.Controls{Release: true}, release)
	if len(charges) != 1 || charges[0] != 4 || pos != 0 {
		t.Fatal("separate device spring", charges, pos)
	}
	pos = 10
	gameplay.Spring(&pos, false, gameplay.Controls{MouseY: 1, MouseFire: true}, release)
	if pos != 10 || len(charges) != 1 {
		t.Fatal("mouse affected invalid spring")
	}
}

func TestINT33VerticalRatioAndRecentering(t *testing.T) {
	var mouse gameplay.Mouse
	if mouse.Motion(7) != 0 || mouse.Motion(1) != 1 {
		t.Fatal("INIT_MOUSE 64 mickeys per 8 pixels")
	}
	if mouse.Motion(800) != 1 || mouse.Motion(0) != 0 {
		t.Fatal("bounded DOS Y/recentre")
	}
	if mouse.Motion(-8) != -1 {
		t.Fatal("upward displacement")
	}
	mouse.Motion(7)
	mouse.Clear()
	if mouse.Motion(1) != 0 {
		t.Fatal("focus retained fractional motion")
	}
}

func TestSpringupSkipsFurtherPullAndMotion(t *testing.T) {
	pos := uint8(12)
	var charge uint8
	gameplay.Spring(&pos, true, gameplay.Controls{MouseFire: true, MouseY: 1, Down: true}, func(c uint8) { charge = c })
	if charge != 12 || pos != 0 {
		t.Fatal("SPRINGUP ran SPRINGSTEEN/SPRINGIT", charge, pos)
	}
}
