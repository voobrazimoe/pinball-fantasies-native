package physics

import (
	"encoding/binary"
	"image"

	"pinballfantasies/internal/assets"
	"pinballfantasies/internal/settings"
)

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
	if g.PresentationSettings().TableY() != 0 {
		// Full-table composition applies SCREENPOSY after rendering. Keep
		// PUTTHEBALL's offset (except HOLDSTILL), so the free ball stays fixed
		// on screen while the table moves, just like SETSCREENSTART.
		viewport = 0
	}
	return initial.SimulationFrameHeight(indices, image.Pt(int(g.Ball.PixelX), int(g.Ball.PixelY+balloffset)), viewport, foreground, g.PresentationSettings().RenderHeight())
}

// ComposeFrame renders the same frame as FramePalette followed by
// presentation.ComposeNative, writing straight into dst. dst is reused when
// it already has the right size, so steady play allocates nothing. matrix
// draws the 33-row dot matrix with off/on lattice indices.
func (g *Game) ComposeFrame(dst *image.RGBA, palette [768]byte, matrix MatrixDrawer, off, on byte) *image.RGBA {
	for i := range g.Flippers {
		g.animateFlipper(i)
	}
	c := g.PresentationSettings()
	var foreground []byte
	if g.Ball.High {
		foreground = g.Table.upperForeground
	}
	balloffset := g.ScreenOffset
	if g.Ball.Hold {
		balloffset = 0
	}
	origin := image.Pt(int(g.Ball.PixelX), int(g.Ball.PixelY+balloffset))
	// Scrolled: the field fills rows 0..h and the matrix follows; rows beyond
	// the table stay transparent like the cropped snapshot. Full table: the
	// matrix comes first and SCREENPOSY shifts the table under black.
	height, fieldY, rows, matrixY := c.FieldHeight()+33, 0, c.FieldHeight(), c.FieldHeight()
	base := int(g.Raster>>4) - 33 + int(g.ScreenOffset)
	outside := uint32(0)
	if c.ScrollMode == settings.ScrollOff {
		height, fieldY, rows, matrixY = 609, 33, 576, 0
		base = int(g.ScreenOffset)
		outside = 0xff000000
	}
	if dst == nil || dst.Rect != image.Rect(0, 0, 320, height) {
		dst = image.NewRGBA(image.Rect(0, 0, 320, height))
	}
	var lut assets.PaletteLUT
	lut.Load(&palette)
	assets.RenderRows(dst, fieldY, g.indices, &lut, base, rows, outside)
	if len(g.SpringGraphics) == 230 {
		g.drawSpring(dst, fieldY, &lut, base, rows)
	}
	g.Table.Initial.DrawBall(dst, fieldY, &lut, origin, foreground, base, rows)
	matrix.DrawAt(dst, palette, off, on, matrixY)
	return dst
}

// MatrixDrawer is presentation.Display's placement-only matrix renderer.
type MatrixDrawer interface {
	DrawAt(out *image.RGBA, p [768]byte, off, on byte, y int)
}

// drawSpring paints FramePalette's spring patch: a cleared 10×17 chute
// window with the spring graphic shifted down by its position.
func (g *Game) drawSpring(dst *image.RGBA, fieldY int, lut *assets.PaletteLUT, base, rows int) {
	real := int(g.SpringPosition)/2 - 3
	srcrow := 0
	if real < 0 {
		srcrow = -real
		real = 0
	}
	for row := 0; row < 17; row++ {
		t := 556 + row
		if t < base || t >= base+rows {
			continue
		}
		out := dst.Pix[(fieldY+t-base)*dst.Stride+304*4:]
		for x := 0; x < 10; x++ {
			c := byte(0)
			if row >= real && row-real < 17-real {
				c = g.SpringGraphics[(srcrow+row-real)*10+x]
			}
			binary.LittleEndian.PutUint32(out[4*x:], lut[c])
		}
	}
}
