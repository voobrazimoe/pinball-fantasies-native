package partyland

import (
	"image"
	"pinballfantasies/internal/assets"
	"pinballfantasies/internal/presentation"
)

func (g *Game) Palette() [768]byte { return g.lampPalette }
func (g *Game) paletteFor(lamps [57]bool) [768]byte {
	p := g.Physics.BasePalette(assets.VGAPalette(g.Physics.Table.Initial.Playfield.Palette))
	for n := 1; n <= 56; n++ {
		g.applyLamp(&p, n, lamps[n])
	}
	return p
}
func (g *Game) applyLamp(p *[768]byte, n int, on bool) {
	c := g.content[n]
	for i, v := range g.Physics.LampRGB(c.rgb, on) {
		p[3*c.start+i] = v
	}
}

// Frame composes the playfield and original VGA matrix split.
// Matrix command timing belongs to the gameplay scheduler.
func (g *Game) Frame() *image.RGBA {
	p := presentation.MatrixPaletteMode(g.Palette(), g.Physics.ReferenceMode, 242)
	if fade, ok := g.demoFadePalette(); ok {
		p = fade
	}
	d := *g.Display
	return presentation.ComposeNative(g.Physics.FramePalette(p), &d, p, 96, 242, g.Physics.PresentationSettings(), g.Physics.ScreenOffset)
}
