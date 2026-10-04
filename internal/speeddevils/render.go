package speeddevils

import (
	"image"
	"image/color"
	"pinballfantasies/internal/assets"
	"pinballfantasies/internal/presentation"
	"strconv"
)

func (g *Game) loadLamps(data []byte) {
	for label, c := range programs.Lamps {
		n, e := strconv.Atoi(label)
		if e != nil {
			panic(e)
		}
		g.content[n] = decodeLamp(data, c.Offset)
	}
}

// decodeLamp owns the packet bytes and validates the DAC extent before use.
func decodeLamp(data []byte, at int) lamp {
	if at < 0 || at > len(data) || len(data)-at < 2 {
		panic("invalid PRG lamp header")
	}
	start, count := int(data[at]), int(data[at+1])
	if count > 256-start || count > (len(data)-at-2)/3 {
		panic("invalid PRG lamp packet")
	}
	return lamp{start, append([]byte(nil), data[at+2:at+2+count*3]...)}
}
func (g *Game) Palette() [768]byte { return g.lampPalette }
func (g *Game) paletteFor(lamps [68]bool) [768]byte {
	p := g.Physics.BasePalette(assets.VGAPalette(g.Physics.Table.Initial.Playfield.Palette))
	for n := 1; n <= 67; n++ {
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
func (g *Game) Frame() *image.RGBA {
	p := presentation.MatrixPaletteMode(g.Palette(), g.Physics.ReferenceMode, 128)
	d := *g.Display
	return presentation.ComposeNative(g.Physics.FramePalette(p), &d, p, 98, 128, g.Physics.PresentationSettings(), g.Physics.ScreenOffset)
}
func (g *Game) text(out *image.RGBA, s string, x, y, h, scale int) {
	font := g.font5
	if h == 13 {
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
			for row := 0; row < h; row++ {
				for col := 0; col < 8; col++ {
					if font[index*h+row]&(128>>uint(col)) != 0 {
						for yy := 0; yy < scale; yy++ {
							for xx := 0; xx < scale; xx++ {
								out.SetRGBA(x+col*scale+xx, y+row*scale+yy, color.RGBA{242, 176, 64, 255})
							}
						}
					}
				}
			}
		}
		x += 8 * scale
	}
}
func (g *Game) AttractFrame(ticks int) *image.RGBA {
	out := g.Frame()
	field := g.Physics.Table.Initial.Playfield
	var lamps [68]bool
	p := g.paletteFor(lamps)
	presentation.AttractLamps(ticks, g.Display.Content.LampFlash, func(n int, on bool) { g.applyLamp(&p, n, on) })
	bottom := int(g.Physics.BottomRaster())
	y := bottom - ticks%(2*bottom)
	if y < 0 {
		y = -y
	}
	if g.Physics.PresentationSettings().TableY() != 0 {
		y = 0
	}
	for row := 0; row < g.Physics.PresentationSettings().RenderHeight(); row++ {
		for x := 0; x < 320; x++ {
			i := int(field.Indices[(y+row)*320+x]) * 3
			o := (row+g.Physics.PresentationSettings().TableY())*out.Stride + x*4
			copy(out.Pix[o:o+3], p[i:i+3])
		}
	}
	return out
}
