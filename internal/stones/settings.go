package stones

import "pinballfantasies/internal/settings"

// Configure is applied once by the frontend factory before audio attachment.
func (g *Game) Configure(c settings.Config) {
	g.Physics.Configure(c, [][2]int16{{0, 10}, {-10, 5}, {0, -10}, {5, 0}, {5, 15}, {-10, 12}, {2, 15}, {-8, 12}, {3, 10}, {4, 13}, {7, 10}})
	g.SetOriginalBallSetting(c.Balls)
	g.palette = g.basePalette()
	for n, on := range g.Lights {
		if on {
			g.applyLamp(&g.palette, n, true)
		}
	}
	g.matrixPalette(true)
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
