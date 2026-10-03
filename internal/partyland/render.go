package partyland

import (
	"image"
	"image/color"
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
	d := *g.Display
	return presentation.ComposeNative(g.Physics.FramePalette(p), &d, p, 96, 242, g.Physics.Settings, g.Physics.ScreenOffset)
}
func (g *Game) text(out *image.RGBA, s string, x, y, height, scale int, c color.RGBA) {
	font := g.font5
	if height == 13 {
		font = g.font13
	}
	for _, ch := range s {
		index := -1
		if ch >= '0' && ch <= '9' {
			index = int(ch - '0')
		} else if ch >= 'A' && ch <= 'Z' {
			index = int(ch-'A') + 10
		}
		if index >= 0 {
			for row := 0; row < height; row++ {
				bits := font[index*height+row]
				for col := 0; col < 8; col++ {
					if bits&(128>>uint(col)) != 0 {
						for yy := 0; yy < scale; yy++ {
							for xx := 0; xx < scale; xx++ {
								out.SetRGBA(x+col*scale+xx, y+row*scale+yy, c)
							}
						}
					}
				}
			}
		}
		x += 8 * scale
	}
}
