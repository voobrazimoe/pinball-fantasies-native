package gameshow

import (
	"pinballfantasies/internal/settings"
	"testing"
)

func TestConfiguredWheelCameraUsesSourceTarget(t *testing.T) {
	for mode := byte(0); mode < 2; mode++ {
		g := game(t)
		c := settings.Legacy()
		c.Resolution = mode
		g.Configure(c)
		g.wheelCapture()
		want := int16(220)
		if mode == 0 {
			want = 270
		}
		if g.Physics.TargetRaster != want || g.ScreenForce != -1 {
			t.Fatal(mode, g.Physics.TargetRaster, g.ScreenForce)
		}
		g.ejectWheel()
		if g.Physics.TargetRaster != -1 {
			t.Fatal("wheel did not release smoothed force")
		}
	}
}
