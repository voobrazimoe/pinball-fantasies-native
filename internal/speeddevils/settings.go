package speeddevils

import "pinballfantasies/internal/settings"

// Configure is applied once by the frontend factory before audio attachment.
func (g *Game) Configure(c settings.Config) {
	g.Physics.Configure(c, [][2]int16{{0, 10}, {0, 15}, {0, 25}, {-1, 10}, {0, 20}, {12, 10}})
	g.SetOriginalBallSetting(c.Balls)
	g.lampPalette = g.paletteFor(g.Lamps)
	if c.Music == 1 && !g.MusicOff {
		g.ToggleMusic()
	}
	if g.inChute {
		g.Physics.Raster = (g.Physics.BottomRaster() + 33) * 16
		g.ScreenForce = -1
	}
}
func (g *Game) SessionSettings() settings.Config { return g.Physics.Settings }
func (g *Game) BaseBalls() byte                  { return g.totalBalls }

func (g *Game) MusicDisabled() bool { return g.MusicOff }
