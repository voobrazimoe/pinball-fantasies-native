package partyland

import (
	"image"
	"pinballfantasies/internal/assets"
	"pinballfantasies/internal/physics"
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
func (g *Game) Frame() *image.RGBA { return g.FrameInto(nil) }

// FrameInto composes the frame into dst, reusing it when its size matches.
// PlayfieldPhysics exposes the ball/camera state the presentation override uses.
func (g *Game) PlayfieldPhysics() *physics.Game { return g.Physics }

func (g *Game) FrameInto(dst *image.RGBA) *image.RGBA {
	p := presentation.MatrixPaletteMode(g.Palette(), g.Physics.ReferenceMode, 242)
	if fade, ok := g.demoFadePalette(); ok {
		p = fade
	}
	return g.Physics.ComposeFrame(dst, p, g.Display, 96, 242)
}
