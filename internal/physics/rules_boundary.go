package physics

import "image"

// SetBall is SETBALLPOS at a table-rule boundary. Hold/gravity are separate
// source variables: callers deliberately retain them unless the rule writes them.
func (g *Game) SetBall(x, y, vx, vy int16, high bool) {
	b := &g.Ball
	b.X = int32(x) * 1024
	b.Y = int32(y) * 1024
	b.PixelX = x
	b.PixelY = y
	b.VX = vx
	b.VY = vy
	b.High = high
}

// PatchMask is MOVE_MASK_DATA: tightly packed source rows to 40-byte map rows.
func (g *Game) PatchMask(high bool, xByte, y, width, height int, data []byte) {
	m := g.mask12
	if high {
		m = g.mask22
	}
	for row := 0; row < height; row++ {
		copy(m[(y+row)*40+xByte:(y+row)*40+xByte+width], data[row*width:(row+1)*width])
	}
}
func (g *Game) MaskRegion(high bool, xByte, y, width, height int) []byte {
	m := g.mask12
	if high {
		m = g.mask22
	}
	out := make([]byte, width*height)
	for row := 0; row < height; row++ {
		copy(out[row*width:], m[(y+row)*40+xByte:(y+row)*40+xByte+width])
	}
	return out
}

// FramePalette uses the original indexed artwork with a game's lamp palette.
// It never mutates the shared decoded asset or the PF1/PF2/PF3 snapshots.
func (g *Game) FramePalette(palette [768]byte) *image.RGBA {
	for i := range g.Flippers {
		g.animateFlipper(i)
	}
	initial := *g.Table.Initial
	field := *initial.Playfield
	field.Palette = palette
	initial.Playfield = &field
	var foreground []byte
	if g.Ball.High {
		foreground = g.Table.upperForeground
	}
	indices := g.indices
	if len(g.SpringGraphics) == 230 {
		indices = append([]byte(nil), indices...)
		real := int(g.SpringPosition)/2 - 3
		srcrow := 0
		if real < 0 {
			srcrow = -real
			real = 0
		}
		for row := 0; row < 17; row++ {
			for x := 0; x < 10; x++ {
				indices[(556+row)*320+304+x] = 0
			}
		}
		for row := 0; row < 17-real; row++ {
			copy(indices[(556+real+row)*320+304:], g.SpringGraphics[(srcrow+row)*10:(srcrow+row+1)*10])
		}
	}
	offset := g.ScreenOffset
	balloffset := offset
	if g.Ball.Hold {
		balloffset = 0
	}
	viewport := int(g.Raster>>4) - 33 + int(offset)
	if g.Settings.TableY() != 0 {
		// Full-table composition applies SCREENPOSY after rendering. Keep
		// PUTTHEBALL's offset (except HOLDSTILL), so the free ball stays fixed
		// on screen while the table moves, just like SETSCREENSTART.
		viewport = 0
	}
	return initial.SimulationFrameHeight(indices, image.Pt(int(g.Ball.PixelX), int(g.Ball.PixelY+balloffset)), viewport, foreground, g.Settings.RenderHeight())
}
